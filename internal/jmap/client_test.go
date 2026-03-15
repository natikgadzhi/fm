package jmap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	gojmap "git.sr.ht/~rockorager/go-jmap"
)

// sessionJSON returns a minimal valid JMAP session response.
func sessionJSON(apiURL string) []byte {
	session := map[string]any{
		"capabilities": map[string]any{
			"urn:ietf:params:jmap:core": map[string]any{
				"maxSizeUpload":          50000000,
				"maxConcurrentUpload":    8,
				"maxSizeRequest":         10000000,
				"maxConcurrentRequests":  8,
				"maxCallsInRequest":      64,
				"maxObjectsInGet":        1000,
				"maxObjectsInSet":        1000,
				"collationAlgorithms":    []string{},
			},
			"urn:ietf:params:jmap:mail": map[string]any{},
		},
		"accounts": map[string]any{
			"u12345": map[string]any{
				"name":               "test@fastmail.com",
				"isPersonal":         true,
				"isReadOnly":         false,
				"accountCapabilities": map[string]any{},
			},
		},
		"primaryAccounts": map[string]any{
			"urn:ietf:params:jmap:core": "u12345",
			"urn:ietf:params:jmap:mail": "u12345",
		},
		"username":       "test@fastmail.com",
		"apiUrl":         apiURL + "/jmap/api",
		"downloadUrl":    apiURL + "/jmap/download/{accountId}/{blobId}/{name}?type={type}",
		"uploadUrl":      apiURL + "/jmap/upload/{accountId}/",
		"eventSourceUrl": apiURL + "/jmap/eventsource/?types={types}&closeafter={closeafter}&ping={ping}",
		"state":          "abc123",
	}
	data, _ := json.Marshal(session)
	return data
}

func TestDiscoverSuccess(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header is present.
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("expected Authorization header, got none")
		}
		if auth != "Bearer fmu1-test-token" {
			t.Errorf("expected Bearer token, got %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)

	err := c.Discover()
	if err != nil {
		t.Fatalf("Discover() failed: %v", err)
	}

	session := c.Session()
	if session == nil {
		t.Fatal("session is nil after Discover()")
	}
	if session.Username != "test@fastmail.com" {
		t.Errorf("expected username 'test@fastmail.com', got '%s'", session.Username)
	}

	// Verify account ID extraction.
	accountID, err := c.PrimaryAccountID()
	if err != nil {
		t.Fatalf("PrimaryAccountID() failed: %v", err)
	}
	if accountID != "u12345" {
		t.Errorf("expected account ID 'u12345', got '%s'", accountID)
	}
}

func TestDiscoverCalledOnce(t *testing.T) {
	var callCount atomic.Int32

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
	)

	// Call Discover multiple times.
	for range 5 {
		if err := c.Discover(); err != nil {
			t.Fatalf("Discover() failed: %v", err)
		}
	}

	if count := callCount.Load(); count != 1 {
		t.Errorf("expected session endpoint to be called once, got %d", count)
	}
}

func TestDiscoverAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"type":"unauthorized","status":401,"detail":"Invalid token"}`))
	}))
	defer server.Close()

	c := NewClient("bad-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithMaxRetries(0), // Don't retry auth failures.
	)

	err := c.Discover()
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestRetryOn429(t *testing.T) {
	var attempts atomic.Int32

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count <= 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"type":"rate-limit","status":429,"detail":"slow down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	var totalSleep time.Duration
	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		withSleepFn(func(d time.Duration) {
			totalSleep += d
		}),
	)

	err := c.Discover()
	if err != nil {
		t.Fatalf("Discover() should succeed after retries, got: %v", err)
	}

	if count := attempts.Load(); count != 3 {
		t.Errorf("expected 3 attempts (2 retries + 1 success), got %d", count)
	}

	// Verify that sleep was called with ~1 second delays (from Retry-After header).
	if totalSleep < 2*time.Second {
		t.Errorf("expected total sleep >= 2s (from Retry-After), got %v", totalSleep)
	}
}

func TestRetryOn500(t *testing.T) {
	var attempts atomic.Int32

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		withSleepFn(func(d time.Duration) {}), // no-op sleep for fast tests
	)

	err := c.Discover()
	if err != nil {
		t.Fatalf("Discover() should succeed after retry, got: %v", err)
	}

	if count := attempts.Load(); count != 2 {
		t.Errorf("expected 2 attempts (1 retry + 1 success), got %d", count)
	}
}

func TestRetryMaxRetriesExhausted(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("always failing"))
	}))
	defer server.Close()

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithMaxRetries(3),
		withSleepFn(func(d time.Duration) {}),
	)

	err := c.Discover()
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}

	// 1 initial + 3 retries = 4 total.
	if count := attempts.Load(); count != 4 {
		t.Errorf("expected 4 attempts, got %d", count)
	}
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server.
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON("http://localhost"))
	}))
	defer server.Close()

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(100*time.Millisecond),
		WithMaxRetries(0),
	)

	err := c.Discover()
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestPrimaryAccountIDNoSession(t *testing.T) {
	c := NewClient("fmu1-test-token")

	_, err := c.PrimaryAccountID()
	if err == nil {
		t.Fatal("expected error when session is nil, got nil")
	}
}

func TestDoRequest(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST with JSON content type.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		resp := gojmap.Response{
			Responses:    []*gojmap.Invocation{},
			SessionState: "abc123",
		}
		data, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
	)

	req := &gojmap.Request{
		Using: []gojmap.URI{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
	}

	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Do() returned nil response")
	}
	if resp.SessionState != "abc123" {
		t.Errorf("expected session state 'abc123', got '%s'", resp.SessionState)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantMin  time.Duration
		wantZero bool
	}{
		{
			name:    "seconds",
			value:   "5",
			wantMin: 5 * time.Second,
		},
		{
			name:    "seconds with whitespace",
			value:   " 3 ",
			wantMin: 3 * time.Second,
		},
		{
			name:     "zero",
			value:    "0",
			wantZero: true,
		},
		{
			name:     "negative",
			value:    "-1",
			wantZero: true,
		},
		{
			name:     "empty",
			value:    "",
			wantZero: true,
		},
		{
			name:     "garbage",
			value:    "not-a-number",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if tt.wantZero && got != 0 {
				t.Errorf("expected 0, got %v", got)
			}
			if !tt.wantZero && got < tt.wantMin {
				t.Errorf("expected >= %v, got %v", tt.wantMin, got)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{201, false},
		{301, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
		{599, true},
	}

	for _, tt := range tests {
		got := isRetryable(tt.code)
		if got != tt.want {
			t.Errorf("isRetryable(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}
