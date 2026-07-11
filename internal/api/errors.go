package api

import (
	"fmt"
	"net/http"
)

// APIError is the typed error returned for any non-2xx response. Its Error() appends a
// status-keyed hint so the user sees a next action ("run `wootctl auth login`"), not just
// "request failed" — the difference between an actionable CLI and an opaque one.
type APIError struct {
	StatusCode int
	Code       string // domain-specific error code if the API supplies one
	Message    string // human message parsed from the body
	Details    string // joined validation errors, when the body carries an errors array
	Body       string // raw body, for --verbose / debugging
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if msg == "" {
		msg = "request failed"
	}
	if e.Code != "" {
		msg = fmt.Sprintf("%s (code: %s)", msg, e.Code)
	}
	if e.Details != "" && e.Details != e.Message {
		msg = fmt.Sprintf("%s: %s", msg, e.Details)
	}
	if hint := hintForStatus(e.StatusCode); hint != "" {
		return fmt.Sprintf("HTTP %d: %s — %s", e.StatusCode, msg, hint)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, msg)
}

// hintForStatus maps an HTTP status to a remediation hint. Keep these specific and
// actionable; a vague hint is no better than none.
func hintForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized: // 401
		return "authentication failed; run `wootctl auth login` to store a valid api_access_token (platform paths need --platform-token)"
	case http.StatusForbidden: // 403
		return "your token lacks access here; check the agent's role (or use an administrator token)"
	case http.StatusNotFound: // 404
		return "not found; verify the id with `wootctl <resource> list` and confirm the account id (--account-id)"
	case http.StatusUnprocessableEntity: // 422
		return "the API rejected the payload; check required fields and value formats"
	case http.StatusTooManyRequests: // 429
		return "rate limited by the server; wootctl slows down automatically — retry shortly or lower --rps"
	}
	if status >= 500 {
		return "server error, usually transient; retry shortly"
	}
	if status == http.StatusBadRequest { // 400
		return "bad request; check required fields and flag values"
	}
	return ""
}

// IsRetryable reports whether an APIError represents a transient condition worth retrying.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}
