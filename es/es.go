// Package es contains a minimal API client to query Elastic Search and export complete indexes.
//
// As of now, it can only return all results for a given index.
package es

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Configuration for http client
const (
	HttpClientTimeout = time.Duration(5 * time.Second)
	HttpUserAgent     = "qvain"
	HttpQueryString   = "_search?size=10000"
)

// Configuration for json path getter
const (
	FilterSourcesPath = "hits.hits.#._source"
	FilterCountPath   = "hits.hits.#"
)

// ESClient describes the parameters needed to query an Elastic Search instance.
type ESClient struct {
	baseUrl    string
	httpClient *http.Client
}

// NewClient creates a new client for Elastic Search.
// It takes a url parameter pointing to the root of the ES API endpoint.
func NewClient(url string) *ESClient {
	// strip trailing slash
	if url != "" && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}

	return &ESClient{
		baseUrl: url,
		httpClient: &http.Client{
			Timeout: HttpClientTimeout,
		},
	}
}

// AllForIndex does a _search query without search parameters for the specified index, returning all results.
func (es *ESClient) All(index, doctype string) ([]byte, error) {
	r, err := es.request(es.baseUrl + "/" + index + "/" + doctype + "/" + HttpQueryString)
	if err != nil {
		return nil, fmt.Errorf("error configuring request: %s", err)
	}

	res, err := es.httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error during request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error: expected http status code 200, got %d", res.StatusCode)
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return nil, fmt.Errorf("invalid content type, expected \"application/json\"")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	return body, nil
}

// Request sets up a http.Request.
func (es *ESClient) request(url string) (*http.Request, error) {
	// not sure what error this could actually cause...
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("UserAgent", HttpUserAgent)
	return r, nil
}

// Filter returns the contents of the Json _sources key from the Elastic Search response, effectively filtering out some ES noise.
func Filter(json []byte) []byte {
	return []byte(gjson.GetBytes(json, FilterSourcesPath).Raw)
}

// FilterAndCount returns the contents of the _sources key just like Filter(), but also returns a count of the number of elements to facilitate sanity checks.
// It is twice as slow as Filter().
func FilterAndCount(json []byte) ([]byte, int64) {
	parsed := gjson.ParseBytes(json)
	return []byte(parsed.Get(FilterSourcesPath).Raw), parsed.Get(FilterCountPath).Int()
}
