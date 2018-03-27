// Package metax provides a client for the CSC MetaX API.
package metax

/*
 *	links:
 *	https://metax-test.csc.fi
 *	https://metax-test.csc.fi/rest/datasets/pid:urn:cr3
 *  https://metax-test.csc.fi/rest/datasets/?owner_id=053bffbcc41edad4853bea91fc42ea18
 */

import (
	"fmt"
	"log"
	"time"
	//"runtime"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/NatLibFi/qvain-api/version"
)

const (
	VERBOSE = false

	DATASETS_ENDPOINT = "/rest/datasets/"
)

var UserAgent string = "qvain " + version.CommitHash

func init() {
	//UserAgent = UserAgent + " (" + runtime.Version() + ")"
}

type MetaxService struct {
	host         string
	baseUrl      string
	client       *http.Client
	disableHttps bool

	urlDatasets string
}

type MetaxOption func(*MetaxService)

func DisableHttps(svc *MetaxService) {
	svc.disableHttps = true
}

// NewMetaxService returns a Metax API client.
func NewMetaxService(host string, params ...MetaxOption) *MetaxService {
	svc := &MetaxService{
		host: host,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	for _, param := range params {
		param(svc)
	}

	if svc.disableHttps {
		svc.baseUrl = "http://" + svc.host
	} else {
		svc.baseUrl = "https://" + svc.host
	}
	//fmt.Println(UserAgent, version.CommitHash)

	svc.makeEndpoints(svc.baseUrl)

	return svc
}

func (api *MetaxService) makeEndpoints(base string) {
	api.urlDatasets = base + DATASETS_ENDPOINT
}

type PaginatedResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []*MetaxRecord `json:"results"`
}

func (api *MetaxService) UrlForId(id string) string {
	return api.urlDatasets + id
}

type DatasetOption func(*http.Request)

func WithOwner(uid string) DatasetOption {
	return func(req *http.Request) {
		// NOTE: might need QueryEscape()
		//req.URL.Values.Set("owner_id", uid)
		// don't set parameter if empty
		if uid == "" {
			return
		}
		qvals := url.Values{}
		qvals.Add("owner_id", uid)
		req.URL.RawQuery = qvals.Encode()
	}
}

func Since(t time.Time) DatasetOption {
	return func(req *http.Request) {
		req.Header.Set("If-Modified-Since", t.UTC().Format(http.TimeFormat))
	}
}

func WithStreaming(req *http.Request) {
	qvals := req.URL.Query()
	qvals.Add("stream", "true")
	qvals.Add("no_pagination", "true")
	req.URL.RawQuery = qvals.Encode()
}

func (api *MetaxService) getRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	return req, nil
}

func (api *MetaxService) Datasets(params ...DatasetOption) (*PaginatedResponse, error) {
	req, err := http.NewRequest("GET", api.urlDatasets, nil)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	for _, param := range params {
		param(req)
	}

	//log.Printf("request headers: %+v\n", req)
	log.Printf("metax: GET %v\n", req.URL)

	start := time.Now()
	count := 0
	defer func() {
		log.Printf("metax: paginated query processed in %v (count: %d)", time.Since(start), count)
	}()

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if VERBOSE {
		log.Printf("response: %+v\n", res)
		log.Println("Content-Length:", res.Header.Get("Content-Length"))
		log.Println("Content-Type:", res.Header["Content-Type"])
		log.Printf("%T\n", res.Header["Content-Type"])
		log.Println("Status code:", res.StatusCode)
	}

	//clientHeaders := []string{"If-Modified-Since", "If-Unmodified-Since", "If-Match", "If-None-Match", "If-Range", "ETag"}
	//serverHeaders := []string{"Last-Modified", "ETag", "Cache-Control"}

	switch res.StatusCode {
	case 200:
	case 404:
		return nil, fmt.Errorf("error: not found (code: %d)", res.StatusCode)
	case 403:
		return nil, fmt.Errorf("error: forbidden (code: %d)", res.StatusCode)
	default:
		return nil, fmt.Errorf("error: can't retrieve record (code: %d)", res.StatusCode)
	}

	if len(res.Header["Content-Type"]) != 1 || res.Header["Content-Type"][0] != "application/json" {
		return nil, fmt.Errorf("unknown content-type, expected json")
	}

	/*
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s", body), nil
	*/

	var page PaginatedResponse
	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		if n, err2 := io.Copy(ioutil.Discard, res.Body); n > 0 || err != nil {
			log.Printf("drained response: bytes=%d, err=%s", n, err2)
		}
		return nil, err
	}
	if n, err := io.Copy(ioutil.Discard, res.Body); n > 0 || err != nil {
		log.Printf("drained response: bytes=%d, err=%s", n, err)
	}
	count = page.Count

	return &page, nil
}

func (api *MetaxService) ReadStream(params ...DatasetOption) ([]MetaxRecord, error) {
	req, err := http.NewRequest("GET", api.urlDatasets+"?stream=true&no_pagination=true", nil)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	for _, param := range params {
		param(req)
	}
	WithStreaming(req)

	log.Printf("request headers: %+v\n", req)

	start := time.Now()
	defer func() {
		log.Printf("metax: stream query processed in %v", time.Since(start))
	}()

	res, err := api.client.Do(req)
	if err != nil {
		return noRecords, err
	}
	defer res.Body.Close()

	//clientHeaders := []string{"If-Modified-Since", "If-Unmodified-Since", "If-Match", "If-None-Match", "If-Range", "ETag"}
	//serverHeaders := []string{"Last-Modified", "ETag", "Cache-Control"}

	switch res.StatusCode {
	case 200:
	case 404:
		return noRecords, fmt.Errorf("error: not found (code: %d)", res.StatusCode)
	case 403:
		return noRecords, fmt.Errorf("error: forbidden (code: %d)", res.StatusCode)
	default:
		return noRecords, fmt.Errorf("error: can't retrieve record (code: %d)", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" && res.Header.Get("Content-Type") != "application/json; charset=utf=8" {
		//if (len(res.Header["Content-Type"]) != 1 || res.Header["Content-Type"][0]!= "application/json") {
		return noRecords, fmt.Errorf("unknown content-type, expected json")
	}

	if res.Header.Get("X-Count") == "" {
		log.Println("metax: missing X-Count header in streaming response")
	} else {
		log.Println("metax: x-count:", res.Header.Get("X-Count"))
	}

	recs := make([]MetaxRecord, 0, 0)

	dec := json.NewDecoder(res.Body)
	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	var rec MetaxRecord
	// while the array contains values
	for dec.More() {
		// decode an array value (Message)
		err := dec.Decode(&rec)
		if err != nil {
			return noRecords, err
		}
		//fmt.Printf("%v: %v\n", m.Name, m.Text)
		recs = append(recs, rec)
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		return noRecords, err
	}
	fmt.Printf("%T: %v\n", t, t)

	return recs, nil
}

func (api *MetaxService) GetId(id string) (string, error) {
	req, err := http.NewRequest("GET", api.UrlForId(id), nil)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	log.Printf("request headers: %+v\n", req)
	res, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if VERBOSE {
		log.Printf("response: %+v\n", res)
		log.Println("Content-Length:", res.Header.Get("Content-Length"))
		log.Println("Content-Type:", res.Header["Content-Type"])
		log.Printf("%T\n", res.Header["Content-Type"])
		log.Println("Status code:", res.StatusCode)
	}

	//clientHeaders := []string{"If-Modified-Since", "If-Unmodified-Since", "If-Match", "If-None-Match", "If-Range", "ETag"}
	//serverHeaders := []string{"Last-Modified", "ETag", "Cache-Control"}

	switch res.StatusCode {
	case 200:
	case 404:
		return "", fmt.Errorf("error: not found (code: %d)", res.StatusCode)
	case 403:
		return "", fmt.Errorf("error: forbidden (code: %d)", res.StatusCode)
	default:
		return "", fmt.Errorf("error: can't retrieve record (code: %d)", res.StatusCode)
	}

	if len(res.Header["Content-Type"]) != 1 || res.Header["Content-Type"][0] != "application/json" {
		return "", fmt.Errorf("unknown content-type, expected json")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", body), nil
}

func (api *MetaxService) ParseRecord(txt string) (*MetaxRecord, error) {
	rec := new(MetaxRecord)
	err := json.Unmarshal([]byte(txt), &rec)
	if err != nil {
		return &MetaxRecord{}, err
	}
	return rec, nil
}

func (api *MetaxService) ParseRootKeys(txt string) ([]string, error) {
	var top map[string]interface{}
	//make(map[string]interface{})
	var keys []string

	err := json.Unmarshal([]byte(txt), &top)
	if err != nil {
		return []string{}, err
	}
	for k := range top {
		keys = append(keys, k)
	}
	return keys, nil
}

func (api *MetaxService) ParseList(txt string) (*PaginatedResponse, error) {
	list := new(PaginatedResponse)
	err := json.Unmarshal([]byte(txt), &list)
	if err != nil {
		return &PaginatedResponse{}, err
	}
	return list, nil
}
