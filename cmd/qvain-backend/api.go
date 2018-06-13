package main

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/models"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
	//"encoding/json"

	// side-effect
	_ "github.com/NatLibFi/qvain-api/metax"
)

type DatasetRouter struct {
	mountedAt string
	db        *psql.DB
	logger    zerolog.Logger
}

func NewDatasetRouter(mountPoint string, db *psql.DB, logger zerolog.Logger) *DatasetRouter {
	return &DatasetRouter{mountedAt: path.Clean(mountPoint) + "/", db: db, logger: logger}
}

func (api *DatasetRouter) Mountpoint() string {
	return api.mountedAt
}

func (api *DatasetRouter) urlForDataset(id uuid.UUID) string {
	return api.mountedAt + id.String()
}

func (api *DatasetRouter) Root(r *http.Request) string {
	if root := strings.TrimPrefix(r.URL.Path, api.mountedAt); len(root) < len(r.URL.Path) {
		return root
	}
	return ""
}

func (api *DatasetRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.Datasets(w, r)
}

func (api *DatasetRouter) Created(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Redirect(w, r, api.urlForDataset(id), http.StatusCreated)
	fmt.Fprintf(w, `{"code":%d,"message":"%s","id":"%s"}}%c`, http.StatusCreated, "created", id.String(), '\n')
}

func head(p string) string {
	//if i := strings.IndexByte(p, '/')+1; i > 0 {
	if i := strings.IndexByte(p, '/'); i > -1 {
		return p[:i]
	}
	return p
}

func headAt(p string) int {
	//if i := strings.IndexByte(p, '/')+1; i > 0 {
	if i := strings.IndexByte(p, '/'); i > -1 {
		return i
	}
	return len(p)
}

func tail(p string) string {
	if len(p) <= 1 {
		return ""
	}
	if i := strings.IndexByte(p[1:], '/') + 2; i > 1 {
		return p[i:]
	}
	return ""
}

func cuthead(p string) string {
	if len(p) <= 1 {
		return ""
	}
	if i := strings.IndexByte(p[1:], '/') + 2; i > 1 {
		return p[i-1:]
	}
	return ""
}

func (api *DatasetRouter) Datasets(w http.ResponseWriter, r *http.Request) {
	var (
		root string
		err  error
	)

	if root = strings.TrimPrefix(r.URL.Path, api.mountedAt); len(root) < len(r.URL.Path) {
		fmt.Println("root:", root)
	} else {
		jsonError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// require authentication from here on
	var user uuid.UUID
	token, ok := jwt.FromContext(r.Context())
	if ok {
		user, err = uuid.FromString(token.Subject())
		if err != nil {
			ok = false
		}
	}
	if !ok {
		jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// handle self
	if root == "" {
		switch r.Method {
		case http.MethodPost:
			api.createDataset(w, r, user)
		default:
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}

	// handle sub
	fmt.Println("WVH:", head(root))
	id, err := uuid.FromString(head(root))
	if err != nil {
		jsonError(w, "bad format for uuid path parameter", http.StatusBadRequest)
		return
	}
	fmt.Println("id:", id)

	api.Dataset(w, r, user, id)
}

// Dataset handles requests for a dataset by UUID. It dispatches to request method specific handlers.
func (api *DatasetRouter) Dataset(w http.ResponseWriter, r *http.Request, user uuid.UUID, id uuid.UUID) {
	path := cuthead(strings.TrimPrefix(r.URL.Path, api.mountedAt))

	switch r.Method {
	case http.MethodGet:
		api.getDataset(w, r, user, id, path)
		return
	/*
		case http.MethodPost:
			w.Write([]byte(http.MethodPost))
	*/
	case http.MethodPut:
		w.Write([]byte(http.MethodPut))
	case http.MethodDelete:
		w.Write([]byte(http.MethodDelete))
	case http.MethodPatch:
		//api.patchDataset(w, r, user, id)
		return

	default:
		//http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte("got Metadata " + r.Method + " request for id " + id.String() + "\n"))
}

// getDataset retrieves a dataset's whole blob or part thereof depending on the path.
// Not all datasets are fully viewable through the API.
func (api *DatasetRouter) getDataset(w http.ResponseWriter, r *http.Request, user uuid.UUID, id uuid.UUID, path string) {
	fmt.Println("id:", id, "path:", path)

	// whole dataset is not visible through this API
	// TODO: plug super-user check
	if false && path == "" {
		jsonError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	//res, err := api.db.SmartGetWithOwner(id, user)
	res, err := api.db.ViewDatasetWithOwner(id, user)
	if err != nil {
		if err == psql.ErrNotOwner {
			jsonError(w, err.Error(), http.StatusForbidden)
			return
		} else if err == psql.ErrNotFound {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}

		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(res)
	return
}

func (api *DatasetRouter) createDataset(w http.ResponseWriter, r *http.Request, creator uuid.UUID) {
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

	typed, err := models.CreateDatasetFromJson(creator, r.Body)
	if err != nil {
		api.logger.Error().Err(err).Msg("create dataset failed")
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Add Fairdata IdP information
	/*
		var identity, organisation string
		if token, ok := jwt.FromContext(r.Context()); ok {
			identity = token.Subject()
			if v, ok := token.Get(`Organisation`); ok {
				organisation = v.(string)
			}
		}
	*/

	err = api.db.Store(typed.Unwrap())
	if err != nil {
		jsonError(w, "store failed", http.StatusBadRequest)
		return
	}

	api.Created(w, r, typed.Unwrap().Id)
	return
}
