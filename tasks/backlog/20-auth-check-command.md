# Task 20: Rename `fm auth status` to `fm auth check` with API Verification

**Phase:** 5 — Refinement
**Blocked by:** none
**Blocks:** none

## Objective

Rename `fm auth status` to `fm auth check` and make it actually verify the token works by making a JMAP session request.

## Acceptance Criteria

- [ ] Rename `authStatusCmd` from `status` to `check` in `cmd/auth.go`
- [ ] `fm auth check` resolves the token, then calls `client.Discover()` to verify it works
- [ ] On success: print token source, masked token, authenticated account ID, and username from JMAP session
- [ ] On failure: print clear error ("Authentication failed. Your API token may be revoked or invalid.")
- [ ] Update README.md to reflect the renamed command
- [ ] Update tests

## Notes

- The JMAP session response includes account info (username, account ID) that should be displayed
- This makes `fm auth check` a real health check, not just "do I have a token somewhere"
