// Package env provides some convenience functions to get configuration values from the environment.
package env

import (
	"os"
)

// Get returns an environment variable. It just calls os.Getenv.
func Get(envvar string) string {
	return os.Getenv(envvar)
}

// GetDefault returns an environment variable, returning a default string if not set.
func GetDefault(envvar, def string) string {
	v, e := os.LookupEnv(envvar)
	if !e {
		return def
	}
	return v
}

// GetBool tries to parse an environment variable into a boolean.
func GetBool(envvar string) bool {
	v := os.Getenv(envvar)
	return isTrue(v)
}

// GetBoolDefault tries to parse an enviroment variable into a boolean, returning a default boolean if not set.
func GetBoolDefault(envvar string, def bool) bool {
	v, e := os.LookupEnv(envvar)
	if !e {
		return def
	}
	return isTrue(v)
}

// isTrue parses a string into a boolean.
func isTrue(s string) bool {
	return s != "" && s != "0" && s != "false" && s != "FALSE" && s != "False" && s != "no" && s != "NO" && s != "No"
}

// isDevelopment parses a string into a boolean indicating a development environment.
func isDevelopment(s string) bool {
	return s == "dev" || s == "DEV" || s == "development" || s == "DEVELOPMENT"
}
