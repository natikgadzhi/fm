# Task 01: Bootstrap Go Project

**Phase:** 0 — Bootstrap
**Blocked by:** none
**Blocks:** 02, 03

## Objective

Initialize the Go project with the full directory structure, cobra root command, and global flags.

## Acceptance Criteria

- [ ] `go.mod` exists with module `github.com/natikgadzhi/fm`
- [ ] Directory structure created: `cmd/`, `internal/jmap/`, `internal/cache/`, `internal/output/`, `internal/config/`
- [ ] `main.go` exists and calls cobra root command
- [ ] `cmd/root.go` defines the root command with:
  - `-o` / `--output` flag (values: `text`, `json`, `markdown`; default: `text`)
  - `--cache-dir` flag (default: empty, resolved later from config)
  - `--timeout` flag (default: `30s`)
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] Running `fm --help` prints usage with all global flags

## Notes

- Use `github.com/spf13/cobra` for CLI framework
- Keep `main.go` minimal — just call `cmd.Execute()`
- Do not add subcommands yet — those come in later tasks
