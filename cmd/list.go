package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/verbose"
	"github.com/spf13/cobra"
)

var listLimit int

var listCmd = &cobra.Command{
	Use:   "list <mailbox>",
	Short: "List emails in a mailbox",
	Long: `List emails in a mailbox by name or JMAP ID, in reverse-chronological order.

Accepts mailbox names (case-insensitive) like "INBOX", "Sent", "Drafts",
or raw JMAP mailbox IDs. Returns the most recent emails first.`,
	Example: `  fm list INBOX
  fm list INBOX -n 10
  fm list Sent -o json
  fm list Drafts -n 5`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "Maximum number of emails to return")
}

// resolveMailboxArg resolves the mailbox argument to a JMAP mailbox ID.
// It first tries to resolve it as a name/role. If the mailbox is not found
// by name, it returns the argument as-is (assuming it's a raw JMAP ID).
// Network, auth, and other errors are propagated to the caller.
func resolveMailboxArg(client *jmap.Client, cmd *cobra.Command, arg string) (string, error) {
	ctx := cmd.Context()

	id, err := client.ResolveMailbox(ctx, arg)
	if err == nil {
		verbose.Log("Resolved mailbox %q to ID %s", arg, id)
		return id, nil
	}

	// Only fall back to raw ID if the mailbox wasn't found by name.
	// Propagate network, auth, and other errors so the user sees the real issue.
	if strings.Contains(err.Error(), "not found") {
		verbose.Log("Mailbox %q not found by name/role, using as raw ID", arg)
		return arg, nil
	}

	return "", err
}

func runList(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	mailboxID, err := resolveMailboxArg(client, cmd, args[0])
	if err != nil {
		return fmt.Errorf("resolving mailbox: %w", err)
	}

	filter := jmap.SearchFilter{
		InMailbox: mailboxID,
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
		fmt.Fprintln(os.Stderr, "No emails found in the specified mailbox.")
		return nil
	}

	format := output.Resolve(cmd)
	renderer := &jmap.EmailListRenderer{Emails: emails}
	if err := output.Print(format, emails, renderer); err != nil {
		return fmt.Errorf("formatting results: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nFound %d email(s).\n", len(emails))

	return nil
}
