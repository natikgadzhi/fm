# fm — Fastmail CLI

A read-only CLI tool for searching and fetching email from Fastmail via the JMAP API, with local Markdown caching.

## Overview

`fm` is a command-line tool that connects to Fastmail using the JMAP protocol (RFC 8620 / RFC 8621) to search, list, and fetch emails. It saves messages locally as Markdown files with YAML frontmatter for metadata. The tool is read-only — it never modifies, sends, or deletes emails.

## Authentication

- Fastmail API tokens are used for authentication
- The JMAP session endpoint is `https://api.fastmail.com/jmap/session`
- Tokens are passed via the `Authorization: Bearer {token}` header
- Token storage: support `FM_API_TOKEN` environment variable, and optionally macOS Keychain / system credential store
- The session response provides the API URL, account IDs, and capability information

## Core Commands

### `fm search <query>`
Search emails using JMAP `Email/query` with filters.
- Support common filters: `from:`, `to:`, `subject:`, `in:` (mailbox), `before:`, `after:`, `has:attachment`
- Free-text search for body/subject content
- Return results as a summary table (date, from, subject, snippet)
- Support `-o json` and `-o markdown` output formats
- Support `--limit N` to control result count (default 25)

### `fm fetch <message-id>`
Fetch a single email by its JMAP message ID.
- Download the full message (headers, text body, HTML body)
- Save as a Markdown file in the cache directory
- Print the message to stdout in the requested format

### `fm mailboxes`
List all mailboxes (folders) in the account.
- Show mailbox name, role, unread count, total count
- Support `-o json` and `-o markdown` output formats

### `fm fetch-thread <thread-id>`
Fetch all emails in a thread.
- Download all messages in the thread
- Save each as a separate Markdown file
- Optionally render as a single threaded conversation

## Output Formats

All commands support `-o <format>` flag:
- `text` (default) — human-readable table/text output
- `json` — structured JSON output
- `markdown` — Markdown formatted output

## Markdown Cache

Fetched emails are cached locally as Markdown files with YAML frontmatter:

```markdown
---
tool: fm
object: email
id: "M1234567890"
thread_id: "T1234567890"
message_id: "<abc@fastmail.com>"
from: "sender@example.com"
to: ["recipient@example.com"]
subject: "Meeting tomorrow"
date: "2025-01-15T10:30:00Z"
mailbox: "INBOX"
cached_at: "2025-01-15T12:00:00Z"
source_url: "https://api.fastmail.com/jmap/api/"
command: "fm fetch M1234567890"
---

# Meeting tomorrow

**From:** sender@example.com
**To:** recipient@example.com
**Date:** January 15, 2025 10:30 AM

Email body content here...
```

### Cache Location
- Default: `~/.local/share/fm/cache/`
- Configurable via `FM_CACHE_DIR` environment variable or `--cache-dir` flag
- Emails are stored as `{id}.md` files

### Cache Behavior
- `fm fetch` checks cache first, use `--no-cache` to bypass
- `fm search` always queries the API (results are ephemeral)
- Cache is append-only — old entries are never automatically deleted

## JMAP Protocol Details

### Key Methods Used
- `Email/query` — search and filter emails, returns email IDs
- `Email/get` — fetch full email objects by ID
- `Mailbox/get` — list mailboxes
- `Thread/get` — fetch thread information
- `Email/query` + `Email/get` are chained via JMAP result references in a single HTTP request

### Rate Limiting
- Respect Fastmail rate limits
- Implement exponential backoff on HTTP 429 responses
- Show progress indicators during long-running fetches
- On partial failure (429 mid-fetch), return all successfully fetched data

### Error Handling
- Clear error messages for authentication failures
- Graceful handling of network errors with retries
- Timeout configuration via `--timeout` flag

## Configuration

Configuration is minimal and environment-driven:
- `FM_API_TOKEN` — Fastmail API token (required)
- `FM_CACHE_DIR` — Cache directory (optional, default `~/.local/share/fm/cache/`)
- `FM_OUTPUT` — Default output format (optional, default `text`)

## Guiding Principles

- **Read-only** — never modify, send, or delete emails
- **Typed language** — built in Go for safety and performance
- **Easy installation** — single binary, no runtime dependencies. Homebrew tap for distribution.
- **Filesystem logging** — all fetched data is cached as Markdown files
- **Portable** — works on macOS and Linux
- **Simple** — minimal configuration, sensible defaults
- **Respect rate limits** — be a good API citizen with backoff and progress indicators
