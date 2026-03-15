//go:build e2e

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mock JMAP server
// ---------------------------------------------------------------------------

// mockJMAP is a configurable mock JMAP server for end-to-end tests.
type mockJMAP struct {
	server *httptest.Server
	mux    *http.ServeMux

	// Test data
	emails    []mockEmail
	mailboxes []mockMailbox
	threads   []mockThread

	// Counters for verifying API calls.
	apiCalls    atomic.Int32
	sessionHits atomic.Int32

	// Error injection
	rateLimitNextN atomic.Int32 // return 429 for the next N API requests
	failAuth       bool         // return 401 on session discovery
}

type mockEmail struct {
	ID        string `json:"id"`
	ThreadID  string `json:"threadId"`
	MessageID string `json:"messageId,omitempty"`
	Subject   string `json:"subject"`
	Preview   string `json:"preview"`
	From      []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"from"`
	To []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"to"`
	SentAt        string            `json:"sentAt"`
	MailboxIDs    map[string]bool   `json:"mailboxIds,omitempty"`
	TextBody      []mockBodyPart    `json:"textBody,omitempty"`
	HTMLBody      []mockBodyPart    `json:"htmlBody,omitempty"`
	BodyValues    map[string]mockBV `json:"bodyValues,omitempty"`
	Size          int               `json:"size"`
	HasAttachment bool              `json:"hasAttachment,omitempty"`
}

type mockBodyPart struct {
	PartID string `json:"partId"`
}

type mockBV struct {
	Value       string `json:"value"`
	IsTruncated bool   `json:"isTruncated"`
}

type mockMailbox struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role,omitempty"`
	TotalEmails  int    `json:"totalEmails"`
	UnreadEmails int    `json:"unreadEmails"`
	ParentID     string `json:"parentId,omitempty"`
}

type mockThread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

func newMockJMAP() *mockJMAP {
	m := &mockJMAP{
		mux: http.NewServeMux(),
	}

	// Default test data.
	sentAt := time.Date(2025, 3, 10, 14, 30, 0, 0, time.UTC).Format(time.RFC3339)
	sentAt2 := time.Date(2025, 3, 11, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)

	m.emails = []mockEmail{
		{
			ID:        "e001",
			ThreadID:  "t001",
			MessageID: "msg001@example.com",
			Subject:   "Test Email One",
			Preview:   "This is a preview of the first test email.",
			From: []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{{Name: "Alice Sender", Email: "alice@example.com"}},
			To: []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{{Name: "Bob Recipient", Email: "bob@example.com"}},
			SentAt:     sentAt,
			MailboxIDs: map[string]bool{"mb-inbox": true},
			TextBody:   []mockBodyPart{{PartID: "1"}},
			BodyValues: map[string]mockBV{
				"1": {Value: "Hello, this is the body of test email one."},
			},
			Size: 1234,
		},
		{
			ID:        "e002",
			ThreadID:  "t001",
			MessageID: "msg002@example.com",
			Subject:   "Re: Test Email One",
			Preview:   "Reply to the first test email.",
			From: []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{{Name: "Bob Recipient", Email: "bob@example.com"}},
			To: []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{{Name: "Alice Sender", Email: "alice@example.com"}},
			SentAt:     sentAt2,
			MailboxIDs: map[string]bool{"mb-inbox": true},
			TextBody:   []mockBodyPart{{PartID: "1"}},
			BodyValues: map[string]mockBV{
				"1": {Value: "Thanks for the email, Alice!"},
			},
			Size: 567,
		},
	}

	m.mailboxes = []mockMailbox{
		{ID: "mb-inbox", Name: "Inbox", Role: "inbox", TotalEmails: 42, UnreadEmails: 5},
		{ID: "mb-sent", Name: "Sent", Role: "sent", TotalEmails: 100, UnreadEmails: 0},
		{ID: "mb-drafts", Name: "Drafts", Role: "drafts", TotalEmails: 3, UnreadEmails: 0},
		{ID: "mb-trash", Name: "Trash", Role: "trash", TotalEmails: 10, UnreadEmails: 0},
		{ID: "mb-archive", Name: "Archive", Role: "archive", TotalEmails: 500, UnreadEmails: 0},
	}

	m.threads = []mockThread{
		{ID: "t001", EmailIDs: []string{"e001", "e002"}},
	}

	m.server = httptest.NewServer(m.mux)

	m.mux.HandleFunc("/jmap/session", m.handleSession)
	m.mux.HandleFunc("/jmap/api", m.handleAPI)

	return m
}

func (m *mockJMAP) Close() {
	m.server.Close()
}

func (m *mockJMAP) URL() string {
	return m.server.URL
}

func (m *mockJMAP) SessionURL() string {
	return m.server.URL + "/jmap/session"
}

func (m *mockJMAP) handleSession(w http.ResponseWriter, r *http.Request) {
	m.sessionHits.Add(1)

	if m.failAuth {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"type":"unauthorized","status":401,"detail":"Invalid token"}`))
		return
	}

	session := map[string]any{
		"capabilities": map[string]any{
			"urn:ietf:params:jmap:core": map[string]any{
				"maxSizeUpload":       50000000,
				"maxConcurrentUpload": 8,
				"maxSizeRequest":      10000000,
				"maxCallsInRequest":   64,
				"maxObjectsInGet":     1000,
				"maxObjectsInSet":     1000,
				"collationAlgorithms": []string{},
			},
			"urn:ietf:params:jmap:mail": map[string]any{},
		},
		"accounts": map[string]any{
			"u12345": map[string]any{
				"name":                "test@fastmail.com",
				"isPersonal":          true,
				"isReadOnly":          false,
				"accountCapabilities": map[string]any{},
			},
		},
		"primaryAccounts": map[string]any{
			"urn:ietf:params:jmap:core": "u12345",
			"urn:ietf:params:jmap:mail": "u12345",
		},
		"username":       "test@fastmail.com",
		"apiUrl":         m.server.URL + "/jmap/api",
		"downloadUrl":    m.server.URL + "/jmap/download/{accountId}/{blobId}/{name}?type={type}",
		"uploadUrl":      m.server.URL + "/jmap/upload/{accountId}/",
		"eventSourceUrl": m.server.URL + "/jmap/eventsource/?types={types}&closeafter={closeafter}&ping={ping}",
		"state":          "mock-state-1",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (m *mockJMAP) handleAPI(w http.ResponseWriter, r *http.Request) {
	m.apiCalls.Add(1)

	// Rate limit injection.
	if n := m.rateLimitNextN.Load(); n > 0 {
		m.rateLimitNextN.Add(-1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"type":"rate-limit","status":429,"detail":"slow down"}`))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req struct {
		Using []string          `json:"using"`
		Calls []json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var responses []json.RawMessage

	for _, raw := range req.Calls {
		var call []json.RawMessage
		if err := json.Unmarshal(raw, &call); err != nil {
			continue
		}
		if len(call) < 3 {
			continue
		}

		var methodName string
		json.Unmarshal(call[0], &methodName)

		var callID string
		json.Unmarshal(call[2], &callID)

		var respArgs any
		switch methodName {
		case "Email/query":
			respArgs = m.handleEmailQuery(call[1])
		case "Email/get":
			respArgs = m.handleEmailGet(call[1])
		case "Mailbox/get":
			respArgs = m.handleMailboxGet()
		case "Thread/get":
			respArgs = m.handleThreadGet(call[1])
		default:
			respArgs = map[string]any{"type": "unknownMethod"}
		}

		inv, _ := json.Marshal([]any{methodName, respArgs, callID})
		responses = append(responses, inv)
	}

	resp := map[string]any{
		"methodResponses": responses,
		"sessionState":    "mock-state-1",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (m *mockJMAP) handleEmailQuery(args json.RawMessage) map[string]any {
	// Return all email IDs (simple mock — ignores filters).
	ids := make([]string, len(m.emails))
	for i, e := range m.emails {
		ids[i] = e.ID
	}
	return map[string]any{
		"accountId":  "u12345",
		"queryState": "qs-1",
		"ids":        ids,
		"position":   0,
		"total":      len(ids),
	}
}

func (m *mockJMAP) handleEmailGet(args json.RawMessage) map[string]any {
	// Parse requested IDs.
	var params struct {
		IDs []string `json:"ids"`
	}
	json.Unmarshal(args, &params)

	// If IDs is nil (result reference), return all emails.
	if params.IDs == nil {
		return m.emailGetResponse(m.emails)
	}

	// Filter to requested IDs.
	idSet := make(map[string]bool)
	for _, id := range params.IDs {
		idSet[id] = true
	}

	var found []mockEmail
	var notFound []string
	for _, e := range m.emails {
		if idSet[e.ID] {
			found = append(found, e)
			delete(idSet, e.ID)
		}
	}
	for id := range idSet {
		notFound = append(notFound, id)
	}

	resp := m.emailGetResponse(found)
	if len(notFound) > 0 {
		resp["notFound"] = notFound
	}
	return resp
}

func (m *mockJMAP) emailGetResponse(emails []mockEmail) map[string]any {
	list := make([]map[string]any, 0, len(emails))
	for _, e := range emails {
		em := map[string]any{
			"id":            e.ID,
			"threadId":      e.ThreadID,
			"subject":       e.Subject,
			"preview":       e.Preview,
			"from":          e.From,
			"to":            e.To,
			"sentAt":        e.SentAt,
			"mailboxIds":    e.MailboxIDs,
			"size":          e.Size,
			"hasAttachment": e.HasAttachment,
		}
		if e.MessageID != "" {
			em["messageId"] = []string{e.MessageID}
		}
		if len(e.TextBody) > 0 {
			em["textBody"] = e.TextBody
		}
		if len(e.HTMLBody) > 0 {
			em["htmlBody"] = e.HTMLBody
		}
		if len(e.BodyValues) > 0 {
			em["bodyValues"] = e.BodyValues
		}
		list = append(list, em)
	}

	return map[string]any{
		"accountId": "u12345",
		"state":     "s-1",
		"list":      list,
		"notFound":  []string{},
	}
}

func (m *mockJMAP) handleMailboxGet() map[string]any {
	list := make([]map[string]any, 0, len(m.mailboxes))
	for _, mb := range m.mailboxes {
		entry := map[string]any{
			"id":           mb.ID,
			"name":         mb.Name,
			"totalEmails":  mb.TotalEmails,
			"unreadEmails": mb.UnreadEmails,
		}
		if mb.Role != "" {
			entry["role"] = mb.Role
		}
		if mb.ParentID != "" {
			entry["parentId"] = mb.ParentID
		}
		list = append(list, entry)
	}

	return map[string]any{
		"accountId": "u12345",
		"state":     "s-1",
		"list":      list,
		"notFound":  []string{},
	}
}

func (m *mockJMAP) handleThreadGet(args json.RawMessage) map[string]any {
	var params struct {
		IDs []string `json:"ids"`
	}
	json.Unmarshal(args, &params)

	idSet := make(map[string]bool)
	for _, id := range params.IDs {
		idSet[id] = true
	}

	var found []map[string]any
	var notFound []string
	for _, t := range m.threads {
		if idSet[t.ID] {
			found = append(found, map[string]any{
				"id":       t.ID,
				"emailIds": t.EmailIDs,
			})
			delete(idSet, t.ID)
		}
	}
	for id := range idSet {
		notFound = append(notFound, id)
	}

	return map[string]any{
		"accountId": "u12345",
		"state":     "s-1",
		"list":      found,
		"notFound":  notFound,
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// fmBinary returns the path to the fm binary. It must be built before tests.
var fmBinary string

func TestMain(m *testing.M) {
	// Build the binary once for all tests.
	tmpDir, err := os.MkdirTemp("", "fm-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fmBinary = filepath.Join(tmpDir, "fm")
	cmd := exec.Command("go", "build", "-o", fmBinary, "..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build fm binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runFM runs the fm CLI with the given args, using the mock server endpoint and token.
// Returns stdout, stderr, and any error.
func runFM(t *testing.T, mock *mockJMAP, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// Prepend the endpoint and token flags.
	fullArgs := append([]string{
		"--endpoint", mock.SessionURL(),
		"--token", "fmu1-test-token",
	}, args...)

	cmd := exec.Command(fmBinary, fullArgs...)
	// Clear environment to avoid picking up real FM_API_TOKEN.
	cmd.Env = []string{
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
	}

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	return outBuf.String(), errBuf.String(), runErr
}

// runFMWithCacheDir is like runFM but also sets --cache-dir.
func runFMWithCacheDir(t *testing.T, mock *mockJMAP, cacheDir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	fullArgs := append([]string{
		"--endpoint", mock.SessionURL(),
		"--token", "fmu1-test-token",
		"--cache-dir", cacheDir,
	}, args...)

	cmd := exec.Command(fmBinary, fullArgs...)
	cmd.Env = []string{
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
	}

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	return outBuf.String(), errBuf.String(), runErr
}

// runFMNoToken runs the fm CLI without a token, to test auth error handling.
func runFMNoToken(t *testing.T, mock *mockJMAP, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	fullArgs := append([]string{
		"--endpoint", mock.SessionURL(),
	}, args...)

	cmd := exec.Command(fmBinary, fullArgs...)
	// Clear environment to ensure no FM_API_TOKEN is set.
	cmd.Env = []string{
		"HOME=" + t.TempDir(),
		"PATH=" + os.Getenv("PATH"),
	}

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	return outBuf.String(), errBuf.String(), runErr
}

// ---------------------------------------------------------------------------
// Auth check command tests
// ---------------------------------------------------------------------------

func TestAuthCheckSuccess(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "auth", "check")
	if err != nil {
		t.Fatalf("auth check command failed: %v", err)
	}

	if !strings.Contains(stdout, "Token source:") {
		t.Error("auth check output should contain token source")
	}
	if !strings.Contains(stdout, "Token:") {
		t.Error("auth check output should contain masked token")
	}
	if !strings.Contains(stdout, "Account ID:") {
		t.Error("auth check output should contain account ID")
	}
	if !strings.Contains(stdout, "u12345") {
		t.Error("auth check output should contain the account ID value")
	}
	if !strings.Contains(stdout, "Username:") {
		t.Error("auth check output should contain username")
	}
	if !strings.Contains(stdout, "test@fastmail.com") {
		t.Error("auth check output should contain the username value")
	}
}

func TestAuthCheckFailure(t *testing.T) {
	mock := newMockJMAP()
	mock.failAuth = true
	defer mock.Close()

	_, stderr, err := runFM(t, mock, "auth", "check")
	if err == nil {
		t.Fatal("expected error when authentication fails")
	}

	if !strings.Contains(stderr, "authentication failed") && !strings.Contains(stderr, "Authentication failed") {
		t.Errorf("error should mention authentication failure, got: %s", stderr)
	}
}

// ---------------------------------------------------------------------------
// Search command tests
// ---------------------------------------------------------------------------

func TestSearchTextOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "search", "test")
	if err != nil {
		t.Fatalf("search command failed: %v", err)
	}

	// Verify text output contains expected headers and data.
	if !strings.Contains(stdout, "DATE") {
		t.Error("text output should contain DATE header")
	}
	if !strings.Contains(stdout, "FROM") {
		t.Error("text output should contain FROM header")
	}
	if !strings.Contains(stdout, "SUBJECT") {
		t.Error("text output should contain SUBJECT header")
	}
	if !strings.Contains(stdout, "Test Email One") {
		t.Error("text output should contain email subject")
	}
	if !strings.Contains(stdout, "Alice Sender") {
		t.Error("text output should contain sender name")
	}
}

func TestSearchJSONOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "search", "-o", "json", "test")
	if err != nil {
		t.Fatalf("search command failed: %v", err)
	}

	// Verify output is valid JSON.
	var result []map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("search JSON output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 emails in JSON output, got %d", len(result))
	}

	// Verify key fields are present.
	if result[0]["subject"] != "Test Email One" {
		t.Errorf("unexpected subject: %v", result[0]["subject"])
	}
}

func TestSearchMarkdownOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "search", "-o", "markdown", "test")
	if err != nil {
		t.Fatalf("search command failed: %v", err)
	}

	// Markdown email list uses a table format with pipe delimiters.
	if !strings.Contains(stdout, "| Date |") {
		t.Error("markdown output should contain table header")
	}
	if !strings.Contains(stdout, "Test Email One") {
		t.Error("markdown output should contain email subject")
	}
}

// ---------------------------------------------------------------------------
// Fetch command tests
// ---------------------------------------------------------------------------

func TestFetchEmail(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "fetch", "e001")
	if err != nil {
		t.Fatalf("fetch command failed: %v", err)
	}

	if !strings.Contains(stdout, "Test Email One") {
		t.Error("fetch output should contain email subject")
	}
	if !strings.Contains(stdout, "alice@example.com") {
		t.Error("fetch output should contain sender email")
	}
	if !strings.Contains(stdout, "Hello, this is the body of test email one.") {
		t.Error("fetch output should contain email body")
	}
}

func TestFetchJSONOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "fetch", "-o", "json", "e001")
	if err != nil {
		t.Fatalf("fetch command failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("fetch JSON output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	if result["id"] != "e001" {
		t.Errorf("expected id 'e001', got %v", result["id"])
	}
}

func TestFetchMarkdownOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "fetch", "-o", "markdown", "e001")
	if err != nil {
		t.Fatalf("fetch command failed: %v", err)
	}

	if !strings.Contains(stdout, "Test Email One") {
		t.Error("markdown output should contain email subject")
	}
}

func TestFetchCaching(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	cacheDir := t.TempDir()

	// First fetch: should hit the API.
	stdout1, _, err := runFMWithCacheDir(t, mock, cacheDir, "fetch", "e001")
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if !strings.Contains(stdout1, "Test Email One") {
		t.Error("first fetch should contain email subject")
	}

	apiCallsAfterFirst := mock.apiCalls.Load()
	if apiCallsAfterFirst == 0 {
		t.Error("first fetch should have hit the API")
	}

	// Verify cache file exists.
	cacheFile := filepath.Join(cacheDir, "e001.md")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Error("cache file should exist after fetch")
	}

	// Second fetch: should use cache (no additional API call).
	stdout2, _, err := runFMWithCacheDir(t, mock, cacheDir, "fetch", "e001")
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}
	if !strings.Contains(stdout2, "Test Email One") {
		t.Error("second fetch should still contain email subject (from cache)")
	}

	apiCallsAfterSecond := mock.apiCalls.Load()
	if apiCallsAfterSecond != apiCallsAfterFirst {
		t.Errorf("second fetch should not hit API (got %d calls, expected %d)",
			apiCallsAfterSecond, apiCallsAfterFirst)
	}
}

func TestFetchNoCacheFlag(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	cacheDir := t.TempDir()

	// First fetch to populate cache.
	_, _, err := runFMWithCacheDir(t, mock, cacheDir, "fetch", "e001")
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}

	apiCallsAfterFirst := mock.apiCalls.Load()

	// Second fetch with --no-cache should hit API again.
	_, _, err = runFMWithCacheDir(t, mock, cacheDir, "fetch", "--no-cache", "e001")
	if err != nil {
		t.Fatalf("second fetch with --no-cache failed: %v", err)
	}

	apiCallsAfterSecond := mock.apiCalls.Load()
	if apiCallsAfterSecond <= apiCallsAfterFirst {
		t.Error("--no-cache should force an API call even when cache exists")
	}
}

func TestFetchCacheFrontmatter(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	cacheDir := t.TempDir()

	_, _, err := runFMWithCacheDir(t, mock, cacheDir, "fetch", "e001")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "e001.md")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("reading cache file: %v", err)
	}

	content := string(data)
	// Verify frontmatter structure.
	if !strings.HasPrefix(content, "---\n") {
		t.Error("cache file should start with YAML frontmatter delimiter")
	}
	if !strings.Contains(content, "tool: fm") {
		t.Error("frontmatter should contain 'tool: fm'")
	}
	if !strings.Contains(content, "id: e001") {
		t.Error("frontmatter should contain email ID")
	}
	if !strings.Contains(content, "subject: Test Email One") {
		t.Error("frontmatter should contain email subject")
	}
	if !strings.Contains(content, "from: alice@example.com") {
		t.Error("frontmatter should contain sender email")
	}
	if !strings.Contains(content, "cached_at:") {
		t.Error("frontmatter should contain cached_at timestamp")
	}
}

// ---------------------------------------------------------------------------
// Mailboxes command tests
// ---------------------------------------------------------------------------

func TestMailboxesTextOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "mailboxes")
	if err != nil {
		t.Fatalf("mailboxes command failed: %v", err)
	}

	// Verify all mailboxes appear in output.
	for _, mb := range mock.mailboxes {
		if !strings.Contains(stdout, mb.Name) {
			t.Errorf("mailboxes output should contain %q", mb.Name)
		}
	}

	// Verify table headers.
	if !strings.Contains(stdout, "NAME") {
		t.Error("text output should contain NAME header")
	}
	if !strings.Contains(stdout, "ROLE") {
		t.Error("text output should contain ROLE header")
	}
}

func TestMailboxesJSONOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "mailboxes", "-o", "json")
	if err != nil {
		t.Fatalf("mailboxes command failed: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("mailboxes JSON output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	if len(result) != 5 {
		t.Errorf("expected 5 mailboxes in JSON output, got %d", len(result))
	}
}

func TestMailboxesMarkdownOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "mailboxes", "-o", "markdown")
	if err != nil {
		t.Fatalf("mailboxes command failed: %v", err)
	}

	if !strings.Contains(stdout, "Inbox") {
		t.Error("markdown output should contain mailbox names")
	}
}

// ---------------------------------------------------------------------------
// Fetch-thread command tests
// ---------------------------------------------------------------------------

func TestFetchThread(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, stderr, err := runFM(t, mock, "fetch-thread", "t001")
	if err != nil {
		t.Fatalf("fetch-thread command failed: %v\nstderr: %s", err, stderr)
	}

	// Both emails should appear in output.
	if !strings.Contains(stdout, "Test Email One") {
		t.Error("fetch-thread output should contain first email subject")
	}
	if !strings.Contains(stdout, "Re: Test Email One") {
		t.Error("fetch-thread output should contain second email subject")
	}

	// Thread summary should appear on stderr.
	if !strings.Contains(stderr, "2 message(s)") {
		t.Errorf("stderr should show message count, got: %s", stderr)
	}
}

func TestFetchThreadJSONOutput(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	stdout, _, err := runFM(t, mock, "fetch-thread", "-o", "json", "t001")
	if err != nil {
		t.Fatalf("fetch-thread command failed: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("fetch-thread JSON output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 emails in JSON output, got %d", len(result))
	}
}

func TestFetchThreadCachesEmails(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	cacheDir := t.TempDir()

	_, _, err := runFMWithCacheDir(t, mock, cacheDir, "fetch-thread", "t001")
	if err != nil {
		t.Fatalf("fetch-thread failed: %v", err)
	}

	// Both emails should be cached.
	for _, id := range []string{"e001", "e002"} {
		cacheFile := filepath.Join(cacheDir, id+".md")
		if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
			t.Errorf("cache file for email %s should exist after fetch-thread", id)
		}
	}
}

// ---------------------------------------------------------------------------
// Error scenario tests
// ---------------------------------------------------------------------------

func TestNoTokenError(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	_, stderr, err := runFMNoToken(t, mock, "search", "test")
	if err == nil {
		t.Fatal("expected error when no token is provided")
	}

	// Should show actionable error message.
	if !strings.Contains(stderr, "No API token found") {
		t.Errorf("error should mention missing token, got: %s", stderr)
	}
	if !strings.Contains(stderr, "fm auth login") {
		t.Errorf("error should suggest 'fm auth login', got: %s", stderr)
	}
}

func TestInvalidEmailIDError(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	_, stderr, err := runFM(t, mock, "fetch", "nonexistent-email-id")
	if err == nil {
		t.Fatal("expected error for invalid email ID")
	}

	if !strings.Contains(stderr, "not found") {
		t.Errorf("error should indicate email not found, got: %s", stderr)
	}
}

func TestInvalidThreadIDError(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	_, stderr, err := runFM(t, mock, "fetch-thread", "nonexistent-thread-id")
	if err == nil {
		t.Fatal("expected error for invalid thread ID")
	}

	if !strings.Contains(stderr, "not found") {
		t.Errorf("error should indicate thread not found, got: %s", stderr)
	}
}

// ---------------------------------------------------------------------------
// Output format validation tests
// ---------------------------------------------------------------------------

func TestAllCommandsJSONValid(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	cacheDir := t.TempDir()

	tests := []struct {
		name string
		args []string
	}{
		{"search", []string{"search", "-o", "json", "test"}},
		{"fetch", []string{"fetch", "-o", "json", "e001"}},
		{"mailboxes", []string{"mailboxes", "-o", "json"}},
		{"fetch-thread", []string{"fetch-thread", "-o", "json", "t001"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runFMWithCacheDir(t, mock, cacheDir, tt.args...)
			if err != nil {
				t.Fatalf("command %v failed: %v\nstderr: %s", tt.args, err, stderr)
			}

			if !json.Valid([]byte(stdout)) {
				t.Errorf("output is not valid JSON:\n%s", stdout)
			}
		})
	}
}

func TestAllOutputFormats(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	formats := []string{"text", "json", "markdown"}

	for _, format := range formats {
		t.Run("search-"+format, func(t *testing.T) {
			stdout, _, err := runFM(t, mock, "search", "-o", format, "test")
			if err != nil {
				t.Fatalf("search -o %s failed: %v", format, err)
			}
			if stdout == "" {
				t.Errorf("search -o %s produced empty output", format)
			}
		})

		t.Run("fetch-"+format, func(t *testing.T) {
			stdout, _, err := runFM(t, mock, "fetch", "-o", format, "e001")
			if err != nil {
				t.Fatalf("fetch -o %s failed: %v", format, err)
			}
			if stdout == "" {
				t.Errorf("fetch -o %s produced empty output", format)
			}
		})

		t.Run("mailboxes-"+format, func(t *testing.T) {
			stdout, _, err := runFM(t, mock, "mailboxes", "-o", format)
			if err != nil {
				t.Fatalf("mailboxes -o %s failed: %v", format, err)
			}
			if stdout == "" {
				t.Errorf("mailboxes -o %s produced empty output", format)
			}
		})

		t.Run("fetch-thread-"+format, func(t *testing.T) {
			stdout, _, err := runFM(t, mock, "fetch-thread", "-o", format, "t001")
			if err != nil {
				t.Fatalf("fetch-thread -o %s failed: %v", format, err)
			}
			if stdout == "" {
				t.Errorf("fetch-thread -o %s produced empty output", format)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// API call counting tests
// ---------------------------------------------------------------------------

func TestSearchHitsAPI(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	callsBefore := mock.apiCalls.Load()
	_, _, err := runFM(t, mock, "search", "test")
	if err != nil {
		t.Fatalf("search command failed: %v", err)
	}

	callsAfter := mock.apiCalls.Load()
	if callsAfter <= callsBefore {
		t.Error("search should have made at least one API call")
	}
}

func TestMailboxesHitsAPI(t *testing.T) {
	mock := newMockJMAP()
	defer mock.Close()

	callsBefore := mock.apiCalls.Load()
	_, _, err := runFM(t, mock, "mailboxes")
	if err != nil {
		t.Fatalf("mailboxes command failed: %v", err)
	}

	callsAfter := mock.apiCalls.Load()
	if callsAfter <= callsBefore {
		t.Error("mailboxes should have made at least one API call")
	}
}

// ---------------------------------------------------------------------------
// Help and version tests (no server needed)
// ---------------------------------------------------------------------------

func TestHelpCommand(t *testing.T) {
	cmd := exec.Command(fmBinary, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "fm") {
		t.Error("help output should mention fm")
	}
	if !strings.Contains(output, "search") {
		t.Error("help output should mention search command")
	}
	if !strings.Contains(output, "fetch") {
		t.Error("help output should mention fetch command")
	}
	if !strings.Contains(output, "mailboxes") {
		t.Error("help output should mention mailboxes command")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(fmBinary, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.Contains(string(out), "fm version") {
		t.Errorf("version output should contain 'fm version', got: %s", string(out))
	}
}
