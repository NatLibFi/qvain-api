package main

import (
	"encoding/hex"
	"os"

	"github.com/NatLibFi/qvain-api/env"
)

// getHostname gets the HTTP hostname from the environment or os, and returns an error on failure.
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

func getScheme() string {
	if env.GetBool("APP_FORCE_HTTP_SCHEME") {
		return "http://"
	}

	return "https:"
}

func getTokenKey() ([]byte, error) {
	key, err := hex.DecodeString(env.Get("APP_TOKEN_KEY"))
	if err != nil {
		return nil, err
	}
	return key, nil
}
