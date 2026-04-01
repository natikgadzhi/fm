package cmd

import (
	"fmt"
	"sort"

	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/table"
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
	emailCmd.AddCommand(mailboxesCmd)
}

func runMailboxes(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	ctx := cmd.Context()
	mailboxes, err := client.GetMailboxes(ctx)
	if err != nil {
		return fmt.Errorf("fetching mailboxes: %w", err)
	}

	// Sort alphabetically by name.
	sort.Slice(mailboxes, func(i, j int) bool {
		return mailboxes[i].Name < mailboxes[j].Name
	})

	format := output.Resolve(cmd)
	if output.IsJSON(format) {
		if err := output.PrintJSON(mailboxes); err != nil {
			return fmt.Errorf("formatting mailboxes: %w", err)
		}
	} else {
		t := table.New()
		renderer := &jmap.MailboxListRenderer{Mailboxes: mailboxes}
		renderer.RenderTable(t)
		if err := t.Flush(); err != nil {
			return fmt.Errorf("formatting mailboxes: %w", err)
		}
	}

	return nil
}
