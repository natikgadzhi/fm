# Task 04: JMAP Types and Email Models

**Phase:** 1 — Core
**Blocked by:** 03
**Blocks:** 05, 06, 07, 08

## Objective

Define Go types for emails, mailboxes, threads, and search filters.

## Acceptance Criteria

- [ ] `internal/jmap/types.go` defines:
  - `Email` struct: Id, ThreadId, MessageId, From ([]Address), To ([]Address), Cc ([]Address), Subject, Date, TextBody, HtmlBody, Preview, MailboxIds, Size
  - `Address` struct: Name, Email
  - `Mailbox` struct: Id, Name, Role, TotalEmails, UnreadEmails, ParentId
  - `Thread` struct: Id, EmailIds
  - `SearchFilter` struct: From, To, Subject, Text, InMailbox, Before, After, HasAttachment
- [ ] Types map correctly to/from JMAP JSON (using go-jmap's type system if applicable)
- [ ] `ParseFilterQuery(query string) SearchFilter` — parses `from:foo to:bar subject:hello some text` into a `SearchFilter`
- [ ] Unit tests for:
  - JSON round-trip serialization
  - Filter query parsing with various combinations
  - Edge cases: empty query, only free text, multiple of same filter

## Notes

- Check if `go-jmap` already defines mail types we can reuse or embed
- The filter parser is important — it maps user-friendly syntax to JMAP FilterCondition
- Filter syntax: `from:`, `to:`, `subject:`, `in:`, `before:YYYY-MM-DD`, `after:YYYY-MM-DD`, `has:attachment`, everything else is free text
