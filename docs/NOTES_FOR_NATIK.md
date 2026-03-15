# Notes for Natik

## Status

**Started:** 2026-03-15
**Current phase:** COMPLETE — All tasks done, v0.1.0 releasing

## Progress Log

### 2026-03-15 — Full build session
- Built the entire `fm` CLI from scratch in one session
- 21 tasks completed across 7 waves of parallel agents
- Every PR went through builder → reviewer → fix → merge cycle
- Fixed 2 security vulnerabilities caught in review (path traversal in attachment downloads)
- Fixed dependency vulnerabilities (updated oauth2, removed deprecated protobuf deps)
- CI is now live on GitHub Actions, GoReleaser configured for releases

## Remaining Requests

- [ ] Add `FM_API_TOKEN` to GH repo secrets (Fastmail read-only API token for integration tests)
- [x] ~~Add `HOMEBREW_TAP_GITHUB_TOKEN` to GH repo secrets~~ (done by Natik)
- [ ] Confirm homebrew tap repo name — currently set to `natikgadzhi/homebrew-taps` (matches gdrive-cli)

## How to Release

Releases are automated via GitHub Actions:
1. **Manual dispatch**: Go to Actions → Release → Run workflow → select major/minor/patch
2. **Tag push**: `git tag v0.1.0 && git push origin v0.1.0` triggers GoReleaser
3. GoReleaser builds binaries (linux/darwin, amd64/arm64), creates GitHub Release, and pushes Homebrew formula to `natikgadzhi/homebrew-taps`

## Architecture Decisions

- Commands in `cmd/` package, business logic in `internal/`
- `go-jmap` library for JMAP protocol, wrapped with retry/backoff
- `zalando/go-keyring` for OS keychain (no C bindings)
- Attachment downloads: `{cache-dir}/attachments/{email-id}/{filename}` with `filepath.Base()` sanitization
- Version/Commit/Date injected via ldflags at build time
- Hidden `--endpoint` flag for e2e test server injection

## Review Findings (all addressed)

- Path traversal in attachment downloads → fixed with `filepath.Base()`
- `--with-attachments` silently failing on cache hits → bypass cache when attachments requested
- Thread emails empty body content → reuse shared `emailProperties` from email.go
- Dependency vulnerabilities → updated oauth2, removed deprecated protobuf

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
| 10 - Fetch command | #14 | Merged (review fixes) |
| 11 - Mailboxes command | #12 | Merged |
| 12 - Fetch-thread command | #15 | Merged (review fixes) |
| 13 - Rate limiting & progress | #16 | Merged |
| 14 - End-to-end tests | #17 | Merged |
| 15 - Error handling & UX | #19 | Merged |
| 16 - README & docs | #18 | Merged |
| 17 - Architecture review | #20 | Merged |
| 18 - CI/CD & release | #22 | Merged |
| 19 - Version command | #21 | Merged |
| 20 - Auth check (renamed) | #24 | Merged |
| 21 - Search show IDs | #23 | Merged |
