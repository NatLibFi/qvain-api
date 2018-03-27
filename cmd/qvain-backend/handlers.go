package main

import (
	"fmt"
	"net/http"

	"github.com/NatLibFi/qvain-api/version"
)

func welcome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Welcome to the Qvain API server.\n"))
}

func myVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "qvain %s at tag %s hash %s\n", version.SemVer, version.CommitTag, version.CommitHash)
}

func protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to a protected url.\n"))
}

func serveApp(w http.ResponseWriter, req *http.Request) {
}
