// Package oidcclient implements a basic oidc client to authenticate users at an OpenID Connect IdP using the Code flow.
package oidc

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NatLibFi/qvain-api/randomkey"

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
)

type OidcClient struct {
	clientID string
	state    string
	logger   zerolog.Logger

	oidcProvider *gooidc.Provider
	oidcVerifier *gooidc.IDTokenVerifier
	oauthConfig  oauth2.Config
	oidcConfig   *gooidc.Config
}

func NewOidcClient(id string, secret string, redirectUrl string, providerUrl string, logger zerolog.Logger) (*OidcClient, error) {
	var err error

	ctx := context.Background()

	client := OidcClient{
		clientID: id,
		logger:   logger,
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

func (client *OidcClient) Auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		client.logger.Debug().Str("state", state).Msg("redirect to IdP")
		//log.Println("redirect:", client.oauthConfig.AuthCodeURL(client.state))
		http.Redirect(w, r, client.oauthConfig.AuthCodeURL(state), http.StatusFound)
	}
}

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

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			client.logger.Error().Err(err).Msg("failed to extract claims")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		client.logger.Info().Str("sub", idToken.Subject).Msg("login")

		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			client.logger.Error().Err(err).Msg("failed to marshal json response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.Redirect(w, r, "/token#"+rawIDToken, http.StatusFound)
		w.Write(data)
	}
}

func (client *OidcClient) DebugCallback() http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("q: %+v\n", r.URL.Query())
		if r.URL.Query().Get("state") != client.state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := client.oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}
		idToken, err := client.oidcVerifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
