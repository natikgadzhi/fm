# Task 19: Version Command

**Phase:** 5 — Infrastructure
**Blocked by:** none
**Blocks:** none

## Objective

Add `fm version` command that displays version, commit, and build date, following the gdrive-cli pattern.

## Acceptance Criteria

- [ ] Add `Commit` and `Date` variables to `cmd/root.go` (alongside existing `Version`)
- [ ] Create `cmd/version.go` with a `version` subcommand
- [ ] Output follows gdrive-cli pattern: supports `-o json` for structured output
- [ ] Default text output: `fm <version>\n  commit: <hash>\n  built: <date>`
- [ ] JSON output: `{"version": "...", "commit": "...", "date": "..."}`
- [ ] Variables injectable via ldflags at build time
- [ ] Unit test for version command

## Notes

- Follow gdrive-cli `cmd/gdrive-cli/version.go` pattern
- ldflags path: `-X github.com/natikgadzhi/fm/cmd.Version=... -X github.com/natikgadzhi/fm/cmd.Commit=... -X github.com/natikgadzhi/fm/cmd.Date=...`
