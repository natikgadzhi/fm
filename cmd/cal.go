package cmd

import (
	"fmt"
	"sort"

	"github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "Calendar commands",
	Long:  "Manage and view Fastmail calendars and events via the JMAP Calendars API.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var calendarsCmd = &cobra.Command{
	Use:   "calendars",
	Short: "List all calendars",
	Long:  "Fetches and displays all subscribed calendars from Fastmail, sorted alphabetically by name.",
	Args:  cobra.NoArgs,
	RunE:  runCalendars,
}

func init() {
	rootCmd.AddCommand(calCmd)
	calCmd.AddCommand(calendarsCmd)
}

func runCalendars(cmd *cobra.Command, args []string) error {
	tok, _, err := auth.ResolveToken(token)
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)

	ctx := cmd.Context()
	calendars, err := client.GetCalendars(ctx)
	if err != nil {
		return fmt.Errorf("fetching calendars: %w", err)
	}

	sort.Slice(calendars, func(i, j int) bool {
		return calendars[i].Name < calendars[j].Name
	})

	format := output.Resolve(cmd)
	renderer := &jmap.CalendarListRenderer{Calendars: calendars}
	if err := output.Print(format, calendars, renderer); err != nil {
		return fmt.Errorf("formatting calendars: %w", err)
	}

	return nil
}
