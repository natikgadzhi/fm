package jmap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
)

// newTestServer creates an httptest.Server that handles /jmap/session and
// /jmap/api. The apiHandler receives the decoded JMAP request and returns
// the response to marshal back.
func newTestServer(t *testing.T, apiHandler func(t *testing.T, req map[string]any) map[string]any) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewUnstartedServer(mux)
	server.Start()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode JMAP request: %v", err)
		}
		resp := apiHandler(t, reqBody)
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(resp)
		w.Write(data)
	})

	return server
}

// newTestClient creates a Client pointing at the test server.
func newTestClient(server *httptest.Server) *Client {
	return NewClient("test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
	)
}

func TestQueryEmails(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		// Verify the method is Email/query.
		calls, _ := req["methodCalls"].([]any)
		if len(calls) != 1 {
			t.Fatalf("expected 1 method call, got %d", len(calls))
		}
		call := calls[0].([]any)
		methodName := call[0].(string)
		if methodName != "Email/query" {
			t.Errorf("expected method Email/query, got %s", methodName)
		}

		// Verify filter contains "from" field.
		args := call[1].(map[string]any)
		filter, _ := args["filter"].(map[string]any)
		if filter["from"] != "alice@example.com" {
			t.Errorf("expected from filter 'alice@example.com', got %v", filter["from"])
		}

		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Email/query",
					map[string]any{
						"accountId":  "u12345",
						"queryState": "state1",
						"ids":        []string{"msg-001", "msg-002", "msg-003"},
						"position":   0,
						"total":      3,
					},
					"0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)

	filter := SearchFilter{From: "alice@example.com"}
	ids, err := c.QueryEmails(context.Background(), filter, 10)
	if err != nil {
		t.Fatalf("QueryEmails() failed: %v", err)
	}

	if len(ids) != 3 {
		t.Fatalf("expected 3 IDs, got %d", len(ids))
	}
	if ids[0] != "msg-001" || ids[1] != "msg-002" || ids[2] != "msg-003" {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestQueryEmailsZeroResults(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Email/query",
					map[string]any{
						"accountId":  "u12345",
						"queryState": "state1",
						"ids":        []string{},
						"position":   0,
						"total":      0,
					},
					"0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)

	ids, err := c.QueryEmails(context.Background(), SearchFilter{Text: "nonexistent"}, 50)
	if err != nil {
		t.Fatalf("QueryEmails() failed: %v", err)
	}

	if len(ids) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(ids))
	}
}

func TestGetEmails(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		calls, _ := req["methodCalls"].([]any)
		call := calls[0].([]any)
		methodName := call[0].(string)
		if methodName != "Email/get" {
			t.Errorf("expected method Email/get, got %s", methodName)
		}

		// Verify IDs are passed correctly.
		args := call[1].(map[string]any)
		ids, _ := args["ids"].([]any)
		if len(ids) != 2 {
			t.Fatalf("expected 2 IDs, got %d", len(ids))
		}

		return emailGetResponse([]map[string]any{
			{
				"id":            "msg-001",
				"threadId":      "thread-001",
				"messageId":     []string{"<abc@example.com>"},
				"from":          []map[string]any{{"name": "Alice", "email": "alice@example.com"}},
				"to":            []map[string]any{{"name": "Bob", "email": "bob@example.com"}},
				"cc":            []map[string]any{},
				"subject":       "Hello World",
				"sentAt":        "2025-01-15T10:30:00Z",
				"preview":       "This is a preview",
				"textBody":      []map[string]any{{"partId": "1"}},
				"htmlBody":      []map[string]any{{"partId": "2"}},
				"bodyValues":    map[string]any{"1": map[string]any{"value": "Hello plain"}, "2": map[string]any{"value": "<p>Hello html</p>"}},
				"mailboxIds":    map[string]any{"inbox-id": true},
				"attachments":   []map[string]any{},
				"hasAttachment": false,
				"size":          1234,
			},
			{
				"id":            "msg-002",
				"threadId":      "thread-002",
				"messageId":     []string{"<def@example.com>"},
				"from":          []map[string]any{{"name": "Charlie", "email": "charlie@example.com"}},
				"to":            []map[string]any{{"name": "Alice", "email": "alice@example.com"}},
				"cc":            nil,
				"subject":       "Re: Hello World",
				"sentAt":        "2025-01-15T11:00:00Z",
				"preview":       "Another preview",
				"textBody":      []map[string]any{{"partId": "1"}},
				"htmlBody":      []map[string]any{},
				"bodyValues":    map[string]any{"1": map[string]any{"value": "Reply text"}},
				"mailboxIds":    map[string]any{"inbox-id": true},
				"attachments":   []map[string]any{{"blobId": "blob-1", "name": "file.pdf", "type": "application/pdf", "size": 5678}},
				"hasAttachment": true,
				"size":          6789,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)

	emails, err := c.GetEmails(context.Background(), []string{"msg-001", "msg-002"})
	if err != nil {
		t.Fatalf("GetEmails() failed: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}

	// Verify first email.
	e1 := emails[0]
	if e1.Id != "msg-001" {
		t.Errorf("expected ID 'msg-001', got '%s'", e1.Id)
	}
	if e1.Subject != "Hello World" {
		t.Errorf("expected subject 'Hello World', got '%s'", e1.Subject)
	}
	if len(e1.From) != 1 || e1.From[0].Email != "alice@example.com" {
		t.Errorf("unexpected From: %v", e1.From)
	}
	if e1.TextBody != "Hello plain" {
		t.Errorf("expected text body 'Hello plain', got '%s'", e1.TextBody)
	}
	if e1.HtmlBody != "<p>Hello html</p>" {
		t.Errorf("expected html body '<p>Hello html</p>', got '%s'", e1.HtmlBody)
	}
	if e1.Preview != "This is a preview" {
		t.Errorf("expected preview 'This is a preview', got '%s'", e1.Preview)
	}
	if _, ok := e1.MailboxIds["inbox-id"]; !ok {
		t.Errorf("expected mailboxIds to contain 'inbox-id', got %v", e1.MailboxIds)
	}

	// Verify second email has attachment.
	e2 := emails[1]
	if !e2.HasAttachment {
		t.Error("expected HasAttachment to be true for second email")
	}
	if len(e2.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(e2.Attachments))
	}
	if e2.Attachments[0].Name != "file.pdf" {
		t.Errorf("expected attachment name 'file.pdf', got '%s'", e2.Attachments[0].Name)
	}
}

func TestGetEmailsEmpty(t *testing.T) {
	c := NewClient("test-token")
	emails, err := c.GetEmails(context.Background(), []string{})
	if err != nil {
		t.Fatalf("GetEmails() with empty IDs should not fail: %v", err)
	}
	if emails != nil {
		t.Errorf("expected nil for empty IDs, got %v", emails)
	}
}

func TestSearchEmails(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		calls, _ := req["methodCalls"].([]any)
		if len(calls) != 2 {
			t.Fatalf("expected 2 method calls (query + get), got %d", len(calls))
		}

		// Verify first call is Email/query.
		queryCall := calls[0].([]any)
		if queryCall[0].(string) != "Email/query" {
			t.Errorf("expected first call to be Email/query, got %s", queryCall[0])
		}

		// Verify second call is Email/get with result reference.
		getCall := calls[1].([]any)
		if getCall[0].(string) != "Email/get" {
			t.Errorf("expected second call to be Email/get, got %s", getCall[0])
		}
		getArgs := getCall[1].(map[string]any)
		refIDs, ok := getArgs["#ids"]
		if !ok {
			t.Fatal("expected Email/get to have #ids result reference")
		}
		ref := refIDs.(map[string]any)
		if ref["resultOf"] != "0" {
			t.Errorf("expected resultOf '0', got '%v'", ref["resultOf"])
		}
		if ref["name"] != "Email/query" {
			t.Errorf("expected name 'Email/query', got '%v'", ref["name"])
		}
		if ref["path"] != "/ids" {
			t.Errorf("expected path '/ids', got '%v'", ref["path"])
		}

		// Return query response + get response.
		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Email/query",
					map[string]any{
						"accountId":  "u12345",
						"queryState": "state1",
						"ids":        []string{"msg-001"},
						"position":   0,
						"total":      1,
					},
					"0",
				},
				[]any{
					"Email/get",
					map[string]any{
						"accountId": "u12345",
						"state":     "state1",
						"list": []map[string]any{
							{
								"id":            "msg-001",
								"threadId":      "thread-001",
								"messageId":     []string{"<search@example.com>"},
								"from":          []map[string]any{{"name": "Sender", "email": "sender@example.com"}},
								"to":            []map[string]any{{"name": "Me", "email": "me@example.com"}},
								"subject":       "Search Result",
								"sentAt":        "2025-02-01T08:00:00Z",
								"preview":       "Found it",
								"textBody":      []map[string]any{{"partId": "1"}},
								"htmlBody":      []map[string]any{},
								"bodyValues":    map[string]any{"1": map[string]any{"value": "Found it body"}},
								"mailboxIds":    map[string]any{"inbox-id": true},
								"hasAttachment": false,
								"size":          512,
							},
						},
						"notFound": []string{},
					},
					"1",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)

	filter := SearchFilter{Subject: "Search Result"}
	emails, err := c.SearchEmails(context.Background(), filter, 10)
	if err != nil {
		t.Fatalf("SearchEmails() failed: %v", err)
	}

	if len(emails) != 1 {
		t.Fatalf("expected 1 email, got %d", len(emails))
	}

	if emails[0].Subject != "Search Result" {
		t.Errorf("expected subject 'Search Result', got '%s'", emails[0].Subject)
	}
	if emails[0].TextBody != "Found it body" {
		t.Errorf("expected text body 'Found it body', got '%s'", emails[0].TextBody)
	}
}

func TestSearchEmailsZeroResults(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Email/query",
					map[string]any{
						"accountId":  "u12345",
						"queryState": "state1",
						"ids":        []string{},
						"position":   0,
						"total":      0,
					},
					"0",
				},
				[]any{
					"Email/get",
					map[string]any{
						"accountId": "u12345",
						"state":     "state1",
						"list":      []map[string]any{},
						"notFound":  []string{},
					},
					"1",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)

	emails, err := c.SearchEmails(context.Background(), SearchFilter{Text: "nonexistent"}, 50)
	if err != nil {
		t.Fatalf("SearchEmails() failed: %v", err)
	}

	if len(emails) != 0 {
		t.Errorf("expected 0 emails, got %d", len(emails))
	}
}

func TestDownloadAttachment(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/download/u12345/blob-abc/report.pdf", func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a GET request.
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("fake-pdf-content"))
	})

	c := newTestClient(server)

	data, err := c.DownloadAttachment(context.Background(), "u12345", "blob-abc", "report.pdf")
	if err != nil {
		t.Fatalf("DownloadAttachment() failed: %v", err)
	}

	if string(data) != "fake-pdf-content" {
		t.Errorf("expected 'fake-pdf-content', got '%s'", string(data))
	}
}

func TestDownloadAttachmentNotFound(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/download/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	c := newTestClient(server)

	_, err := c.DownloadAttachment(context.Background(), "u12345", "blob-nonexistent", "missing.txt")
	if err == nil {
		t.Fatal("expected error for 404 download, got nil")
	}
}

func TestToFilterCondition(t *testing.T) {
	before := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	filter := SearchFilter{
		From:          "alice@example.com",
		To:            "bob@example.com",
		Subject:       "meeting notes",
		Text:          "quarterly",
		InMailbox:     "inbox-001",
		Before:        &before,
		After:         &after,
		HasAttachment: true,
	}

	fc := toFilterCondition(filter)

	if fc.From != "alice@example.com" {
		t.Errorf("expected From 'alice@example.com', got '%s'", fc.From)
	}
	if fc.To != "bob@example.com" {
		t.Errorf("expected To 'bob@example.com', got '%s'", fc.To)
	}
	if fc.Subject != "meeting notes" {
		t.Errorf("expected Subject 'meeting notes', got '%s'", fc.Subject)
	}
	if fc.Text != "quarterly" {
		t.Errorf("expected Text 'quarterly', got '%s'", fc.Text)
	}
	if string(fc.InMailbox) != "inbox-001" {
		t.Errorf("expected InMailbox 'inbox-001', got '%s'", fc.InMailbox)
	}
	if fc.Before == nil || !fc.Before.Equal(before) {
		t.Errorf("expected Before %v, got %v", before, fc.Before)
	}
	if fc.After == nil || !fc.After.Equal(after) {
		t.Errorf("expected After %v, got %v", after, fc.After)
	}
	if !fc.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
}

func TestToFilterConditionEmpty(t *testing.T) {
	fc := toFilterCondition(SearchFilter{})

	if fc.From != "" || fc.To != "" || fc.Subject != "" || fc.Text != "" {
		t.Error("expected all string fields to be empty for empty filter")
	}
	if string(fc.InMailbox) != "" {
		t.Error("expected InMailbox to be empty")
	}
	if fc.Before != nil || fc.After != nil {
		t.Error("expected Before and After to be nil")
	}
	if fc.HasAttachment {
		t.Error("expected HasAttachment to be false")
	}
}

func TestMapEmailBodyValues(t *testing.T) {
	// Test that body values are correctly extracted from BodyPart -> BodyValues mapping.
	e := &email.Email{
		ID:      "test-id",
		Subject: "Body Test",
		TextBody: []*email.BodyPart{
			{PartID: "part-1"},
		},
		HTMLBody: []*email.BodyPart{
			{PartID: "part-2"},
		},
		BodyValues: map[string]*email.BodyValue{
			"part-1": {Value: "Plain text content"},
			"part-2": {Value: "<p>HTML content</p>"},
		},
	}

	result := mapEmail(e)

	if result.TextBody != "Plain text content" {
		t.Errorf("expected text body 'Plain text content', got '%s'", result.TextBody)
	}
	if result.HtmlBody != "<p>HTML content</p>" {
		t.Errorf("expected html body '<p>HTML content</p>', got '%s'", result.HtmlBody)
	}
}

func TestMapEmailNoBodyValues(t *testing.T) {
	// Test graceful handling when body parts exist but no body values.
	e := &email.Email{
		ID:      "test-id",
		Subject: "No Body Values",
		TextBody: []*email.BodyPart{
			{PartID: "part-1"},
		},
		BodyValues: nil,
	}

	result := mapEmail(e)

	if result.TextBody != "" {
		t.Errorf("expected empty text body, got '%s'", result.TextBody)
	}
}

// emailGetResponse builds a standard Email/get JMAP response.
func emailGetResponse(emails []map[string]any) map[string]any {
	return map[string]any{
		"methodResponses": []any{
			[]any{
				"Email/get",
				map[string]any{
					"accountId": "u12345",
					"state":     "state1",
					"list":      emails,
					"notFound":  []string{},
				},
				"0",
			},
		},
		"sessionState": "abc123",
	}
}

func TestGetEmailsBatchedSingleBatch(t *testing.T) {
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return emailGetResponse([]map[string]any{
			{
				"id": "msg-001", "threadId": "t1", "subject": "First",
				"from": []map[string]any{{"name": "A", "email": "a@test.com"}},
				"to": []map[string]any{}, "sentAt": "2025-01-15T10:00:00Z",
				"textBody": []map[string]any{}, "htmlBody": []map[string]any{},
				"bodyValues": map[string]any{}, "mailboxIds": map[string]any{},
				"hasAttachment": false, "size": 100,
			},
			{
				"id": "msg-002", "threadId": "t2", "subject": "Second",
				"from": []map[string]any{{"name": "B", "email": "b@test.com"}},
				"to": []map[string]any{}, "sentAt": "2025-01-15T11:00:00Z",
				"textBody": []map[string]any{}, "htmlBody": []map[string]any{},
				"bodyValues": map[string]any{}, "mailboxIds": map[string]any{},
				"hasAttachment": false, "size": 200,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)

	var progressCount int
	emails, err := c.GetEmailsBatched(context.Background(), []string{"msg-001", "msg-002"}, 50, func() {
		progressCount++
	})
	if err != nil {
		t.Fatalf("GetEmailsBatched() failed: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}
	if progressCount != 2 {
		t.Errorf("expected progress callback called 2 times, got %d", progressCount)
	}
}

func TestGetEmailsBatchedEmptyIDs(t *testing.T) {
	c := NewClient("test-token")
	emails, err := c.GetEmailsBatched(context.Background(), []string{}, 10, nil)
	if err != nil {
		t.Fatalf("GetEmailsBatched() with empty IDs should not fail: %v", err)
	}
	if emails != nil {
		t.Errorf("expected nil for empty IDs, got %v", emails)
	}
}

func TestGetEmailsBatchedPartialFailure(t *testing.T) {
	var callCount int
	mux := http.NewServeMux()
	server := httptest.NewUnstartedServer(mux)
	server.Start()
	defer server.Close()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(sessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First batch succeeds.
			resp := emailGetResponse([]map[string]any{
				{
					"id": "msg-001", "threadId": "t1", "subject": "Success",
					"from": []map[string]any{{"name": "A", "email": "a@test.com"}},
					"to": []map[string]any{}, "sentAt": "2025-01-15T10:00:00Z",
					"textBody": []map[string]any{}, "htmlBody": []map[string]any{},
					"bodyValues": map[string]any{}, "mailboxIds": map[string]any{},
					"hasAttachment": false, "size": 100,
				},
			})
			data, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		} else {
			// Second batch fails with 500.
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}
	})

	c := NewClient("test-token",
		WithBaseURL(server.URL+"/jmap/session"),
		WithTimeout(5*time.Second),
		WithMaxRetries(0), // No retries so we fail fast.
	)

	_, err := c.GetEmailsBatched(context.Background(), []string{"msg-001", "msg-002"}, 1, nil)
	if err == nil {
		t.Fatal("expected error from second batch, got nil")
	}

	var partialErr *PartialResultError
	if !errors.As(err, &partialErr) {
		t.Fatalf("expected PartialResultError, got %T: %v", err, err)
	}

	if partialErr.Fetched != 1 {
		t.Errorf("expected 1 fetched email, got %d", partialErr.Fetched)
	}
	if partialErr.Total != 2 {
		t.Errorf("expected 2 total emails, got %d", partialErr.Total)
	}
	if len(partialErr.Emails) != 1 {
		t.Errorf("expected 1 email in partial result, got %d", len(partialErr.Emails))
	}
	if partialErr.Emails[0].Id != "msg-001" {
		t.Errorf("expected partial email ID 'msg-001', got '%s'", partialErr.Emails[0].Id)
	}
}

func TestGetEmailsBatchedDefaultBatchSize(t *testing.T) {
	// Test that passing 0 for batchSize uses DefaultBatchSize.
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return emailGetResponse([]map[string]any{
			{
				"id": "msg-001", "threadId": "t1", "subject": "Test",
				"from": []map[string]any{{"name": "A", "email": "a@test.com"}},
				"to": []map[string]any{}, "sentAt": "2025-01-15T10:00:00Z",
				"textBody": []map[string]any{}, "htmlBody": []map[string]any{},
				"bodyValues": map[string]any{}, "mailboxIds": map[string]any{},
				"hasAttachment": false, "size": 100,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	emails, err := c.GetEmailsBatched(context.Background(), []string{"msg-001"}, 0, nil)
	if err != nil {
		t.Fatalf("GetEmailsBatched() with default batch size failed: %v", err)
	}
	if len(emails) != 1 {
		t.Errorf("expected 1 email, got %d", len(emails))
	}
}

// Ensure the email package init() registers the methods we need.
// This is a compile-time test that the import is working correctly.
var _ gojmap.Method = (*email.Query)(nil)
var _ gojmap.Method = (*email.Get)(nil)
