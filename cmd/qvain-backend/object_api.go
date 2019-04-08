package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
)

type ObjectApi struct {
	db     *psql.DB
	logger zerolog.Logger
}

func NewObjectApi(db *psql.DB, logger zerolog.Logger) *ObjectApi {
	return &ObjectApi{db: db, logger: logger}
}

func (api *ObjectApi) CreateObject(user uuid.UUID, family int, schema string, objtype string, val []byte) error {
	api.logger.Info().
		Int("family", family).
		Str("schema", schema).
		Str("type", objtype).
		Msg("creating object")

	return nil
}

func (api *ObjectApi) ListObjects(user uuid.UUID) []byte {
	return []byte(`[
		{"name": "Wouter Van Hemel", "email": "wvh@example.com", "https://orcid.org/0000-0001-7695-4511"},
		{"name": "Esa-Pekka Keskitalo", "email": "epk@example.com", "https://orcid.org/0000-0002-4411-8452"},
		{"name": "Jessica Parland-von Essen", "email": "jpve@example.com", "https://orcid.org/0000-0003-4460-3906"},
	]`)
}

func (api *ObjectApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string

	ShiftUrl(r)
	head = ShiftUrl(r)
	//api.logger.Debug().Str("head", head).Str("path", r.URL.Path).Msg("request at Router()")

	if r.URL.Path != "/" {
		oid, err := strconv.Atoi(head)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid object %q", head), http.StatusBadRequest)
			return
		}
		api.logger.Debug().Int("oid", oid).Msg("request")
	}

	switch r.Method {
	case "GET":
		//fmt.Println("here")
		w.Write(api.ListObjects(uuid.MustNewUUID()))
	case "POST":
		api.CreateObject(uuid.MustNewUUID(), 2, "att-ida", "person", []byte(`{"name": "jack", "email": "jack@example.com"}`))
	default:
		http.Error(w, "Only GET and PUT are allowed", http.StatusMethodNotAllowed)
	}
	//http.Error(w, "Not Found", http.StatusNotFound)
}

func (api *ObjectApi) Create(w http.ResponseWriter, r *http.Request) {}
func (api *ObjectApi) List(w http.ResponseWriter, r *http.Request)   {}

func (api *ObjectApi) Get(w http.ResponseWriter, r *http.Request)    {}
func (api *ObjectApi) Put(w http.ResponseWriter, r *http.Request)    {}
func (api *ObjectApi) Delete(w http.ResponseWriter, r *http.Request) {}
