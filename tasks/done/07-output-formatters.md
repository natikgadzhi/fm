# Task 07: Output Formatters

**Phase:** 2 — Implementation
**Blocked by:** 04
**Blocks:** 09, 10, 11, 12

## Objective

Implement output formatters for text, JSON, and Markdown formats.

## Acceptance Criteria

- [ ] `internal/output/formatter.go` defines:
  - `Formatter` interface with methods:
    - `FormatEmailList(emails []jmap.Email) (string, error)`
    - `FormatEmail(email jmap.Email) (string, error)`
    - `FormatMailboxes(mailboxes []jmap.Mailbox) (string, error)`
  - `New(format string) (Formatter, error)` — factory function
- [ ] `internal/output/text.go` — `TextFormatter`:
  - Email list: tabular format with Date, From, Subject, Preview columns
  - Single email: headers + body as plain text
  - Mailboxes: tabular format with Name, Role, Unread, Total columns
- [ ] `internal/output/json.go` — `JSONFormatter`:
  - Pretty-printed JSON output for all types
- [ ] `internal/output/markdown.go` — `MarkdownFormatter`:
  - Email list: Markdown table
  - Single email: Markdown with headers as bold, body as content
  - Mailboxes: Markdown table
- [ ] Unit tests for each formatter with sample data

## Notes

- Text formatter should handle long subjects/names gracefully (truncate with `...`)
- JSON formatter should use `json.MarshalIndent` for readability
- Date formatting: use a human-friendly format like `Jan 02, 2006 3:04 PM`
