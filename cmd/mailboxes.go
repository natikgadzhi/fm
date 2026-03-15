package cmd

import (
	"fmt"
	"sort"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/output"
	"github.com/spf13/cobra"
)

var mailboxesCmd = &cobra.Command{
	Use:   "mailboxes",
	Short: "List all mailboxes (folders)",
	Long:  "Fetches and displays all mailboxes from Fastmail, sorted alphabetically by name.",
	Args:  cobra.NoArgs,
	RunE:  runMailboxes,
}

func init() {
	rootCmd.AddCommand(mailboxesCmd)
}

func runMailboxes(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return fmt.Errorf("resolving token: %w", err)
	}

	client := jmap.NewClient(tok, jmap.WithTimeout(timeout))

	ctx := cmd.Context()
	if err := client.Discover(); err != nil {
		return fmt.Errorf("connecting to Fastmail: %w", err)
	}

	mailboxes, err := client.GetMailboxes(ctx)
	if err != nil {
		return fmt.Errorf("fetching mailboxes: %w", err)
	}

	// Sort alphabetically by name.
	sort.Slice(mailboxes, func(i, j int) bool {
		return mailboxes[i].Name < mailboxes[j].Name
	})

	formatter, err := output.New(outputFormat)
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}

	result, err := formatter.FormatMailboxes(mailboxes)
	if err != nil {
		return fmt.Errorf("formatting mailboxes: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), result)
	return nil
}
