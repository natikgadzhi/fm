# Notes for Natik

## Status

**Started:** 2026-03-15
**Current phase:** Wave 6 — Rate limiting, e2e tests, polish, docs

## Progress Log

### 2026-03-15 — Session Start
- Reviewed PROJECT_PROMPT.md, PLAN.md, all 16 task specs
- Studied slack-cli reference implementation at `../scripts/slack-cli/` for patterns
- Updated task specs per your requirements:
  - Task 09 (search): Added `--from`, `--to`, `--has-attachments` CLI flags
  - Task 10 (fetch): Added `--with-attachments` flag for downloading attachment blobs
  - Task 12 (fetch-thread): Added `--with-attachments` flag
  - Task 04 (types): Added `Attachment` struct
  - Task 05 (email ops): Added `DownloadAttachment` method, attachment properties
- Starting Wave 1: Tasks 01 (bootstrap) → 02 (config/auth)
- Then Wave 2: Tasks 03, 04 in parallel (JMAP client + types)
- Then Wave 3: Tasks 05, 06, 07, 08 in parallel (email ops, mailbox/thread, formatters, cache)
- Then Wave 4: Tasks 09, 10, 11, 12 in parallel (all CLI commands)
- Then Wave 5: Tasks 13, 14, 15, 16 (rate limiting, e2e tests, polish, docs)

## Questions for Natik

1. **Integration test secret**: When you add the Fastmail test account token to GH secrets, please use env var name `FM_API_TOKEN`. The e2e tests will look for this.

2. **Homebrew tap**: I'll set up the release workflow (GoReleaser + GitHub Actions). Will need `HOMEBREW_TAP_GITHUB_TOKEN` in repo secrets for the tap push. The tap repo should be `natikgadzhi/homebrew-tap` (or let me know if you want a different name).

3. **Test account**: For integration tests, the test account should have a few emails in the inbox. Ideally some with attachments, some threads with 2+ messages, and a few different mailboxes/labels.

## Requests

- [ ] Add `FM_API_TOKEN` to GH repo secrets (Fastmail read-only API token for test account)
- [ ] Add `HOMEBREW_TAP_GITHUB_TOKEN` to GH repo secrets
- [ ] Confirm homebrew tap repo name (I'll assume `natikgadzhi/homebrew-tap`)

## Architecture Decisions Made

- Following slack-cli patterns: minimal entry point, commands in `internal/commands/` (deviation from PLAN.md's `cmd/` — kept as `cmd/` per the plan since it's already established)
- Using `go-jmap` library as planned for JMAP protocol
- Using `zalando/go-keyring` for keychain (no C bindings, good for static binary)
- Attachment downloads saved to `{cache-dir}/attachments/{email-id}/{filename}`
- Version injected via ldflags at build time
- Makefile for build, test, vet targets

## Review Findings (addressed)

- **PR #14 (fetch)**: Path traversal vulnerability in attachment downloads — fixed with `filepath.Base()`
- **PR #14 (fetch)**: `--with-attachments` silently failed on cache hits — fixed by bypassing cache when attachments requested
- **PR #15 (fetch-thread)**: Same path traversal vulnerability — fixed
- **PR #15 (fetch-thread)**: Thread emails had empty body content (missing properties in Email/get) — fixed by reusing shared `emailProperties`
- **PR #15 (fetch-thread)**: Variable naming conflict with PR #14 — renamed to `threadWithAttachments`
- **GitHub Dependabot**: 9 dependency vulnerabilities flagged (3 high, 6 moderate) — will address in architecture review

## Completed Tasks

| Task | PR | Status |
|------|-----|--------|
| 01 - Bootstrap Go project | #2 | Merged |
| 02 - Configuration and auth | #3 | Merged |
| 03 - JMAP session and client | #5 | Merged |
| 04 - JMAP types and email models | #4 | Merged |
| 05 - Email query and get | #10 | Merged |
| 06 - Mailbox and thread operations | #11 | Merged |
| 07 - Output formatters | #8 | Merged |
| 08 - Markdown cache | #9 | Merged |
| 09 - Search command | #13 | Merged |
| 10 - Fetch command | #14 | Merged (review fixes applied) |
| 11 - Mailboxes command | #12 | Merged |
| 12 - Fetch-thread command | #15 | Merged (review fixes applied) |
