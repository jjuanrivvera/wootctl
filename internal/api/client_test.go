package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TokenPerPathClass(t *testing.T) {
	var gotHeader []string
	var gotHas []bool
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		v, ok := r.Header[http.CanonicalHeaderKey(AuthHeader)]
		gotHas = append(gotHas, ok)
		gotHeader = append(gotHeader, strings.Join(v, ","))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	c.platformToken = "platform-token"

	require.NoError(t, c.GetJSON(t.Context(), "api/v1/profile", nil, nil))
	require.NoError(t, c.GetJSON(t.Context(), "platform/api/v1/users/1", nil, nil))
	require.NoError(t, c.GetJSON(t.Context(), "public/api/v1/inboxes/abc", nil, nil))
	require.NoError(t, c.GetJSON(t.Context(), "survey/responses/uuid-1", nil, nil))

	require.Len(t, gotHeader, 4)
	assert.Equal(t, "test-token", gotHeader[0], "application path uses the user token")
	assert.Equal(t, "platform-token", gotHeader[1], "platform path uses the platform token")
	assert.False(t, gotHas[2], "public path sends NO auth header")
	assert.False(t, gotHas[3], "survey path sends NO auth header")
}

func TestClient_FailsFastWithoutCredential(t *testing.T) {
	// No server: the request must fail before any HTTP happens.
	c := New("http://127.0.0.1:0", "")
	err := c.GetJSON(t.Context(), "api/v1/profile", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth login")

	c2 := New("http://127.0.0.1:0", "user-token") // user token present, platform absent
	err = c2.GetJSON(t.Context(), "platform/api/v1/users", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "platform")

	c2.AccountID = ""
	err = c2.GetJSON(t.Context(), c2.AccountPath("labels"), nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account id")
}

func TestClient_AccountPaths(t *testing.T) {
	c := New("https://x.example", "t")
	c.AccountID = "42"
	assert.Equal(t, "api/v1/accounts/42/labels", c.AccountPath("labels"))
	assert.Equal(t, "api/v2/accounts/42/reports", c.AccountPathV2("/reports"))
}

func TestClient_DryRunPrintsCurlAndSkipsRequest(t *testing.T) {
	hits := 0
	var buf bytes.Buffer
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { hits++ })
	c.DryRun = true
	c.DryRunOut = &buf

	err := c.Send(t.Context(), http.MethodPost, c.AccountPath("labels"), url.Values{"page": {"1"}}, map[string]string{"title": "vip"}, nil)
	require.NoError(t, err)
	assert.Zero(t, hits, "dry-run must not hit the server")

	out := buf.String()
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "/api/v1/accounts/1/labels?page=1")
	assert.Contains(t, out, AuthHeader+": REDACTED")
	assert.NotContains(t, out, "test-token", "token must be redacted by default")
	assert.Contains(t, out, `{"title":"vip"}`)
}

func TestClient_DryRunShowToken(t *testing.T) {
	var buf bytes.Buffer
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	c.DryRun = true
	c.ShowToken = true
	c.DryRunOut = &buf
	require.NoError(t, c.GetJSON(t.Context(), "api/v1/profile", nil, nil))
	assert.Contains(t, buf.String(), "api_access_token: test-token")
}

func TestClient_DryRunPublicPathHasNoAuthHeader(t *testing.T) {
	var buf bytes.Buffer
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	c.DryRun = true
	c.DryRunOut = &buf
	require.NoError(t, c.GetJSON(t.Context(), "public/api/v1/inboxes/abc", nil, nil))
	assert.NotContains(t, buf.String(), AuthHeader)
}

func TestClient_RetriesIdempotentOn5xx(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	require.NoError(t, c.GetJSON(t.Context(), "api/v1/profile", nil, nil))
	assert.Equal(t, int32(3), calls.Load())
}

func TestClient_NeverRetriesPOST(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	})
	err := c.Send(t.Context(), http.MethodPost, "api/v1/profile", nil, map[string]string{}, nil)
	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load(), "POST must not be auto-retried")
}

func TestClient_HonorsRetryAfterSeconds(t *testing.T) {
	var calls atomic.Int32
	var firstRetryGap time.Duration
	var last time.Time
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		if calls.Add(1) == 1 {
			last = now
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		firstRetryGap = now.Sub(last)
		_, _ = w.Write([]byte(`{}`))
	})
	start := time.Now()
	require.NoError(t, c.GetJSON(t.Context(), "api/v1/profile", nil, nil))
	require.GreaterOrEqual(t, time.Since(start), 900*time.Millisecond, "Retry-After must outrank the tiny test backoff")
	assert.GreaterOrEqual(t, firstRetryGap, 900*time.Millisecond)
}

func TestClient_RetryAfterHTTPDate(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", time.Now().Add(2*time.Second).UTC().Format(http.TimeFormat))
	d := retryAfterDelay(resp)
	assert.Greater(t, d, time.Second)
	assert.LessOrEqual(t, d, 2*time.Second+100*time.Millisecond)

	resp.Header.Set("Retry-After", "not-a-date")
	assert.Zero(t, retryAfterDelay(resp))
	resp.Header.Set("Retry-After", "-3")
	assert.Zero(t, retryAfterDelay(resp))
	assert.Zero(t, retryAfterDelay(nil))
}

func TestClient_CtrlCCancelsBackoff(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := c.GetJSON(ctx, "api/v1/profile", nil, nil)
	require.Error(t, err)
	assert.Less(t, time.Since(start), 5*time.Second, "cancellation must cut the 30s Retry-After sleep short")
}

func TestClient_APIErrorShapes(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		wantMsg  string
		wantCode string
	}{
		{"error key", 401, `{"error":"Invalid token"}`, "Invalid token", ""},
		{"message key", 404, `{"message":"Resource could not be found"}`, "Resource could not be found", ""},
		{"description key", 403, `{"description":"You are not authorized to do this action"}`, "You are not authorized to do this action", ""},
		{"errors string array", 422, `{"errors":["Name can't be blank","Email invalid"]}`, "Name can't be blank; Email invalid", ""},
		{"errors object array", 422, `{"errors":[{"message":"Invalid","code":"invalid_record"}]}`, "Invalid", "invalid_record"},
		{"plain text", 500, `boom`, "boom", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := parseAPIError(tc.status, []byte(tc.body))
			assert.Equal(t, tc.status, e.StatusCode)
			assert.Equal(t, tc.wantMsg, e.Message)
			assert.Equal(t, tc.wantCode, e.Code)
		})
	}
}

func TestClient_RawSurfacesTypedError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Resource could not be found"}`))
	})
	_, _, _, err := c.Raw(t.Context(), http.MethodGet, "api/v1/profile", nil, nil)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Contains(t, apiErr.Error(), "wootctl")
}

func TestClient_VerboseLogsRoundTrip(t *testing.T) {
	var vbuf bytes.Buffer
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) })
	c.Verbose = true
	c.VerboseOut = &vbuf
	require.NoError(t, c.GetJSON(t.Context(), "api/v1/profile", nil, nil))
	assert.Contains(t, vbuf.String(), "GET api/v1/profile -> 200")
}

func TestClient_QueryEncodingSortedAndEscaped(t *testing.T) {
	c := New("https://x.example", "t")
	u := c.buildURL("api/v1/x", url.Values{"b": {"2"}, "a": {"1 3"}})
	assert.Equal(t, "https://x.example/api/v1/x?a=1+3&b=2", u)
}

func TestClient_DoMultipart(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "note.txt")
	require.NoError(t, os.WriteFile(fpath, []byte("hello"), 0o600))

	var gotContentType, gotContent, gotFile string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		require.NoError(t, r.ParseMultipartForm(1<<20))
		gotContent = r.FormValue("content")
		f, hdr, err := r.FormFile("attachments[]")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		gotFile = hdr.Filename
		_, _ = w.Write([]byte(`{"id":7}`))
	})
	var out struct {
		ID ID `json:"id"`
	}
	err := c.DoMultipart(t.Context(), http.MethodPost, c.AccountPath("conversations/5/messages"),
		url.Values{"content": {"see attachment"}}, []UploadFile{{Field: "attachments[]", Path: fpath}}, &out)
	require.NoError(t, err)
	assert.Contains(t, gotContentType, "multipart/form-data")
	assert.Equal(t, "see attachment", gotContent)
	assert.Equal(t, "note.txt", gotFile)
	assert.Equal(t, ID("7"), out.ID)
}

func TestClient_DoMultipartDryRun(t *testing.T) {
	var buf bytes.Buffer
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { t.Fatal("must not be called") })
	c.DryRun = true
	c.DryRunOut = &buf
	err := c.DoMultipart(t.Context(), http.MethodPost, c.AccountPath("conversations/5/messages"),
		url.Values{"content": {"hi"}}, []UploadFile{{Field: "attachments[]", Path: "/tmp/x.png"}}, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "-F 'content=hi'")
	assert.Contains(t, out, "-F 'attachments[]=@/tmp/x.png'")
	assert.Contains(t, out, "REDACTED")
}

func TestClient_DoMultipartErrorStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Content missing"}`))
	})
	err := c.DoMultipart(t.Context(), http.MethodPost, c.AccountPath("conversations/5/messages"), url.Values{}, nil, nil)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 422, apiErr.StatusCode)
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'a b'`, shellQuote("a b"))
	assert.Equal(t, `'it'\''s'`, shellQuote("it's"))
}

func TestClient_SendDecodesEnvelopeAndRaw(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"payload":{"id":3,"name":"n"}}`)
	})
	var rec testRec
	require.NoError(t, c.Send(t.Context(), http.MethodGet, "api/v1/x", nil, nil, &rec))
	assert.Equal(t, ID("3"), rec.ID)

	var raw json.RawMessage
	require.NoError(t, c.Send(t.Context(), http.MethodGet, "api/v1/x", nil, nil, &raw))
	assert.JSONEq(t, `{"payload":{"id":3,"name":"n"}}`, string(raw))
}
