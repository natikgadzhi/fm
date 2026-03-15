# Task 14: End-to-End Tests

**Phase:** 3 — Integration
**Blocked by:** 09, 10, 11, 12, 13
**Blocks:** 15

## Objective

Create a comprehensive end-to-end test suite using a mock JMAP server.

## Acceptance Criteria

- [ ] Mock JMAP server that simulates Fastmail's API:
  - Session discovery endpoint
  - JMAP API endpoint handling Email/query, Email/get, Mailbox/get, Thread/get
  - Configurable responses, error injection, rate limit simulation
- [ ] E2E tests covering full command flows:
  - `fm search` → results displayed correctly
  - `fm fetch` → email cached → subsequent fetch uses cache
  - `fm mailboxes` → all mailboxes listed
  - `fm fetch-thread` → all thread emails fetched and cached
- [ ] Error scenario tests:
  - No token anywhere (no flag, no env, no keychain) → actionable error with setup instructions
  - Invalid API token → clear auth failure message
  - Network timeout → retry + error
  - Rate limited → backoff + partial results
  - Invalid email ID → clear error
- [ ] Auth flow tests:
  - `fm auth login` → token stored in keychain → subsequent commands use it
  - `fm auth status` → shows correct token source
  - `fm auth logout` → token removed → commands fail with auth error
- [ ] Output format tests:
  - Each command with `-o text`, `-o json`, `-o markdown`
  - Verify JSON is valid, Markdown is well-formed
- [ ] Cache integration tests:
  - Fetch → verify cache file exists with correct frontmatter
  - Fetch with `--no-cache` → always hits API
  - Cache file corruption → graceful fallback to API

## Notes

- Use `httptest.NewServer` for the mock JMAP server
- Tests should be runnable without any Fastmail credentials
- Consider using `testify` for assertions if the team prefers, but stdlib is fine
