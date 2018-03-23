package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	
)



func TestCommonHandlers(t *testing.T) {
	tests := []struct {
		name string
		url string
		status int
	}{
		{ name: "root url", url: "/", status: http.StatusOK },
		{ name: "non-existent url", url: "/non-existent", status: http.StatusNotFound },
		{ name: "protected url without token", url: "/protected", status: http.StatusUnauthorized },
		{ name: "api endpoint", url: "/api", status: http.StatusOK },
		{ name: "api endpoint with slash", url: "/api/", status: http.StatusNotFound },
		{ name: "datasets endpoint without token", url: "/api/dataset", status: http.StatusMovedPermanently },
	}
	
	mux := makeMux()
	
	for _, exp := range tests {
		t.Run(exp.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", exp.url, nil)
			if err != nil {
				t.Fatal("NewRequest():", err)
			}
			
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			
			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != exp.status {
				t.Errorf("expected status: %d, got %d", exp.status, res.StatusCode)
			}
		})
	}
}
