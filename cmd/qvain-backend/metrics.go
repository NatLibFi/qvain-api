package main

import (
	"expvar"
	"os"
	"runtime"
	"time"
)

var (
	// api counters
	datasetsC expvar.Int
	objectsC  expvar.Int
	sessionsC expvar.Int
	authC     expvar.Int
	proxyC    expvar.Int
	lookupC   expvar.Int
	versionC  expvar.Int

	// map containers
	metricsState = expvar.NewMap("app.state")
	metricsApis  = expvar.NewMap("app.apis")

	// startup time
	startupTime = time.Now()
	startupVar  expvar.String
)

// calculateUptime gives the time since process start-up in seconds.
func calculateUptime() interface{} {
	return time.Since(startupTime) / time.Second
	//return time.Since(startup).String()
}

// getOpenFHs returns the number of open file handles for the current process or -1 in case of error.
func getOpenFHs() interface{} {
	dir, err := os.Open("/proc/self/fd")
	if err != nil {
		return -1
	}
	defer dir.Close()

	fhs, err := dir.Readdirnames(-1)
	if err != nil {
		return -1
	}

	return len(fhs)
}

// getNumGoroutine returns the number of existing Go routines.
func getNumGoroutine() interface{} {
	return runtime.NumGoroutine()
}

// getNumCgoCall returns the number of C go calls.
func getNumCgoCall() interface{} {
	return runtime.NumCgoCall()
}

func init() {
	metricsApis.Set("datasets", &datasetsC)
	metricsApis.Set("objects", &objectsC)
	metricsApis.Set("sessions", &sessionsC)
	metricsApis.Set("auth", &authC)
	metricsApis.Set("proxy", &proxyC)
	metricsApis.Set("lookup", &lookupC)
	metricsApis.Set("version", &versionC)

	startupVar.Set(startupTime.UTC().Format(time.RFC3339))
	metricsState.Set("startup", &startupVar)
	metricsState.Set("uptime", expvar.Func(calculateUptime))
	metricsState.Set("fds", expvar.Func(getOpenFHs))
	metricsState.Set("goroutines", expvar.Func(getNumGoroutine))
	metricsState.Add("gomaxprocs", int64(runtime.GOMAXPROCS(0)))
	metricsState.Add("cpus", int64(runtime.NumCPU()))
	metricsState.Set("cgocalls", expvar.Func(getNumCgoCall))
}
