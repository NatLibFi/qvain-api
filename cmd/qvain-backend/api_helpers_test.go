package main

import (
	"encoding/json"
	//"io"
	"io/ioutil"
	//"net/http"
	"bytes"
	"net/http/httptest"
	"testing"
)

type JsonError struct {
	Status int              `json:"status"`
	Msg    string           `json:"msg"`
	Origin string           `json:"origin"`
	Extra  *json.RawMessage `json:"more,omitempty"`
}

func TestJsonErrors(t *testing.T) {
	var tests = []struct {
		status int
		msg    string
		origin string
		extra  []byte
	}{
		{
			status: 200,
			msg:    "OK",
			extra:  nil,
		},
		{
			status: 200,
			msg:    "that worked",
			origin: "someservice",
			extra:  []byte(`"extra string field"`),
		},
		{
			status: 400,
			msg:    "boom",
			origin: "someservice",
			extra:  []byte(`{"much":"key","so":"data"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonError(w, test.msg, test.status)

			response := w.Result()
			body, _ := ioutil.ReadAll(response.Body)

			if response.StatusCode != test.status {
				t.Errorf("statuscode (header): expected %d, got %d", test.status, response.StatusCode)
			}

			if response.Header.Get("Content-Type") != "application/json" {
				t.Errorf("content-type: expected %s, got %s", "application/json", response.Header.Get("Content-Type"))
			}

			var parsed JsonError
			if err := json.Unmarshal(body, &parsed); err != nil {
				t.Error("error response failed to unmarshal:", err)
			}

			if parsed.Status != test.status {
				t.Errorf("status (body): expected %d, got %d", test.status, parsed.Status)
			}

			if parsed.Msg != test.msg {
				t.Errorf("message: expected %s, got %s", test.msg, parsed.Msg)
			}
		})

		t.Run(test.msg+"_payload", func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonErrorWithPayload(w, test.msg, test.origin, test.extra, test.status)

			response := w.Result()
			body, _ := ioutil.ReadAll(response.Body)

			if response.StatusCode != test.status {
				t.Errorf("statuscode (header): expected %d, got %d", test.status, response.StatusCode)
			}

			if response.Header.Get("Content-Type") != "application/json" {
				t.Errorf("content-type: expected %s, got %s", "application/json", response.Header.Get("Content-Type"))
			}

			var parsed JsonError
			if err := json.Unmarshal(body, &parsed); err != nil {
				t.Error("error response failed to unmarshal:", err)
			}

			t.Log(string(body))

			if parsed.Status != test.status {
				t.Errorf("status (body): expected %d, got %d", test.status, parsed.Status)
			}

			if parsed.Msg != test.msg {
				t.Errorf("message: expected %s, got %s", test.msg, parsed.Msg)
			}

			if parsed.Origin != test.origin {
				t.Errorf("origin: expected %s, got %s", test.origin, parsed.Origin)
			}

			if parsed.Extra != nil && !bytes.Equal(*parsed.Extra, test.extra) {
				t.Errorf("extra: expected %v, got %v", test.extra, parsed.Extra)
			}
		})

	}
}
