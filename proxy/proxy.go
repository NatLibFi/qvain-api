package proxy

import (
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
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

		// combine query values
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		// get rid of the original request's user-agent
		if config.UserAgent != "" {
			req.Header.Set("User-Agent", config.UserAgent)
		} else {
			req.Header.Del("User-Agent")
		}

		// add auth creds if necessary
		if config.AuthHeader != "" {
			req.Header.Set("Authorization", config.AuthHeader)
		}

		req.Host = target.Host
		//log.Printf("sent: %+v", req)
	}

	// shorter timeouts and higher per-host concurrency
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
			//Timeout:   20 * time.Millisecond,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        16,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
	}
	//return &httputil.ReverseProxy{Director: director, BufferPool: NewBufferPool() }
	return &httputil.ReverseProxy{Director: director, Transport: transport, ErrorHandler: config.ErrorHandler}
}

// proxyConfig holds the fields that can be configured for a proxy.
type proxyConfig struct {
	UserAgent    string
	AuthHeader   string
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
}

// proxyOption is an option that can be configured when creating a proxy instance.
type proxyOption func(*proxyConfig)

// WithUserAgent sets the user-agent of outgoing requests.
func WithUserAgent(ua string) proxyOption {
	return func(config *proxyConfig) {
		config.UserAgent = ua
	}
}

// WithErrorHandler sets the ErrorHandler callback function in Go's ReverseProxy.
func WithErrorHandler(handler func(http.ResponseWriter, *http.Request, error)) proxyOption {
	return func(config *proxyConfig) {
		config.ErrorHandler = handler
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
