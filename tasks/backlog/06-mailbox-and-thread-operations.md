# Task 06: Mailbox and Thread Operations

**Phase:** 1 — Core
**Blocked by:** 03, 04
**Blocks:** 11, 12

## Objective

Implement Mailbox/get and Thread/get JMAP operations.

## Acceptance Criteria

- [ ] `internal/jmap/mailbox.go` implements:
  - `GetMailboxes(ctx) ([]Mailbox, error)` — fetches all mailboxes
  - `ResolveMailbox(ctx, name string) (string, error)` — resolves a mailbox name to its JMAP ID (for `in:` filter)
- [ ] `internal/jmap/thread.go` implements:
  - `GetThread(ctx, threadId string) (*Thread, error)` — fetches thread metadata
  - `GetThreadEmails(ctx, threadId string) ([]Email, error)` — fetches all emails in a thread (chains Thread/get + Email/get)
- [ ] Mailbox resolution is case-insensitive
- [ ] Thread email fetch uses result references where possible
- [ ] Unit tests with mock responses:
  - List mailboxes with various roles
  - Resolve mailbox by name
  - Resolve non-existent mailbox (error)
  - Fetch thread with multiple emails

## Notes

- Mailbox roles include: inbox, archive, drafts, sent, trash, junk, etc.
- Mailbox resolution should match on name or role
- Thread emails should be sorted by date
