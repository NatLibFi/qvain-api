// Package orcid implements the ORCID authentication API.
//
// See also:
//   https://github.com/ORCID/ORCID-Source/blob/master/orcid-model/src/main/resources/record_2.0/README.md#scopes
package orcid

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

// Endpoints for ORCID production API
const (
	AuthUrl  = "https://orcid.org/oauth/authorize"
	TokenUrl = "https://orcid.org/oauth/token"
)

type OrcidClient struct {
	id          string
	secret      string
	redirectUrl string
	logger      zerolog.Logger

	oauthConfig oauth2.Config
}

func NewOrcidClient(redirectUrl string) (*OrcidClient, error) {
	//ctx := context.Background()

	id := os.Getenv("APP_ORCID_CLIENT_ID")
	secret := os.Getenv("APP_ORCID_CLIENT_SECRET")

	return &OrcidClient{
		id:          id,
		secret:      secret,
		redirectUrl: redirectUrl,
		logger:      zerolog.Nop(),

		oauthConfig: oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  redirectUrl,
			Scopes:       []string{"/authenticate"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  AuthUrl,
				TokenURL: TokenUrl,
			},
		},
	}, nil
}

// SetLogger assigns a logger to the service.
// It is not save to call this after instantiating any HTTP handlers.
func (client *OrcidClient) SetLogger(logger zerolog.Logger) {
	client.logger = logger
}

// Auth sends the user to the OAuth auth endpoint.
func (client *OrcidClient) Auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client.logger.Info().Str("state", "state").Str("url", client.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)).Msg("redirect to auth")
		http.Redirect(w, r, client.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline), http.StatusFound)
	}
}

// Callback handles the OAuth callback from the token endpoint.
func (client *OrcidClient) Callback() http.HandlerFunc {
	ctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request) {
		token, err := client.oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			client.logger.Error().Err(err).Msg("token exchange failed")
			http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
			return
		}

		client.logger.Info().Str("token", token.AccessToken).Msg("token exchanged")
		fmt.Fprintf(w, "successfully authenticated with ORCID!\nToken:\n%+v\n", token)

		// TODO: do something with token credentials...
	}
}

// GetApiClient takes a context and token and returns a ready HTTP client for the oauth API.
func (client *OrcidClient) GetApiClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return client.oauthConfig.Client(ctx, token)
}
