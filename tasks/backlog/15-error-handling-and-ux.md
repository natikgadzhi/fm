# Task 15: Error Handling and UX Polish

**Phase:** 4 — Polish
**Blocked by:** 14
**Blocks:** 16

## Objective

Review and improve error handling, add debug mode, and polish the user experience.

## Acceptance Criteria

- [ ] All error messages are clear and actionable:
  - Auth errors: "FM_API_TOKEN not set. Get an API token at https://www.fastmail.com/settings/security/tokens"
  - Network errors: "Failed to connect to Fastmail API. Check your internet connection."
  - Not found: "Email M123 not found. Verify the message ID is correct."
- [ ] `--verbose` flag adds detailed logging to stderr:
  - JMAP requests/responses (redacting tokens)
  - Cache hits/misses
  - Retry attempts
- [ ] Graceful Ctrl+C handling:
  - Cancel in-flight requests
  - Return partial results if available
- [ ] Input validation:
  - Validate email ID format
  - Validate output format values
  - Validate date formats in filters
- [ ] Exit codes: 0 success, 1 general error, 2 auth error

## Notes

- Use `context.WithCancel` for Ctrl+C handling
- Log to stderr so stdout remains clean for piping
- Never log the actual API token value
