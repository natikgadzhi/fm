package cmd

import (
	"testing"
	"time"
)

func TestCalCommandExists(t *testing.T) {
	if calCmd == nil {
		t.Fatal("calCmd should not be nil")
	}
	if calCmd.Use != "cal" {
		t.Errorf("expected Use 'cal', got %q", calCmd.Use)
	}
}

func TestCalendarsCommandNoArgs(t *testing.T) {
	cmd := calendarsCmd
	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
	err := cmd.Args(cmd, []string{"extra"})
	if err == nil {
		t.Error("expected error when passing args to calendars command")
	}
}

func TestEventsCommandNoArgs(t *testing.T) {
	cmd := eventsCmd
	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
	err := cmd.Args(cmd, []string{"extra"})
	if err == nil {
		t.Error("expected error when passing args to events command")
	}
}

func TestEventsGetCommandArgs(t *testing.T) {
	cmd := eventsGetCmd
	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error for zero args")
	}
	if err := cmd.Args(cmd, []string{"evt-123"}); err != nil {
		t.Errorf("expected no error for one arg, got %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error for two args")
	}
}

func TestResolveEventDateRangeDefaults(t *testing.T) {
	// Reset all flags.
	eventsToday, eventsTomorrow = false, false
	eventsThisWeek, eventsNextWeek = false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""

	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if from == nil || to == nil {
		t.Fatal("expected non-nil from and to")
	}

	expectedFrom := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeToday(t *testing.T) {
	eventsToday = true
	eventsTomorrow, eventsThisWeek, eventsNextWeek = false, false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsToday = false }()

	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeTomorrow(t *testing.T) {
	eventsTomorrow = true
	eventsToday, eventsThisWeek, eventsNextWeek = false, false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsTomorrow = false }()

	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 6, 17, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeThisWeek(t *testing.T) {
	eventsThisWeek = true
	eventsToday, eventsTomorrow, eventsNextWeek = false, false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsThisWeek = false }()

	// Sunday June 15, 2025 - weekday=0 (Sunday).
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	// Next Monday is June 16.
	expectedTo := time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeNextWeek(t *testing.T) {
	eventsNextWeek = true
	eventsToday, eventsTomorrow, eventsThisWeek = false, false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsNextWeek = false }()

	// Wednesday June 18, 2025 - weekday=3.
	now := time.Date(2025, 6, 18, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Next Monday is June 23.
	expectedFrom := time.Date(2025, 6, 23, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeThisMonth(t *testing.T) {
	eventsThisMonth = true
	eventsToday, eventsTomorrow, eventsThisWeek = false, false, false
	eventsNextWeek, eventsNextMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsThisMonth = false }()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeNextMonth(t *testing.T) {
	eventsNextMonth = true
	eventsToday, eventsTomorrow, eventsThisWeek = false, false, false
	eventsNextWeek, eventsThisMonth = false, false
	eventsFrom, eventsTo = "", ""
	defer func() { eventsNextMonth = false }()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeExplicitFlags(t *testing.T) {
	eventsToday, eventsTomorrow = false, false
	eventsThisWeek, eventsNextWeek = false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom = "2025-03-01"
	eventsTo = "2025-04-01"
	defer func() { eventsFrom, eventsTo = "", "" }()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *to)
	}
}

func TestResolveEventDateRangeInvalidDate(t *testing.T) {
	eventsToday, eventsTomorrow = false, false
	eventsThisWeek, eventsNextWeek = false, false
	eventsThisMonth, eventsNextMonth = false, false
	eventsFrom = "not-a-date"
	eventsTo = ""
	defer func() { eventsFrom = "" }()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	_, _, err := resolveEventDateRange(now)
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestResolveEventDateRangeShortcutOverridesExplicit(t *testing.T) {
	// Shortcut should win over explicit --from/--to.
	eventsToday = true
	eventsFrom = "2025-01-01"
	eventsTo = "2025-12-31"
	eventsTomorrow, eventsThisWeek, eventsNextWeek = false, false, false
	eventsThisMonth, eventsNextMonth = false, false
	defer func() {
		eventsToday = false
		eventsFrom, eventsTo = "", ""
	}()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	from, to, err := resolveEventDateRange(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFrom := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Errorf("shortcut should override --from: expected %v, got %v", expectedFrom, *from)
	}
	if !to.Equal(expectedTo) {
		t.Errorf("shortcut should override --to: expected %v, got %v", expectedTo, *to)
	}
}
