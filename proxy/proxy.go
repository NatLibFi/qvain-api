package proxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"sync"
	"encoding/base64"
)

const BufferSize = 8192

type bufferPool struct {
	pool *sync.Pool
}

func NewBufferPool() httputil.BufferPool {
	return &bufferPool{
		pool: new(sync.Pool),
	}
}

func (b *bufferPool) Get() []byte {
	buf := b.pool.Get()
	if buf == nil {
		return make([]byte, BufferSize)
	}
	return buf.([]byte)
}

func (b *bufferPool) Put(buf []byte) {
	b.pool.Put(buf)
}


func NewSingleHostReverseProxy(target *url.URL, opts ...proxyOption) *httputil.ReverseProxy {
	config := new(proxyConfig)

	for _, f := range opts {
		f(config)
	}

	targetQuery := target.RawQuery

	target.Path = strings.TrimSuffix(target.Path, "/")

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		//req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.URL.Path = target.Path + req.URL.Path

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		req.Header.Set("User-Agent", config.UserAgent)
		if config.AuthHeader != "" {
			req.Header.Set("Authorization", config.AuthHeader)
		}

		req.Host = target.Host
		fmt.Printf("sent: %+v\n", req)
	}
	//return &httputil.ReverseProxy{Director: director, BufferPool: NewBufferPool() }
	return &httputil.ReverseProxy{Director: director}
}

func Get(url string) {
	fmt.Println("GET", url)
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	duration := time.Since(start)

	fmt.Printf("%s", b)
	fmt.Printf("%v\n", duration)
}

// proxyConfig holds the fields that can be configured for a proxy.
type proxyConfig struct {
	UserAgent string
	AuthHeader string
}

// proxyOption is an option that can be configured when creating a proxy instance.
type proxyOption func(*proxyConfig)

// WithUserAgent sets the user-agent of outgoing requests.
func WithUserAgent(ua string) proxyOption {
	return func(config *proxyConfig) {
		config.UserAgent = ua
	}
}

// WithBasicAuth sets the basic auth header for the given user and password.
func WithBasicAuth(user, pass string) proxyOption {
	return func(config *proxyConfig) {
		if user != "" && pass != "" {
			config.AuthHeader = "Basic " + basicAuth(user, pass)
		}
	}
}

// basicAuth creates a basic auth header.
// This was lifted from the Go http library so we can cache and reuse the result.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// singleJoiningSlash makes sure the paths get appended without double slash between them.
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
		case aslash && bslash:
			return a + b[1:]
		case !aslash && !bslash:
			return a + "/" + b
	}
	return a + b
}
