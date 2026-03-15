package jmap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// jmapAPIResponse builds a JSON-encoded JMAP response with the given method
// responses. Each entry in methodResponses should be [name, args, callID].
func jmapAPIResponse(methodResponses []any) []byte {
	resp := map[string]any{
		"methodResponses": methodResponses,
		"sessionState":    "state-1",
	}
	data, _ := json.Marshal(resp)
	return data
}

func TestGetMailboxes(t *testing.T) {
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
				"Mailbox/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "mbox-state-1",
					"list": []map[string]any{
						{
							"id":           "mbox-inbox",
							"name":         "Inbox",
							"role":         "inbox",
							"totalEmails":  42,
							"unreadEmails": 5,
						},
						{
							"id":           "mbox-sent",
							"name":         "Sent",
							"role":         "sent",
							"totalEmails":  100,
							"unreadEmails": 0,
						},
						{
							"id":           "mbox-drafts",
							"name":         "Drafts",
							"role":         "drafts",
							"totalEmails":  3,
							"unreadEmails": 3,
						},
						{
							"id":           "mbox-archive",
							"name":         "Archive",
							"role":         "archive",
							"totalEmails":  999,
							"unreadEmails": 0,
						},
						{
							"id":           "mbox-trash",
							"name":         "Trash",
							"role":         "trash",
							"totalEmails":  10,
							"unreadEmails": 0,
						},
						{
							"id":       "mbox-custom",
							"name":     "My Folder",
							"parentId": "mbox-inbox",
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

	mailboxes, err := c.GetMailboxes(context.Background())
	if err != nil {
		t.Fatalf("GetMailboxes() failed: %v", err)
	}

	if len(mailboxes) != 6 {
		t.Fatalf("expected 6 mailboxes, got %d", len(mailboxes))
	}

	// Verify the inbox mailbox.
	inbox := mailboxes[0]
	if inbox.Id != "mbox-inbox" {
		t.Errorf("expected inbox ID 'mbox-inbox', got %q", inbox.Id)
	}
	if inbox.Name != "Inbox" {
		t.Errorf("expected inbox name 'Inbox', got %q", inbox.Name)
	}
	if inbox.Role != "inbox" {
		t.Errorf("expected inbox role 'inbox', got %q", inbox.Role)
	}
	if inbox.TotalEmails != 42 {
		t.Errorf("expected 42 total emails, got %d", inbox.TotalEmails)
	}
	if inbox.UnreadEmails != 5 {
		t.Errorf("expected 5 unread emails, got %d", inbox.UnreadEmails)
	}

	// Verify the custom folder with parent.
	custom := mailboxes[5]
	if custom.Id != "mbox-custom" {
		t.Errorf("expected custom ID 'mbox-custom', got %q", custom.Id)
	}
	if custom.Name != "My Folder" {
		t.Errorf("expected name 'My Folder', got %q", custom.Name)
	}
	if custom.ParentId != "mbox-inbox" {
		t.Errorf("expected parentId 'mbox-inbox', got %q", custom.ParentId)
	}

	// Verify different roles are present.
	roles := make(map[string]bool)
	for _, m := range mailboxes {
		if m.Role != "" {
			roles[m.Role] = true
		}
	}
	for _, expected := range []string{"inbox", "sent", "drafts", "archive", "trash"} {
		if !roles[expected] {
			t.Errorf("expected role %q to be present", expected)
		}
	}
}

func TestResolveMailboxByName(t *testing.T) {
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
				"Mailbox/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "mbox-state-1",
					"list": []map[string]any{
						{
							"id":   "mbox-inbox",
							"name": "Inbox",
							"role": "inbox",
						},
						{
							"id":   "mbox-sent",
							"name": "Sent Mail",
							"role": "sent",
						},
						{
							"id":   "mbox-custom",
							"name": "Project Alpha",
						},
					},
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

	tests := []struct {
		name     string
		query    string
		wantID   string
		wantErr  bool
	}{
		{
			name:   "exact name match",
			query:  "Inbox",
			wantID: "mbox-inbox",
		},
		{
			name:   "case-insensitive name match",
			query:  "inbox",
			wantID: "mbox-inbox",
		},
		{
			name:   "case-insensitive name match uppercase",
			query:  "INBOX",
			wantID: "mbox-inbox",
		},
		{
			name:   "match by role",
			query:  "sent",
			wantID: "mbox-sent",
		},
		{
			name:   "match by display name with spaces",
			query:  "Sent Mail",
			wantID: "mbox-sent",
		},
		{
			name:   "case-insensitive custom folder",
			query:  "project alpha",
			wantID: "mbox-custom",
		},
		{
			name:    "non-existent mailbox",
			query:   "nonexistent",
			wantErr: true,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := c.ResolveMailbox(context.Background(), tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("expected ID %q, got %q", tt.wantID, id)
			}
		})
	}
}

func TestGetMailboxesEmptyResponse(t *testing.T) {
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
				"Mailbox/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "mbox-state-1",
					"list":      []map[string]any{},
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

	mailboxes, err := c.GetMailboxes(context.Background())
	if err != nil {
		t.Fatalf("GetMailboxes() failed: %v", err)
	}

	if len(mailboxes) != 0 {
		t.Errorf("expected 0 mailboxes, got %d", len(mailboxes))
	}
}
