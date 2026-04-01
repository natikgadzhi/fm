// Package jmap implements the JMAP protocol client for Fastmail.
package jmap

import (
	"fmt"
	"strings"
	"sync"
	"time"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"github.com/natikgadzhi/cli-kit/debug"
	clierrors "github.com/natikgadzhi/cli-kit/errors"
	"github.com/natikgadzhi/cli-kit/ratelimit"
)

const (
	// DefaultSessionEndpoint is the Fastmail JMAP session URL.
	DefaultSessionEndpoint = "https://api.fastmail.com/jmap/session"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// Client wraps the go-jmap Client with retry logic and session management.
type Client struct {
	// inner is the underlying go-jmap client.
	inner *gojmap.Client

	// token is the API bearer token.
	token string

	// retryTransport is the cli-kit retry transport used for HTTP retries.
	retryTransport *ratelimit.RetryTransport

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
		c.retryTransport.MaxRetries = n
	}
}

// WithBaseURL overrides the JMAP session endpoint (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.inner.SessionEndpoint = url
	}
}

// WithOnRetry sets a callback that is invoked before each retry sleep.
func WithOnRetry(fn func(attempt int, delay time.Duration, statusCode int)) Option {
	return func(c *Client) {
		c.retryTransport.OnRetry = fn
	}
}

// NewClient creates a new JMAP client configured with bearer token authentication.
// The token should be resolved before calling this function (e.g. via cli-kit/auth).
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		inner: &gojmap.Client{
			SessionEndpoint: DefaultSessionEndpoint,
		},
		token: token,
	}

	// Set up the inner client with bearer token auth. This sets up an
	// oauth2 transport that adds the Authorization header.
	c.inner.WithAccessToken(token)

	// Save the authenticated transport, then wrap it with retry logic.
	authTransport := c.inner.HttpClient.Transport

	// Set a default timeout on the http client.
	c.inner.HttpClient.Timeout = DefaultTimeout

	// Create the retry transport with cli-kit defaults.
	c.retryTransport = ratelimit.NewRetryTransport(authTransport)
	c.retryTransport.BaseDelay = 500 * time.Millisecond
	c.retryTransport.MaxDelay = 30 * time.Second

	// Wire up debug logging for retries.
	c.retryTransport.OnRetry = func(attempt int, delay time.Duration, statusCode int) {
		debug.Log("HTTP %d: retrying in %v (attempt %d/%d)", statusCode, delay, attempt+1, c.retryTransport.MaxRetries+1)
	}

	// Apply user options (they may change retry config, timeout, etc.)
	for _, opt := range opts {
		opt(c)
	}

	// Install the retry transport on the HTTP client.
	c.inner.HttpClient.Transport = c.retryTransport

	return c
}

// Discover fetches the JMAP session from the server. It is safe to call
// multiple times; the session is only fetched once and cached for the
// client's lifetime.
func (c *Client) Discover() error {
	c.sessionOnce.Do(func() {
		debug.Log("JMAP session discovery: %s", c.inner.SessionEndpoint)
		c.sessionErr = c.inner.Authenticate()
		if c.sessionErr != nil {
			debug.Log("session discovery failed: %v", c.sessionErr)
			c.sessionErr = classifyError(c.sessionErr)
		} else {
			debug.Log("session discovery succeeded")
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
	if debug.Enabled() {
		for _, inv := range req.Calls {
			debug.Log("JMAP request: %s", inv.Name)
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
		return clierrors.WrapAuth(err,
			"Authentication failed. Your API token may be revoked or invalid",
			"Run 'fm auth login' to set a new token",
		)
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
