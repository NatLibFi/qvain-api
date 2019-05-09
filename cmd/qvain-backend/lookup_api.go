package main

import (
	"net/http"
	"strings"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/francoispqt/gojay"
	"github.com/wvh/uuid"
)

const (
	// caching time for "found" lookups
	defaultLookupPositiveCacheSecs = "28800" // 8h

	// caching time for "not found" lookups
	defaultLookupNegativeCacheSecs = "300" // 5m
)

// LookupApi holds the configuration for the identifier lookup service.
type LookupApi struct {
	db          *psql.DB
	frontendURL string
	apiURL      string
}

// NewLookupApi sets up a basic identifier lookup service.
func NewLookupApi(db *psql.DB) *LookupApi {
	return &LookupApi{
		db:          db,
		frontendURL: "/dataset/",
		apiURL:      "/api/datasets/",
	}
}

// ServeHTTP is the main entry point for the Lookup API.
func (api *LookupApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, rest := ShiftPath(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if head == "" {
			jsonError(w, "expected lookup field name", http.StatusBadRequest)
			return
		}
		// leading slash hasn't been stripped yet
		if len(rest) <= 1 {
			jsonError(w, "expected lookup value", http.StatusBadRequest)
			return
		}
		api.Lookup(w, r, head, rest[1:])

	case http.MethodOptions:
		apiWriteOptions(w, "GET, OPTIONS")

	default:
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	//jsonError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	return
}

// Lookup queries the database for indexed identifiers and redirects the response to the relevant endpoint if found.
// The response can be in HTML or JSON format depending on the request's Accept header.
func (api *LookupApi) Lookup(w http.ResponseWriter, r *http.Request, field, value string) {
	var (
		wantJson = strings.HasPrefix(r.Header.Get("Accept"), "application/json")
		ctError  func(http.ResponseWriter, string, int)
		id       uuid.UUID
		err      error
	)

	if wantJson {
		ctError = jsonError
	} else {
		ctError = http.Error
	}

	switch field {
	case "qvain":
		id, err = uuid.FromString(value)
		if err != nil {
			ctError(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = api.db.LookupByQvainId(id)

	case "fairdata":
		if !isValidIdentifier(value) {
			ctError(w, "that doesn't look like a valid dataset identifier", http.StatusBadRequest)
			return
		}
		id, err = api.db.LookupByFairdataIdentifier(value)

	default:
		ctError(w, "unknown index field", http.StatusBadRequest)
		return
	}
	if err != nil {
		w.Header().Set("Vary", "Accept")

		if err == psql.ErrNotFound {
			w.Header().Set("Cache-Control", "public, max-age="+defaultLookupNegativeCacheSecs)
			ctError(w, err.Error(), http.StatusNotFound)
		} else {
			ctError(w, "database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Vary", "Accept")
	w.Header().Set("Cache-Control", "public, max-age="+defaultLookupPositiveCacheSecs)
	if wantJson {
		apiWriteHeaders(w)
	}
	http.Redirect(w, r, api.makeRedirURL(id.String(), wantJson), http.StatusSeeOther)

	if wantJson {
		enc := gojay.BorrowEncoder(w)
		defer enc.Release()

		enc.AppendByte('{')
		enc.AddIntKey("status", http.StatusOK)
		enc.AddStringKey("msg", "found")
		enc.AddStringKey("id", id.String())
		enc.AppendByte('}')
		enc.Write()
	}
	return
}

// makeRedirURL makes a redirect URL to a given dataset's frontend or backend endpoint.
func (api *LookupApi) makeRedirURL(id string, wantAPI bool) string {
	if wantAPI {
		return api.apiURL + id
	}
	return api.frontendURL + id
}

// isValidIdentifier is a dumb character filter that "validates" UUID and URI style identifiers.
//
// The prime purpose is to prevent strings that clearly can't be a valid identifier going into SQL or JSON.
// Specifically, it does not try to parse URN, URI or other syntaxes.
//
// Accepted characters: 0-9 A-Z a-z - . / : _ { | }
func isValidIdentifier(id string) bool {
	var c byte
	for i := 0; i < len(id); i++ {
		c = id[i]

		//    -./ 0-9 :               @ A-Z                  _           a-z { | }
		if !((c >= 45 && c <= 58) || (c >= 64 && c <= 90) || c == 95 || (c >= 97 && c <= 125)) {
			return false
		}
	}
	return true
}
