// Package oidcclient implements a basic oidc client to authenticate users at an OpenID Connect IdP using the Code flow.
package oidc

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NatLibFi/qvain-api/internal/randomkey"

	gooidc "github.com/coreos/go-oidc"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	// DefaultLoginTimeout is the age, in seconds, of the state cookie during OIDC login.
	DefaultLoginTimeout = 600 // 10m

	// DefaultCookiePath sets the URL path cookies from this package are valid for.
	DefaultCookiePath = "/api/auth"

	// skipRedirect dumps the token to the end-user's browser instead of redirecting back to the frontend.
	skipRedirect = false
)

// OidcClient holds the OpenID Connect and OAuth2 configuration for an authentication provider.
type OidcClient struct {
	Name        string
	clientID    string
	frontendUrl string
	state       string
	logger      zerolog.Logger

	oidcProvider *gooidc.Provider
	oidcVerifier *gooidc.IDTokenVerifier
	oauthConfig  oauth2.Config
	oidcConfig   *gooidc.Config

	//OnLogin func(w http.ResponseWriter, r *http.Request, sub string, exp time.Time) error
	//OnLogin func(http.ResponseWriter, *http.Request, *oauth2.Token, *gooidc.IDToken) error
	OnLogin func(http.ResponseWriter, *http.Request, *oauth2.Token, *gooidc.IDToken) error
}

// NewOidcClient creates a new OpenID Connect client for the given provider and credentials.
func NewOidcClient(name string, id string, secret string, redirectUrl string, providerUrl string, frontendUrl string) (*OidcClient, error) {
	var err error

	ctx := context.Background()

	client := OidcClient{
		Name:        name,
		clientID:    id,
		frontendUrl: frontendUrl,
		logger:      zerolog.Nop(),
	}

	client.oidcProvider, err = gooidc.NewProvider(ctx, providerUrl)
	if err != nil {
		return nil, err
	}

	client.oidcConfig = &gooidc.Config{
		ClientID: id,
	}

	client.oidcVerifier = client.oidcProvider.Verifier(client.oidcConfig)

	client.oauthConfig = oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     client.oidcProvider.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       []string{gooidc.ScopeOpenID, "profile", "email"},
	}

	client.state = "foobar"

	return &client, nil
}

// SetLogger sets the logger for the OIDC client.
// It is probably not safe to call this after the handlers are instantiated.
func (client *OidcClient) SetLogger(logger zerolog.Logger) {
	client.logger = logger
}

// Auth is a HTTP handler that forwards the OIDC client to the Authorization endpoint.
func (client *OidcClient) Auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nonce := r.URL.RawQuery

		key, err := randomkey.Random16()
		if err != nil {
			client.logger.Error().Err(err).Msg("can't create state parameter")
			http.Error(w, "can't create state parameter", http.StatusInternalServerError)
			return
		}
		state := key.Base64()

		http.SetCookie(w, &http.Cookie{
			Name:  "state",
			Value: state,
			Path:  DefaultCookiePath,
			// old browsers such as IE<=8 don't understand MaxAge; use Expires or leave it unset to make this a "session cookie"
			Expires:  time.Now().Add(DefaultLoginTimeout * time.Second),
			MaxAge:   DefaultLoginTimeout,
			Secure:   true,
			HttpOnly: true,
		})

		client.logger.Debug().Str("state", state).Bool("withNonce", len(nonce) > 0).Msg("redirect to IdP")
		http.Redirect(w, r, client.oauthConfig.AuthCodeURL(state, gooidc.Nonce(nonce)), http.StatusFound)
	}
}

// Callback is a HTTP handler that takes the callback from the OIDC token endpoint.
func (client *OidcClient) Callback() http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("q: %+v\n", r.URL.Query())

		cookie, err := r.Cookie("state")
		if err != nil {
			client.logger.Debug().Msg("no state cookie")
			http.Error(w, "login session expired", http.StatusBadRequest)
			return
		}

		if r.URL.Query().Get("state") != cookie.Value {
			client.logger.Debug().Str("param", r.URL.Query().Get("state")).Str("cookie", cookie.Value).Msg("state did not match")
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := client.oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			client.logger.Error().Err(err).Msg("token exchange failed")
			http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			client.logger.Error().Msg("id_token missing from IdP response")
			http.Error(w, "IdP did not sent an id token", http.StatusInternalServerError)
			return
		}
		idToken, err := client.oidcVerifier.Verify(ctx, rawIDToken)
		if err != nil {
			client.logger.Error().Err(err).Msg("id token does not verify")
			http.Error(w, "id token verification failed", http.StatusInternalServerError)
			return
		}

		// client is now successfully logged in
		client.logger.Info().Str("sub", idToken.Subject).Msg("login")

		// debug response by dumping it to the browser instead of redirecting
		if skipRedirect {
			client.DumpToken(w, oauth2Token, idToken)
			return
		}

		// OnLogin callback; don't write to the response before this as it might try to set a cookie
		//if client.OnLogin != nil && client.OnLogin(w, r, idToken.Subject, oauth2Token.Expiry) != nil {
		if client.OnLogin != nil {
			if err := client.OnLogin(w, r, oauth2Token, idToken); err != nil {
				client.logger.Error().Err(err).Str("sub", idToken.Subject).Msg("OnLogin callback failed")
				http.Error(w, "Login failed", http.StatusInternalServerError)
				return
			}
		}

		// done; redirect to frontend with token in fragment
		http.Redirect(w, r, client.frontendUrl+"#"+rawIDToken, http.StatusFound)
	}
}

func (client *OidcClient) DumpToken(w http.ResponseWriter, token *oauth2.Token, idToken *gooidc.IDToken) {
	// censor access token
	if token.AccessToken != "" {
		token.AccessToken = "***"
	}

	// censor refresh token
	if token.RefreshToken != "" {
		token.RefreshToken = "***"
	}

	out := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage
	}{token, new(json.RawMessage)}

	if err := idToken.Claims(&out.IDTokenClaims); err != nil {
		client.logger.Error().Err(err).Msg("failed to extract claims")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		client.logger.Error().Err(err).Msg("failed to marshal json response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return
}
