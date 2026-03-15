package jmap

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEmailJSONRoundTrip(t *testing.T) {
	date := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	original := Email{
		Id:        "email-001",
		ThreadId:  "thread-001",
		MessageId: "<abc123@example.com>",
		From:      []Address{{Name: "Alice", Email: "alice@example.com"}},
		To:        []Address{{Name: "Bob", Email: "bob@example.com"}},
		Cc:        []Address{{Name: "Charlie", Email: "charlie@example.com"}},
		Subject:   "Hello World",
		Date:      date,
		TextBody:  "Plain text body",
		HtmlBody:  "<p>HTML body</p>",
		Preview:   "Plain text body",
		MailboxIds: map[string]bool{
			"inbox-id": true,
		},
		Size:          4096,
		HasAttachment: true,
		Attachments: []Attachment{
			{
				BlobId:  "blob-001",
				Name:    "report.pdf",
				Type:    "application/pdf",
				Size:    1024,
				Charset: "",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Email
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Id != original.Id {
		t.Errorf("Id: got %q, want %q", decoded.Id, original.Id)
	}
	if decoded.ThreadId != original.ThreadId {
		t.Errorf("ThreadId: got %q, want %q", decoded.ThreadId, original.ThreadId)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject: got %q, want %q", decoded.Subject, original.Subject)
	}
	if !decoded.Date.Equal(original.Date) {
		t.Errorf("Date: got %v, want %v", decoded.Date, original.Date)
	}
	if decoded.Size != original.Size {
		t.Errorf("Size: got %d, want %d", decoded.Size, original.Size)
	}
	if decoded.HasAttachment != original.HasAttachment {
		t.Errorf("HasAttachment: got %v, want %v", decoded.HasAttachment, original.HasAttachment)
	}
	if len(decoded.From) != 1 || decoded.From[0].Email != "alice@example.com" {
		t.Errorf("From: got %v, want alice@example.com", decoded.From)
	}
	if len(decoded.To) != 1 || decoded.To[0].Email != "bob@example.com" {
		t.Errorf("To: got %v, want bob@example.com", decoded.To)
	}
	if len(decoded.Cc) != 1 || decoded.Cc[0].Email != "charlie@example.com" {
		t.Errorf("Cc: got %v, want charlie@example.com", decoded.Cc)
	}
	if len(decoded.Attachments) != 1 {
		t.Fatalf("Attachments: got %d, want 1", len(decoded.Attachments))
	}
	if decoded.Attachments[0].Name != "report.pdf" {
		t.Errorf("Attachment Name: got %q, want %q", decoded.Attachments[0].Name, "report.pdf")
	}
	if !decoded.MailboxIds["inbox-id"] {
		t.Errorf("MailboxIds missing inbox-id")
	}
}

func TestAddressJSONRoundTrip(t *testing.T) {
	original := Address{Name: "Test User", Email: "test@example.com"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Address
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

func TestMailboxJSONRoundTrip(t *testing.T) {
	original := Mailbox{
		Id:           "mbox-001",
		Name:         "Inbox",
		Role:         "inbox",
		TotalEmails:  150,
		UnreadEmails: 3,
		ParentId:     "",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Mailbox
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

func TestThreadJSONRoundTrip(t *testing.T) {
	original := Thread{
		Id:       "thread-001",
		EmailIds: []string{"email-001", "email-002", "email-003"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Thread
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Id != original.Id {
		t.Errorf("Id: got %q, want %q", decoded.Id, original.Id)
	}
	if len(decoded.EmailIds) != len(original.EmailIds) {
		t.Fatalf("EmailIds length: got %d, want %d", len(decoded.EmailIds), len(original.EmailIds))
	}
	for i := range original.EmailIds {
		if decoded.EmailIds[i] != original.EmailIds[i] {
			t.Errorf("EmailIds[%d]: got %q, want %q", i, decoded.EmailIds[i], original.EmailIds[i])
		}
	}
}

func TestAttachmentJSONRoundTrip(t *testing.T) {
	original := Attachment{
		BlobId:  "blob-123",
		Name:    "document.txt",
		Type:    "text/plain",
		Size:    512,
		Charset: "utf-8",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Attachment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

func TestSearchFilterJSONRoundTrip(t *testing.T) {
	before := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	original := SearchFilter{
		From:          "alice@example.com",
		To:            "bob@example.com",
		Subject:       "meeting",
		Text:          "quarterly review",
		InMailbox:     "INBOX",
		Before:        &before,
		After:         &after,
		HasAttachment: true,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SearchFilter
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.From != original.From {
		t.Errorf("From: got %q, want %q", decoded.From, original.From)
	}
	if decoded.To != original.To {
		t.Errorf("To: got %q, want %q", decoded.To, original.To)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject: got %q, want %q", decoded.Subject, original.Subject)
	}
	if decoded.Text != original.Text {
		t.Errorf("Text: got %q, want %q", decoded.Text, original.Text)
	}
	if decoded.InMailbox != original.InMailbox {
		t.Errorf("InMailbox: got %q, want %q", decoded.InMailbox, original.InMailbox)
	}
	if decoded.HasAttachment != original.HasAttachment {
		t.Errorf("HasAttachment: got %v, want %v", decoded.HasAttachment, original.HasAttachment)
	}
	if decoded.Before == nil || !decoded.Before.Equal(*original.Before) {
		t.Errorf("Before: got %v, want %v", decoded.Before, original.Before)
	}
	if decoded.After == nil || !decoded.After.Equal(*original.After) {
		t.Errorf("After: got %v, want %v", decoded.After, original.After)
	}
}

func TestSearchFilterJSONOmitsEmpty(t *testing.T) {
	filter := SearchFilter{}

	data, err := json.Marshal(filter)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// An empty filter should serialize to just "{}"
	if string(data) != "{}" {
		t.Errorf("empty filter should be {}, got %s", string(data))
	}
}

func TestEmailWithNilSlices(t *testing.T) {
	// Ensure an email with nil slices serializes and deserializes cleanly
	original := Email{
		Id:      "email-minimal",
		Subject: "Minimal email",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Email
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Id != original.Id {
		t.Errorf("Id: got %q, want %q", decoded.Id, original.Id)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject: got %q, want %q", decoded.Subject, original.Subject)
	}
}
