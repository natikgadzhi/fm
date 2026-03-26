package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/natikgadzhi/cli-kit/derived"
	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/cache"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/spf13/cobra"
)

var threadWithAttachments bool

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
	emailCmd.AddCommand(fetchThreadCmd)
	fetchThreadCmd.Flags().BoolVar(&threadWithAttachments, "with-attachments", false,
		"Download attachments for all emails in the thread")
}

func runFetchThread(cmd *cobra.Command, args []string) error {
	threadID := args[0]

	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)
	ctx := cmd.Context()

	// Fetch all emails in the thread, sorted chronologically.
	// GetThreadEmails calls Discover internally.
	emails, err := client.GetThreadEmails(ctx, threadID)

	// Check for partial results — display what we got with a warning.
	var partialErr *jmap.PartialResultError
	if errors.As(err, &partialErr) {
		emails = partialErr.Emails
		fmt.Fprintf(os.Stderr, "Warning: partial results — fetched %d of %d emails: %v\n",
			partialErr.Fetched, partialErr.Total, partialErr.Err)
	} else if err != nil {
		return fmt.Errorf("fetching thread: %w", err)
	}

	if len(emails) == 0 {
		return fmt.Errorf("thread %q contains no emails", threadID)
	}

	// Resolve derived directory for caching.
	derivedDir := derived.Resolve(cmd, "fm")

	// Cache each email individually.
	c := cache.NewCache(derivedDir)
	for _, email := range emails {
		if err := c.Put(email, "fm email fetch-thread "+threadID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cache email %s: %v\n", email.Id, err)
		}
	}

	// Download attachments if requested.
	if threadWithAttachments {
		for _, email := range emails {
			if len(email.Attachments) == 0 {
				continue
			}
			if dlErr := downloadAttachments(cmd, client, derivedDir, email); dlErr != nil {
				return dlErr
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
	format := output.Resolve(cmd)
	renderer := &jmap.EmailListRenderer{Emails: emails}
	if err := output.Print(format, emails, renderer); err != nil {
		return fmt.Errorf("formatting thread emails: %w", err)
	}

	return nil
}

// collectParticipants extracts unique participant names/emails from all
// emails in a thread (from From, To, and Cc fields).
func collectParticipants(emails []jmap.Email) []string {
	seen := make(map[string]bool)
	var participants []string

	addAddresses := func(addrs []jmap.Address) {
		for _, addr := range addrs {
			if seen[addr.Email] {
				continue
			}
			seen[addr.Email] = true
			if addr.Name != "" {
				participants = append(participants, fmt.Sprintf("%s <%s>", addr.Name, addr.Email))
			} else {
				participants = append(participants, addr.Email)
			}
		}
	}

	for _, email := range emails {
		addAddresses(email.From)
		addAddresses(email.To)
		addAddresses(email.Cc)
	}

	return participants
}
