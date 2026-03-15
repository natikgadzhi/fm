# Task 09: `fm search` Command

**Phase:** 2 — Implementation
**Blocked by:** 05, 07
**Blocks:** 13, 14

## Objective

Implement the `fm search <query>` CLI command.

## Acceptance Criteria

- [ ] `cmd/search.go` registers a `search` subcommand on the root command
- [ ] Takes positional args as the search query (joined with spaces)
- [ ] Supports `--limit N` flag (default: 25)
- [ ] Parses query string into JMAP filters using `ParseFilterQuery`
- [ ] Calls `SearchEmails` and formats output with the configured formatter
- [ ] Outputs results to stdout
- [ ] Shows count of results found
- [ ] Handles zero results gracefully with a message
- [ ] Unit tests for:
  - Command registration and flag parsing
  - Query string → filter mapping
  - Output with various formats

## Notes

- Example usage: `fm search "from:boss@company.com subject:urgent after:2025-01-01"`
- Example usage: `fm search meeting notes`
- The search always hits the API (no cache involvement)
