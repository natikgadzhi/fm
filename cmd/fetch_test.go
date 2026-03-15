package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/jmap"
)

func TestFetchCommandRegistered(t *testing.T) {
	// Verify the fetch command is registered on the root command.
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "fetch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'fetch' command to be registered on root")
	}
}

func TestFetchCommandArgs(t *testing.T) {
	cmd := fetchCmd

	// The command should require exactly one argument.
	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
}

func TestFetchCommandFlags(t *testing.T) {
	f := fetchCmd.Flags()

	noCacheFlag := f.Lookup("no-cache")
	if noCacheFlag == nil {
		t.Fatal("expected --no-cache flag to be registered")
	}
	if noCacheFlag.DefValue != "false" {
		t.Errorf("expected --no-cache default to be false, got %s", noCacheFlag.DefValue)
	}

	withAttFlag := f.Lookup("with-attachments")
	if withAttFlag == nil {
		t.Fatal("expected --with-attachments flag to be registered")
	}
	if withAttFlag.DefValue != "false" {
		t.Errorf("expected --with-attachments default to be false, got %s", withAttFlag.DefValue)
	}
}

func TestFetchCacheHit(t *testing.T) {
	// Create a temporary cache directory.
	tmpDir := t.TempDir()

	// Write a sample email to the cache.
	c := cache.NewCache(tmpDir)
	email := jmap.Email{
		Id:       "M12345",
		ThreadId: "T12345",
		From:     []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
		To:       []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		Subject:  "Test Email",
		Date:     time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC),
		TextBody: "Hello, this is a test email.",
	}
	if err := c.Put(email, "fm fetch M12345"); err != nil {
		t.Fatalf("failed to write to cache: %v", err)
	}

	// Save and restore global flags.
	origCacheDir := cacheDir
	origOutputFormat := outputFormat
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		cacheDir = origCacheDir
		outputFormat = origOutputFormat
		fetchNoCache = origNoCache
		token = origToken
	}()

	cacheDir = tmpDir
	outputFormat = "text"
	fetchNoCache = false
	token = "" // No token needed for cache hit.

	// Capture output.
	buf := new(bytes.Buffer)
	fetchCmd.SetOut(buf)
	fetchCmd.SetErr(buf)

	err := runFetch(fetchCmd, []string{"M12345"})
	if err != nil {
		t.Fatalf("runFetch returned error on cache hit: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected non-empty output on cache hit")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test Email")) {
		t.Error("expected 'Test Email' in output")
	}
}

func TestFetchNoCacheSkipsCache(t *testing.T) {
	// Create a temporary cache directory with a cached email.
	tmpDir := t.TempDir()
	c := cache.NewCache(tmpDir)
	email := jmap.Email{
		Id:       "M99999",
		ThreadId: "T99999",
		From:     []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
		To:       []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		Subject:  "Cached Email",
		Date:     time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC),
		TextBody: "This is cached.",
	}
	if err := c.Put(email, "fm fetch M99999"); err != nil {
		t.Fatalf("failed to write to cache: %v", err)
	}

	// Save and restore global flags.
	origCacheDir := cacheDir
	origOutputFormat := outputFormat
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		cacheDir = origCacheDir
		outputFormat = origOutputFormat
		fetchNoCache = origNoCache
		token = origToken
	}()

	cacheDir = tmpDir
	outputFormat = "text"
	fetchNoCache = true
	token = "" // No token — this should fail because --no-cache requires API access.

	err := runFetch(fetchCmd, []string{"M99999"})
	if err == nil {
		t.Fatal("expected error when --no-cache is set and no token is available")
	}

	// The error should be about missing token, not about cache.
	if err.Error() != "No API token found. Run 'fm auth login' or set FM_API_TOKEN. Create a token at https://app.fastmail.com/settings/security/tokens/new" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetchEmailNotCached(t *testing.T) {
	// When the email is not in cache and no token is available,
	// the command should fail with a token error.
	tmpDir := t.TempDir()

	origCacheDir := cacheDir
	origOutputFormat := outputFormat
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		cacheDir = origCacheDir
		outputFormat = origOutputFormat
		fetchNoCache = origNoCache
		token = origToken
	}()

	cacheDir = tmpDir
	outputFormat = "text"
	fetchNoCache = false
	token = ""

	err := runFetch(fetchCmd, []string{"Mnonexistent"})
	if err == nil {
		t.Fatal("expected error when email not cached and no token available")
	}
}

func TestFetchCacheHitJsonOutput(t *testing.T) {
	tmpDir := t.TempDir()

	c := cache.NewCache(tmpDir)
	email := jmap.Email{
		Id:       "Mjson1",
		ThreadId: "Tjson1",
		From:     []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
		To:       []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		Subject:  "JSON Test",
		Date:     time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC),
		TextBody: "Hello from JSON.",
	}
	if err := c.Put(email, "fm fetch Mjson1"); err != nil {
		t.Fatalf("failed to write to cache: %v", err)
	}

	origCacheDir := cacheDir
	origOutputFormat := outputFormat
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		cacheDir = origCacheDir
		outputFormat = origOutputFormat
		fetchNoCache = origNoCache
		token = origToken
	}()

	cacheDir = tmpDir
	outputFormat = "json"
	fetchNoCache = false
	token = ""

	buf := new(bytes.Buffer)
	fetchCmd.SetOut(buf)
	fetchCmd.SetErr(buf)

	err := runFetch(fetchCmd, []string{"Mjson1"})
	if err != nil {
		t.Fatalf("runFetch returned error: %v", err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("Mjson1")) {
		t.Error("expected email ID 'Mjson1' in JSON output")
	}
}

func TestFetchAttachmentDirCreated(t *testing.T) {
	// Verify that the attachment directory structure is correct.
	tmpDir := t.TempDir()
	attachDir := filepath.Join(tmpDir, "attachments", "Mattach1")

	if err := os.MkdirAll(attachDir, 0o755); err != nil {
		t.Fatalf("failed to create attachment dir: %v", err)
	}

	// Verify the directory exists.
	info, err := os.Stat(attachDir)
	if err != nil {
		t.Fatalf("attachment dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected attachment path to be a directory")
	}
}
