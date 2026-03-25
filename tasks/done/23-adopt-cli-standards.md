# Task 23: Adopt CLI Standards and cli-kit Library

## Objective

Migrate `fm` to use the `cli-kit` shared library and conform to the unified CLI UX standards.

## Status: Done

Merged in PR #26. Key changes:
- Replaced `internal/output` with `cli-kit/output` (table + json formats, TTY auto-detection)
- Replaced `internal/config` with `cli-kit/derived` (-d/--derived flag, new default path)
- Replaced `cmd/version.go` with `cli-kit/version` (always JSON output)
- Replaced custom progress with `cli-kit/progress`
- Updated error handling to use `cli-kit/errors` exit codes
- Updated all tests and e2e tests
- Net reduction of ~1,250 lines of code
