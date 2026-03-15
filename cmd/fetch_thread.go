package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/config"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/output"
	"github.com/spf13/cobra"
)

var withAttachments bool

var fetchThreadCmd = &cobra.Command{
	Use:   "fetch-thread <thread-id>",
	Short: "Fetch all emails in a thread",
	Long: `Fetch all emails in a JMAP thread by thread ID.

Retrieves the thread metadata and all emails in the thread, caches each
email individually, and displays them in chronological order. Shows a
thread summary (number of messages, participants) on stderr.

Use --with-attachments to download all attachments for every email in the thread.`,
	Args: cobra.ExactArgs(1),
	RunE: runFetchThread,
}

func init() {
	rootCmd.AddCommand(fetchThreadCmd)
	fetchThreadCmd.Flags().BoolVar(&withAttachments, "with-attachments", false,
		"Download attachments for all emails in the thread")
}

func runFetchThread(cmd *cobra.Command, args []string) error {
	threadID := args[0]

	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return fmt.Errorf("resolving token: %w", err)
	}

	client := jmap.NewClient(tok, jmap.WithTimeout(timeout))
	ctx := cmd.Context()

	if err := client.Discover(); err != nil {
		return fmt.Errorf("session discovery: %w", err)
	}

	// Fetch all emails in the thread, sorted chronologically.
	emails, err := client.GetThreadEmails(ctx, threadID)
	if err != nil {
		return fmt.Errorf("fetching thread: %w", err)
	}

	if len(emails) == 0 {
		return fmt.Errorf("thread %q contains no emails", threadID)
	}

	// Load config for cache directory and output format.
	cfg, err := config.Load(cacheDir, outputFormat)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Cache each email individually.
	c := cache.NewCache(cfg.CacheDir)
	for _, email := range emails {
		if err := c.Put(email, "fm fetch-thread "+threadID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cache email %s: %v\n", email.Id, err)
		}
	}

	// Download attachments if requested.
	if withAttachments {
		accountID, err := client.PrimaryAccountID()
		if err != nil {
			return fmt.Errorf("getting account ID for attachment download: %w", err)
		}

		for _, email := range emails {
			for _, att := range email.Attachments {
				attDir := filepath.Join(cfg.CacheDir, "attachments", email.Id)
				if err := os.MkdirAll(attDir, 0o755); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create attachment dir: %v\n", err)
					continue
				}

				name := att.Name
				if name == "" {
					name = att.BlobId
				}

				data, err := client.DownloadAttachment(ctx, string(accountID), att.BlobId, name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to download attachment %q: %v\n", name, err)
					continue
				}

				attPath := filepath.Join(attDir, name)
				if err := os.WriteFile(attPath, data, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save attachment %q: %v\n", attPath, err)
					continue
				}

				fmt.Fprintf(os.Stderr, "Saved attachment: %s\n", attPath)
			}
		}
	}

	// Print thread summary to stderr.
	participants := collectParticipants(emails)
	fmt.Fprintf(os.Stderr, "Thread: %d message(s), %d participant(s)\n", len(emails), len(participants))
	if len(participants) > 0 {
		fmt.Fprintf(os.Stderr, "Participants: %s\n", strings.Join(participants, ", "))
	}

	// Format and display all emails.
	formatter, err := output.New(cfg.OutputFormat)
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}

	result, err := formatter.FormatEmailList(emails)
	if err != nil {
		return fmt.Errorf("formatting thread emails: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), result)
	return nil
}

// collectParticipants extracts unique participant names/emails from all
// emails in a thread (from From, To, and Cc fields).
func collectParticipants(emails []jmap.Email) []string {
	seen := make(map[string]bool)
	var participants []string

	for _, email := range emails {
		for _, addr := range email.From {
			key := addr.Email
			if !seen[key] {
				seen[key] = true
				if addr.Name != "" {
					participants = append(participants, fmt.Sprintf("%s <%s>", addr.Name, addr.Email))
				} else {
					participants = append(participants, addr.Email)
				}
			}
		}
		for _, addr := range email.To {
			key := addr.Email
			if !seen[key] {
				seen[key] = true
				if addr.Name != "" {
					participants = append(participants, fmt.Sprintf("%s <%s>", addr.Name, addr.Email))
				} else {
					participants = append(participants, addr.Email)
				}
			}
		}
		for _, addr := range email.Cc {
			key := addr.Email
			if !seen[key] {
				seen[key] = true
				if addr.Name != "" {
					participants = append(participants, fmt.Sprintf("%s <%s>", addr.Name, addr.Email))
				} else {
					participants = append(participants, addr.Email)
				}
			}
		}
	}

	return participants
}
