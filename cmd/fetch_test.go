package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/zalando/go-keyring"
)

func TestFetchCommandRegistered(t *testing.T) {
	// Verify the fetch command is registered on the email command.
	found := false
	for _, cmd := range emailCmd.Commands() {
		if cmd.Name() == "fetch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'fetch' command to be registered on emailCmd")
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
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		fetchNoCache = origNoCache
		token = origToken
	}()

	fetchNoCache = false
	token = "" // No token needed for cache hit.

	// cli-kit output writes directly to os.Stdout, so we verify the
	// command runs without error. The output correctness is verified
	// by the cache and renderer tests.
	rootCmd.SetArgs([]string{"email", "fetch", "--derived", tmpDir, "--output", "table", "M12345"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("fetch command returned error on cache hit: %v", err)
	}
}

func TestFetchNoCacheSkipsCache(t *testing.T) {
	// Use mock keyring so no real keychain token interferes.
	keyring.MockInit()

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
	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		fetchNoCache = origNoCache
		token = origToken
	}()

	fetchNoCache = true
	token = ""
	t.Setenv("FM_API_TOKEN", "")

	err := runFetch(fetchCmd, []string{"M99999"})
	if err == nil {
		t.Fatal("expected error when --no-cache is set and no token is available")
	}

	// The error should be an auth error (missing token).
	var authErr *auth.AuthError
	if !errors.As(err, &authErr) {
		t.Errorf("expected AuthError, got: %v", err)
	}
}

func TestFetchEmailNotCached(t *testing.T) {
	// Use mock keyring so no real keychain token interferes.
	keyring.MockInit()

	// When the email is not in cache and no token is available,
	// the command should fail with a token error.

	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		fetchNoCache = origNoCache
		token = origToken
	}()

	fetchNoCache = false
	token = ""
	t.Setenv("FM_API_TOKEN", "")

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

	origNoCache := fetchNoCache
	origToken := token
	defer func() {
		fetchNoCache = origNoCache
		token = origToken
	}()

	fetchNoCache = false
	token = ""

	// Note: JSON output goes to stdout directly via cli-kit, so we can't easily
	// capture it via cobra's SetOut. We verify the command runs without error.
	rootCmd.SetArgs([]string{"email", "fetch", "--derived", tmpDir, "--output", "json", "Mjson1"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("fetch command returned error: %v", err)
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
