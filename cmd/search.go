package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/output"
	"github.com/spf13/cobra"
)

var (
	searchLimit         int
	searchFrom          string
	searchTo            string
	searchHasAttachment bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search emails by query",
	Long: `Search emails using a query string with optional filter syntax.

Supported query syntax:
  from:user@example.com       Filter by sender
  to:user@example.com         Filter by recipient
  subject:meeting notes       Filter by subject
  in:INBOX                    Filter by mailbox
  before:2025-01-01           Emails before date
  after:2025-01-01            Emails after date
  has:attachment              Only emails with attachments
  <any text>                  Free text search

CLI flags (--from, --to, --has-attachments) override inline query filters.`,
	Example: `  fm search "from:boss@company.com subject:urgent after:2025-01-01"
  fm search --from boss@company.com --has-attachments
  fm search --from alice@example.com "subject:report"
  fm search meeting notes`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().IntVar(&searchLimit, "limit", 25, "Maximum number of results to return")
	searchCmd.Flags().StringVar(&searchFrom, "from", "", "Filter by sender email address")
	searchCmd.Flags().StringVar(&searchTo, "to", "", "Filter by recipient email address")
	searchCmd.Flags().BoolVar(&searchHasAttachment, "has-attachments", false, "Filter to only emails with attachments")
}

// MergeFilterFlags applies CLI flag values to a SearchFilter, overriding any
// values that were parsed from the query string. This is exported for testing.
func MergeFilterFlags(filter jmap.SearchFilter, from, to string, hasAttachments bool) jmap.SearchFilter {
	if from != "" {
		filter.From = from
	}
	if to != "" {
		filter.To = to
	}
	if hasAttachments {
		filter.HasAttachment = true
	}
	return filter
}

func runSearch(cmd *cobra.Command, args []string) error {
	// 1. Resolve token.
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	// 2. Create JMAP client.
	client := jmap.NewClient(tok, clientOpts()...)

	// 3. Parse query string into SearchFilter.
	query := strings.Join(args, " ")
	filter, err := jmap.ParseFilterQuery(query)
	if err != nil {
		return fmt.Errorf("invalid search query: %w", err)
	}

	// 4. Merge CLI flags into filter (flags override query string).
	filter = MergeFilterFlags(filter, searchFrom, searchTo, searchHasAttachment)

	// 5. Call SearchEmails using the command context (supports Ctrl+C cancellation).
	ctx := cmd.Context()
	emails, err := client.SearchEmails(ctx, filter, searchLimit)

	// Check for partial results — display what we got with a warning.
	var partialErr *jmap.PartialResultError
	if errors.As(err, &partialErr) {
		emails = partialErr.Emails
		fmt.Fprintf(os.Stderr, "Warning: partial results — fetched %d of %d emails: %v\n",
			partialErr.Fetched, partialErr.Total, partialErr.Err)
	} else if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// 6. Handle zero results.
	if len(emails) == 0 {
		fmt.Fprintln(os.Stderr, "No emails found matching the search criteria.")
		return nil
	}

	// 7. Create formatter from outputFormat global flag.
	formatter, err := output.New(outputFormat)
	if err != nil {
		return err
	}

	// 8. Format and print results to stdout.
	out, err := formatter.FormatEmailList(emails)
	if err != nil {
		return fmt.Errorf("formatting results: %w", err)
	}
	fmt.Fprint(cmd.OutOrStdout(), out)

	// 9. Print count to stderr.
	fmt.Fprintf(os.Stderr, "\nFound %d email(s).\n", len(emails))

	return nil
}
