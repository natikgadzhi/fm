# Task 03: JMAP Session and Client

**Phase:** 1 — Core
**Blocked by:** 01, 02
**Blocks:** 04, 05, 06

## Objective

Implement the core JMAP client that handles session discovery, authentication, and request execution.

## Acceptance Criteria

- [ ] `internal/jmap/client.go` implements a `Client` struct with:
  - `NewClient(token string, opts ...Option)` constructor (token comes from `auth.ResolveToken`)
  - `Discover()` method — fetches JMAP session from `https://api.fastmail.com/jmap/session`
  - `Do(request)` method — executes a JMAP request and returns the response
  - Account ID extraction from session
- [ ] Bearer token authentication via `Authorization` header
- [ ] Retry logic with exponential backoff:
  - Retry on HTTP 429 (parse `Retry-After` header)
  - Retry on HTTP 5xx
  - Max 5 retries with jitter
- [ ] Configurable timeout
- [ ] Unit tests with httptest mock server covering:
  - Successful session discovery
  - Auth failure (401)
  - Retry on 429
  - Retry on 500
  - Timeout

## Notes

- Use `git.sr.ht/~rockorager/go-jmap` as the foundation
- The go-jmap library handles much of the JMAP protocol details — wrap it with our retry/backoff logic
- Store the discovered session for reuse within the client's lifetime
- Token is resolved by `internal/auth` before constructing the client — this package only needs the token string
- `fm auth status` will use `Discover()` to validate the token works
