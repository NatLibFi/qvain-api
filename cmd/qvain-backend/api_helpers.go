package main

import (
	"fmt"
	"net/http"
	"strings"
	//"encoding/json"

	"github.com/NatLibFi/qvain-api/psql"
	"github.com/NatLibFi/qvain-api/version"
)

// jsonError generates an HTTP error but in json format.
// NOTE: This function uses simple string formatting which is faster than json encoding for small responses; no json escaping is done.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	fmt.Fprintf(w, `{"error":{"code":%d,"message":"%s"}}%c`, code, msg, '\n')
}

func smartError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if strings.HasPrefix(r.Header.Get("Accept"), "application/json") {
		jsonError(w, msg, code)
		return
	}
	http.Error(w, msg, code)
}

// apiHello catches all requests to the bare api endpoint.
func apiHello(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/api" && r.RequestURI != "/api/" {
		jsonError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`"` + version.Id + ` api"` + "\n"))
}

// apiVersion returns the version information that was (hopefully) linked in at build time.
func apiVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", `"`+version.CommitHash+`"`)
	fmt.Fprintf(w, `{"name":"%s","description":"%s","version":"%s","tag":"%s","hash":"%s","repo":"%s"}%c`, version.Name, version.Description, version.SemVer, version.CommitTag, version.CommitHash, version.CommitRepo, '\n')
}

func apiDatabaseCheck(db *psql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err := db.Check()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			//w.WriteHeader()
			fmt.Fprintf(w, `{"alive":false,"error":"%s"}%c`, err, '\n')
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alive":true}` + "\n"))
	})
}

// responseNotModified
/*
func responseNotModified(w http.ResponseWriter, r *http.Request, etag string) {
	//w.Header().Set("Content-Type", "application/json")
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	w.WriteHeader(http.StatusNotModified)
}
*/
