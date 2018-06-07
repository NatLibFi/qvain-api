package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
)

// HashModeString needs to be appended to the base url for web apps using the fragment part for routing.
const HashModeString = "#/"

// LoginHandler contains the configuration for the front-end login handler.
type LoginHandler struct {
	logger              zerolog.Logger
	maybeHashModeString string
	host                string
}

// NewLoginHandler returns a configured LoginHandler.
func NewLoginHandler(logger zerolog.Logger, needsHashMode bool) *LoginHandler {
	return &LoginHandler{
		logger: logger,
		host:   "localhost:8080",
		maybeHashModeString: func() string {
			if needsHashMode {
				return HashModeString
			}
			return ""
		}(),
	}
}

func (lh *LoginHandler) Callback(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("Welcome to the cb url.\n"))
	fmt.Printf("\n%#v\n%+v\n\n", r.URL)

	q := r.URL.Query()
	if q.Get("id_token") == "" {
		//w.Write([]byte("id_token missing\n"))
		http.Error(w, "id_token missing", http.StatusForbidden)
		return
	}

	//w.Write([]byte("id_token: " + q.Get("id_token") + "\n"))
	//http.Redirect(w, r, r.URL.Scheme+"localhost:8080/"+lh.maybeHashModeString+"token#"+q.Get("id_token"), http.StatusFound)
	lh.logger.Debug().Str("scheme", r.URL.Scheme).Msgf("%#v", r.URL)
	lh.logger.Debug().Msgf("%#v", r)
	http.Redirect(w, r, "http://localhost:8080/"+lh.maybeHashModeString+"token#"+q.Get("id_token"), http.StatusFound)
}
