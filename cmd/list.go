package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/output"
	"github.com/spf13/cobra"
)

var (
	listLimit  int
	listAfter  string
	listBefore string
)

var listCmd = &cobra.Command{
	Use:   "list <mailbox>",
	Short: "List emails in a mailbox sorted by recency",
	Long: `List all emails in a mailbox, sorted by most recent first.

Use --after and --before to restrict results to a time interval, for example
to fetch all emails from the past week and then retrieve each one by ID.`,
	Example: `  # List the 50 most recent emails in INBOX
  fm list INBOX --limit 50

  # List emails received in the last week
  fm list INBOX --after 2025-06-01

  # List emails within a date range
  fm list Archive --after 2025-01-01 --before 2025-02-01

  # Output as JSON for scripting
  fm list INBOX --after 2025-06-01 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntVar(&listLimit, "limit", 50, "Maximum number of emails to return")
	listCmd.Flags().StringVar(&listAfter, "after", "", "Only emails received after this date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listBefore, "before", "", "Only emails received before this date (YYYY-MM-DD)")
}

func runList(cmd *cobra.Command, args []string) error {
	mailbox := args[0]

	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	filter := jmap.SearchFilter{
		InMailbox: mailbox,
	}

	if listAfter != "" {
		t, err := time.Parse("2006-01-02", listAfter)
		if err != nil {
			return fmt.Errorf("invalid --after date %q: expected YYYY-MM-DD", listAfter)
		}
		filter.After = &t
	}

	if listBefore != "" {
		t, err := time.Parse("2006-01-02", listBefore)
		if err != nil {
			return fmt.Errorf("invalid --before date %q: expected YYYY-MM-DD", listBefore)
		}
		filter.Before = &t
	}

	ctx := cmd.Context()
	emails, err := client.SearchEmails(ctx, filter, listLimit)

	var partialErr *jmap.PartialResultError
	if errors.As(err, &partialErr) {
		emails = partialErr.Emails
		fmt.Fprintf(os.Stderr, "Warning: partial results — fetched %d of %d emails: %v\n",
			partialErr.Fetched, partialErr.Total, partialErr.Err)
	} else if err != nil {
		return fmt.Errorf("listing emails: %w", err)
	}

	if len(emails) == 0 {
		fmt.Fprintln(os.Stderr, "No emails found.")
		return nil
	}

	formatter, err := output.New(outputFormat)
	if err != nil {
		return err
	}

	out, err := formatter.FormatEmailList(emails)
	if err != nil {
		return fmt.Errorf("formatting results: %w", err)
	}
	fmt.Fprint(cmd.OutOrStdout(), out)

	fmt.Fprintf(os.Stderr, "\nFound %d email(s).\n", len(emails))

	return nil
}
