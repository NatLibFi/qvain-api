package main

import (
	"net/http"
)


func welcome(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.Write([]byte("Welcome to the Qvain API server.\n"))
}


func protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to a protected url.\n"))
}


func serveApp(w http.ResponseWriter, req *http.Request) {
}

