package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/NatLibFi/qvain-api/env"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/rs/zerolog"
)

// Config holds the configuration for the application.
// It's probably not safe to change settings during operation as they might have already have been injected into components.
type Config struct {
	Hostname      string
	Port          string
	Standalone    bool
	ForceHttpOnly bool
	Debug         bool
	UseHttpErrors bool
	Logger        zerolog.Logger

	tokenKey         []byte
	oidcClientID     string
	oidcClientSecret string
	oidcProviderUrl  string

	db *psql.DB
}

// ConfigFromEnv() creates the application configuration by reading in environment variables.
// If this function returns an error, it would be wise to exit the program with a non-zero exit code.
func ConfigFromEnv() (*Config, error) {
	// get hostname; refuse to start without one
	hostname, err := getHostname()
	if err != nil {
		return nil, fmt.Errorf("can't get hostname: %s", err)
	}

	// get token key; refuse to start without one
	key, err := getTokenKey()
	if err != nil {
		return nil, fmt.Errorf("invalid token key: %s", err)
	}
	if len(key) < 1 {
		return nil, fmt.Errorf("no token key set in environment")
	}

	return &Config{
		Hostname:         hostname,
		Port:             env.GetDefault("APP_HTTP_PORT", HttpProxyPort),
		Standalone:       env.GetBool("APP_HTTP_STANDALONE"),
		ForceHttpOnly:    *forceHttpOnly,
		Debug:            *appDebug,
		Logger:           createAppLogger(*appDebug),
		UseHttpErrors:    env.GetBool("APP_HTTP_ERRORS"),
		tokenKey:         key,
		oidcClientID:     env.Get("APP_OIDC_CLIENT_ID"),
		oidcClientSecret: env.Get("APP_OIDC_CLIENT_SECRET"),
		oidcProviderUrl:  env.Get("APP_OIDC_PROVIDER_URL"),
	}, nil
}

// NewLogger creates a new logger based on the main logger with the given component name.
func (config *Config) NewLogger(name string) zerolog.Logger {
	return config.Logger.With().Str("component", name).Logger()
}

// initDB initialises a new database pool to be used across the application.
func (config *Config) initDB(logger zerolog.Logger) (err error) {
	config.db, err = psql.NewPoolServiceFromEnv()
	if err == nil {
		config.db.SetLogger(logger)
	}
	return err
}

// getHostname gets the HTTP hostname from the environment or os, and returns an error on failure.
// The hostname is used as vhost in http and in token audience checks, so it is important to get this right.
func getHostname() (string, error) {
	if h := env.Get("APP_HOSTNAME"); h != "" {
		return h, nil
	}

	h, err := os.Hostname()
	if err == nil {
		return h, nil
	}

	return "", err
}

// getScheme sets the URL scheme used for links and redirects.
// It is used to enforce non-SSL links for local development where we don't have certificates.
func getScheme() string {
	if env.GetBool("APP_FORCE_HTTP_SCHEME") {
		return "http://"
	}

	return "https://"
}

// getTokenKey gets the token secret in hex from the environment and decodes it.
func getTokenKey() ([]byte, error) {
	key, err := hex.DecodeString(env.Get("APP_TOKEN_KEY"))
	if err != nil {
		return nil, err
	}
	return key, nil
}
