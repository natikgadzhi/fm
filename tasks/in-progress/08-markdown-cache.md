# Task 08: Markdown Cache

**Phase:** 2 — Implementation
**Blocked by:** 04
**Blocks:** 10, 12

## Objective

Implement the Markdown file cache for storing fetched emails locally.

## Acceptance Criteria

- [ ] `internal/cache/frontmatter.go` defines:
  - `Frontmatter` struct matching PROJECT_PROMPT.md spec:
    - Tool, Object, Id, ThreadId, MessageId, From, To, Subject, Date, Mailbox, CachedAt, SourceURL, Command
  - `Marshal(fm Frontmatter) ([]byte, error)` — render YAML frontmatter
  - `Unmarshal(data []byte) (*Frontmatter, *string, error)` — parse frontmatter + body
- [ ] `internal/cache/cache.go` defines:
  - `Cache` struct with configurable directory
  - `NewCache(dir string) *Cache`
  - `Get(id string) (*jmap.Email, error)` — read cached email, return nil if not found
  - `Put(email jmap.Email, command string) error` — write email as `{id}.md`
  - `Exists(id string) bool` — check if email is cached
- [ ] Email body rendered as Markdown:
  - Headers (From, To, Date, Subject) as bold labels
  - Text body as content
  - Fallback to HTML body stripped of tags if no text body
- [ ] Cache directory is created automatically if it doesn't exist
- [ ] Unit tests using `t.TempDir()`:
  - Write and read back an email
  - Cache miss returns nil
  - Frontmatter round-trip
  - File content matches expected format

## Notes

- Use `github.com/adrg/frontmatter` or manual YAML marshaling
- File names: `{email-id}.md` — JMAP IDs may contain special characters, sanitize if needed
- Cache is append-only — never delete or overwrite existing entries
