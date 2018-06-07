package main

import (
	"fmt"
	"net/http"
	"strings"
)

/*
 *	Google style API responses
 *	https://developers.google.com/drive/v3/web/handle-errors
 *	{
 *		"error": {
 *			"errors": [
 *			{
 *				"domain": "global",
 *				"reason": "badRequest",
 *				"message": "Bad Request"
 *			}
 *			],
 *			"code": 400,
 *			"message": "Bad Request"
 *		}
 *	}
 */

// jsonError generates an HTTP error but in json format.
// NOTE: This function uses simple string formatting which is faster than json encoding for small responses; no json escaping is done.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	fmt.Fprintf(w, `{"error":{"code":%d,"message":"%s"}}%c`, code, msg, '\n')
}

func smartError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if strings.HasPrefix(r.Header.Get("Accept"), "application/json") {
		jsonError(w, msg, code)
		return
	}
	http.Error(w, msg, code)
}

func responseNotModified(w http.ResponseWriter, r *http.Request, etag string) {
	//w.Header().Set("Content-Type", "application/json")
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	w.WriteHeader(http.StatusNotModified)
}
