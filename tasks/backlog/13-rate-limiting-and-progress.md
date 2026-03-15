# Task 13: Rate Limiting and Progress Indicators

**Phase:** 3 — Integration
**Blocked by:** 09, 10
**Blocks:** 14

## Objective

Add robust rate limiting, partial result handling, and progress indicators across all commands.

## Acceptance Criteria

- [ ] Rate limiting in the JMAP client:
  - Detect HTTP 429 responses
  - Parse `Retry-After` header (seconds or date)
  - Exponential backoff: 1s, 2s, 4s, 8s, 16s with ±25% jitter
  - Maximum 5 retries before giving up
- [ ] Partial result handling:
  - If a batch fetch hits 429 mid-way, return all successfully fetched emails + warning
  - Warning message includes how many were fetched vs. total requested
- [ ] Progress indicators:
  - Show progress bar or counter during multi-email fetches (search, fetch-thread)
  - Format: `Fetching emails... [12/50]` or similar
  - Only show on TTY (not when piped)
- [ ] Integration tests simulating rate limit scenarios

## Notes

- Use `github.com/schollz/progressbar/v3` or a simple custom counter
- Check `os.Stdout.Fd()` with `term.IsTerminal()` to detect TTY
- The partial result behavior is critical — never silently drop data
