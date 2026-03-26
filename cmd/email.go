package cmd

import (
	"github.com/spf13/cobra"
)

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Email commands (search, fetch, list, mailboxes)",
	Long: `Commands for searching, fetching, and listing emails from Fastmail.

Use "fm email <subcommand>" to interact with your email.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(emailCmd)
}
