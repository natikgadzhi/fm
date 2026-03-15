# fm — Implementation Plan

## Architecture Overview

```
fm/
├── cmd/                    # CLI entry points (cobra commands)
│   ├── root.go             # Root command, global flags (-o, --cache-dir, --timeout)
│   ├── search.go           # fm search <query>
│   ├── fetch.go            # fm fetch <message-id>
│   ├── fetch_thread.go     # fm fetch-thread <thread-id>
│   └── mailboxes.go        # fm mailboxes
├── internal/
│   ├── jmap/               # JMAP protocol client
│   │   ├── client.go       # HTTP client, session discovery, auth
│   │   ├── client_test.go
│   │   ├── types.go        # JMAP request/response types
│   │   ├── types_test.go
│   │   ├── email.go        # Email/query, Email/get wrappers
│   │   ├── email_test.go
│   │   ├── mailbox.go      # Mailbox/get wrapper
│   │   ├── mailbox_test.go
│   │   ├── thread.go       # Thread/get wrapper
│   │   └── thread_test.go
│   ├── cache/              # Markdown file cache
│   │   ├── cache.go        # Read/write cached emails
│   │   ├── cache_test.go
│   │   ├── frontmatter.go  # YAML frontmatter parsing/rendering
│   │   └── frontmatter_test.go
│   ├── output/             # Output formatting
│   │   ├── formatter.go    # Interface + factory (text, json, markdown)
│   │   ├── formatter_test.go
│   │   ├── text.go         # Human-readable table output
│   │   ├── json.go         # JSON output
│   │   └── markdown.go     # Markdown output
│   └── config/             # Configuration loading
│       ├── config.go       # Env vars, defaults
│       └── config_test.go
├── go.mod
├── go.sum
├── main.go                 # Entry point
├── CLAUDE.md
├── PROJECT_PROMPT.md
├── PLAN.md
└── README.md
```

## Dependencies

- **CLI framework**: `github.com/spf13/cobra` — standard Go CLI framework
- **JMAP client**: `git.sr.ht/~rockorager/go-jmap` — most complete Go JMAP library, supports Email/query, Email/get, Mailbox/get, Thread/get, and result references
- **YAML frontmatter**: `github.com/adrg/frontmatter` — parse/render YAML frontmatter in Markdown files
- **Table output**: `github.com/olekukonez/tablewriter` or simple fmt-based formatting
- **Progress indicators**: `github.com/schollz/progressbar/v3` or similar

## Phases & Tasks

### Phase 0: Bootstrap

**Task 01 — Bootstrap Go project**
- Initialize Go module (`go mod init github.com/natikgadzhi/fm`)
- Create directory structure: `cmd/`, `internal/jmap/`, `internal/cache/`, `internal/output/`, `internal/config/`
- Add `main.go` with cobra root command
- Add global flags: `-o`/`--output`, `--cache-dir`, `--timeout`
- Verify: `go build ./...` succeeds
- Acceptance: running `fm --help` shows usage with all global flags

**Task 02 — Configuration and auth**
- Implement `internal/config/config.go`: load `FM_API_TOKEN`, `FM_CACHE_DIR`, `FM_OUTPUT` from environment
- Validate token is present, return clear error if missing
- Provide defaults for cache dir (`~/.local/share/fm/cache/`) and output format (`text`)
- Unit tests for config loading with various env combinations

### Phase 1: Core — JMAP Client

**Task 03 — JMAP session and client**
- Implement `internal/jmap/client.go`: create JMAP client using `go-jmap` library
- Session discovery at `https://api.fastmail.com/jmap/session`
- Bearer token authentication
- Extract account ID and API URL from session
- Implement retry with exponential backoff on transient errors (429, 5xx)
- Unit tests with mock HTTP server

**Task 04 — JMAP types and email models**
- Define Go types in `internal/jmap/types.go` for:
  - `Email` (id, threadId, messageId, from, to, cc, subject, date, textBody, htmlBody, preview, mailboxIds)
  - `Mailbox` (id, name, role, totalEmails, unreadEmails)
  - `Thread` (id, emailIds)
  - Search filter types
- Unit tests for JSON serialization/deserialization

**Task 05 — Email query and get**
- Implement `internal/jmap/email.go`:
  - `QueryEmails(filters, limit)` — calls `Email/query`
  - `GetEmails(ids, properties)` — calls `Email/get`
  - `SearchEmails(filters, limit)` — chains `Email/query` + `Email/get` via result references in a single JMAP request
- Parse filter strings (`from:`, `to:`, `subject:`, `in:`, `before:`, `after:`, `has:attachment`, free text)
- Handle pagination
- Unit tests with mock JMAP responses

**Task 06 — Mailbox and thread operations**
- Implement `internal/jmap/mailbox.go`:
  - `GetMailboxes()` — calls `Mailbox/get`, returns all mailboxes
  - Resolve mailbox names to IDs for `in:` filter
- Implement `internal/jmap/thread.go`:
  - `GetThread(threadId)` — calls `Thread/get`
  - `GetThreadEmails(threadId)` — gets thread, then fetches all emails in it
- Unit tests

### Phase 2: Implementation — Commands & Features

**Task 07 — Output formatters**
- Implement `internal/output/formatter.go`: `Formatter` interface with `FormatEmails`, `FormatMailboxes`, `FormatEmail` methods
- Implement `text.go`: human-readable table output for search results, single email display
- Implement `json.go`: structured JSON output
- Implement `markdown.go`: Markdown formatted output
- Factory function to create formatter from format string
- Unit tests for each formatter

**Task 08 — Markdown cache**
- Implement `internal/cache/frontmatter.go`: YAML frontmatter struct matching PROJECT_PROMPT.md spec
- Implement `internal/cache/cache.go`:
  - `Get(id)` — read cached email by ID, return nil if not found
  - `Put(email)` — write email as Markdown with frontmatter
  - `CacheDir()` — resolve cache directory (flag > env > default)
- Render email body as Markdown with headers
- Unit tests using temp directories

**Task 09 — `fm search` command**
- Implement `cmd/search.go`: parse query string, call JMAP search, format output
- Parse filter syntax from query args
- Support `--limit` flag
- Display results in requested output format
- Show progress indicator for long searches
- Unit tests for query parsing

**Task 10 — `fm fetch` command**
- Implement `cmd/fetch.go`: fetch single email by ID
- Check cache first (unless `--no-cache`)
- Fetch from JMAP if not cached
- Save to cache
- Print in requested output format
- Unit tests

**Task 11 — `fm mailboxes` command**
- Implement `cmd/mailboxes.go`: list all mailboxes
- Format as table (name, role, unread, total)
- Support `-o json` and `-o markdown`
- Unit tests

**Task 12 — `fm fetch-thread` command**
- Implement `cmd/fetch_thread.go`: fetch all emails in a thread
- Fetch thread info, then all emails
- Cache each email individually
- Display as threaded conversation or individual messages
- Unit tests

### Phase 3: Integration

**Task 13 — Rate limiting and progress**
- Implement rate limit handling across all commands:
  - Detect HTTP 429, parse `Retry-After` header
  - Exponential backoff with jitter
  - On 429 mid-batch, return partial results + warning
- Add progress indicators for multi-email fetches
- Integration tests with rate limit simulation

**Task 14 — End-to-end tests**
- Create e2e test suite that runs against a mock JMAP server
- Test full command flows: search → fetch → cache hit
- Test error scenarios: bad token, network error, rate limit
- Test output format correctness for all formats
- Test cache read/write cycle

### Phase 4: Polish

**Task 15 — Error handling and UX polish**
- Review all error paths for clear, actionable messages
- Add `--verbose` / `--debug` flag for detailed output
- Ensure graceful handling of Ctrl+C during long operations
- Validate all user inputs at CLI boundaries

**Task 16 — README and documentation**
- Write README.md with:
  - Installation instructions (go install, binary release)
  - Quick start guide
  - Command reference
  - Configuration reference
  - Examples

## Task Dependency Graph

```
Phase 0:  [01-bootstrap] ──→ [02-config]
              │                    │
Phase 1:      └──→ [03-client] ←──┘
                      │
              ┌───────┼───────┐
              ↓       ↓       ↓
          [04-types] [05-email] [06-mailbox]
              │       │         │
Phase 2:      ↓       ↓         ↓
          [07-output] │    [08-cache]
              │       │         │
              ↓       ↓         ↓
          [09-search] [10-fetch] [11-mailboxes] [12-fetch-thread]
              │         │           │               │
Phase 3:      └─────────┴───────────┴───────────────┘
                              │
                        [13-ratelimit]
                              │
                        [14-e2e-tests]
                              │
Phase 4:              [15-error-handling]
                              │
                        [16-readme]
```

## Key Design Decisions

1. **go-jmap library** — Use `git.sr.ht/~rockorager/go-jmap` as the JMAP client foundation. It's the most actively maintained Go JMAP library and supports result references for chaining `Email/query` + `Email/get` in a single HTTP request.

2. **Cobra for CLI** — Standard Go CLI framework. Provides subcommands, flags, help generation out of the box.

3. **Read-only by design** — No methods that modify server state. The JMAP client should only import capabilities needed for reading.

4. **Cache as Markdown** — Emails cached as `.md` files with YAML frontmatter. This makes them grep-able, readable in any editor, and useful as context for other tools.

5. **Filter parsing** — Parse Gmail-style filter syntax (`from:`, `to:`, `subject:`, etc.) into JMAP `FilterCondition` objects. Free text maps to JMAP's `text` filter.

6. **Partial results on rate limit** — If a batch fetch hits 429 mid-way, return everything fetched so far along with a warning, rather than failing the entire operation.
