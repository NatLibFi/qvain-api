package main

import (
	"net/http"

	"github.com/NatLibFi/qvain-api/internal/jwt"
	"github.com/NatLibFi/qvain-api/internal/oidc"
	"github.com/NatLibFi/qvain-api/internal/orcid"
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

	// OIDC client
	oidcLogger := config.NewLogger("oidc")
	oidcClient, err := oidc.NewOidcClient(
		config.oidcProviderName,
		config.oidcClientID,
		config.oidcClientSecret,
		"https://"+config.Hostname+"/api/auth/cb",
		//"https://qvain-test.csc.fi/api/auth/cb",
		config.oidcProviderUrl,
		"/token",
	)
	if err != nil {
		oidcLogger.Error().Err(err).Msg("oidc configuration failed")
	} else {
		oidcClient.SetLogger(oidcLogger)
		//oidcClient.OnLogin = MakeSessionHandlerForExternalService(config.sessions, config.db, config.Logger, "fd")
		oidcClient.OnLogin = MakeSessionHandlerForFairdata(config.sessions, config.db, config.Logger, "fd")
		mux.HandleFunc("/api/auth/login", oidcClient.Auth())
		mux.HandleFunc("/api/auth/cb", oidcClient.Callback())
	}

	// ORCID client
	orcidLogger := config.NewLogger("orcid")
	if orcidClient, err := orcid.NewOrcidClient("https://developers.google.com/oauthplayground/"); err != nil {
		orcidLogger.Error().Err(err).Msg("orcid configuration failed")
	} else {
		orcidClient.SetLogger(orcidLogger)
		mux.HandleFunc("/api/auth/orcid/login", orcidClient.Auth())
		mux.HandleFunc("/api/auth/orcid/cb", orcidClient.Callback())
	}

	// dataset endpoints
	//datasetApi := NewDatasetApi(config.db, config.sessions, config.NewLogger("dataset"))
	//mux.Handle("/api/dataset/", datasetApi)

	// object storage endpoints
	//objectApi := NewObjectApi(config.db, config.NewLogger("objects"))
	//mux.Handle("/api/objects/", objectApi)

	// views
	//viewApi := &ViewApi{db: config.db, logger: config.NewLogger("views")}
	//mux.HandleFunc("/api/views/byowner", viewApi.ByOwner())

	return mux
}
