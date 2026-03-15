# Task 02: Configuration and Auth

**Phase:** 0 — Bootstrap
**Blocked by:** 01
**Blocks:** 03

## Objective

Implement configuration loading from environment variables with sensible defaults.

## Acceptance Criteria

- [ ] `internal/config/config.go` implements a `Config` struct with fields:
  - `APIToken string` — from `FM_API_TOKEN` (required)
  - `CacheDir string` — from `FM_CACHE_DIR` (default: `~/.local/share/fm/cache/`)
  - `OutputFormat string` — from `FM_OUTPUT` (default: `text`)
- [ ] `Load()` function reads env vars and applies defaults
- [ ] Returns clear error message if `FM_API_TOKEN` is not set
- [ ] `CacheDir` expands `~` to user home directory
- [ ] CLI flag overrides (from root command) take precedence over env vars
- [ ] Unit tests cover:
  - Loading with all env vars set
  - Loading with defaults (no optional env vars)
  - Missing token error
  - Home directory expansion

## Notes

- Do not implement keychain/credential store — just env var for now
- Config should be passed down to commands via cobra's context or a shared package-level variable
