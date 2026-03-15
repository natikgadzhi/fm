package jmap

import (
	"context"
	"fmt"
	"strings"

	gojmap "git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
)

// GetMailboxes fetches all mailboxes from the server using Mailbox/get.
func (c *Client) GetMailboxes(ctx context.Context) ([]Mailbox, error) {
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
	req.Invoke(&mailbox.Get{
		Account: accountID,
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Mailbox/get request failed: %w", err)
	}

	if len(resp.Responses) == 0 {
		return nil, fmt.Errorf("Mailbox/get: empty response")
	}

	getResp, ok := resp.Responses[0].Args.(*mailbox.GetResponse)
	if !ok {
		return nil, fmt.Errorf("Mailbox/get: unexpected response type %T", resp.Responses[0].Args)
	}

	result := make([]Mailbox, 0, len(getResp.List))
	for _, m := range getResp.List {
		result = append(result, mailboxFromJMAP(m))
	}

	return result, nil
}

// ResolveMailbox resolves a mailbox name or role to its JMAP ID.
// The match is case-insensitive and checks both the Name and Role fields.
func (c *Client) ResolveMailbox(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("mailbox name must not be empty")
	}

	mailboxes, err := c.GetMailboxes(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving mailbox %q: %w", name, err)
	}

	lower := strings.ToLower(name)
	for _, m := range mailboxes {
		if strings.ToLower(m.Name) == lower || strings.ToLower(m.Role) == lower {
			return m.Id, nil
		}
	}

	return "", fmt.Errorf("mailbox %q not found", name)
}

// mailboxFromJMAP converts a go-jmap Mailbox to our domain Mailbox type.
func mailboxFromJMAP(m *mailbox.Mailbox) Mailbox {
	return Mailbox{
		Id:           string(m.ID),
		Name:         m.Name,
		Role:         string(m.Role),
		TotalEmails:  int(m.TotalEmails),
		UnreadEmails: int(m.UnreadEmails),
		ParentId:     string(m.ParentID),
	}
}
