package main

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/NatLibFi/qvain-api/internal/secmsg"
	"github.com/NatLibFi/qvain-api/models"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/NatLibFi/qvain-api/sessions"

	"github.com/francoispqt/gojay"
	"github.com/rs/zerolog"
)

type SessionApi struct {
	sessions    *sessions.Manager
	db          *psql.DB
	messenger   *secmsg.MessageService
	logger      zerolog.Logger
	allowCreate bool
}

func NewSessionApi(sessions *sessions.Manager, db *psql.DB, msgsvc *secmsg.MessageService, logger zerolog.Logger) *SessionApi {
	return &SessionApi{sessions: sessions, db: db, messenger: msgsvc, logger: logger}
}

func (api *SessionApi) AllowCreate(allowed bool) {
	api.allowCreate = allowed
}

func (api *SessionApi) Login(w http.ResponseWriter, r *http.Request) {
	uid, isNew, err := api.db.RegisterIdentity("qvain", "wvh@example.com")
	if err != nil {
		//jsonError(w, err.Error(), http.StatusBadRequest)
		dbError(w, err)
		return
	}

	apiWriteHeaders(w)
	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	enc.AppendByte('{')
	enc.AddStringKey("uid", uid.String())
	enc.AddBoolKey("isNew", isNew)
	enc.AppendByte('}')
	enc.Write()
}

func (api *SessionApi) Check(w http.ResponseWriter, r *http.Request) {
	uid, err := api.db.GetUidForIdentity("qvain", "jack@example.com")
	if err != nil {
		dbError(w, err)
		return
	}

	apiWriteHeaders(w)
	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	enc.AppendByte('{')
	enc.AddStringKey("uid", uid.String())
	enc.AppendByte('}')
	enc.Write()
}

// Create creates a new session using the information from a secure message.
//
// This function does NOT create the application user in the database; it is meant for existing users.
// This function can be disabled by the application configuration.
func (api *SessionApi) Create(w http.ResponseWriter, r *http.Request) {
	// check if we got a request body
	if r.Body == nil {
		jsonError(w, "missing request body", http.StatusBadRequest)
		return
	}

	// don't check the content-type for now
	/*
		ct := r.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		} else if !strings.HasPrefix(ct, "application/octet-stream") {
			jsonError(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}
	*/

	// server will close request body
	rdr := io.LimitReader(r.Body, 4096)
	msg, err := ioutil.ReadAll(rdr)
	if err != nil || len(msg) < 1 {
		api.logger.Debug().Err(err).Bool("nil", msg == nil).Int("len", len(msg)).Msg("request body read failed")
		jsonError(w, "request body read failed", http.StatusBadRequest)
		return
	}

	// decode message
	obj, err := api.messenger.Decode(msg, -1)
	if err != nil {
		api.logger.Warn().Err(err).Msg("invalid message")
		jsonError(w, "invalid or expired message", http.StatusForbidden)
		return
	}
	api.logger.Debug().Str("obj", string(obj)).Int("len", len(obj)).Msg("obj")

	// decode message contents
	user, err := models.UserFromJson(obj)
	if err != nil {
		api.logger.Debug().Err(err).Msg("failed to decode decode message contents")
		jsonError(w, "failed to decode message contents", http.StatusBadRequest)
		return
	}

	// do registration?
	doRegistration := r.URL.RawQuery == "register" && user.Identity != "" && user.Service != ""
	var isNew bool
	if doRegistration {
		user.Uid, isNew, err = api.db.RegisterIdentity(user.Service, user.Identity)
		if err != nil {
			dbError(w, err)
			return
		}
	}

	// create session
	sid, err := api.sessions.NewLoginWithCookie(w, &user.Uid, user)
	if err != nil {
		api.logger.Debug().Err(err).Msg("failed to create session")
		jsonError(w, err.Error(), http.StatusForbidden)
		return
	}

	if doRegistration {
		api.logger.Info().Str("uid", user.Uid.String()).Str("sid", sid).Bool("registered", doRegistration).Bool("new", isNew).Msg("new session")
	} else {
		api.logger.Info().Str("uid", user.Uid.String()).Str("sid", sid).Msg("new session")
	}

	apiWriteHeaders(w)
	w.WriteHeader(http.StatusCreated)
	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	enc.AppendByte('{')
	enc.AddIntKey("status", http.StatusCreated)
	enc.AddStringKey("msg", "created")
	enc.AddStringKey("sid", sid)
	enc.AddStringKey("uid", user.Uid.String())
	if doRegistration {
		enc.AddBoolKey("registered", doRegistration)
		enc.AddBoolKey("new", isNew)
	}
	enc.AppendByte('}')
	enc.Write()
}

// Current dumps the (public) data from the current session in json format to the response.
func (api *SessionApi) Current(w http.ResponseWriter, r *http.Request) {
	session, err := api.sessions.SessionFromRequest(r)
	if err != nil {
		api.logger.Debug().Err(err).Msg("no current session")
		jsonError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	apiWriteHeaders(w)
	err = enc.EncodeObject(session.Public())
	if err != nil {
		api.logger.Error().Err(err).Msg("failed to encode public session")
		jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// ServeHTTP satisfies the http.Handler interface; it is the main endpoint for the session api.
func (api *SessionApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug().Str("path", r.URL.Path).Msg("request path")

	if r.URL.Path != "/" {
		jsonError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		api.Current(w, r)
	case http.MethodPost:
		if api.allowCreate {
			api.Create(w, r)
		} else {
			jsonError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}
	case http.MethodOptions:
		apiWriteOptions(w, "GET, POST, OPTIONS")
	default:
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
