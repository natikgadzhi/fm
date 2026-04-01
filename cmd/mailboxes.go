package cmd

import (
	"fmt"
	"sort"

	cliauth "github.com/natikgadzhi/cli-kit/auth"
	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/cli-kit/progress"
	"github.com/natikgadzhi/cli-kit/table"
	"github.com/natikgadzhi/fm/internal/jmap"
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
	tok, _, err := cliauth.ResolveToken(tokenSource())
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	format := output.Resolve(cmd)
	spinner := progress.NewSpinner("Fetching mailboxes...", format)
	spinner.Start()

	ctx := cmd.Context()
	mailboxes, err := client.GetMailboxes(ctx)
	if err != nil {
		spinner.Finish()
		return fmt.Errorf("fetching mailboxes: %w", err)
	}

	spinner.Finish()

	// Sort alphabetically by name.
	sort.Slice(mailboxes, func(i, j int) bool {
		return mailboxes[i].Name < mailboxes[j].Name
	})

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
