package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/models"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/NatLibFi/qvain-api/version"
	"github.com/wvh/uuid"
	//"encoding/json"
)

/*
var fakeDatasetMap = map[uuid.UUID]string{
	uuid.MustFromString("12345678901234567890123456789012"): "this is fake dataset 12345678901234567890123456789012",
	uuid.MustFromString("66666666666666666666666666666666"): "this is fake dataset 66666666666666666666666666666666",
	uuid.MustFromString("00000000000000000000000000000001"): "this is fake dataset 00000000000000000000000000000001",
}
*/
var owner = uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")

var fakeDatasets = []*models.Dataset{
	{Id: uuid.MustFromString("12345678901234567890123456789012"), Creator: owner, Owner: owner},
	{Id: uuid.MustFromString("12345678901234567890123456789012"), Creator: owner, Owner: owner},
	{Id: uuid.MustFromString("12345678901234567890123456789012"), Creator: owner, Owner: owner},
	{Id: uuid.MustFromString("056bffbcc41edad4853bea9100000001"), Creator: owner, Owner: owner},
}

var fakeDatasetMap map[uuid.UUID]*models.Dataset

func init() {
	fakeDatasetMap = make(map[uuid.UUID]*models.Dataset)
	for _, dataset := range fakeDatasets {
		fakeDatasetMap[dataset.Id] = dataset
	}
}

func checkDatasetExists(id uuid.UUID) bool {
	_, e := fakeDatasetMap[id]
	return e
}

/*
func needsDataset(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("doc") == "" {
			//http.Error(w, "missing dataset id", http.StatusUnauthorized)
			jsonError(w, "missing dataset id", http.StatusBadRequest)
			return
		}
		if !checkDatasetExists(r.URL.Query().Get("doc")) {
			jsonError(w, "dataset not found", http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}
*/

func apiDatasetCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(http.MethodGet))
	case http.MethodPost:
		w.Write([]byte(http.MethodPost))
	case http.MethodPut:
		w.Write([]byte(http.MethodPut))
	case http.MethodDelete:
		w.Write([]byte(http.MethodDelete))
	default:
		//http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte("done\n"))
}

type DatasetRouter struct {
	mountedAt string
	db        *psql.DB
}

func NewDatasetRouter(mountPoint string, db *psql.DB) *DatasetRouter {
	return &DatasetRouter{mountedAt: path.Clean(mountPoint) + "/", db: db}
}

func (api *DatasetRouter) Mountpoint() string {
	return api.mountedAt
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
		/*
			case http.MethodGet:
				w.Write([]byte(http.MethodGet))
		*/
		case http.MethodPost:
			//w.Write([]byte(http.MethodPost))
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
	if _, e := fakeDatasetMap[id]; !e {
		jsonError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

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
		api.patchDataset(w, r, user, id)
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

	// determine what sort of dataset family we're dealing with
	//if fam, found := models.LookupFamily(2); found {
	if fam := models.LookupFamily(2); fam != nil {
		fmt.Println("found fam:", fam.Name)
		// this part of the dataset is not public
		if !fam.IsPathPublic(path) {
			jsonError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}
	// else: panic?

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

	decoder := json.NewDecoder(r.Body)
	var blob json.RawMessage
	err = decoder.Decode(&blob)
	defer r.Body.Close()
	if err != nil {
		//jsonError(w, err.Error(), 500)
		jsonError(w, "invalid json", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	dataset, err := models.NewDataset(creator)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataset.SetData(2, "metax", blob)

	err = api.db.Store(dataset)
	if err != nil {
		jsonError(w, "store failed", http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("%s %d\n", creator, r.ContentLength)))
	return
}

//func (api *DatasetRouter) putDataset()

// patchDataset allows changing a dataset's top fields.
func (api *DatasetRouter) patchDataset(w http.ResponseWriter, r *http.Request, user uuid.UUID, id uuid.UUID) {
	w.Write([]byte(http.MethodPatch))

	decoder := json.NewDecoder(r.Body)
	var t = struct {
		Owner   uuid.UUID `json:"owner"`
		Creator uuid.UUID `json:"creator"`
	}{}
	err := decoder.Decode(&t)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	fmt.Println(t.Owner)
}

func PatchDataset(id uuid.UUID) error {
	return nil
}

func (api *DatasetRouter) ViewDatasets(w http.ResponseWriter, r *http.Request, user uuid.UUID) {

}

// ChangeOwner sets the owner to another allowed UUID value, either the user's own or one of the user's groups.
// This is a higher-level function that is not in the model since the storage layer is not aware of group memberships.
func ChangeOwner(id, owner uuid.UUID) error {
	fmt.Printf("changing owner for dataset %s to %s\n", id, owner)
	return nil
}

func ViewMetadata(w http.ResponseWriter, r *http.Request, id string) {

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
