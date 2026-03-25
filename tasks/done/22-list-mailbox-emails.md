# Task 22: Implement `fm list` command

## Objective

Add a `fm list <mailbox>` command that fetches emails in a given mailbox (by name or ID) in reverse-chronological order. Primary use-case: grab the 10-20 most recent emails in INBOX for summarization workflows.

## Acceptance Criteria

1. `fm list INBOX` returns emails from INBOX, newest first
2. `fm list Sent` works with mailbox names (case-insensitive)
3. `fm list <mailbox-id>` works with raw JMAP mailbox IDs
4. `--limit N` flag controls how many emails to return (default: 20)
5. Supports all output formats (`-o text`, `-o json`, `-o markdown`)
6. Prints count to stderr like `search` does
7. Handles partial results gracefully (like search/fetch-thread do)
8. Has unit tests for the command

## Implementation Notes

- Reuse `SearchEmails` with a `SearchFilter{InMailbox: resolvedID}` — it already sorts by `receivedAt` descending
- Use `ResolveMailbox` to convert name → ID, but also accept raw IDs directly
- Follow the same patterns as `search.go` and `fetch_thread.go`
- New file: `cmd/list.go` and `cmd/list_test.go`

## Dependencies

None — all JMAP client methods already exist.
