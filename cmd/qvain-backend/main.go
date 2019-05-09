// Command qvain-backend is the backend server for the Qvain API.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/CSCfi/qvain-api/internal/version"
	"github.com/CSCfi/qvain-api/pkg/env" //"github.com/CSCfi/qvain-api/internal/jwt"
)

const (
	// service name, used for instance in logs
	ServiceName = "qvain"

	// http server setup
	HttpProxyPort = "8080"

	// timeouts
	HttpReadTimeout  = 10 * time.Second
	HttpWriteTimeout = 25 * time.Second
	HttpIdleTimeout  = 120 * time.Second

	// additional info message when Go web server returns
	strHttpServerPanic = "http server crashed"
)

// startHttpsRedirector spawns a background HTTP server that redirects to https://.
// NOTE: This function returns immediately.
func startHttpsRedirector(config *Config) {
	logger := config.NewLogger("main")
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + r.Host + r.URL.String()
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}),
		ReadTimeout:  HttpReadTimeout,
		WriteTimeout: HttpWriteTimeout,
		ErrorLog:     adaptToStdlibLogger(config.NewLogger("go.http")),
	}
	srv.SetKeepAlivesEnabled(false)
	logger.Info().Msg("starting https redirect server")
	go func() { logger.Fatal().Err(srv.ListenAndServe()).Msg(strHttpServerPanic) }()
}

// Defined command line flags (some of which take their defaults from the environment).
var (
	appDebug       = flag.Bool("d", env.GetBool("APP_DEBUG"), "log debug output (env APP_DEBUG)")
	appDevMode     = flag.Bool("dev", env.GetBool("APP_DEV_MODE"), "dev mode: debug, http-only, CORS:all (env APP_DEV_MODE)")
	disableLogging = flag.Bool("q", false, "quiet: disable all logging")
	disableHttpLog = flag.Bool("nrl", false, "disable http request logging")
	forceHttpOnly  = flag.Bool("http", env.GetBool("APP_FORCE_HTTP_SCHEME"), "use http for generated links (env APP_FORCE_HTTP_SCHEME)")
	appHttpPort    = flag.String("port", env.GetDefault("APP_HTTP_PORT", HttpProxyPort), "port to run web server on (env APP_HTTP_PORT)")
)

func main() {
	flag.Parse()

	// configure application from environment; exit if there was an error
	config, err := ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}

	// logger just for this main() function
	logger := config.NewLogger("main")
	setStdlibLogger(config.NewLogger("log"))

	if env.Get("APP_ENV_CHECK") == "" {
		logger.Warn().Msg("environment variable APP_ENV_CHECK is not set")
	}

	// initialise database pool
	err = config.initDB(config.NewLogger("psql"))
	if err != nil {
		logger.Error().Err(err).Msg("daba baad")
	}

	// initialise session manager
	err = config.initSessions()
	if err != nil {
		logger.Error().Err(err).Msg("session manager failed")
	}

	// initialise secure messaging service
	err = config.initMessenger()
	if err != nil {
		logger.Error().Err(err).Msg("secure messaging service initialisation failed")
	}

	// set up default handlers
	mux := makeMux(config)
	var handler http.Handler = mux
	_ = handler

	apis := NewApis(config)
	_ = apis

	// default server, without TLSConfig
	srv := &http.Server{
		//Handler:           authMux,
		Handler:           Root(config),
		ReadTimeout:       HttpReadTimeout,
		ReadHeaderTimeout: HttpReadTimeout,
		WriteTimeout:      HttpWriteTimeout,
		IdleTimeout:       HttpIdleTimeout,
		ErrorLog:          adaptToStdlibLogger(config.NewLogger("go.http")),
	}

	// if standalone, run on 443 and start redirecting port 80; else run on 8080 or whatever is configured above
	var listen string
	if config.Standalone {
		if can, err := canNetBindService(); err == nil {
			if !can {
				// print to STDERR, because the server will crash
				fmt.Fprintln(os.Stderr, "warning: need cap_net_bind_service capability to run stand-alone")
			}
		} else {
			logger.Error().Err(err).Msg("capability check returned error")
		}

		srv.TLSConfig = tlsIntermediateConfig
		listen = "*"
		config.Port = "https"
		startHttpsRedirector(config)
	} else {
		listen = "localhost"
		srv.Addr = listen + ":" + config.Port
	}

	logger.Info().
		Str("hash", version.CommitHash).
		Str("tag", version.CommitTag).
		Str("port", config.Port).
		Str("host", config.Hostname).
		Str("iface", listen).
		Bool("standalone", config.Standalone).
		Bool("debug", config.Debug).
		Bool("dev", config.DevMode).
		Msg("starting http server")
	logger.Fatal().Err(srv.ListenAndServe()).Msg(strHttpServerPanic)
}
