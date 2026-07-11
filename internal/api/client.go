// Package api is the generic HTTP core: one client + one generic Resource[T] power every
// resource. Adding a resource never touches this package.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// AuthHeader is the one header every authenticated Chatwoot API group uses,
	// regardless of token class (user / platform app / agent bot).
	AuthHeader = "api_access_token"
	redacted   = "REDACTED"
)

// Client is the authenticated HTTP client for one Chatwoot instance. It is safe for
// sequential use by the CLI; the rate limiter serializes requests internally.
type Client struct {
	BaseURL   string // instance root, e.g. https://app.chatwoot.com (no trailing slash)
	AccountID string // account scope for /api/v1/accounts/{id} paths

	token         string // user token (application API)
	platformToken string // platform app token (platform API); empty unless configured

	httpClient *http.Client
	limiter    *rateLimiter
	retry      retryPolicy

	// DryRun, when true, prints the equivalent curl to DryRunOut and performs no request.
	DryRun    bool
	ShowToken bool // reveal the token in dry-run/curl output instead of redacting
	DryRunOut io.Writer

	// Verbose enables request/response logging to VerboseOut.
	Verbose    bool
	VerboseOut io.Writer
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying *http.Client (used by tests).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithRateLimit sets the base requests-per-second.
func WithRateLimit(rps float64) Option { return func(c *Client) { c.limiter = newRateLimiter(rps) } }

// WithDryRun toggles dry-run mode and its output sink.
func WithDryRun(on bool, out io.Writer) Option {
	return func(c *Client) { c.DryRun = on; c.DryRunOut = out }
}

// WithPlatformToken supplies the platform app token used by /platform paths.
func WithPlatformToken(t string) Option { return func(c *Client) { c.platformToken = t } }

// New builds a Client for an instance root URL. Chatwoot has no universal SaaS default
// worth hardcoding beyond the cloud host; the CLI always resolves baseURL from the profile.
func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		limiter:    newRateLimiter(5),
		retry:      defaultRetryPolicy(),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// AccountPath prefixes a resource path with the profile's application-API account scope.
func (c *Client) AccountPath(sub string) string {
	return "api/v1/accounts/" + url.PathEscape(c.AccountID) + "/" + strings.TrimLeft(sub, "/")
}

// AccountPathV2 is AccountPath for the /api/v2 (reports) group.
func (c *Client) AccountPathV2(sub string) string {
	return "api/v2/accounts/" + url.PathEscape(c.AccountID) + "/" + strings.TrimLeft(sub, "/")
}

// tokenFor selects the credential class from the path shape: /platform → platform app
// token, /public and /survey → unauthenticated by design, everything else → user token.
// The bool reports whether an auth header should be sent at all.
func (c *Client) tokenFor(path string) (string, bool) {
	p := strings.TrimLeft(path, "/")
	switch {
	case strings.HasPrefix(p, "platform/"):
		return c.platformToken, true
	case strings.HasPrefix(p, "public/"), strings.HasPrefix(p, "survey/"):
		return "", false
	default:
		return c.token, true
	}
}

// checkCreds fails fast (before any HTTP) when the path needs a credential or account
// scope the profile doesn't have — a clearer failure than the server's eventual 401/404.
func (c *Client) checkCreds(path string) error {
	p := strings.TrimLeft(path, "/")
	if strings.Contains(p, "/accounts//") {
		return fmt.Errorf("account id missing — run `wootctl auth login` (it captures the account id), `wootctl config set account_id <id>`, or pass --account-id")
	}
	tok, need := c.tokenFor(p)
	if !need || tok != "" {
		return nil
	}
	if strings.HasPrefix(p, "platform/") {
		return fmt.Errorf("platform API needs a platform app token — run `wootctl auth login --platform-token <token>` or set WOOTCTL_PLATFORM_TOKEN")
	}
	return fmt.Errorf("no API token — run `wootctl auth login` or set WOOTCTL_API_KEY")
}

// buildURL joins the base URL, path, and query params deterministically (sorted keys).
func (c *Client) buildURL(path string, query url.Values) string {
	u := c.BaseURL + "/" + strings.TrimLeft(path, "/")
	if len(query) == 0 {
		return u
	}
	return u + "?" + encodeSorted(query)
}

// encodeSorted encodes query params with sorted keys so output (and the dry-run curl) is
// stable across runs — never rely on map iteration order.
func encodeSorted(v url.Values) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		for _, val := range v[k] {
			if b.Len() > 0 {
				b.WriteByte('&')
			}
			b.WriteString(url.QueryEscape(k))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(val))
		}
	}
	return b.String()
}

// Do performs an authenticated request with retry, rate limiting, and dry-run support.
// In dry-run it prints the equivalent curl and returns (nil, nil). The caller owns closing
// resp.Body when a response is returned.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		bodyBytes = b
	}
	return c.doBytes(ctx, method, path, query, bodyBytes, "application/json")
}

func (c *Client) doBytes(ctx context.Context, method, path string, query url.Values, bodyBytes []byte, contentType string) (*http.Response, error) {
	if err := c.checkCreds(path); err != nil {
		return nil, err
	}
	fullURL := c.buildURL(path, query)

	if c.DryRun {
		c.printCurl(method, path, fullURL, bodyBytes, contentType)
		return nil, nil
	}

	var lastErr error
	var delay time.Duration
	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, delay); err != nil {
				return nil, err
			}
		}
		if err := c.limiter.wait(ctx); err != nil {
			return nil, err
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
		if err != nil {
			return nil, err
		}
		if tok, need := c.tokenFor(path); need {
			req.Header.Set(AuthHeader, tok)
		}
		if bodyBytes != nil && contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := c.httpClient.Do(req)
		c.limiter.observe(resp)

		if shouldRetry(method, resp, err) && attempt < c.retry.MaxRetries {
			// The server's Retry-After (seconds or HTTP-date) outranks our jittered backoff.
			delay = c.retry.backoff(attempt)
			if ra := retryAfterDelay(resp); ra > 0 {
				delay = ra
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
			lastErr = err
			continue
		}
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request to %s failed after retries", fullURL)
}

// doJSON performs a request and decodes a JSON response into out (if non-nil), turning any
// non-2xx into a typed APIError. Returns the response headers for metadata.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body io.Reader, out any) (http.Header, error) {
	status, header, data, err := c.Raw(ctx, method, path, query, body)
	if err != nil || status == 0 { // error or dry-run
		return header, err
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return header, fmt.Errorf("decode response: %w", err)
		}
	}
	return header, nil
}

// Raw performs a request and returns (status, headers, body) with non-2xx mapped to a
// typed APIError. It backs the `api` escape hatch and non-JSON endpoints (CSAT page).
// A zero status with nil error means dry-run.
func (c *Client) Raw(ctx context.Context, method, path string, query url.Values, body io.Reader) (int, http.Header, []byte, error) {
	resp, err := c.Do(ctx, method, path, query, body)
	if err != nil {
		return 0, nil, nil, err
	}
	if resp == nil { // dry-run
		return 0, nil, nil, nil
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}
	if c.Verbose && c.VerboseOut != nil {
		_, _ = fmt.Fprintf(c.VerboseOut, "%s %s -> %d\n", method, path, resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, resp.Header, data, parseAPIError(resp.StatusCode, data)
	}
	return resp.StatusCode, resp.Header, data, nil
}

// GetJSON is a convenience GET-and-decode used by read paths.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.doJSON(ctx, http.MethodGet, path, query, nil, out)
	return err
}

// UploadFile names one local file for a multipart field (e.g. attachments[]).
type UploadFile struct {
	Field string
	Path  string
}

// DoMultipart POSTs multipart/form-data (message attachments, contact avatars). Multipart
// bodies are never auto-retried: they are POSTs. File paths come from the user's own flags,
// not from data files, so they are intentionally not path-confined.
func (c *Client) DoMultipart(ctx context.Context, method, path string, fields url.Values, files []UploadFile, out any) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	// Deterministic field order for the dry-run curl and tests.
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range fields[k] {
			if err := w.WriteField(k, v); err != nil {
				return err
			}
		}
	}
	if c.DryRun {
		_ = w.Close()
		c.printCurlMultipart(method, path, fields, files)
		return nil
	}
	for _, f := range files {
		fh, err := os.Open(f.Path) // #nosec G304 -- the user's own --attachment/--avatar flag value
		if err != nil {
			return err
		}
		part, err := w.CreateFormFile(f.Field, filepath.Base(f.Path))
		if err == nil {
			_, err = io.Copy(part, fh)
		}
		_ = fh.Close()
		if err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	resp, err := c.doBytes(ctx, method, path, nil, buf.Bytes(), w.FormDataContentType())
	if err != nil || resp == nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, data)
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// parseAPIError extracts a message/code from the error body shapes Chatwoot actually
// returns: {"error": "..."}, {"message": "..."}, {"description": "..."},
// {"errors": ["..."]}, or {"errors": [{"message": "...", "code": ...}]}.
func parseAPIError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status, Body: string(body)}
	var obj struct {
		Message     string          `json:"message"`
		Error       string          `json:"error"`
		Description string          `json:"description"`
		Errors      json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(body, &obj) == nil {
		switch {
		case obj.Message != "":
			e.Message = obj.Message
		case obj.Error != "":
			e.Message = obj.Error
		case obj.Description != "":
			e.Message = obj.Description
		}
		if len(obj.Errors) > 0 {
			var strs []string
			if json.Unmarshal(obj.Errors, &strs) == nil && len(strs) > 0 {
				e.Details = strings.Join(strs, "; ")
			} else {
				var objs []struct {
					Message string `json:"message"`
					Code    any    `json:"code"`
				}
				if json.Unmarshal(obj.Errors, &objs) == nil && len(objs) > 0 {
					parts := make([]string, 0, len(objs))
					for _, o := range objs {
						parts = append(parts, o.Message)
					}
					e.Details = strings.Join(parts, "; ")
					if objs[0].Code != nil {
						e.Code = fmt.Sprint(objs[0].Code)
					}
				}
			}
			if e.Message == "" {
				e.Message = e.Details
			}
		}
	}
	if e.Message == "" {
		if s := strings.TrimSpace(string(body)); s != "" && len(s) < 200 {
			e.Message = s
		}
	}
	return e
}

// printCurl writes the equivalent, shell-escaped curl command, redacting the auth header
// unless ShowToken is set. Indispensable for debugging and teaching.
func (c *Client) printCurl(method, path, fullURL string, body []byte, contentType string) {
	out := c.DryRunOut
	if out == nil {
		return
	}
	var b strings.Builder
	b.WriteString("curl -X ")
	b.WriteString(method)
	fmt.Fprintf(&b, " %s", shellQuote(fullURL))
	if tok, need := c.tokenFor(path); need {
		if !c.ShowToken {
			tok = redacted
		}
		fmt.Fprintf(&b, " -H %s", shellQuote(AuthHeader+": "+tok))
	}
	if len(body) > 0 {
		fmt.Fprintf(&b, " -H %s", shellQuote("Content-Type: "+contentType))
		fmt.Fprintf(&b, " -d %s", shellQuote(string(body)))
	}
	_, _ = fmt.Fprintln(out, b.String())
}

// printCurlMultipart is printCurl for multipart bodies: -F per field and @file part.
func (c *Client) printCurlMultipart(method, path string, fields url.Values, files []UploadFile) {
	out := c.DryRunOut
	if out == nil {
		return
	}
	var b strings.Builder
	b.WriteString("curl -X ")
	b.WriteString(method)
	fmt.Fprintf(&b, " %s", shellQuote(c.buildURL(path, nil)))
	if tok, need := c.tokenFor(path); need {
		if !c.ShowToken {
			tok = redacted
		}
		fmt.Fprintf(&b, " -H %s", shellQuote(AuthHeader+": "+tok))
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range fields[k] {
			fmt.Fprintf(&b, " -F %s", shellQuote(k+"="+v))
		}
	}
	for _, f := range files {
		fmt.Fprintf(&b, " -F %s", shellQuote(f.Field+"=@"+f.Path))
	}
	_, _ = fmt.Fprintln(out, b.String())
}

// shellQuote single-quotes a string for safe pasting into a POSIX shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
