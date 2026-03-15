package jmap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetThread(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		// Verify the request contains Thread/get.
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		body := jmapAPIResponse([]any{
			[]any{
				"Thread/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "thread-state-1",
					"list": []map[string]any{
						{
							"id":       "thread-1",
							"emailIds": []string{"email-a", "email-b", "email-c"},
						},
					},
					"notFound": []string{},
				},
				"0",
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)

	thread, err := c.GetThread(context.Background(), "thread-1")
	if err != nil {
		t.Fatalf("GetThread() failed: %v", err)
	}

	if thread.Id != "thread-1" {
		t.Errorf("expected thread ID 'thread-1', got %q", thread.Id)
	}

	if len(thread.EmailIds) != 3 {
		t.Fatalf("expected 3 email IDs, got %d", len(thread.EmailIds))
	}

	expectedIDs := []string{"email-a", "email-b", "email-c"}
	for i, id := range thread.EmailIds {
		if id != expectedIDs[i] {
			t.Errorf("email ID[%d]: expected %q, got %q", i, expectedIDs[i], id)
		}
	}
}

func TestGetThreadNotFound(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		body := jmapAPIResponse([]any{
			[]any{
				"Thread/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "thread-state-1",
					"list":      []map[string]any{},
					"notFound":  []string{"nonexistent-thread"},
				},
				"0",
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)

	_, err := c.GetThread(context.Background(), "nonexistent-thread")
	if err == nil {
		t.Fatal("expected error for non-existent thread, got nil")
	}
}

func TestGetThreadEmails(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		// This handler responds to the chained Thread/get + Email/get call.
		body := jmapAPIResponse([]any{
			// Thread/get response
			[]any{
				"Thread/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "thread-state-1",
					"list": []map[string]any{
						{
							"id":       "thread-1",
							"emailIds": []string{"email-c", "email-a", "email-b"},
						},
					},
					"notFound": []string{},
				},
				"0",
			},
			// Email/get response (emails in arbitrary server order)
			[]any{
				"Email/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "email-state-1",
					"list": []map[string]any{
						{
							"id":       "email-c",
							"threadId": "thread-1",
							"subject":  "Re: Re: Hello",
							"sentAt":   "2025-01-03T12:00:00Z",
							"from": []map[string]any{
								{"name": "Charlie", "email": "charlie@example.com"},
							},
							"to": []map[string]any{
								{"name": "Alice", "email": "alice@example.com"},
							},
							"preview":   "Thanks for the update.",
							"size":      3000,
							"mailboxIds": map[string]bool{"mbox-inbox": true},
						},
						{
							"id":       "email-a",
							"threadId": "thread-1",
							"subject":  "Hello",
							"sentAt":   "2025-01-01T10:00:00Z",
							"from": []map[string]any{
								{"name": "Alice", "email": "alice@example.com"},
							},
							"to": []map[string]any{
								{"name": "Bob", "email": "bob@example.com"},
							},
							"preview":   "Hi Bob, how are you?",
							"size":      1500,
							"mailboxIds": map[string]bool{"mbox-sent": true},
						},
						{
							"id":       "email-b",
							"threadId": "thread-1",
							"subject":  "Re: Hello",
							"sentAt":   "2025-01-02T14:30:00Z",
							"from": []map[string]any{
								{"name": "Bob", "email": "bob@example.com"},
							},
							"to": []map[string]any{
								{"name": "Alice", "email": "alice@example.com"},
							},
							"cc": []map[string]any{
								{"name": "Charlie", "email": "charlie@example.com"},
							},
							"preview":       "I'm doing well!",
							"size":          2000,
							"hasAttachment": true,
							"mailboxIds":    map[string]bool{"mbox-inbox": true},
						},
					},
					"notFound": []string{},
				},
				"1",
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)

	emails, err := c.GetThreadEmails(context.Background(), "thread-1")
	if err != nil {
		t.Fatalf("GetThreadEmails() failed: %v", err)
	}

	if len(emails) != 3 {
		t.Fatalf("expected 3 emails, got %d", len(emails))
	}

	// Verify emails are sorted by date ascending.
	if emails[0].Id != "email-a" {
		t.Errorf("expected first email to be 'email-a' (earliest), got %q", emails[0].Id)
	}
	if emails[1].Id != "email-b" {
		t.Errorf("expected second email to be 'email-b', got %q", emails[1].Id)
	}
	if emails[2].Id != "email-c" {
		t.Errorf("expected third email to be 'email-c' (latest), got %q", emails[2].Id)
	}

	// Verify first email fields.
	if emails[0].Subject != "Hello" {
		t.Errorf("expected subject 'Hello', got %q", emails[0].Subject)
	}
	if len(emails[0].From) != 1 || emails[0].From[0].Email != "alice@example.com" {
		t.Errorf("expected from alice@example.com, got %+v", emails[0].From)
	}
	if emails[0].ThreadId != "thread-1" {
		t.Errorf("expected threadId 'thread-1', got %q", emails[0].ThreadId)
	}

	// Verify second email has CC and attachment.
	if len(emails[1].Cc) != 1 || emails[1].Cc[0].Email != "charlie@example.com" {
		t.Errorf("expected CC charlie@example.com, got %+v", emails[1].Cc)
	}
	if !emails[1].HasAttachment {
		t.Error("expected email-b to have attachment")
	}

	// Verify dates are in order.
	for i := 1; i < len(emails); i++ {
		if emails[i].Date.Before(emails[i-1].Date) {
			t.Errorf("emails not sorted by date: email[%d] (%v) before email[%d] (%v)",
				i, emails[i].Date, i-1, emails[i-1].Date)
		}
	}
}

func TestGetThreadEmailsNotFound(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		body := jmapAPIResponse([]any{
			[]any{
				"Thread/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "thread-state-1",
					"list":      []map[string]any{},
					"notFound":  []string{"nonexistent-thread"},
				},
				"0",
			},
			// The Email/get response would be empty since the reference resolved to nothing.
			[]any{
				"Email/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "email-state-1",
					"list":      []map[string]any{},
					"notFound":  []string{},
				},
				"1",
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	c := NewClient("fmu1-test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)

	_, err := c.GetThreadEmails(context.Background(), "nonexistent-thread")
	if err == nil {
		t.Fatal("expected error for non-existent thread, got nil")
	}
}
