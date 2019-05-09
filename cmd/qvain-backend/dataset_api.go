package main

import (
	"net/http"
	"strings"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/internal/sessions"
	"github.com/CSCfi/qvain-api/internal/shared"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/CSCfi/qvain-api/pkg/models"

	"github.com/francoispqt/gojay"
	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
)

// DefaultIdentity is the user identity to show to the outside world.
const DefaultIdentity = "fairdata"

type DatasetApi struct {
	db       *psql.DB
	sessions *sessions.Manager
	metax    *metax.MetaxService
	logger   zerolog.Logger

	identity string
}

func NewDatasetApi(db *psql.DB, sessions *sessions.Manager, metax *metax.MetaxService, logger zerolog.Logger) *DatasetApi {
	return &DatasetApi{
		db:       db,
		sessions: sessions,
		metax:    metax,
		logger:   logger,
		identity: DefaultIdentity,
	}
}

// SetIdentity sets the identity to show to the outside world.
// It is not safe to call this method after instantiation.
func (api *DatasetApi) SetIdentity(identity string) {
	api.identity = identity
}

func (api *DatasetApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// authenticated api
	session, err := api.sessions.SessionFromRequest(r)
	if err != nil {
		api.logger.Error().Err(err).Msg("no session from request")
		jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user := session.User

	head := ShiftUrlWithTrailing(r)
	api.logger.Debug().Str("head", head).Str("path", r.URL.Path).Str("method", r.Method).Msg("datasets")

	// root
	if head == "" {
		// handle self
		switch r.Method {
		case http.MethodGet:
			api.ListDatasets(w, r, user)
		case http.MethodPost:
			api.createDataset(w, r, user)
		case http.MethodOptions:
			apiWriteOptions(w, "GET, POST, OPTIONS")
			return
		default:
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}

	// dataset uuid
	id, err := GetUuidParam(head)
	if err != nil {
		jsonError(w, "bad format for uuid path parameter", http.StatusBadRequest)
		return
	}

	// delegate to dataset handler
	api.Dataset(w, r, user, id)
}

func (api *DatasetApi) ListDatasets(w http.ResponseWriter, r *http.Request, user *models.User) {
	switch r.URL.RawQuery {
	case "":
	case "fetch":
		api.logger.Debug().Str("op", "fetch").Msg("datasets")
		err := shared.Fetch(api.metax, api.db, api.logger, user.Uid, user.Identity)
		if err != nil {
			// TODO: handle mixed error
			jsonError(w, err.Error(), http.StatusBadRequest)
			//dbError(w, err)
			return
		}
	case "fetchall":
		api.logger.Debug().Str("op", "fetchall").Msg("datasets")
		shared.FetchAll(api.metax, api.db, api.logger, user.Uid, user.Identity)
	default:
		jsonError(w, "invalid parameter", http.StatusBadRequest)
		return
	}

	jsondata, err := api.db.ViewDatasetsByOwner(user.Uid)
	if err != nil {
		api.logger.Error().Err(err).Str("uid", user.Uid.String()).Msg("error listing datasets")
		dbError(w, err)
		return
	}

	apiWriteHeaders(w)
	//w.Header().Set("Cache-Control", "private, max-age=300")
	w.Write(jsondata)
}

// Dataset handles requests for a dataset by UUID. It dispatches to request method specific handlers.
func (api *DatasetApi) Dataset(w http.ResponseWriter, r *http.Request, user *models.User, id uuid.UUID) {
	//api.logger.Debug().Str("head", "").Str("path", r.URL.Path).Msg("dataset")
	hasTrailing := r.URL.Path == "/"
	op := ShiftUrlWithTrailing(r)
	api.logger.Debug().Bool("hasTrailing", hasTrailing).Str("head", op).Str("path", r.URL.Path).Str("dataset", id.String()).Msg("dataset")

	// root; don't accept trailing
	if op == "" && !hasTrailing {
		// handle self
		switch r.Method {
		case http.MethodGet:
			api.getDataset(w, r, user.Uid, id, "")
			return
		case http.MethodPut:
			api.updateDataset(w, r, user, id)
			return
		case http.MethodDelete:
			api.deleteDataset(w, r, user.Uid, id)
			return
		case http.MethodOptions:
			apiWriteOptions(w, "GET, PUT, DELETE, OPTIONS")
			return

		default:
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	}

	// dataset operations
	switch op {
	case "export":
		// TODO: assess security implementations before enabling this
		jsonError(w, "export not implemented", http.StatusNotImplemented)
		return
	case "versions":
		if checkMethod(w, r, http.MethodGet) {
			api.ListVersions(w, r, user.Uid, id)
		}
		return
	case "publish":
		if checkMethod(w, r, http.MethodPost) {
			api.publishDataset(w, r, user.Uid, id)
		}
		return
	default:
		jsonError(w, "invalid dataset operation", http.StatusNotFound)
		return
	}
	return
}

// getDataset retrieves a dataset's whole blob or part thereof depending on the path.
// Not all datasets are fully viewable through the API.
func (api *DatasetApi) getDataset(w http.ResponseWriter, r *http.Request, owner uuid.UUID, id uuid.UUID, path string) {
	// whole dataset is not visible through this API
	// TODO: plug super-user check
	if false && path == "" {
		jsonError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	res, err := api.db.ViewDatasetWithOwner(id, owner, api.identity)
	if dbError(w, err) {
		return
	}

	apiWriteHeaders(w)
	w.Write(res)
	return
}

func (api *DatasetApi) createDataset(w http.ResponseWriter, r *http.Request, creator *models.User) {
	var err error

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		jsonError(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	if r.Body == nil || r.Body == http.NoBody {
		jsonError(w, "empty body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	typed, err := models.CreateDatasetFromJson(creator.Uid, r.Body, map[string]string{"identity": creator.Identity, "org": creator.Organisation})
	if err != nil {
		api.logger.Error().Err(err).Msg("create dataset failed")
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.db.Create(typed.Unwrap())
	if err != nil {
		//jsonError(w, "store failed", http.StatusBadRequest)
		dbError(w, err)
		return
	}

	api.Created(w, r, typed.Unwrap().Id)
}

func (api *DatasetApi) updateDataset(w http.ResponseWriter, r *http.Request, owner *models.User, id uuid.UUID) {
	var err error

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		jsonError(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	if r.Body == nil || r.Body == http.NoBody {
		jsonError(w, "empty body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	typed, err := models.UpdateDatasetFromJson(owner.Uid, r.Body, nil)
	if err != nil {
		api.logger.Error().Err(err).Str("dataset", id.String()).Str("user", owner.Uid.String()).Msg("update dataset failed")
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.logger.Debug().Str("json", string(typed.Unwrap().Blob())).Msg("new json")

	api.logger.Debug().Str("owner", owner.Uid.String()).Msg("owner")

	err = api.db.SmartUpdateWithOwner(id, typed.Unwrap().Blob(), owner.Uid)
	if err != nil {
		dbError(w, err)
		return
	}

	api.Created(w, r, typed.Unwrap().Id)
}

func (api *DatasetApi) publishDataset(w http.ResponseWriter, r *http.Request, owner uuid.UUID, id uuid.UUID) {
	vId, nId, qId, err := shared.Publish(api.metax, api.db, id, owner)
	if err != nil {
		switch t := err.(type) {
		case *metax.ApiError:
			api.logger.Warn().Err(err).Str("dataset", id.String()).Str("owner", owner.String()).Str("origin", "api").Msg("publish failed")
			jsonErrorWithPayload(w, t.Error(), "metax", t.OriginalError(), convertExternalStatusCode(t.StatusCode()))
		case *psql.DatabaseError:
			api.logger.Error().Err(err).Str("dataset", id.String()).Str("owner", owner.String()).Str("origin", "database").Msg("publish failed")
			dbError(w, err)
		default:
			api.logger.Error().Err(err).Str("dataset", id.String()).Str("owner", owner.String()).Str("origin", "other").Msg("publish failed")
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	api.Published(w, r, id, vId, qId, nId)
}

func (api *DatasetApi) deleteDataset(w http.ResponseWriter, r *http.Request, owner uuid.UUID, id uuid.UUID) {
	err := api.db.Delete(id, &owner)
	if err != nil {
		dbError(w, err)
		return
	}

	// deleted, return 204 No Content
	apiWriteHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

// ListVersions lists an array of existing versions for a given dataset and owner.
func (api *DatasetApi) ListVersions(w http.ResponseWriter, r *http.Request, user uuid.UUID, id uuid.UUID) {
	jsondata, err := api.db.ViewVersions(user, id)
	if err != nil {
		api.logger.Error().Err(err).Str("uid", user.String()).Str("dataset", id.String()).Msg("error getting versions")
		dbError(w, err)
		return
	}

	apiWriteHeaders(w)
	w.Write(jsondata)
}

// redirectToNew redirects to the location of a newly created (POST) or updated (PUT) resource.
// Note that http.Redirect() will write and send the headers, so set ours before.
func (api *DatasetApi) redirectToNew(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	api.logger.Debug().Str("r.URL.Path", r.URL.Path).Str("r.RequestURI", r.RequestURI).Str("r.URL.RawPath", r.URL.RawPath).Str("r.Method", r.Method).Msg("available URL information")

	// either: POST /parent/ or PUT /parent/id; so check we can check either the method or if we've got a trailing slash
	// NOTE: r.RequestURI is insecure, but http.Redirect escapes it for us anyway.
	if r.Method == http.MethodPost {
		http.Redirect(w, r, r.RequestURI+id.String(), http.StatusCreated)
		return
	}
	http.Redirect(w, r, r.RequestURI, http.StatusCreated)
}

// Created returns a success response for created (POST, 201) or updated (PUT, 204) resources.
//
// TODO: Perhaps 200 OK with a body is a better response to a PUT request?
func (api *DatasetApi) Created(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiWriteHeaders(w)

	// new, return 201 Created
	if r.Method == http.MethodPost {
		api.redirectToNew(w, r, id)

		enc := gojay.BorrowEncoder(w)
		defer enc.Release()

		enc.AppendByte('{')
		enc.AddIntKey("status", http.StatusCreated)
		enc.AddStringKey("msg", "created")
		enc.AddStringKey("id", id.String())
		enc.AppendByte('}')
		enc.Write()

		return
	}

	// update, return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

func (api *DatasetApi) Published(w http.ResponseWriter, r *http.Request, id uuid.UUID, extid string, newId *uuid.UUID, newExtid string) {
	apiWriteHeaders(w)
	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	enc.AppendByte('{')
	enc.AddIntKey("status", http.StatusOK)
	enc.AddStringKey("msg", "dataset published")
	enc.AddStringKey("id", id.String())
	enc.AddStringKey("extid", extid)
	if newId != nil {
		enc.AddStringKey("new_id", newId.String())
	}
	if newExtid != "" {
		enc.AddStringKey("new_extid", newExtid)
	}
	enc.AppendByte('}')
	enc.Write()

	return
}
