# Task 17: Architecture Review and Simplify

**Phase:** 5 — Final Review
**Blocked by:** 09, 10, 11, 12, 13, 14, 15, 16
**Blocks:** none

## Objective

Run a comprehensive architecture review of the entire codebase, then simplify and refine for clarity, consistency, and maintainability.

## Acceptance Criteria

- [ ] Full architecture review covering:
  - Consistent patterns across all packages (error handling, naming, types)
  - No dead code, unused imports, or unnecessary abstractions
  - Proper separation of concerns between packages
  - Consistent API surfaces (method signatures, option patterns)
  - No duplicate type mapping code between task 05 and task 06 (email mapping)
  - Security: tokens never logged, input validated at CLI boundaries
- [ ] Run /simplify agent on all recently changed code
- [ ] All identified issues fixed
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` all pass after changes
- [ ] Code is clean, consistent, and ready for release

## Notes

- This is the final quality gate before release
- Focus on reuse opportunities (e.g., email mapping helpers used in both email.go and thread.go)
- Check for consistent error wrapping patterns
- Verify all public API surfaces are intentional
