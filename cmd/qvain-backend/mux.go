package main

import (
	"net/http"

	//"github.com/NatLibFi/qvain-api/env"
	"github.com/NatLibFi/qvain-api/jwt"
)

// makeMux sets up the default handlers and returns a mux that can also be used for testing.
func makeMux(config *Config) *http.ServeMux {
	mux := http.NewServeMux()

	// static endpoints
	mux.HandleFunc("/", welcome)

	// token middleware
	jwt := jwt.NewJwtHandler(config.tokenKey, config.Hostname, jwt.Verbose, jwt.RequireJwtID, jwt.WithErrorFunc(jsonError))
	mux.Handle("/auth/check", jwt.MustToken(http.HandlerFunc(protected)))

	// login callback
	lh := NewLoginHandler(config.NewLogger("auth"), true)
	mux.HandleFunc("/auth/oidc/cb", lh.Callback)

	// api endpoint
	mux.HandleFunc("/api", apiHello)
	mux.HandleFunc("/api/", apiHello)

	// api endpoint, show version
	mux.HandleFunc("/api/version", apiVersion)

	// api endpoint, database check
	mux.Handle("/api/db", apiDatabaseCheck(config.db))

	// dataset endpoints
	dsRouter := NewDatasetRouter("/api/dataset/", config.db)
	mux.Handle("/api/dataset/", dsRouter)

	return mux
}
