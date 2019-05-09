package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/internal/secmsg"
	"github.com/CSCfi/qvain-api/internal/sessions"
	"github.com/CSCfi/qvain-api/pkg/env"
	"github.com/CSCfi/qvain-api/pkg/models"
)

// Config holds the configuration for the application.
// It's probably not safe to change settings during operation as they might have already have been injected into components.
type Config struct {
	// application settings
	Hostname      string
	Port          string
	Standalone    bool
	ForceHttpOnly bool
	Debug         bool
	DevMode       bool
	Logging       bool
	LogRequests   bool
	UseHttpErrors bool
	Logger        zerolog.Logger

	// Metax service related settings
	MetaxApiHost string
	metaxApiUser string
	metaxApiPass string

	// session settings
	tokenKey         []byte
	oidcProviderName string
	oidcProviderUrl  string
	oidcClientID     string
	oidcClientSecret string

	// configured service instances
	db        *psql.DB
	sessions  *sessions.Manager
	messenger *secmsg.MessageService
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

	if *appDevMode {
		*appDebug = true
		*forceHttpOnly = true

		// slight hack: if in dev mode, set default API headers to include CORS allow all
		enableCORS()

		// create fake session
		if env.Get("APP_DEV_USER") != "" {
			b, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(env.Get("APP_DEV_USER"))
			if err != nil {
				return nil, fmt.Errorf("can't decode APP_DEV_USER: %s", err)
			}
			var user models.User
			user.UnmarshalJSON(b)
		}
	}

	return &Config{
		Hostname:         hostname,
		Port:             *appHttpPort,
		Standalone:       env.GetBool("APP_HTTP_STANDALONE"),
		ForceHttpOnly:    *forceHttpOnly,
		Debug:            *appDebug,
		DevMode:          *appDevMode,
		Logging:          !*disableLogging,
		LogRequests:      !*disableHttpLog,
		Logger:           createAppLogger(ServiceName, *appDebug, *disableLogging),
		UseHttpErrors:    env.GetBool("APP_HTTP_ERRORS"),
		tokenKey:         key,
		oidcProviderName: env.Get("APP_OIDC_PROVIDER_NAME"),
		oidcProviderUrl:  env.Get("APP_OIDC_PROVIDER_URL"),
		oidcClientID:     env.Get("APP_OIDC_CLIENT_ID"),
		oidcClientSecret: env.Get("APP_OIDC_CLIENT_SECRET"),
		MetaxApiHost:     env.Get("APP_METAX_API_HOST"),
		metaxApiUser:     env.Get("APP_METAX_API_USER"),
		metaxApiPass:     env.Get("APP_METAX_API_PASS"),
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

// initSessions initialises the session manager.
func (config *Config) initSessions() error {
	config.sessions = sessions.NewManager(sessions.WithRequireCSCUserName(!config.DevMode))
	return nil
}

// initMessenger initialises the secure message service.
func (config *Config) initMessenger() (err error) {
	config.messenger, err = secmsg.NewMessageService(config.tokenKey)
	return
}

// NewMetaxService initialises a metax service.
// TODO: worth putting it here?
//func (config *Config) NewMetaxService() *metax.MetaxService {}

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
	encoded := env.Get("APP_TOKEN_KEY")
	if encoded == "" {
		return nil, fmt.Errorf("no token key set in environment")
	}

	key, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(key) < 32 {
		return nil, fmt.Errorf("token key too short, expected at least 32 bytes")
	}
	return key, nil
}
