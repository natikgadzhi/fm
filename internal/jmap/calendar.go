package jmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/natikgadzhi/fm/internal/verbose"
)

// Calendar represents a JMAP Calendar object.
type Calendar struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	Description  string `json:"description"`
	IsReadOnly   bool   `json:"isReadOnly"`
	IsSubscribed bool   `json:"isSubscribed"`
}

// CalendarEvent represents a JMAP CalendarEvent object (JSCalendar / RFC 8984).
type CalendarEvent struct {
	Id              string            `json:"id"`
	CalendarIds     map[string]bool   `json:"calendarIds"`
	Title           string            `json:"title"`
	Start           string            `json:"start"`
	TimeZone        string            `json:"timeZone"`
	Duration        string            `json:"duration"`
	ShowWithoutTime bool              `json:"showWithoutTime"`
	Location        string            `json:"location,omitempty"`
	Description     string            `json:"description"`
	Participants    map[string]Participant `json:"participants,omitempty"`
	RecurrenceRules []RecurrenceRule  `json:"recurrenceRules,omitempty"`
	Status          string            `json:"status"`
}

// Participant represents a calendar event participant.
type Participant struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Kind  string `json:"kind"`
	Roles map[string]bool `json:"roles"`
}

// RecurrenceRule represents a JMAP recurrence rule.
type RecurrenceRule struct {
	Frequency string `json:"frequency"`
	Interval  int    `json:"interval,omitempty"`
	Until     string `json:"until,omitempty"`
	Count     int    `json:"count,omitempty"`
}

// CalendarFilter holds filter criteria for querying calendar events.
type CalendarFilter struct {
	CalendarIds []string
	After       *time.Time
	Before      *time.Time
}

// calendarProperties is the set of Calendar properties we request.
var calendarProperties = []string{
	"id", "name", "color", "description", "isReadOnly", "isSubscribed",
}

// calendarEventProperties is the set of CalendarEvent properties we request.
var calendarEventProperties = []string{
	"id", "calendarIds", "title", "start", "timeZone", "duration",
	"showWithoutTime", "location", "description", "participants",
	"recurrenceRules", "status",
}

// CalendarAccountID returns the primary account ID for the calendar capability.
// Falls back to PrimaryAccountID if calendar capability is not separately listed.
func (c *Client) CalendarAccountID() (string, error) {
	session := c.Session()
	if session == nil {
		return "", fmt.Errorf("session not discovered; call Discover() first")
	}

	calURI := "urn:ietf:params:jmap:calendars"
	for uri, id := range session.PrimaryAccounts {
		if string(uri) == calURI {
			return string(id), nil
		}
	}

	// Fall back to mail/core account.
	id, err := c.PrimaryAccountID()
	if err != nil {
		return "", err
	}
	return string(id), nil
}

// doRawJMAP sends a raw JMAP request body to the API URL and returns the parsed response.
func (c *Client) doRawJMAP(ctx context.Context, body map[string]any) (map[string]any, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	session := c.Session()
	if session == nil {
		return nil, fmt.Errorf("no session available")
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling JMAP request: %w", err)
	}

	if verbose.Enabled() {
		calls, _ := body["methodCalls"].([]any)
		for _, call := range calls {
			if arr, ok := call.([]any); ok && len(arr) > 0 {
				verbose.Log("JMAP raw request: %s", arr[0])
			}
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, session.APIURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating JMAP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.inner.HttpClient.Do(httpReq)
	if err != nil {
		return nil, classifyError(err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("JMAP request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	respData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading JMAP response: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parsing JMAP response: %w", err)
	}

	return result, nil
}

// GetCalendars fetches all calendars from the server using Calendar/get.
func (c *Client) GetCalendars(ctx context.Context) ([]Calendar, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.CalendarAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting calendar account ID: %w", err)
	}

	body := map[string]any{
		"using": []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:calendars",
			"https://www.fastmail.com/dev/calendars",
		},
		"methodCalls": []any{
			[]any{
				"Calendar/get",
				map[string]any{
					"accountId":  accountID,
					"properties": calendarProperties,
				},
				"c0",
			},
		},
	}

	resp, err := c.doRawJMAP(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("Calendar/get request failed: %w", err)
	}

	responses, ok := resp["methodResponses"].([]any)
	if !ok || len(responses) == 0 {
		return nil, fmt.Errorf("Calendar/get: empty response")
	}

	inv, ok := responses[0].([]any)
	if !ok || len(inv) < 2 {
		return nil, fmt.Errorf("Calendar/get: malformed response")
	}

	args, ok := inv[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Calendar/get: unexpected args type")
	}

	list, ok := args["list"].([]any)
	if !ok {
		return nil, fmt.Errorf("Calendar/get: no list in response")
	}

	calendars := make([]Calendar, 0, len(list))
	for _, item := range list {
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var cal Calendar
		if err := json.Unmarshal(data, &cal); err != nil {
			continue
		}
		// Only include subscribed calendars.
		if cal.IsSubscribed {
			calendars = append(calendars, cal)
		}
	}

	return calendars, nil
}

// ResolveCalendar resolves a calendar name or ID to its JMAP Calendar ID.
// The match is case-insensitive on the Name field; if no match, treat input as raw ID.
func (c *Client) ResolveCalendar(ctx context.Context, nameOrID string) (string, error) {
	nameOrID = strings.TrimSpace(nameOrID)
	if nameOrID == "" {
		return "", fmt.Errorf("calendar name must not be empty")
	}

	calendars, err := c.GetCalendars(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving calendar %q: %w", nameOrID, err)
	}

	lower := strings.ToLower(nameOrID)
	for _, cal := range calendars {
		if strings.ToLower(cal.Name) == lower {
			return cal.Id, nil
		}
	}

	// Treat as raw ID — check if it exists.
	for _, cal := range calendars {
		if cal.Id == nameOrID {
			return cal.Id, nil
		}
	}

	return "", fmt.Errorf("calendar %q not found", nameOrID)
}

// QueryCalendarEvents queries calendar events matching the given filter.
func (c *Client) QueryCalendarEvents(ctx context.Context, filter CalendarFilter, sort string, limit int) ([]string, error) {
	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.CalendarAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting calendar account ID: %w", err)
	}

	jmapFilter := make(map[string]any)

	if filter.After != nil {
		jmapFilter["after"] = filter.After.UTC().Format("2006-01-02T15:04:05")
	}
	if filter.Before != nil {
		jmapFilter["before"] = filter.Before.UTC().Format("2006-01-02T15:04:05")
	}
	if len(filter.CalendarIds) > 0 {
		calMap := make(map[string]bool)
		for _, id := range filter.CalendarIds {
			calMap[id] = true
		}
		jmapFilter["inCalendars"] = filter.CalendarIds
	}

	isAscending := true
	if sort == "desc" {
		isAscending = false
	}

	queryArgs := map[string]any{
		"accountId": accountID,
		"sort": []map[string]any{
			{"property": "start", "isAscending": isAscending},
		},
	}

	if len(jmapFilter) > 0 {
		queryArgs["filter"] = jmapFilter
	}
	if limit > 0 {
		queryArgs["limit"] = limit
	}

	body := map[string]any{
		"using": []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:calendars",
			"https://www.fastmail.com/dev/calendars",
		},
		"methodCalls": []any{
			[]any{
				"CalendarEvent/query",
				queryArgs,
				"c0",
			},
		},
	}

	resp, err := c.doRawJMAP(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("CalendarEvent/query request failed: %w", err)
	}

	responses, ok := resp["methodResponses"].([]any)
	if !ok || len(responses) == 0 {
		return nil, fmt.Errorf("CalendarEvent/query: empty response")
	}

	inv, ok := responses[0].([]any)
	if !ok || len(inv) < 2 {
		return nil, fmt.Errorf("CalendarEvent/query: malformed response")
	}

	args, ok := inv[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("CalendarEvent/query: unexpected args type")
	}

	rawIDs, ok := args["ids"].([]any)
	if !ok {
		return nil, nil
	}

	ids := make([]string, 0, len(rawIDs))
	for _, raw := range rawIDs {
		if id, ok := raw.(string); ok {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// GetCalendarEvents fetches full CalendarEvent objects by their IDs.
func (c *Client) GetCalendarEvents(ctx context.Context, ids []string) ([]CalendarEvent, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	if err := c.Discover(); err != nil {
		return nil, fmt.Errorf("session discovery failed: %w", err)
	}

	accountID, err := c.CalendarAccountID()
	if err != nil {
		return nil, fmt.Errorf("getting calendar account ID: %w", err)
	}

	body := map[string]any{
		"using": []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:calendars",
			"https://www.fastmail.com/dev/calendars",
		},
		"methodCalls": []any{
			[]any{
				"CalendarEvent/get",
				map[string]any{
					"accountId":  accountID,
					"ids":        ids,
					"properties": calendarEventProperties,
				},
				"c0",
			},
		},
	}

	resp, err := c.doRawJMAP(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("CalendarEvent/get request failed: %w", err)
	}

	responses, ok := resp["methodResponses"].([]any)
	if !ok || len(responses) == 0 {
		return nil, fmt.Errorf("CalendarEvent/get: empty response")
	}

	inv, ok := responses[0].([]any)
	if !ok || len(inv) < 2 {
		return nil, fmt.Errorf("CalendarEvent/get: malformed response")
	}

	args, ok := inv[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("CalendarEvent/get: unexpected args type")
	}

	list, ok := args["list"].([]any)
	if !ok {
		return nil, fmt.Errorf("CalendarEvent/get: no list in response")
	}

	events := make([]CalendarEvent, 0, len(list))
	for _, item := range list {
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var evt CalendarEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			continue
		}

		// Handle location as either string or object.
		if rawItem, ok := item.(map[string]any); ok {
			evt.Location = parseLocation(rawItem["location"])
		}

		events = append(events, evt)
	}

	return events, nil
}

// SearchCalendarEvents queries and fetches calendar events in two steps.
func (c *Client) SearchCalendarEvents(ctx context.Context, filter CalendarFilter, sort string, limit int) ([]CalendarEvent, error) {
	ids, err := c.QueryCalendarEvents(ctx, filter, sort, limit)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return c.GetCalendarEvents(ctx, ids)
}

// parseLocation extracts a location string from the JMAP location field,
// which may be a string, a map (JSCalendar Location object), or nil.
func parseLocation(v any) string {
	if v == nil {
		return ""
	}
	switch loc := v.(type) {
	case string:
		return loc
	case map[string]any:
		// JSCalendar locations: map of locationId -> Location object.
		// Each Location has a "name" field.
		var names []string
		for _, val := range loc {
			if locObj, ok := val.(map[string]any); ok {
				if name, ok := locObj["name"].(string); ok && name != "" {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%v", v)
}
