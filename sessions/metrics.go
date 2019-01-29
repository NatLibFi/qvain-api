package sessions

import (
	"expvar"
)

var (
	activeSessionsC expvar.Int
	maxSessionsC    expvar.Int
	sessionMetrics  = expvar.NewMap("app.sessions")
)

func calculateMaxSessions() interface{} {
	if activeSessionsC.Value() > maxSessionsC.Value() {
		maxSessionsC.Set(activeSessionsC.Value())
	}
	return maxSessionsC.Value()
}

func init() {
	sessionMetrics.Set("active", &activeSessionsC)
	sessionMetrics.Set("max", expvar.Func(calculateMaxSessions))
	//expvar.Publish("sessionsActive", &sessionsC)
}
