package main

import (
	"io"
	stdlog "log"
	"net/http"
	"os"

	"github.com/CSCfi/qvain-api/internal/caller"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type locationHook struct {
	name          string
	stackInfoFunc func() string
}

func newLocationHook(name string) *locationHook {
	return &locationHook{name: name, stackInfoFunc: caller.CreateStackInfoFunc(skipFrameCount, true)}
}

func (h locationHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	e.Str(h.name, h.stackInfoFunc())
}

// createAppLogger returns a zerolog logger configured according to the process environment:
// - if APP_DEBUG is set, debugging output is enabled;
// - if the output is to a terminal, use a coloured console writer.
func createAppLogger(service string, debugging bool, disabled bool) (logger zerolog.Logger) {
	var out io.Writer = os.Stdout

	zerolog.MessageFieldName = "msg"
	//zerolog.TimeFieldFormat = ""

	if disabled {
		return zerolog.Nop()
	}

	// use colour output if logging to console
	if isatty.IsTerminal(os.Stdout.Fd()) {
		zerolog.TimeFieldFormat = "15:04:05.000000"
		out = zerolog.ConsoleWriter{Out: out, TimeFormat: "15:04:05.000000"}
	}

	if debugging {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = zerolog.New(out).Hook(newLocationHook("at")).With().Timestamp().Str("service", service).Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = zerolog.New(out).With().Timestamp().Str("service", service).Logger()
	}

	log.Logger = logger

	return logger
}

// adaptToStdlibLogger takes a zerolog logger and sets it as writer for a stdlib logger.
// The returned log.Logger can be passed to libraries that require a stdlib logger.
func adaptToStdlibLogger(logger zerolog.Logger) *stdlog.Logger {
	return stdlog.New(logger, "", 0)
}

// setStdlibLogger sets zerolog as output for anything that may use the stdlib logger.
func setStdlibLogger(logger zerolog.Logger) {
	stdlog.SetFlags(0)
	stdlog.SetOutput(logger)
}

// addUserToRequest adds the user to the logger stored in a request's context.
func addUserToRequest(r *http.Request, user string) {
	zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("user", user)
	})
}

// addApiToRequest adds the called API to the logger stored in a request's context.
func addApiToRequest(r *http.Request, api string) {
	zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("api", api)
	})
}

// newLoggingHandler is a middleware wrapper that adds a logger to the request context.
// See also the zerolog/hlog package.
func newLoggingHandler(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a copy of the logger (including internal context slice)
			// to prevent data race when using UpdateContext.
			l := logger.With().Logger()
			r = r.WithContext(l.WithContext(r.Context()))
			next.ServeHTTP(w, r)
			l.Info().Msg("request")
		})
	}
}
