package jmap

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
)

// emailProperties is the set of Email properties we request from the server.
var emailProperties = []string{
	"id", "threadId", "messageId",
	"from", "to", "cc",
	"subject", "sentAt", "preview",
	"textBody", "htmlBody", "bodyValues",
	"mailboxIds", "attachments", "hasAttachment",
	"size",
}

// QueryEmails calls Email/query and returns matching email IDs.
func (c *Client) QueryEmails(ctx context.Context, filter SearchFilter, limit int) ([]string, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.PrimaryAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting account ID: %w", err)
	}

	query := &email.Query{
		Account: accountID,
		Filter:  toFilterCondition(filter),
		Sort: []*email.SortComparator{
			{Property: "receivedAt", IsAscending: false},
		},
	}
	if limit > 0 {
		query.Limit = uint64(limit)
	}

	req := &gojmap.Request{
		Context: ctx,
	}
	req.Invoke(query)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Email/query request failed: %w", err)
	}

	for _, inv := range resp.Responses {
		if qr, ok := inv.Args.(*email.QueryResponse); ok {
			ids := make([]string, len(qr.IDs))
			for i, id := range qr.IDs {
				ids[i] = string(id)
			}
			return ids, nil
		}
	}

	return nil, fmt.Errorf("Email/query: no query response in server reply")
}

// GetEmails calls Email/get and returns full Email objects for the given IDs.
func (c *Client) GetEmails(ctx context.Context, ids []string) ([]Email, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.PrimaryAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting account ID: %w", err)
	}

	jmapIDs := make([]gojmap.ID, len(ids))
	for i, id := range ids {
		jmapIDs[i] = gojmap.ID(id)
	}

	get := &email.Get{
		Account:            accountID,
		IDs:                jmapIDs,
		Properties:         emailProperties,
		FetchAllBodyValues: true,
	}

	req := &gojmap.Request{
		Context: ctx,
	}
	req.Invoke(get)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Email/get request failed: %w", err)
	}

	for _, inv := range resp.Responses {
		if gr, ok := inv.Args.(*email.GetResponse); ok {
			return mapEmails(gr.List, gr), nil
		}
	}

	return nil, fmt.Errorf("Email/get: no get response in server reply")
}

// SearchEmails chains Email/query and Email/get in a single JMAP request
// using result references, so that Email/get fetches the IDs returned by
// Email/query without a second round trip.
func (c *Client) SearchEmails(ctx context.Context, filter SearchFilter, limit int) ([]Email, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.PrimaryAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting account ID: %w", err)
	}

	query := &email.Query{
		Account: accountID,
		Filter:  toFilterCondition(filter),
		Sort: []*email.SortComparator{
			{Property: "receivedAt", IsAscending: false},
		},
	}
	if limit > 0 {
		query.Limit = uint64(limit)
	}

	req := &gojmap.Request{
		Context: ctx,
	}

	// First call: Email/query — returns email IDs.
	queryCallID := req.Invoke(query)

	// Second call: Email/get — references the query result IDs.
	req.Invoke(&email.Get{
		Account:    accountID,
		Properties: emailProperties,
		ReferenceIDs: &gojmap.ResultReference{
			ResultOf: queryCallID,
			Name:     "Email/query",
			Path:     "/ids",
		},
		FetchAllBodyValues: true,
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SearchEmails request failed: %w", err)
	}

	// Find the Email/get response (second invocation).
	for _, inv := range resp.Responses {
		if gr, ok := inv.Args.(*email.GetResponse); ok {
			return mapEmails(gr.List, gr), nil
		}
	}

	return nil, fmt.Errorf("SearchEmails: no get response in server reply")
}

// DownloadAttachment downloads a blob by its ID using the JMAP download URL
// template from the session. It returns the raw bytes of the attachment.
func (c *Client) DownloadAttachment(ctx context.Context, accountID, blobID, name string) ([]byte, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	session := c.Session()
	if session == nil {
		return nil, fmt.Errorf("no session available")
	}

	// The JMAP download URL is a template like:
	// https://api.example.com/jmap/download/{accountId}/{blobId}/{name}?type={type}
	url := session.DownloadURL
	url = strings.ReplaceAll(url, "{accountId}", accountID)
	url = strings.ReplaceAll(url, "{blobId}", blobID)
	url = strings.ReplaceAll(url, "{name}", name)
	url = strings.ReplaceAll(url, "{type}", "application/octet-stream")

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}

	httpResp, err := c.inner.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("downloading attachment: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", httpResp.StatusCode)
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading attachment body: %w", err)
	}

	return data, nil
}

// toFilterCondition converts our SearchFilter to go-jmap's email.FilterCondition.
func toFilterCondition(f SearchFilter) *email.FilterCondition {
	fc := &email.FilterCondition{
		From:          f.From,
		To:            f.To,
		Subject:       f.Subject,
		Text:          f.Text,
		HasAttachment: f.HasAttachment,
		Before:        f.Before,
		After:         f.After,
	}

	if f.InMailbox != "" {
		fc.InMailbox = gojmap.ID(f.InMailbox)
	}

	return fc
}

// mapEmails converts a slice of go-jmap email.Email objects to our Email type.
// The GetResponse is needed to access bodyValues for extracting body content.
func mapEmails(src []*email.Email, gr *email.GetResponse) []Email {
	result := make([]Email, 0, len(src))
	for _, e := range src {
		result = append(result, mapEmail(e))
	}
	return result
}

// mapEmail converts a single go-jmap email.Email to our Email type.
func mapEmail(e *email.Email) Email {
	em := Email{
		Id:            string(e.ID),
		ThreadId:      string(e.ThreadID),
		Subject:       e.Subject,
		Preview:       e.Preview,
		HasAttachment: e.HasAttachment,
		Size:          int64(e.Size),
	}

	// MessageId — go-jmap stores as []string, we take the first.
	if len(e.MessageID) > 0 {
		em.MessageId = e.MessageID[0]
	}

	// Date — use SentAt if available.
	if e.SentAt != nil {
		em.Date = *e.SentAt
	}

	// From addresses.
	em.From = mapAddresses(e.From)
	em.To = mapAddresses(e.To)
	em.Cc = mapAddresses(e.CC)

	// MailboxIds.
	if len(e.MailboxIDs) > 0 {
		em.MailboxIds = make(map[string]bool, len(e.MailboxIDs))
		for id, v := range e.MailboxIDs {
			em.MailboxIds[string(id)] = v
		}
	}

	// Body content — extract from bodyValues using part IDs.
	if len(e.TextBody) > 0 && e.BodyValues != nil {
		for _, part := range e.TextBody {
			if bv, ok := e.BodyValues[part.PartID]; ok {
				em.TextBody = bv.Value
				break
			}
		}
	}
	if len(e.HTMLBody) > 0 && e.BodyValues != nil {
		for _, part := range e.HTMLBody {
			if bv, ok := e.BodyValues[part.PartID]; ok {
				em.HtmlBody = bv.Value
				break
			}
		}
	}

	// Attachments.
	if len(e.Attachments) > 0 {
		em.Attachments = make([]Attachment, 0, len(e.Attachments))
		for _, a := range e.Attachments {
			em.Attachments = append(em.Attachments, Attachment{
				BlobId:  string(a.BlobID),
				Name:    a.Name,
				Type:    a.Type,
				Size:    int64(a.Size),
				Charset: a.Charset,
			})
		}
	}

	return em
}

// mapAddresses converts go-jmap mail.Address pointers to our Address type.
func mapAddresses(addrs []*mail.Address) []Address {
	if len(addrs) == 0 {
		return nil
	}
	result := make([]Address, len(addrs))
	for i, a := range addrs {
		result[i] = Address{Name: a.Name, Email: a.Email}
	}
	return result
}
