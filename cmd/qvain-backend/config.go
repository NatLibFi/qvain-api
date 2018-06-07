package main

import (
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/rs/zerolog"
)

type Config struct {
	Hostname      string
	Port          string
	Standalone    bool
	ForceHttpOnly bool
	Debug         bool
	UseHttpErrors bool
	Logger        zerolog.Logger
	tokenKey      []byte
	db            *psql.DB
}

func (config *Config) NewLogger(name string) zerolog.Logger {
	return config.Logger.With().Str("component", name).Logger()
}

func (config *Config) initDB(logger zerolog.Logger) (err error) {
	config.db, err = psql.NewPoolServiceFromEnv()
	if err == nil {
		config.db.SetLogger(logger)
	}
	return err
}
