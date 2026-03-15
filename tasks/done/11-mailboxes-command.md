# Task 11: `fm mailboxes` Command

**Phase:** 2 — Implementation
**Blocked by:** 06, 07
**Blocks:** 14

## Objective

Implement the `fm mailboxes` CLI command.

## Acceptance Criteria

- [ ] `cmd/mailboxes.go` registers a `mailboxes` subcommand on the root command
- [ ] No positional arguments
- [ ] Calls `GetMailboxes` and formats output with the configured formatter
- [ ] Default text output shows a table: Name, Role, Unread, Total
- [ ] Mailboxes sorted alphabetically by name
- [ ] Supports `-o json` and `-o markdown`
- [ ] Unit tests for command registration and output formatting

## Notes

- This is the simplest command — good for verifying end-to-end flow
- Mailbox roles may be empty (user-created folders have no role)
