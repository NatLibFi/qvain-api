package main

import (
	"io"
	stdlog "log"
	"os"

	"github.com/NatLibFi/qvain-api/caller"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type locationHook struct {
	name          string
	stackInfoFunc func() string
}

func newLocationHook(name string) *locationHook {
	return &locationHook{name: name, stackInfoFunc: caller.CreateStackInfoFunc(4, true)}
}

func (h locationHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	e.Str(h.name, h.stackInfoFunc())
}

// createAppLogger returns a zerolog logger configured according to the process environment:
// - if APP_DEBUG is set, debugging output is enabled;
// - if the output is to a terminal, use a coloured console writer.
func createAppLogger(isDebugging bool) (logger zerolog.Logger) {
	var out io.Writer = os.Stdout

	zerolog.MessageFieldName = "msg"
	//zerolog.TimeFieldFormat = ""

	// use colour output if logging to console
	if isatty.IsTerminal(os.Stdout.Fd()) {
		zerolog.TimeFieldFormat = "15:04:05.000000"
		out = zerolog.ConsoleWriter{Out: out}
	}

	if isDebugging {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = zerolog.New(out).Hook(newLocationHook("at")).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = zerolog.New(out).With().Timestamp().Logger()
	}
	log.Logger = logger
	return logger
}

// adaptToStdlibLogger takes a zerolog logger and sets it as writer for a stdlib logger.
// The returned log.Logger can be passed to libraries that require a stdlib logger.
func adaptToStdlibLogger(logger zerolog.Logger) *stdlog.Logger {
	return stdlog.New(logger, "", 0)
}
