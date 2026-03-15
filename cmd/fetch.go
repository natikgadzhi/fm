package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/config"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/output"
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
Use --with-attachments to download email attachments to the cache directory.`,
	Args: cobra.ExactArgs(1),
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().BoolVar(&fetchNoCache, "no-cache", false,
		"Bypass the cache and fetch directly from the server")
	fetchCmd.Flags().BoolVar(&fetchWithAttachments, "with-attachments", false,
		"Download email attachments to the cache directory")
}

func runFetch(cmd *cobra.Command, args []string) error {
	emailID := args[0]

	cfg, err := config.Load(cacheDir, outputFormat)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	formatter, err := output.New(cfg.OutputFormat)
	if err != nil {
		return err
	}

	c := cache.NewCache(cfg.CacheDir)

	// Try the cache first (unless --no-cache).
	// When --with-attachments is requested, skip the cache because cached
	// emails do not store attachment metadata (HasAttachment, Attachments
	// are empty), so we must always fetch from the API to get blob IDs.
	if !fetchNoCache && !fetchWithAttachments {
		if cached, err := c.Get(emailID); err == nil && cached != nil {
			out, err := formatter.FormatEmail(*cached)
			if err != nil {
				return fmt.Errorf("formatting email: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		}
	}

	// Resolve token and create JMAP client.
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	ctx := context.Background()
	emails, err := client.GetEmails(ctx, []string{emailID})
	if err != nil {
		return fmt.Errorf("fetching email: %w", err)
	}

	if len(emails) == 0 {
		return fmt.Errorf("email not found: %s", emailID)
	}

	email := emails[0]

	// Cache the email.
	if putErr := c.Put(email, "fm fetch "+emailID); putErr != nil {
		// Log but don't fail — caching is best-effort.
		fmt.Fprintf(os.Stderr, "Warning: failed to cache email: %v\n", putErr)
	}

	// Download attachments if requested.
	if fetchWithAttachments && len(email.Attachments) > 0 {
		if dlErr := downloadAttachments(cmd, cfg, email); dlErr != nil {
			return dlErr
		}
	}

	out, err := formatter.FormatEmail(email)
	if err != nil {
		return fmt.Errorf("formatting email: %w", err)
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

// downloadAttachments downloads each attachment for the given email
// and saves them to {cache-dir}/attachments/{email-id}/{filename}.
func downloadAttachments(cmd *cobra.Command, cfg *config.Config, email jmap.Email) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	accountID, err := client.PrimaryAccountID()
	if err != nil {
		return fmt.Errorf("getting account ID for attachment download: %w", err)
	}

	ctx := context.Background()
	attachDir := filepath.Join(cfg.CacheDir, "attachments", email.Id)
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
		// A malicious name like "../../../etc/passwd" would otherwise
		// write outside the intended cache directory.
		safeName := filepath.Base(name)
		path := filepath.Join(attachDir, safeName)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save attachment %q: %v\n", name, err)
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Attachment saved: %s\n", path)
	}

	return nil
}
