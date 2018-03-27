// Command qvain-backend is the backend server for the Qvain API.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	//"github.com/NatLibFi/qvain-api/models"
	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/version"
)

const (
	//DO_HTTPS_REDIRECT = true
	RUN_STAND_ALONE   = false
	HTTP_PROXIED_PORT = ":8080"

	HTTP_READ_TIMEOUT  = 5 * time.Second
	HTTP_WRITE_TIMEOUT = 5 * time.Second
	HTTP_IDLE_TIMEOUT  = 120 * time.Second
)

// startHttpsRedirector spawns a background HTTP server that redirects to https://.
// NOTE: This function returns immediately.
func startHttpsRedirector() {
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + r.Host + r.URL.String()
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}),
		ReadTimeout:  HTTP_READ_TIMEOUT,
		WriteTimeout: HTTP_WRITE_TIMEOUT,
	}
	srv.SetKeepAlivesEnabled(false)
	log.Println("starting https redirect service")
	go func() { log.Fatal(srv.ListenAndServe()) }()
}

// makeMux() sets up the default handlers and returns a mux that can also be used for testing.
func makeMux() *http.ServeMux {
	mux := http.NewServeMux()

	// static endpoints
	mux.HandleFunc("/", welcome)
	mux.HandleFunc("/echo", echo)
	mux.Handle("/Qvain/", http.FileServer(http.Dir("/home/wouter/Code/Javascript/")))
	//mux.HandleFunc("/api/dataset/", apiDataset)
	//mux.Handle("/api/dataset/meta", needsDataset(http.HandlerFunc(apiMetadata)))

	// token middleware
	jwt := jwt.NewJwtHandler([]byte("secret"), "service.example.com", jwt.Verbose, jwt.RequireJwtID, jwt.WithErrorFunc(jsonError))
	mux.Handle("/protected", jwt.MustToken(http.HandlerFunc(protected)))

	// api endpoint, show version
	mux.HandleFunc("/api", apiVersion)

	// dataset endpoints
	dsRouter := NewDatasetRouter("/api/dataset/")
	mux.Handle("/api/dataset/", dsRouter)

	return mux
}

func main() {
	fmt.Println("qvain backend // hash:", version.CommitHash, version.CommitTag)

	// set up default handlers
	mux := makeMux()

	// default server
	srv := &http.Server{
		Handler: mux,
		//TLSConfig:         tlsConfig,
		ReadTimeout: HTTP_READ_TIMEOUT,
		//ReadHeaderTimeout: HTTP_READ_TIMEOUT,
		WriteTimeout: HTTP_WRITE_TIMEOUT,
		IdleTimeout:  HTTP_IDLE_TIMEOUT,
	}

	// if standalone, run on 443 and start redirecting port 80; else run on 8080 or whatever is configured above
	if RUN_STAND_ALONE {
		if can, err := canNetBindService(); err == nil {
			if !can {
				fmt.Println("warning: need cap_net_bind_service capability to run stand-alone")
			}
		} else {
			fmt.Println(err)
		}

		startHttpsRedirector()
		srv.TLSConfig = tlsIntermediateConfig
		log.Println("starting stand-alone server on default http/https ports")
		log.Fatal(srv.ListenAndServe()) // TODO: certificate
	} else {
		srv.Addr = HTTP_PROXIED_PORT
		log.Println("starting proxied server on port", HTTP_PROXIED_PORT)
		log.Fatal(srv.ListenAndServe())
	}
}
