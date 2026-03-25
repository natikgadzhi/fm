package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natikgadzhi/cli-kit/derived"
	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/verbose"
	"github.com/spf13/cobra"
)

var (
	fetchNoCache         bool
	fetchWithAttachments bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch <email-id>",
	Short: "Fetch a single email by its JMAP ID",
	Long: `Fetch downloads and displays a single email by its JMAP email ID.

By default, fetched emails are cached locally as Markdown files.
Subsequent fetches of the same ID are served from the cache.

Use --no-cache to bypass the cache and always fetch from the server.
Use --with-attachments to download email attachments to the derived directory.`,
	Args: cobra.ExactArgs(1),
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().BoolVar(&fetchNoCache, "no-cache", false,
		"Bypass the cache and fetch directly from the server")
	fetchCmd.Flags().BoolVar(&fetchWithAttachments, "with-attachments", false,
		"Download email attachments to the derived directory")
}

func runFetch(cmd *cobra.Command, args []string) error {
	emailID := args[0]

	derivedDir := derived.Resolve(cmd, "fm")
	format := output.Resolve(cmd)

	c := cache.NewCache(derivedDir)

	// Try the cache first (unless --no-cache).
	// When --with-attachments is requested, skip the cache because cached
	// emails do not store attachment metadata (HasAttachment, Attachments
	// are empty), so we must always fetch from the API to get blob IDs.
	if !fetchNoCache && !fetchWithAttachments {
		if cached, err := c.Get(emailID); err == nil && cached != nil {
			verbose.Log("cache hit for email %s", emailID)
			renderer := &jmap.EmailRenderer{Email: *cached}
			if err := output.Print(format, *cached, renderer); err != nil {
				return fmt.Errorf("formatting email: %w", err)
			}
			return nil
		}
		verbose.Log("cache miss for email %s", emailID)
	}

	// Resolve token and create JMAP client.
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	ctx := cmd.Context()
	emails, err := client.GetEmails(ctx, []string{emailID})

	// Check for partial results.
	var partialErr *jmap.PartialResultError
	if errors.As(err, &partialErr) {
		emails = partialErr.Emails
		fmt.Fprintf(os.Stderr, "Warning: partial results — %v\n", partialErr.Err)
	} else if err != nil {
		return fmt.Errorf("fetching email: %w", err)
	}

	if len(emails) == 0 {
		return fmt.Errorf("email %q not found. Verify the message ID is correct", emailID)
	}

	email := emails[0]

	// Cache the email.
	if putErr := c.Put(email, "fm fetch "+emailID); putErr != nil {
		// Log but don't fail — caching is best-effort.
		fmt.Fprintf(os.Stderr, "Warning: failed to cache email: %v\n", putErr)
	}

	// Download attachments if requested.
	if fetchWithAttachments && len(email.Attachments) > 0 {
		if dlErr := downloadAttachments(cmd, client, derivedDir, email); dlErr != nil {
			return dlErr
		}
	}

	renderer := &jmap.EmailRenderer{Email: email}
	if err := output.Print(format, email, renderer); err != nil {
		return fmt.Errorf("formatting email: %w", err)
	}
	return nil
}

// downloadAttachments downloads each attachment for the given email
// and saves them to {derived-dir}/attachments/{email-id}/{filename}.
// The client must have an active session (Discover already called).
func downloadAttachments(cmd *cobra.Command, client *jmap.Client, derivedBaseDir string, email jmap.Email) error {
	accountID, err := client.PrimaryAccountID()
	if err != nil {
		return fmt.Errorf("getting account ID for attachment download: %w", err)
	}

	ctx := cmd.Context()
	attachDir := filepath.Join(derivedBaseDir, "attachments", email.Id)
	if err := os.MkdirAll(attachDir, 0o755); err != nil {
		return fmt.Errorf("creating attachment directory: %w", err)
	}

	for _, att := range email.Attachments {
		name := att.Name
		if name == "" {
			name = att.BlobId
		}

		data, err := client.DownloadAttachment(ctx, string(accountID), att.BlobId, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to download attachment %q: %v\n", name, err)
			continue
		}

		// Sanitize the filename to prevent path traversal attacks.
		safeName := filepath.Base(name)
		path := filepath.Join(attachDir, safeName)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save attachment %q: %v\n", name, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "Saved attachment: %s\n", path)
	}

	return nil
}
