package main

import (
	"net/http"

	//"github.com/NatLibFi/qvain-api/env"
	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/oidc"
)

// makeMux sets up the default handlers and returns a mux that can also be used for testing.
func makeMux(config *Config) *http.ServeMux {
	mux := http.NewServeMux()

	// static endpoints
	mux.HandleFunc("/", welcome)

	// api endpoint
	mux.HandleFunc("/api", apiHello)
	mux.HandleFunc("/api/", apiHello)

	// api endpoint, show version
	mux.HandleFunc("/api/version", apiVersion)

	// api endpoint, database check
	mux.Handle("/api/db", apiDatabaseCheck(config.db))

	// token middleware
	jwt := jwt.NewJwtHandler(config.tokenKey, config.Hostname, jwt.Verbose, jwt.RequireJwtID, jwt.WithErrorFunc(jsonError))
	mux.Handle("/auth/check", jwt.MustToken(http.HandlerFunc(protected)))

	// login callback
	lh := NewLoginHandler(config.NewLogger("auth"), true)
	mux.HandleFunc("/auth/oidc/cb", lh.Callback)

	// OIDC client
	oidcLogger := config.NewLogger("oidc")
	oidcClient, err := oidc.NewOidcClient(
		config.oidcClientID,
		config.oidcClientSecret,
		"https://"+config.Hostname+"/api/auth/cb",
		//"https://qvain-test.csc.fi/api/auth/cb",
		config.oidcProviderUrl,
		oidcLogger,
	)
	if err != nil {
		oidcLogger.Error().Err(err).Msg("oidc configuration failed")
	} else {
		mux.HandleFunc("/api/auth/login", oidcClient.Auth())
		mux.HandleFunc("/api/auth/cb", oidcClient.Callback())
	}

	// dataset endpoints
	dsRouter := NewDatasetRouter("/api/dataset/", config.db)
	mux.Handle("/api/dataset/", dsRouter)

	// views
	views := Views{db: config.db, logger: config.NewLogger("views")}
	mux.HandleFunc("/api/views/byowner", views.ByOwner())

	return mux
}
