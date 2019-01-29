package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"syscall"

	"github.com/NatLibFi/qvain-api/proxy"
	"github.com/NatLibFi/qvain-api/sessions"
	"github.com/NatLibFi/qvain-api/version"

	"github.com/rs/zerolog"
)

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

func (api *ApiProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug().Str("path", r.URL.Path).Msg("request path")

	// auth
	session, err := api.sessions.UserSessionFromRequest(r)
	if err != nil {
		sessionError(w, err)
		return
	}

	if len(session.User.Projects) < 1 {
		jsonError(w, "access denied: no projects", http.StatusForbidden)
		return
	}

	//project := ShiftUrlWithTrailing(r)
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
