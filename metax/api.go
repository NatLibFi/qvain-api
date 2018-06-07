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
	//"log"
	"time"
	//"runtime"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	//"net/url"
	"strings"
	"bytes"
	"context"
	"errors"

	"github.com/NatLibFi/qvain-api/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	VERBOSE = false

	DATASETS_ENDPOINT = "/rest/datasets/"
)

var UserAgent string = "qvain " + version.CommitHash

var (
	errStreamMustBeArray = errors.New("stream is not a json array")
	errEmptyDataset = errors.New("dataset is empty")
)

func init() {
	//UserAgent = UserAgent + " (" + runtime.Version() + ")"
}

// MetaxService represents the Metax API server.
type MetaxService struct {
	host         string
	baseUrl      string
	client       *http.Client
	disableHttps bool
	logger       zerolog.Logger

	urlDatasets string

	user         string
	pass         string
}

type MetaxOption func(*MetaxService)

func DisableHttps(svc *MetaxService) {
	svc.disableHttps = true
}

func WithLogger(logger zerolog.Logger) MetaxOption {
	return func(svc *MetaxService) {
		svc.logger = logger
	}
}

func WithCredentials(user, pass string) MetaxOption {
	return func(svc *MetaxService) {
		svc.user = user
		svc.pass = pass
	}
}

// NewMetaxService returns a Metax API client.
func NewMetaxService(host string, params ...MetaxOption) *MetaxService {
	svc := &MetaxService{
		host: host,
		logger: zerolog.Nop(),
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

	//if svc.logger == nil {
	/*
	if svc.logger == (zerolog.Logger{}) {
		svc.logger = log.Logger
	}
	*/
	//svc.logger = zerolog.Nop()
	svc.logger = log.Logger
	//_ = log.Logger

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
		//qvals := url.Values{}
		qvals := req.URL.Query()
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

	api.logger.Printf("metax: GET %v\n", req.URL)

	start := time.Now()
	count := 0
	defer func() {
		api.logger.Printf("metax: paginated query processed in %v (count: %d)", time.Since(start), count)
	}()

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if VERBOSE {
		api.logger.Printf("response: %+v\n", res)
		api.logger.Print("Content-Length:", res.Header.Get("Content-Length"))
		api.logger.Print("Content-Type:", res.Header["Content-Type"])
		api.logger.Printf("%T\n", res.Header["Content-Type"])
		api.logger.Print("Status code:", res.StatusCode)
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
			api.logger.Printf("drained response: bytes=%d, err=%s", n, err2)
		}
		return nil, err
	}
	if n, err := io.Copy(ioutil.Discard, res.Body); n > 0 || err != nil {
		api.logger.Printf("drained response: bytes=%d, err=%s", n, err)
	}
	count = page.Count

	return &page, nil
}

// drainBody discards the response body when it's not read until the end, so the next keep-alive request can be handled with the same connection.
// Murky stuff; is this still/again needed in whatever Go version this is compiled with?
func (api *MetaxService) drainBody(body io.ReadCloser) {
	if n, err := io.Copy(ioutil.Discard, body); n > 0 || err != nil {
		api.logger.Printf("drained response: bytes=%d, err=%s", n, err)
	}
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

	api.logger.Printf("request headers: %+v\n", req)

	start := time.Now()
	defer func() {
		api.logger.Printf("metax: stream query processed in %v", time.Since(start))
	}()

	res, err := api.client.Do(req)
	if err != nil {
		return noRecords, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 404:
		return noRecords, fmt.Errorf("error: not found (code: %d)", res.StatusCode)
	case 403:
		return noRecords, fmt.Errorf("error: forbidden (code: %d)", res.StatusCode)
	default:
		return noRecords, fmt.Errorf("error: can't retrieve record (code: %d)", res.StatusCode)
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return noRecords, fmt.Errorf("unknown content-type, expected json")
	}

	if res.Header.Get("X-Count") == "" {
		api.logger.Print("metax: missing X-Count header in streaming response")
	} else {
		api.logger.Print("metax: x-count:", res.Header.Get("X-Count"))
	}

	recs := make([]MetaxRecord, 0, 0)

	dec := json.NewDecoder(res.Body)

	// start stream
	t, err := dec.Token()
	if err != nil {
		api.logger.Print(err)
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim.String() != "[" {
		return noRecords, errStreamMustBeArray
	}

	// streaming array values...
	for dec.More() {
		var rec MetaxRecord

		err := dec.Decode(&rec)
		if err != nil {
			return noRecords, err
		}
		recs = append(recs, rec)
	}

	// end stream
	t, err = dec.Token()
	if err != nil {
		return noRecords, err
	}
	fmt.Printf("%T: %v\n", t, t)

	return recs, nil
}

func (api *MetaxService) ReadStreamChannel(ctx context.Context, params ...DatasetOption) (chan *MetaxRawRecord, chan error, error) {
	req, err := http.NewRequest("GET", api.urlDatasets, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	for _, param := range params {
		param(req)
	}
	WithStreaming(req)

	api.logger.Printf("request headers: %+v\n", req)

	start := time.Now()
	defer func() {
		api.logger.Printf("metax: stream query processed in %v", time.Since(start))
	}()

	res, err := api.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	// WARNING: go routine below is responsible for closing the response body

	switch res.StatusCode {
		case 200:
		case 404:
			return nil, nil, fmt.Errorf("error: not found (code: %d)", res.StatusCode)
		case 403:
			return nil, nil, fmt.Errorf("error: forbidden (code: %d)", res.StatusCode)
		default:
			return nil, nil, fmt.Errorf("error: can't retrieve record (code: %d)", res.StatusCode)
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return nil, nil, fmt.Errorf("unknown content-type, expected json")
	}

	if res.Header.Get("X-Count") == "" {
		api.logger.Print("metax: missing X-Count header in streaming response")
	} else {
		api.logger.Print("metax: x-count:", res.Header.Get("X-Count"))
	}

	outc := make(chan *MetaxRawRecord)
	errc := make(chan error, 1)

	go func(stream io.ReadCloser) {
		defer func() {
			api.drainBody(stream)
			stream.Close()
			api.logger.Printf("metax: stream data processed in %v", time.Since(start))
		}()

		dec := json.NewDecoder(stream)

		// start stream
		t, err := dec.Token()
		if err != nil {
			api.logger.Print(err)
			errc <- err
			return
		}
		if delim, ok := t.(json.Delim); !ok || delim.String() != "[" {
			errc <- errStreamMustBeArray
			return
		}

		// while records in the array stream...
		for dec.More() {
			var rec MetaxRawRecord

			err := dec.Decode(&rec)
			if err != nil {
				errc <- err
				return
			}

			select {
			case outc <- &rec:
			case <-ctx.Done():
				api.logger.Print("context timeout triggered")
				return
			}
		}

		// end stream
		t, err = dec.Token()
		if err != nil {
			errc <- err
		}
		close(outc)
	}(res.Body)

	return outc, errc, nil
}


func (api *MetaxService) Create(ctx context.Context, blob json.RawMessage) (json.RawMessage, error) {
	if blob == nil || len(blob) < 1 {
		return nil, errEmptyDataset
	}

	req, err := http.NewRequest(http.MethodPost, api.urlDatasets, bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.SetBasicAuth(api.user, api.pass)

	fmt.Printf("%v\n", req)
	start := time.Now()
	defer func() {
		api.logger.Printf("metax: create processed in %v", time.Since(start))
	}()

	res, err := api.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	hasJson := strings.HasPrefix(res.Header.Get("Content-Type"), "application/json")
	var body []byte
	if hasJson {
		body, _ = ioutil.ReadAll(res.Body)
	}

	switch res.StatusCode {
		case 201:
			return body, nil
		case 400:
			return nil, &ApiError{"invalid dataset", body}
		case 401:
			return nil, &ApiError{"authorisation required", nil}
		default:
			return nil, &ApiError{fmt.Sprintf("API returned status code %d", res.StatusCode), nil}
	}

	fmt.Println("response Status:", res.Status)
	fmt.Println("response Headers:", res.Header)
	fmt.Println("response Body:", string(body))
	if strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return body, nil
	}

	return nil, nil
}

func (api *MetaxService) GetId(id string) (string, error) {
	req, err := http.NewRequest("GET", api.UrlForId(id), nil)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	api.logger.Printf("request headers: %+v\n", req)
	res, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

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
