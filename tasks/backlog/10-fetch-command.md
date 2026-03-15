# Task 10: `fm fetch` Command

**Phase:** 2 — Implementation
**Blocked by:** 05, 07, 08
**Blocks:** 13, 14

## Objective

Implement the `fm fetch <message-id>` CLI command.

## Acceptance Criteria

- [ ] `cmd/fetch.go` registers a `fetch` subcommand on the root command
- [ ] Takes a single positional arg: the JMAP email ID
- [ ] Supports `--no-cache` flag to bypass cache
- [ ] Default behavior:
  1. Check cache for the email
  2. If cached, read from cache and display
  3. If not cached, fetch from JMAP API
  4. Save to cache
  5. Display in requested output format
- [ ] Prints the full email (headers + body) to stdout
- [ ] Unit tests for:
  - Cache hit path
  - Cache miss → fetch → cache write path
  - `--no-cache` flag behavior

## Notes

- The message ID is the JMAP `Email` id (e.g., `Mxxxxxxxx`)
- When fetching, save to cache as a side effect
- Error clearly if the ID doesn't exist on the server
