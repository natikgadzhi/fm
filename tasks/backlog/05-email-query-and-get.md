# Task 05: Email Query and Get

**Phase:** 1 — Core
**Blocked by:** 03, 04
**Blocks:** 09, 10, 12

## Objective

Implement Email/query and Email/get JMAP operations, including chained requests via result references.

## Acceptance Criteria

- [ ] `internal/jmap/email.go` implements:
  - `QueryEmails(ctx, filter SearchFilter, limit int) ([]string, error)` — returns email IDs matching filter
  - `GetEmails(ctx, ids []string) ([]Email, error)` — fetches full email objects by ID
  - `SearchEmails(ctx, filter SearchFilter, limit int) ([]Email, error)` — chains query+get in a single JMAP request using result references
- [ ] Result references work correctly — `Email/get` references `Email/query` result IDs
- [ ] Handles pagination when results exceed limit
- [ ] Properties requested: Id, ThreadId, MessageId, From, To, Cc, Subject, Date, Preview, TextBody, HtmlBody, MailboxIds
- [ ] Unit tests with mock JMAP responses:
  - Search returning multiple results
  - Search returning zero results
  - Get by specific IDs
  - Chained query+get via result references

## Notes

- The `go-jmap` library supports result references — use its API for chaining
- For `SearchEmails`, build both method calls and submit in a single `Do()` request
- TextBody and HtmlBody use JMAP's body value structure — may need special handling
