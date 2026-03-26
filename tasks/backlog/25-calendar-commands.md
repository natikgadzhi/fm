# Task 25: Implement `fm cal` calendar commands

## Objective

Add a `fm cal` subcommand group that lets users list their Fastmail calendars and query/fetch calendar events via the JMAP Calendars API (`urn:ietf:params:jmap:calendars`).

**New commands:**

```
fm cal calendars                          # list all calendars
fm cal events [--calendar <name|id>]      # list events (optionally filtered to one calendar)
              [--from <date>]             # events starting on or after this date
              [--to <date>]               # events starting before this date
              [--sort asc|desc]           # sort by start time (default: asc)
fm cal events get <event-id>              # show full details for one event
```

## Acceptance Criteria

1. `fm cal calendars` lists all calendars in the account. Table columns: Name, Color, Description, Read-only.
2. `fm cal events` without flags returns upcoming events (default window: today → today+30d).
3. `--from` and `--to` accept dates in `YYYY-MM-DD` format and narrow the time window.
4. `--calendar` accepts a calendar name (case-insensitive) or JMAP Calendar ID and filters events to that calendar.
5. `--sort asc|desc` controls chronological vs. reverse-chronological order (default `asc`).
6. `fm cal events get <id>` prints full event details (title, calendar, start, end, location, description, attendees, recurrence rule if present).
7. All three commands support `-o table` and `-o json`.
8. `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
9. Unit tests cover: JMAP calendar type mapping, calendar name resolution, date flag parsing, and the `CalendarFilter` → JMAP filter conversion.

## Implementation Notes

### JMAP Calendar capability

Fastmail exposes `urn:ietf:params:jmap:calendars`. The `go-jmap` library does **not** include calendar types, so implement them from scratch in a new package `internal/jmap/calendar/` (or directly in `internal/jmap/` alongside `email.go`). Use raw JMAP method invocations via `gojmap.RawInvoke` / custom `Args` structs — the same approach the email package uses.

Key JMAP methods:
- `Calendar/get` — returns all calendars for the account. No filter needed; just pass `ids: null` to get all.
- `CalendarEvent/query` — filter by `calendarIds`, `after` (UTCStart >= date), `before` (UTCStart < date). Returns event IDs.
- `CalendarEvent/get` — fetch full event objects by ID.

Fastmail's calendar API follows RFC 8984 (JSCalendar) for event shapes. The minimum fields to request:

**Calendar properties:** `id`, `name`, `color`, `description`, `isReadOnly`, `isSubscribed`

**CalendarEvent properties:** `id`, `calendarIds`, `title`, `start`, `timeZone`, `duration`, `showWithoutTime`, `location`, `description`, `participants`, `recurrenceRules`, `status`

### Account ID for calendars

`PrimaryAccountID()` looks up `urn:ietf:params:jmap:mail`. Add a new helper `CalendarAccountID()` on the `Client` that looks up `urn:ietf:params:jmap:calendars` instead, falling back to `PrimaryAccountID()` if the calendar capability isn't separately listed (Fastmail uses the same account for both).

### Calendar name resolution

Add `ResolveCalendar(ctx, nameOrID) (calendarID string, err error)` on the client — analogous to `ResolveMailbox` in `mailbox.go`. Fetches all calendars, does a case-insensitive name match, falls back to treating the input as a raw ID.

### Date parsing

Accept `YYYY-MM-DD`; parse with `time.Parse("2006-01-02", ...)` and convert to UTC midnight. Produce a clear error if the format is wrong.

### File structure

```
internal/jmap/calendar.go        # Calendar and CalendarEvent types + JMAP client methods
internal/jmap/calendar_test.go   # unit tests for mapping and filter logic
cmd/cal.go                       # calCmd parent + calendarsCmd
cmd/cal_events.go                # eventsCmd (list) + eventsGetCmd
cmd/cal_events_test.go           # tests for flag parsing and command wiring
```

### CLI wiring

Register `calCmd` on `rootCmd` in `cmd/cal.go` `init()`, similar to how `emailCmd` will be registered in task 24. This task can be implemented independently of task 24 (the `cal` group is a sibling of `email`, not nested under it).

## Dependencies

- Task 24 (`fm email` prefix) is a **sibling**, not a prerequisite. `fm cal` registers directly on `rootCmd`.
- No new Go module dependencies are expected — use `go-jmap`'s raw invocation support.
