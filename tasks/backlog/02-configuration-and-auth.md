# Task 02: Configuration and Auth

**Phase:** 0 — Bootstrap
**Blocked by:** 01
**Blocks:** 03

## Objective

Implement configuration loading, API token resolution with a 3-source priority chain (flag → env → OS keychain), and `fm auth` subcommands.

## Acceptance Criteria

### Config (`internal/config/config.go`)

- [ ] `Config` struct with fields:
  - `CacheDir string` — from `FM_CACHE_DIR` (default: `~/.local/share/fm/cache/`)
  - `OutputFormat string` — from `FM_OUTPUT` (default: `text`)
- [ ] `Load()` function reads env vars and applies defaults
- [ ] `CacheDir` expands `~` to user home directory
- [ ] CLI flag overrides take precedence over env vars

### Token Resolution (`internal/auth/auth.go`)

- [ ] `ResolveToken(flagValue string) (token string, source string, err error)` function with priority:
  1. `flagValue` (from `--token` flag) — highest priority
  2. `FM_API_TOKEN` environment variable
  3. OS keychain via `github.com/zalando/go-keyring` (service: `"fm"`, key: `"api-token"`)
- [ ] Returns `(token, source, nil)` where source is `"flag"`, `"environment"`, or `"keychain"`
- [ ] If no token found, returns error: `"No API token found. Run 'fm auth login' or set FM_API_TOKEN. Create a token at https://app.fastmail.com/settings/security/tokens/new"`
- [ ] `StoreToken(token string) error` — saves token to OS keychain
- [ ] `DeleteToken() error` — removes token from OS keychain

### Auth Commands (`cmd/auth.go`)

- [ ] `fm auth login` — prompts for API token (reads from stdin), validates format (starts with `fmu1-`), stores in OS keychain, confirms success
- [ ] `fm auth status` — resolves token, makes a JMAP session request to verify it works, prints:
  - Authenticated user (from session response)
  - Token source (flag / environment / keychain)
  - Account ID
- [ ] `fm auth logout` — removes token from OS keychain, confirms removal

### Tests

- [ ] Unit tests for `ResolveToken` using `keyring.MockInit()`:
  - Token from flag takes precedence over env and keychain
  - Token from env takes precedence over keychain
  - Token from keychain when no flag or env
  - Error when no token available anywhere
- [ ] Unit tests for `StoreToken` and `DeleteToken` with mock keyring
- [ ] Unit tests for config loading with various env combinations

## Notes

- Use `github.com/zalando/go-keyring` — no C bindings, works with static binaries
- `keyring.MockInit()` replaces the OS keychain with an in-memory store for tests
- The `fm auth login` prompt should mask input or at minimum not echo it
- Token format validation is a soft check (warn if it doesn't start with `fmu1-`, don't block)
- Config should be passed down to commands via cobra's persistent pre-run or a shared package-level variable
