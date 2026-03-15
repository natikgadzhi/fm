# fm

A read-only CLI tool for searching and fetching email from Fastmail via the JMAP API, with local Markdown caching.

## Why?

`fm` makes Fastmail emails accessible from the command line. Search your inbox, fetch individual messages or entire threads, and cache them locally as Markdown files with YAML frontmatter. It is designed for scripting, piping into other tools, or building workflows with AI tools that can read Markdown.

`fm` is read-only by design -- it never modifies, sends, or deletes emails.

## Installation

### Homebrew

```sh
brew install natikgadzhi/taps/fm
```

### `go install`

```sh
go install github.com/natikgadzhi/fm@latest
```

### From source

```sh
git clone https://github.com/natikgadzhi/fm
cd fm
make build
# The binary is at ./fm — move it to somewhere in your $PATH.
```

## Quick start

1. Create a Fastmail API token at <https://app.fastmail.com/settings/security/tokens/new>.
   Grant **JMAP Core** and **Mail** scopes (read-only is sufficient).
2. Store the token in your OS keychain:
   ```sh
   fm auth login
   # Paste your token (starts with fmu1-) when prompted
   ```
3. Search your inbox:
   ```sh
   fm search "from:boss@company.com subject:quarterly report"
   ```
4. Fetch a single email by its JMAP ID (shown in search results):
   ```sh
   fm fetch Mabcdef1234567890
   ```

## Commands

### `fm auth login`

Prompts for a Fastmail API token and stores it securely in the OS keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager).

### `fm auth check`

Verifies that your API token is valid by making a JMAP session request. On success, displays the token source, masked token, account ID, and username.

```sh
$ fm auth check
Token source: keychain
Token:        fmu1-****7890
Account ID:   u12345
Username:     user@fastmail.com
```

### `fm auth logout`

Removes the stored API token from the OS keychain.

### `fm search <query>`

Search emails using Fastmail's JMAP query interface. Returns a summary table with date, sender, subject, and snippet.

```sh
# Free text search
fm search meeting notes

# Filter by sender
fm search "from:alice@example.com"

# Combine filters
fm search "from:boss@company.com subject:urgent after:2025-01-01"

# Use CLI flags instead of (or in addition to) query syntax
fm search --from boss@company.com --has-attachments

# Limit results and output as JSON
fm search "in:INBOX" --limit 10 -o json
```

**Supported query syntax:**

| Filter              | Example                    | Description                    |
|---------------------|----------------------------|--------------------------------|
| `from:`             | `from:user@example.com`    | Filter by sender               |
| `to:`               | `to:user@example.com`      | Filter by recipient            |
| `subject:`          | `subject:meeting notes`    | Filter by subject              |
| `in:`               | `in:INBOX`                 | Filter by mailbox name         |
| `before:`           | `before:2025-06-01`        | Emails before a date           |
| `after:`            | `after:2025-01-01`         | Emails after a date            |
| `has:attachment`    | `has:attachment`           | Only emails with attachments   |
| Free text           | `meeting notes`            | Search body and subject        |

**CLI flags** (`--from`, `--to`, `--has-attachments`) override inline query filters when both are provided.

### `fm fetch <email-id>`

Fetch a single email by its JMAP ID. The email is cached locally as a Markdown file and printed to stdout.

```sh
# Fetch and display an email
fm fetch Mabcdef1234567890

# Bypass the cache and re-fetch from the server
fm fetch --no-cache Mabcdef1234567890

# Fetch with attachments saved to the cache directory
fm fetch --with-attachments Mabcdef1234567890

# Output as JSON
fm fetch Mabcdef1234567890 -o json
```

### `fm fetch-thread <thread-id>`

Fetch all emails in a thread, displayed in chronological order. Each email is cached individually.

```sh
# Fetch an entire thread
fm fetch-thread Tabcdef1234567890

# Fetch a thread and download all attachments
fm fetch-thread --with-attachments Tabcdef1234567890

# Output the thread as Markdown
fm fetch-thread Tabcdef1234567890 -o markdown
```

### `fm mailboxes`

List all mailboxes (folders) in the account, sorted alphabetically.

```sh
# List mailboxes as a table
fm mailboxes

# List mailboxes as JSON (useful for scripting)
fm mailboxes -o json
```

## Output formats

All commands support the `-o` flag for output format:

| Format     | Flag           | Description                          |
|------------|----------------|--------------------------------------|
| Text       | `-o text`      | Human-readable table (default)       |
| JSON       | `-o json`      | Structured JSON                      |
| Markdown   | `-o markdown`  | Markdown formatted output            |

You can set a default output format with the `FM_OUTPUT` environment variable.

## Configuration

`fm` is configured through environment variables and CLI flags. Flags take precedence over environment variables, which take precedence over defaults.

| Environment variable | CLI flag       | Default                       | Description                    |
|----------------------|----------------|-------------------------------|--------------------------------|
| `FM_API_TOKEN`       | `--token`      | OS keychain                   | Fastmail API token             |
| `FM_CACHE_DIR`       | `--cache-dir`  | `~/.local/share/fm/cache/`    | Cache directory path           |
| `FM_OUTPUT`          | `-o`           | `text`                        | Default output format          |
| --                   | `--timeout`    | `30s`                         | HTTP request timeout           |

### Token resolution order

The API token is resolved from these sources, in priority order:

1. `--token` flag
2. `FM_API_TOKEN` environment variable
3. OS keychain (set via `fm auth login`)

## Cache

Fetched emails are cached locally as Markdown files with YAML frontmatter at `~/.local/share/fm/cache/` (configurable via `FM_CACHE_DIR` or `--cache-dir`).

```
~/.local/share/fm/cache/
    Mabcdef1234567890.md
    Mxyz9876543210.md
    attachments/
        Mabcdef1234567890/
            report.pdf
            screenshot.png
```

Each cached file looks like:

```markdown
---
tool: fm
object: email
id: "Mabcdef1234567890"
thread_id: "Tabcdef1234567890"
message_id: "<abc@fastmail.com>"
from: "sender@example.com"
to: ["recipient@example.com"]
subject: "Meeting tomorrow"
date: "2025-01-15T10:30:00Z"
mailbox: "INBOX"
cached_at: "2025-01-15T12:00:00Z"
source_url: "https://api.fastmail.com/jmap/api/"
command: "fm fetch Mabcdef1234567890"
---

# Meeting tomorrow

**From:** sender@example.com
**To:** recipient@example.com
**Date:** January 15, 2025 3:30 PM

Hey, are we still on for the meeting tomorrow at 10am?
Let me know if the time works for you.
```

**Cache behavior:**

- `fm fetch` checks the cache first. Use `--no-cache` to bypass it.
- `fm search` always queries the API (search results are not cached).
- The cache is append-only -- old entries are never automatically deleted.

## License

[MIT](LICENSE)
