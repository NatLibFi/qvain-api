package main

import (
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog"
)

func welcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the Qvain API server.\n"))
}

func protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to a protected url.\n"))
}

// makeLoggingHandler takes a handler and logger and then wraps the given handler with request logging middleware.
func makeLoggingHandler(prefix string, wrapped http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we eat the url, so make a copy
		url := prefix + r.URL.String()
		h := httpsnoop.CaptureMetrics(wrapped, w, r)

		logger.Log().Str("method", r.Method).Str("url", url).Int("status", h.Code).Dur("⌛", h.Duration).Str("Δt", h.Duration.String()).Int64("written", h.Written).Msg("request")
	})
}
