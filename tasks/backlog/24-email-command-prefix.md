# Task 24: Move email commands under `fm email` prefix

## Objective

Group all email-related commands under an `fm email` subcommand prefix so the CLI has a clear namespace for email vs. future `fm calendar` and `fm contacts` command groups.

**Before:**
```
fm search <query>
fm fetch <email-id>
fm fetch-thread <thread-id>
fm mailboxes
fm list <mailbox>
```

**After:**
```
fm email search <query>
fm email fetch <email-id>
fm email fetch-thread <thread-id>
fm email mailboxes
fm email list <mailbox>
```

`fm auth` and `fm version` are not email-specific and stay at the root level.

## Acceptance Criteria

1. All five commands are reachable under `fm email <subcommand>`.
2. `fm email` with no subcommand prints help listing all subcommands.
3. Old top-level aliases (`fm search`, `fm fetch`, etc.) are removed — the old names should produce a "unknown command" error, not silently work.
4. `go build ./...`, `go vet ./...`, and `go test ./...` all pass with no changes to test logic (only command paths in tests may need updating).
5. README.md is updated to reflect the new command structure.
6. Shell completion still works correctly for the new nested structure.

## Implementation Notes

- Add a new `cmd/email.go` file that defines the `emailCmd` parent `*cobra.Command` with `Use: "email"` and registers it on `rootCmd`.
- Move `searchCmd`, `fetchCmd`, `fetchThreadCmd`, `mailboxesCmd`, and `listCmd` from `rootCmd` to `emailCmd` — each command's `init()` should call `emailCmd.AddCommand(...)` instead of `rootCmd.AddCommand(...)`.
- The `PersistentPreRunE` on `rootCmd` propagates to subcommands automatically in Cobra, so no changes needed there.
- Global flags (`--token`, `--debug`, `-o`, `-d`) are inherited by subcommands already.
- Update any integration / end-to-end tests that invoke the CLI by name (e.g. `exec.Command("fm", "search", ...)`) to use the new paths.

## Dependencies

None — this is a pure CLI restructuring with no JMAP or cache changes.
