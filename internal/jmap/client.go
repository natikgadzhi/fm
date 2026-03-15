// Package jmap implements the JMAP protocol client for Fastmail.
package jmap

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/verbose"
)

const (
	// DefaultSessionEndpoint is the Fastmail JMAP session URL.
	DefaultSessionEndpoint = "https://api.fastmail.com/jmap/session"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default maximum number of retry attempts.
	DefaultMaxRetries = 5

	// DefaultBaseDelay is the initial delay for exponential backoff.
	DefaultBaseDelay = 500 * time.Millisecond

	// DefaultMaxDelay caps the backoff delay.
	DefaultMaxDelay = 30 * time.Second
)

// Client wraps the go-jmap Client with retry logic and session management.
type Client struct {
	// inner is the underlying go-jmap client.
	inner *gojmap.Client

	// token is the API bearer token.
	token string

	// maxRetries is the maximum number of retry attempts.
	maxRetries int

	// baseDelay is the initial delay for exponential backoff.
	baseDelay time.Duration

	// maxDelay caps the backoff delay.
	maxDelay time.Duration

	// sleepFn is the function used for sleeping (overridable for tests).
	sleepFn func(time.Duration)

	// sessionOnce ensures Discover is only called once.
	sessionOnce sync.Once

	// sessionErr stores any error from session discovery.
	sessionErr error
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.inner.HttpClient.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithBaseURL overrides the JMAP session endpoint (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.inner.SessionEndpoint = url
	}
}

// withSleepFn overrides the sleep function (used in tests).
func withSleepFn(fn func(time.Duration)) Option {
	return func(c *Client) {
		c.sleepFn = fn
	}
}

// NewClient creates a new JMAP client configured with bearer token authentication.
// The token should be resolved by internal/auth before calling this function.
func NewClient(token string, opts ...Option) *Client {
	// Create a retrying HTTP transport that wraps the default transport.
	// We'll set retryTransport's parent after creating the Client so it can
	// access retry configuration.
	c := &Client{
		inner: &gojmap.Client{
			SessionEndpoint: DefaultSessionEndpoint,
		},
		token:      token,
		maxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
		maxDelay:   DefaultMaxDelay,
		sleepFn:    time.Sleep,
	}

	// Set up the inner client with bearer token auth. This sets up an
	// oauth2 transport that adds the Authorization header.
	c.inner.WithAccessToken(token)

	// Save the authenticated transport, then wrap it with retry logic.
	authTransport := c.inner.HttpClient.Transport

	// Set a default timeout on the http client.
	c.inner.HttpClient.Timeout = DefaultTimeout

	// Apply user options first (they may change maxRetries, baseDelay, etc.)
	for _, opt := range opts {
		opt(c)
	}

	// Now wrap the transport with retry logic using the final config.
	c.inner.HttpClient.Transport = &retryTransport{
		base:    authTransport,
		client:  c,
	}

	return c
}

// Discover fetches the JMAP session from the server. It is safe to call
// multiple times; the session is only fetched once and cached for the
// client's lifetime.
func (c *Client) Discover() error {
	c.sessionOnce.Do(func() {
		verbose.Log("JMAP session discovery: %s", c.inner.SessionEndpoint)
		c.sessionErr = c.inner.Authenticate()
		if c.sessionErr != nil {
			verbose.Log("session discovery failed: %v", c.sessionErr)
			c.sessionErr = classifyError(c.sessionErr)
		} else {
			verbose.Log("session discovery succeeded")
		}
	})
	return c.sessionErr
}

// Do executes a JMAP request and returns the response.
// It automatically discovers the session if it hasn't been fetched yet.
func (c *Client) Do(req *gojmap.Request) (*gojmap.Response, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	// Log the JMAP method names being invoked.
	if verbose.Enabled() {
		for _, inv := range req.Calls {
			verbose.Log("JMAP request: %s", inv.Name)
		}
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, classifyError(err)
	}
	return resp, nil
}

// Session returns the discovered JMAP session, or nil if Discover has not
// been called yet.
func (c *Client) Session() *gojmap.Session {
	c.inner.Lock()
	defer c.inner.Unlock()
	return c.inner.Session
}

// PrimaryAccountID returns the primary account ID for the mail capability.
// Discover must be called first.
func (c *Client) PrimaryAccountID() (gojmap.ID, error) {
	session := c.Session()
	if session == nil {
		return "", fmt.Errorf("session not discovered; call Discover() first")
	}

	// Try mail capability first, then core.
	mailURI := gojmap.URI("urn:ietf:params:jmap:mail")
	if id, ok := session.PrimaryAccounts[mailURI]; ok {
		return id, nil
	}
	if id, ok := session.PrimaryAccounts[gojmap.CoreURI]; ok {
		return id, nil
	}

	// Fall back to the first account if any exist.
	for id := range session.Accounts {
		return id, nil
	}

	return "", fmt.Errorf("no accounts found in session")
}

// retryTransport wraps an http.RoundTripper with retry logic for
// transient errors (HTTP 429 and 5xx).
type retryTransport struct {
	base   http.RoundTripper
	client *Client
}

// RoundTrip implements http.RoundTripper with retry logic.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := range t.client.maxRetries + 1 {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			// Network-level errors are not retried (e.g., DNS failure, timeout).
			verbose.Log("HTTP request error (attempt %d): %v", attempt+1, err)
			return nil, err
		}

		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		// Don't retry on the last attempt.
		if attempt == t.client.maxRetries {
			verbose.Log("HTTP %d: max retries exhausted after %d attempts", resp.StatusCode, attempt+1)
			return resp, nil
		}

		// Calculate backoff delay.
		delay := t.backoffDelay(attempt, resp)
		verbose.Log("HTTP %d: retrying in %v (attempt %d/%d)", resp.StatusCode, delay, attempt+1, t.client.maxRetries+1)

		// Drain and close the body before retrying.
		resp.Body.Close()

		t.client.sleepFn(delay)
	}

	return resp, err
}

// isRetryable returns true if the HTTP status code should trigger a retry.
func isRetryable(statusCode int) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	return statusCode >= 500 && statusCode < 600
}

// backoffDelay calculates the delay for the given attempt, respecting
// Retry-After headers and applying jitter.
func (t *retryTransport) backoffDelay(attempt int, resp *http.Response) time.Duration {
	// Check Retry-After header (used with 429 responses).
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if delay := parseRetryAfter(ra); delay > 0 {
			return delay
		}
	}

	// Exponential backoff with jitter.
	backoff := float64(t.client.baseDelay) * math.Pow(2, float64(attempt))
	if backoff > float64(t.client.maxDelay) {
		backoff = float64(t.client.maxDelay)
	}

	// Add jitter: ±25% of the backoff (random value between 0.75x and 1.25x).
	jitter := backoff * (0.75 + 0.5*rand.Float64())
	return time.Duration(jitter)
}

// parseRetryAfter parses the Retry-After header value.
// It supports both delay-seconds and HTTP-date formats.
func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)

	// Try parsing as seconds first.
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date.
	if t, err := http.ParseTime(value); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay
		}
	}

	return 0
}

// classifyError examines an error and wraps it with an actionable message
// when possible. It detects authentication failures (401), network errors,
// and other common failure modes.
func classifyError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for HTTP 401 Unauthorized — the go-jmap library surfaces this
	// in the error string when session discovery fails.
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") {
		return &auth.AuthError{
			Message: "Authentication failed. Your API token may be revoked or invalid. Run 'fm auth login' to set a new token",
			Err:     err,
		}
	}

	// Check for common network errors.
	if isNetworkError(err) {
		return fmt.Errorf("failed to connect to Fastmail API. Check your internet connection: %w", err)
	}

	return err
}

// isNetworkError checks if the error looks like a network connectivity issue.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	patterns := []string{
		"no such host",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"i/o timeout",
		"dial tcp",
		"TLS handshake",
	}
	for _, p := range patterns {
		if strings.Contains(errStr, p) {
			return true
		}
	}
	return false
}
