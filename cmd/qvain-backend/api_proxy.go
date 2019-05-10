package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"syscall"

	"github.com/CSCfi/qvain-api/internal/sessions"
	"github.com/CSCfi/qvain-api/internal/version"
	"github.com/CSCfi/qvain-api/pkg/proxy"
	"github.com/rs/zerolog"
)

// ApiProxy is a reverse proxy.
type ApiProxy struct {
	proxy    *httputil.ReverseProxy
	sessions *sessions.Manager
	logger   zerolog.Logger
}

// makeProxyErrorHandler makes a callback function to handle errors happening inside the proxy.
func makeProxyErrorHandler(logger zerolog.Logger) func(http.ResponseWriter, *http.Request, error) {
	// log only every N proxy error
	//logger = logger.Sample(&zerolog.BasicSampler{N: 3})
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Info().Err(err).Msg("upstream error")
		jsonError(w, convertNetError(err), http.StatusBadGateway)
	}
}

// NewApiProxy creates a reverse web proxy that uses HTTP Basic Authentication. Used for allowing
// the front-end user access to the Metax files api. Since this allows the user to access Metax using
// Qvain service credentials, care needs to be taken that users cannot perform actions they shouldn't
// have access to.
func NewApiProxy(upstreamURL string, user string, pass string, sessions *sessions.Manager, logger zerolog.Logger) *ApiProxy {
	upUrl, err := url.Parse(upstreamURL)
	if err != nil {
		logger.Error().Err(err).Str("url", upstreamURL).Msg("can't parse upstream url")
	}

	return &ApiProxy{
		proxy: proxy.NewSingleHostReverseProxy(
			upUrl,
			proxy.WithBasicAuth(user, pass),
			proxy.WithErrorHandler(makeProxyErrorHandler(logger)),
			proxy.WithUserAgent(version.Id+"/"+version.CommitTag),
		),
		sessions: sessions,
		logger:   logger,
	}
}

// ServeHTTP proxies user requests to Metax so the front-end can query project information from Metax.
// The query is checked against the user session to make sure that users can only query projects
// they have access to.
func (api *ApiProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug().Str("path", r.URL.Path).Msg("request path")

	// make sure the user is authenticated
	session, err := api.sessions.UserSessionFromRequest(r)
	if err != nil {
		sessionError(w, err)
		return
	}

	if len(session.User.Projects) < 1 {
		jsonError(w, "access denied: no projects", http.StatusForbidden)
		return
	}

	// allow users to query only projects in their login token
	if !session.User.HasProject(r.URL.Query().Get("project")) {
		api.logger.Debug().Strs("projects", session.User.Projects).Str("wanted", r.URL.Query().Get("project")).Msg("project check")
		jsonError(w, "access denied: invalid project", http.StatusForbidden)
		return
	}

	api.proxy.ServeHTTP(w, r)
}

// convertNetError tries to catch (package) net and syscall errors and give a friendlier description.
// TODO: move this elsewhere?
func convertNetError(err error) string {
	if err == nil {
		return "no error"
	}

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return "connection timeout"
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			return "unknown host"
		}
		if t.Op == "read" {
			return "connection refused"
		}
	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return "connection refused"
		}
	}

	// fallback to simple Bad Gateway error
	return http.StatusText(http.StatusBadGateway)
}
