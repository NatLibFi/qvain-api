// Package oidcclient implements a basic oidc client to authenticate users at an OpenID Connect IdP using the Code flow.
package oidc

import (
	"encoding/json"
	"log"
	"net/http"

	gooidc "github.com/coreos/go-oidc"

	"github.com/rs/zerolog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
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
		log.Println("redirect:", client.oauthConfig.AuthCodeURL(client.state))
		http.Redirect(w, r, client.oauthConfig.AuthCodeURL(client.state), http.StatusFound)
	}
}

func (client *OidcClient) Callback() http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("q: %+v\n", r.URL.Query())
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
