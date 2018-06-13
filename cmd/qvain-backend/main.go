// Command qvain-backend is the backend server for the Qvain API.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	//"github.com/NatLibFi/qvain-api/models"
	"github.com/NatLibFi/qvain-api/env"
	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/version"
)

const (
	// http server setup
	HttpProxyPort = "8080"

	// timeouts
	HttpReadTimeout  = 5 * time.Second
	HttpWriteTimeout = 5 * time.Second
	HttpIdleTimeout  = 120 * time.Second

	// additional info message when Go web server returns
	strHttpServerPanic = "http server crashed"
)

var appConfig Config

/*
var (
	// forceHttpOnly is a flag to override the default http scheme.
	forceHttpOnly bool

	// appDebug is a flag that sets debugging mode.
	appDebug bool

	// appLogger is the base logger to derive others from.
	appLogger zerolog.Logger

	// logger defines the logger for this (main) file.
	logger zerolog.Logger
)
*/

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

/*
func init() {
	flag.BoolVar(&appDebug, "d", env.GetBool("APP_DEBUG"), "set debugging mode (env APP_DEBUG)")
	flag.BoolVar(&forceHttpOnly, "http", env.GetBool("APP_FORCE_HTTP_SCHEME"), "force links to http:// (env APP_FORCE_HTTP_SCHEME)")
	flag.Parse()

	appLogger = createAppLogger(appDebug)
	logger = appLogger.With().Str("component", "main").Logger()
}
*/
var (
	appDebug      = flag.Bool("d", env.GetBool("APP_DEBUG"), "set debugging mode (env APP_DEBUG)")
	forceHttpOnly = flag.Bool("http", env.GetBool("APP_FORCE_HTTP_SCHEME"), "force links to http:// (env APP_FORCE_HTTP_SCHEME)")
)

func main() {
	flag.Parse()

	/*
		// get hostname; refuse to start without one
		hostname, err := getHostname()
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't get hostname:", err)
			os.Exit(1)
		}

		// get token key; refuse to start without one
		key, err := getTokenKey()
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid token key:", err)
			os.Exit(1)
		}
		if len(key) < 1 {
			fmt.Fprintln(os.Stderr, "fatal: no token key set in environment")
			os.Exit(1)
		}
		// From here on we don't exit anymore

		config := &Config{
			Hostname:         hostname,
			Port:             env.GetDefault("APP_HTTP_PORT", HttpProxyPort),
			Standalone:       env.GetBool("APP_HTTP_STANDALONE"),
			ForceHttpOnly:    *forceHttpOnly,
			Debug:            *appDebug,
			Logger:           createAppLogger(*appDebug),
			UseHttpErrors:    useHttpErrors,
			tokenKey:         key,
			oidcClientID:     env.Get("APP_OIDC_CLIENT_ID"),
			oidcClientSecret: env.Get("APP_OIDC_CLIENT_SECRET"),
			oidcProviderUrl:  env.Get("APP_OIDC_PROVIDER_URL"),
		}
	*/

	// configure application from environment; exit if there was an error
	config, err := ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}

	logger := config.NewLogger("main")
	setStdlibLogger(config.NewLogger("log"))

	if env.Get("APP_ENV_CHECK") == "" {
		logger.Warn().Msg("environment variable APP_ENV_CHECK is not set")
	}

	err = config.initDB(config.NewLogger("psql"))
	if err != nil {
		logger.Error().Err(err).Msg("daba baad")
	}

	// set up default handlers
	mux := makeMux(config)

	// adding logging middleware
	loggingMux := makeLoggingHandler(mux, config.NewLogger("http"))

	// add auth middleware
	jwt := jwt.NewJwtHandler(config.tokenKey, config.Hostname, jwt.Verbose, jwt.RequireJwtID, jwt.WithErrorFunc(jsonError))
	//authMux := jwt.MustToken(loggingMux)
	authMux := jwt.AppendUser(loggingMux)

	// default server, without TLSConfig
	srv := &http.Server{
		Handler:           authMux,
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
				fmt.Fprintln(os.Stderr, "warning: need cap_net_bind_service capability to run stand-alone")
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
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
		Msg("starting http server")
	logger.Fatal().Err(srv.ListenAndServe()).Msg(strHttpServerPanic)
}
