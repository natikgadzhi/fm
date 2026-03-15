# Task 21: Show Email and Thread IDs in Search Results

**Phase:** 5 — Refinement
**Blocked by:** none
**Blocks:** none

## Objective

The default text output of `fm search` should include email ID and thread ID columns so users can directly use them with `fm fetch` and `fm fetch-thread`.

## Acceptance Criteria

- [ ] Text formatter's `FormatEmailList` includes ID and ThreadId columns
- [ ] Column order: ID, ThreadId, Date, From, Subject (or similar logical order)
- [ ] IDs are not truncated (they're needed for copy-paste)
- [ ] JSON and Markdown formatters already include IDs (verify)
- [ ] Update tests for text formatter output

## Notes

- Without IDs in the output, search results are not actionable — users can't pipe IDs into `fm fetch`
- Keep columns compact; IDs are typically short (M + 8-12 chars)
