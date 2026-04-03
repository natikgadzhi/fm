package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/spf13/cobra"
)

var (
	eventsCalendar string
	eventsFrom     string
	eventsTo       string
	eventsSort     string
	eventsLimit    int

	// Shortcut flags.
	eventsToday     bool
	eventsTomorrow  bool
	eventsThisWeek  bool
	eventsNextWeek  bool
	eventsThisMonth bool
	eventsNextMonth bool
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List calendar events",
	Long: `List calendar events within a time window.

By default, shows events from today through the next 30 days.
Use --from and --to flags with YYYY-MM-DD dates to customize the window,
or use shortcut flags like --today, --this-week, --next-month, etc.

Shortcut flags:
  --today         Events for today only
  --tomorrow      Events for tomorrow only
  --this-week     Events from today through end of this week (Sunday)
  --next-week     Events for next week (Monday through Sunday)
  --this-month    Events from today through end of this month
  --next-month    Events for all of next month

Shortcut flags override --from and --to when provided.`,
	Example: `  fm cal events
  fm cal events --today
  fm cal events --next-week
  fm cal events --this-month
  fm cal events --calendar Personal
  fm cal events --from 2025-01-01 --to 2025-02-01
  fm cal events --calendar Work --sort desc -n 10`,
	Args: cobra.NoArgs,
	RunE: runEvents,
}

var eventsGetCmd = &cobra.Command{
	Use:   "get <event-id>",
	Short: "Show full details for a calendar event",
	Long:  "Fetches and displays all details for a single calendar event by its JMAP ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runEventsGet,
}

func init() {
	calCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(eventsGetCmd)

	eventsCmd.Flags().StringVar(&eventsCalendar, "calendar", "", "Filter by calendar name or ID")
	eventsCmd.Flags().StringVar(&eventsFrom, "from", "", "Events starting on or after this date (YYYY-MM-DD)")
	eventsCmd.Flags().StringVar(&eventsTo, "to", "", "Events starting before this date (YYYY-MM-DD)")
	eventsCmd.Flags().StringVar(&eventsSort, "sort", "asc", "Sort by start time: asc or desc")
	eventsCmd.Flags().IntVarP(&eventsLimit, "limit", "n", 50, "Maximum number of events to return")

	// Shortcut flags.
	eventsCmd.Flags().BoolVar(&eventsToday, "today", false, "Show events for today only")
	eventsCmd.Flags().BoolVar(&eventsTomorrow, "tomorrow", false, "Show events for tomorrow only")
	eventsCmd.Flags().BoolVar(&eventsThisWeek, "this-week", false, "Show events from today through end of this week")
	eventsCmd.Flags().BoolVar(&eventsNextWeek, "next-week", false, "Show events for next week")
	eventsCmd.Flags().BoolVar(&eventsThisMonth, "this-month", false, "Show events from today through end of this month")
	eventsCmd.Flags().BoolVar(&eventsNextMonth, "next-month", false, "Show events for all of next month")
}

// resolveEventDateRange determines the from/to date range based on flags.
// Shortcut flags take priority over --from/--to.
// Returns (from, to) as *time.Time pointers.
func resolveEventDateRange(now time.Time) (*time.Time, *time.Time, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check shortcut flags (first match wins).
	switch {
	case eventsToday:
		from := today
		to := today.AddDate(0, 0, 1)
		return &from, &to, nil

	case eventsTomorrow:
		from := today.AddDate(0, 0, 1)
		to := today.AddDate(0, 0, 2)
		return &from, &to, nil

	case eventsThisWeek:
		from := today
		// Find end of week (Sunday 23:59:59 → next Monday at 00:00).
		daysUntilMonday := (7 - int(today.Weekday()) + 1) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		to := today.AddDate(0, 0, daysUntilMonday)
		return &from, &to, nil

	case eventsNextWeek:
		// Next Monday.
		daysUntilMonday := (7 - int(today.Weekday()) + 1) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		from := today.AddDate(0, 0, daysUntilMonday)
		to := from.AddDate(0, 0, 7)
		return &from, &to, nil

	case eventsThisMonth:
		from := today
		to := time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location())
		return &from, &to, nil

	case eventsNextMonth:
		from := time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location())
		to := time.Date(today.Year(), today.Month()+2, 1, 0, 0, 0, 0, today.Location())
		return &from, &to, nil
	}

	// Fall back to --from/--to flags.
	var fromTime, toTime *time.Time

	if eventsFrom != "" {
		t, err := time.Parse("2006-01-02", eventsFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --from date %q (expected YYYY-MM-DD): %w", eventsFrom, err)
		}
		fromTime = &t
	}
	if eventsTo != "" {
		t, err := time.Parse("2006-01-02", eventsTo)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --to date %q (expected YYYY-MM-DD): %w", eventsTo, err)
		}
		toTime = &t
	}

	// Default: today → today + 30 days.
	if fromTime == nil && toTime == nil {
		from := today
		to := today.AddDate(0, 0, 30)
		fromTime = &from
		toTime = &to
	}

	return fromTime, toTime, nil
}

func runEvents(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)
	ctx := cmd.Context()

	// Build filter.
	filter := jmap.CalendarFilter{}

	fromTime, toTime, err := resolveEventDateRange(time.Now())
	if err != nil {
		return err
	}
	filter.After = fromTime
	filter.Before = toTime

	// Resolve calendar filter.
	if eventsCalendar != "" {
		calID, err := client.ResolveCalendar(ctx, eventsCalendar)
		if err != nil {
			return fmt.Errorf("resolving calendar: %w", err)
		}
		filter.CalendarIds = []string{calID}
	}

	// Validate sort.
	if eventsSort != "asc" && eventsSort != "desc" {
		return fmt.Errorf("invalid --sort value %q: must be 'asc' or 'desc'", eventsSort)
	}

	events, err := client.SearchCalendarEvents(ctx, filter, eventsSort, eventsLimit)
	if err != nil {
		return fmt.Errorf("fetching events: %w", err)
	}

	if len(events) == 0 {
		fmt.Fprintln(os.Stderr, "No calendar events found in the specified time range.")
		return nil
	}

	format := output.Resolve(cmd)
	renderer := &jmap.CalendarEventListRenderer{Events: events}
	if err := output.Print(format, events, renderer); err != nil {
		return fmt.Errorf("formatting events: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nFound %d event(s).\n", len(events))
	return nil
}

func runEventsGet(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)
	ctx := cmd.Context()

	eventID := args[0]
	events, err := client.GetCalendarEvents(ctx, []string{eventID})
	if err != nil {
		return fmt.Errorf("fetching event: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("event %q not found", eventID)
	}

	format := output.Resolve(cmd)
	renderer := &jmap.CalendarEventRenderer{Event: events[0]}
	if err := output.Print(format, events[0], renderer); err != nil {
		return fmt.Errorf("formatting event: %w", err)
	}

	return nil
}
