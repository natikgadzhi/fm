package jmap

import (
	"context"
	"fmt"
	"sort"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/thread"
)

// GetThread fetches a thread by its ID using Thread/get.
func (c *Client) GetThread(ctx context.Context, threadId string) (*Thread, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery: %w", err)
	}

	accountID, err := c.PrimaryAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting account ID: %w", err)
	}

	req := &gojmap.Request{
		Context: ctx,
	}
	req.Invoke(&thread.Get{
		Account: accountID,
		IDs:     []gojmap.ID{gojmap.ID(threadId)},
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Thread/get request failed: %w", err)
	}

	if len(resp.Responses) == 0 {
		return nil, fmt.Errorf("Thread/get: empty response")
	}

	getResp, ok := resp.Responses[0].Args.(*thread.GetResponse)
	if !ok {
		return nil, fmt.Errorf("Thread/get: unexpected response type %T", resp.Responses[0].Args)
	}

	if len(getResp.NotFound) > 0 {
		return nil, fmt.Errorf("thread %q not found", threadId)
	}

	if len(getResp.List) == 0 {
		return nil, fmt.Errorf("thread %q not found", threadId)
	}

	t := getResp.List[0]
	return threadFromJMAP(t), nil
}

// GetThreadEmails fetches a thread and all its emails, sorted by date ascending.
// It uses a single JMAP request with result references to chain Thread/get and Email/get.
func (c *Client) GetThreadEmails(ctx context.Context, threadId string) ([]Email, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery: %w", err)
	}

	accountID, err := c.PrimaryAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting account ID: %w", err)
	}

	req := &gojmap.Request{
		Context: ctx,
	}

	// First call: Thread/get to retrieve email IDs.
	threadCallID := req.Invoke(&thread.Get{
		Account: accountID,
		IDs:     []gojmap.ID{gojmap.ID(threadId)},
	})

	// Second call: Email/get with a result reference to the thread's email IDs.
	req.Invoke(&email.Get{
		Account: accountID,
		ReferenceIDs: &gojmap.ResultReference{
			ResultOf: threadCallID,
			Name:     "Thread/get",
			Path:     "/list/*/emailIds",
		},
		Properties:         emailProperties,
		FetchAllBodyValues: true,
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Thread+Email/get request failed: %w", err)
	}

	if len(resp.Responses) < 2 {
		return nil, fmt.Errorf("GetThreadEmails: expected 2 responses, got %d", len(resp.Responses))
	}

	// Check the thread response for not-found.
	threadResp, ok := resp.Responses[0].Args.(*thread.GetResponse)
	if !ok {
		return nil, fmt.Errorf("Thread/get: unexpected response type %T", resp.Responses[0].Args)
	}
	if len(threadResp.NotFound) > 0 {
		return nil, fmt.Errorf("thread %q not found", threadId)
	}
	if len(threadResp.List) == 0 {
		return nil, fmt.Errorf("thread %q not found", threadId)
	}

	// Parse the email response.
	emailResp, ok := resp.Responses[1].Args.(*email.GetResponse)
	if !ok {
		return nil, fmt.Errorf("Email/get: unexpected response type %T", resp.Responses[1].Args)
	}

	emails := make([]Email, 0, len(emailResp.List))
	for _, e := range emailResp.List {
		emails = append(emails, emailFromJMAP(e))
	}

	// Sort by date ascending.
	sort.Slice(emails, func(i, j int) bool {
		return emails[i].Date.Before(emails[j].Date)
	})

	return emails, nil
}

// threadFromJMAP converts a go-jmap Thread to our domain Thread type.
func threadFromJMAP(t *thread.Thread) *Thread {
	emailIds := make([]string, len(t.EmailIDs))
	for i, id := range t.EmailIDs {
		emailIds[i] = string(id)
	}
	return &Thread{
		Id:       string(t.ID),
		EmailIds: emailIds,
	}
}

// emailFromJMAP converts a go-jmap Email to our domain Email type.
func emailFromJMAP(e *email.Email) Email {
	from := make([]Address, 0, len(e.From))
	for _, a := range e.From {
		from = append(from, Address{Name: a.Name, Email: a.Email})
	}

	to := make([]Address, 0, len(e.To))
	for _, a := range e.To {
		to = append(to, Address{Name: a.Name, Email: a.Email})
	}

	cc := make([]Address, 0, len(e.CC))
	for _, a := range e.CC {
		cc = append(cc, Address{Name: a.Name, Email: a.Email})
	}

	mailboxIds := make(map[string]bool, len(e.MailboxIDs))
	for id, v := range e.MailboxIDs {
		mailboxIds[string(id)] = v
	}

	// Extract text body from body values if available.
	var textBody string
	if len(e.TextBody) > 0 && e.BodyValues != nil {
		if bv, ok := e.BodyValues[e.TextBody[0].PartID]; ok {
			textBody = bv.Value
		}
	}

	// Extract HTML body from body values if available.
	var htmlBody string
	if len(e.HTMLBody) > 0 && e.BodyValues != nil {
		if bv, ok := e.BodyValues[e.HTMLBody[0].PartID]; ok {
			htmlBody = bv.Value
		}
	}

	var messageId string
	if len(e.MessageID) > 0 {
		messageId = e.MessageID[0]
	}

	attachments := make([]Attachment, 0, len(e.Attachments))
	for _, att := range e.Attachments {
		attachments = append(attachments, Attachment{
			BlobId:  string(att.BlobID),
			Name:    att.Name,
			Type:    att.Type,
			Size:    int64(att.Size),
			Charset: att.Charset,
		})
	}

	var date = e.SentAt
	if date == nil {
		date = e.ReceivedAt
	}

	result := Email{
		Id:            string(e.ID),
		ThreadId:      string(e.ThreadID),
		MessageId:     messageId,
		From:          from,
		To:            to,
		Cc:            cc,
		Subject:       e.Subject,
		TextBody:      textBody,
		HtmlBody:      htmlBody,
		Preview:       e.Preview,
		MailboxIds:    mailboxIds,
		Size:          int64(e.Size),
		HasAttachment: e.HasAttachment,
		Attachments:   attachments,
	}

	if date != nil {
		result.Date = *date
	}

	return result
}
