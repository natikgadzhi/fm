# Task 12: `fm fetch-thread` Command

**Phase:** 2 — Implementation
**Blocked by:** 05, 06, 07, 08
**Blocks:** 14

## Objective

Implement the `fm fetch-thread <thread-id>` CLI command.

## Acceptance Criteria

- [ ] `cmd/fetch_thread.go` registers a `fetch-thread` subcommand on the root command
- [ ] Takes a single positional arg: the JMAP thread ID
- [ ] Supports `--with-attachments` flag to download attachments for all emails in the thread (same behavior as `fm fetch --with-attachments`)
- [ ] Fetches thread metadata, then all emails in the thread
- [ ] Caches each email individually
- [ ] Displays all emails in chronological order
- [ ] Shows thread summary (number of messages, participants)
- [ ] Supports all output formats
- [ ] Unit tests for:
  - Single-message thread
  - Multi-message thread
  - Cache interaction (some emails cached, some not)

## Notes

- Thread ID format: `Txxxxxxxx`
- Each email in the thread is cached separately as its own `.md` file
- The conversation view in text format should show clear separators between messages
