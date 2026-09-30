package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The demo server exists so a recording can show the CLI working without an
// account. It is small, but it is the thing the recording depends on: if it stops
// answering, the demo silently becomes a screen of errors.
func TestServeDemoAnswersTheRecordedEndpoints(t *testing.T) {
	for _, path := range []string{
		"/api/v1/accounts/999/conversations",
		"/api/v1/accounts/999/conversations/9001/messages",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			serveDemo(rec, httptest.NewRequest(http.MethodGet, path, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "json") {
				t.Errorf("content type %q", ct)
			}
			var body any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("not JSON: %v\n%s", err, rec.Body.String())
			}
		})
	}
}

// An unknown path 404s rather than answering with empty data, so a mistyped
// command while recording fails visibly instead of looking like "no results".
func TestServeDemoUnknownPath(t *testing.T) {
	rec := httptest.NewRecorder()
	serveDemo(rec, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/999/nothing", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
}

// The write commands echo what they were sent, which is what makes the recording
// show a real round trip rather than a canned line.
func TestServeDemoEchoesWhatItIsSent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/accounts/999/conversations/9001/messages",
		strings.NewReader(`{"content":"an invented reply"}`))
	serveDemo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["content"] != "an invented reply" {
		t.Errorf("content = %v, want what was sent", body["content"])
	}

	// A malformed body is rejected rather than silently accepted.
	bad := httptest.NewRecorder()
	serveDemo(bad, httptest.NewRequest(http.MethodPost,
		"/api/v1/accounts/999/conversations/9001/messages", strings.NewReader("not json")))
	if bad.Code != http.StatusBadRequest {
		t.Errorf("malformed body got status %d, want 400", bad.Code)
	}
}
