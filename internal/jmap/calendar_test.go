package jmap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// calendarSessionJSON returns a JMAP session response that includes calendar capabilities.
func calendarSessionJSON(apiURL string) []byte {
	session := map[string]any{
		"capabilities": map[string]any{
			"urn:ietf:params:jmap:core":      map[string]any{},
			"urn:ietf:params:jmap:mail":      map[string]any{},
			"urn:ietf:params:jmap:calendars": map[string]any{},
		},
		"accounts": map[string]any{
			"u12345": map[string]any{
				"name":               "test@fastmail.com",
				"isPersonal":         true,
				"isReadOnly":         false,
				"accountCapabilities": map[string]any{},
			},
		},
		"primaryAccounts": map[string]any{
			"urn:ietf:params:jmap:core":      "u12345",
			"urn:ietf:params:jmap:mail":      "u12345",
			"urn:ietf:params:jmap:calendars": "u12345",
		},
		"username":       "test@fastmail.com",
		"apiUrl":         apiURL + "/jmap/api",
		"downloadUrl":    apiURL + "/jmap/download/{accountId}/{blobId}/{name}?type={type}",
		"uploadUrl":      apiURL + "/jmap/upload/{accountId}/",
		"eventSourceUrl": apiURL + "/jmap/eventsource/?types={types}&closeafter={closeafter}&ping={ping}",
		"state":          "abc123",
	}
	data, _ := json.Marshal(session)
	return data
}

// newCalendarTestServer creates a test server with calendar-aware session.
func newCalendarTestServer(t *testing.T, apiHandler func(t *testing.T, req map[string]any) map[string]any) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	server := httptest.NewUnstartedServer(mux)
	server.Start()

	mux.HandleFunc("/jmap/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(calendarSessionJSON(server.URL))
	})

	mux.HandleFunc("/jmap/api", func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode JMAP request: %v", err)
		}
		resp := apiHandler(t, reqBody)
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(resp)
		w.Write(data)
	})

	return server
}

func TestGetCalendars(t *testing.T) {
	server := newCalendarTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		calls, _ := req["methodCalls"].([]any)
		if len(calls) != 1 {
			t.Fatalf("expected 1 method call, got %d", len(calls))
		}
		call := calls[0].([]any)
		if call[0].(string) != "Calendar/get" {
			t.Errorf("expected Calendar/get, got %s", call[0])
		}

		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Calendar/get",
					map[string]any{
						"accountId": "u12345",
						"list": []any{
							map[string]any{
								"id":           "cal-1",
								"name":         "Personal",
								"color":        "#0078d4",
								"description":  "My personal calendar",
								"isReadOnly":   false,
								"isSubscribed": true,
							},
							map[string]any{
								"id":           "cal-2",
								"name":         "Work",
								"color":        "#e74c3c",
								"description":  "Work events",
								"isReadOnly":   false,
								"isSubscribed": true,
							},
							map[string]any{
								"id":           "cal-3",
								"name":         "Holidays",
								"color":        "#2ecc71",
								"description":  "Public holidays",
								"isReadOnly":   true,
								"isSubscribed": false,
							},
						},
					},
					"c0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	calendars, err := c.GetCalendars(context.Background())
	if err != nil {
		t.Fatalf("GetCalendars() failed: %v", err)
	}

	// Should only include subscribed calendars.
	if len(calendars) != 2 {
		t.Fatalf("expected 2 subscribed calendars, got %d", len(calendars))
	}

	if calendars[0].Name != "Personal" {
		t.Errorf("expected first calendar 'Personal', got %q", calendars[0].Name)
	}
	if calendars[1].Name != "Work" {
		t.Errorf("expected second calendar 'Work', got %q", calendars[1].Name)
	}
}

func TestResolveCalendar(t *testing.T) {
	server := newCalendarTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Calendar/get",
					map[string]any{
						"accountId": "u12345",
						"list": []any{
							map[string]any{
								"id": "cal-1", "name": "Personal",
								"isSubscribed": true,
							},
							map[string]any{
								"id": "cal-2", "name": "Work",
								"isSubscribed": true,
							},
						},
					},
					"c0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	ctx := context.Background()

	// Case-insensitive name match.
	id, err := c.ResolveCalendar(ctx, "personal")
	if err != nil {
		t.Fatalf("ResolveCalendar('personal') failed: %v", err)
	}
	if id != "cal-1" {
		t.Errorf("expected cal-1, got %s", id)
	}

	// Exact ID match.
	id, err = c.ResolveCalendar(ctx, "cal-2")
	if err != nil {
		t.Fatalf("ResolveCalendar('cal-2') failed: %v", err)
	}
	if id != "cal-2" {
		t.Errorf("expected cal-2, got %s", id)
	}
}

func TestResolveCalendarNotFound(t *testing.T) {
	server := newCalendarTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return map[string]any{
			"methodResponses": []any{
				[]any{
					"Calendar/get",
					map[string]any{
						"accountId": "u12345",
						"list":      []any{},
					},
					"c0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.ResolveCalendar(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent calendar")
	}
}

func TestQueryCalendarEvents(t *testing.T) {
	server := newCalendarTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		calls, _ := req["methodCalls"].([]any)
		call := calls[0].([]any)
		if call[0].(string) != "CalendarEvent/query" {
			t.Errorf("expected CalendarEvent/query, got %s", call[0])
		}

		args := call[1].(map[string]any)
		filter, _ := args["filter"].(map[string]any)
		if filter["after"] == nil {
			t.Error("expected 'after' filter to be set")
		}

		return map[string]any{
			"methodResponses": []any{
				[]any{
					"CalendarEvent/query",
					map[string]any{
						"accountId": "u12345",
						"ids":       []string{"evt-1", "evt-2"},
						"total":     2,
					},
					"c0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	filter := CalendarFilter{After: &after}

	ids, err := c.QueryCalendarEvents(context.Background(), filter, "asc", 10)
	if err != nil {
		t.Fatalf("QueryCalendarEvents() failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 event IDs, got %d", len(ids))
	}
}

func TestGetCalendarEvents(t *testing.T) {
	server := newCalendarTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		calls, _ := req["methodCalls"].([]any)
		call := calls[0].([]any)
		if call[0].(string) != "CalendarEvent/get" {
			t.Errorf("expected CalendarEvent/get, got %s", call[0])
		}

		return map[string]any{
			"methodResponses": []any{
				[]any{
					"CalendarEvent/get",
					map[string]any{
						"accountId": "u12345",
						"list": []any{
							map[string]any{
								"id":              "evt-1",
								"calendarIds":     map[string]any{"cal-1": true},
								"title":           "Team standup",
								"start":           "2025-01-15T10:00:00",
								"timeZone":        "America/New_York",
								"duration":        "PT30M",
								"showWithoutTime": false,
								"description":     "Daily standup meeting",
								"status":          "confirmed",
							},
						},
					},
					"c0",
				},
			},
			"sessionState": "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	events, err := c.GetCalendarEvents(context.Background(), []string{"evt-1"})
	if err != nil {
		t.Fatalf("GetCalendarEvents() failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.Title != "Team standup" {
		t.Errorf("expected title 'Team standup', got %q", evt.Title)
	}
	if evt.Start != "2025-01-15T10:00:00" {
		t.Errorf("expected start '2025-01-15T10:00:00', got %q", evt.Start)
	}
	if evt.Duration != "PT30M" {
		t.Errorf("expected duration 'PT30M', got %q", evt.Duration)
	}
}

func TestCalendarFilterToJMAP(t *testing.T) {
	after := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	filter := CalendarFilter{
		CalendarIds: []string{"cal-1"},
		After:       &after,
		Before:      &before,
	}

	if len(filter.CalendarIds) != 1 || filter.CalendarIds[0] != "cal-1" {
		t.Errorf("unexpected CalendarIds: %v", filter.CalendarIds)
	}
	if filter.After == nil || !filter.After.Equal(after) {
		t.Errorf("unexpected After: %v", filter.After)
	}
	if filter.Before == nil || !filter.Before.Equal(before) {
		t.Errorf("unexpected Before: %v", filter.Before)
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, ""},
		{"string", "Office Room 42", "Office Room 42"},
		{
			"jscalendar location object",
			map[string]any{
				"loc1": map[string]any{"name": "Conference Room A"},
			},
			"Conference Room A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLocation(tt.input)
			if got != tt.want {
				t.Errorf("parseLocation(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCalendarAccountIDFallback(t *testing.T) {
	// Use the standard session (without calendar capability) to test fallback.
	server := newTestServer(t, func(t *testing.T, req map[string]any) map[string]any {
		return map[string]any{
			"methodResponses": []any{},
			"sessionState":    "abc123",
		}
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.Discover(); err != nil {
		t.Fatalf("Discover() failed: %v", err)
	}

	id, err := c.CalendarAccountID()
	if err != nil {
		t.Fatalf("CalendarAccountID() failed: %v", err)
	}

	// Should fall back to mail account.
	if id != "u12345" {
		t.Errorf("expected fallback to u12345, got %s", id)
	}
}
