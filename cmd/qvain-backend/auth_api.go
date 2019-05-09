package main

import (
	"net/http"

	//"github.com/CSCfi/qvain-api/internal/jwt"
	"github.com/CSCfi/qvain-api/internal/oidc"
	//"github.com/CSCfi/qvain-api/orcid"

	"github.com/francoispqt/gojay"
	"github.com/rs/zerolog"
)

// AuthApi holds authentication handlers configured for this application.
type AuthApi struct {
	oidc struct {
		client           *oidc.OidcClient
		authorizeHandler http.Handler
		callbackHandler  http.Handler
	}
	ServeHTTP http.HandlerFunc
	logger    zerolog.Logger
}

// NewAuthApi sets up external authentication services such as OpenID Connect endpoints.
//
// TODO: Too hard-coded: right now the OIDC configuration is flat; perhaps come up with better,
// more dynamic config that allows more than one provider.
func NewAuthApi(config *Config, onLogin loginHook, logger zerolog.Logger) *AuthApi {
	api := AuthApi{
		logger: logger,
	}

	// main OIDC client
	oidcLogger := config.NewLogger("oidc").With().Str("idp", config.oidcProviderName).Logger()
	oidcClient, err := oidc.NewOidcClient(
		config.oidcProviderName,
		config.oidcClientID,
		config.oidcClientSecret,
		"https://"+config.Hostname+"/api/auth/cb",
		config.oidcProviderUrl,
		"/token",
	)
	if err != nil {
		logger.Error().Err(err).Str("idp", config.oidcProviderName).Msg("oidc configuration failed")
		api.ServeHTTP = func(w http.ResponseWriter, r *http.Request) {
			jsonError(w, "no authentication endpoints configured", http.StatusNotFound)
		}
	} else {
		oidcClient.SetLogger(oidcLogger)
		oidcClient.OnLogin = MakeSessionHandlerForFairdata(config.sessions, config.db, onLogin, config.Logger, config.oidcProviderName)
		api.oidc.client = oidcClient
		api.oidc.authorizeHandler = oidcClient.Auth()
		api.oidc.callbackHandler = oidcClient.Callback()
		api.ServeHTTP = api.authHandler
	}
	return &api
}

// authHandler is the main http handler for the auth API.
func (api *AuthApi) authHandler(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug().Str("path", r.URL.Path).Msg("auth api request")
	head := ShiftUrlWithTrailing(r)
	api.logger.Debug().Str("path", r.URL.Path).Str("head", head).Msg("auth api request")

	switch head {
	case "":
		ifGet(w, r, api.listProviders)
		return
	case "login":
		api.oidc.authorizeHandler.ServeHTTP(w, r)
		return
	case "cb":
		api.oidc.callbackHandler.ServeHTTP(w, r)
		return
	}
	jsonError(w, "unknown authentication method", http.StatusNotFound)
	return
}

// listProviders lists configured providers at the auth endpoint.
func (api *AuthApi) listProviders(w http.ResponseWriter, r *http.Request) {
	apiWriteHeaders(w)

	enc := gojay.BorrowEncoder(w)
	defer enc.Release()

	enc.AppendByte('{')
	enc.StringKey("api", "auth")
	enc.ArrayKey("IdPs", gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
		for _, provider := range []string{api.oidc.client.Name} {
			enc.AddString(provider)
		}
	}))
	enc.AppendByte('}')
	enc.Write()
}
