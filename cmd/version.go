package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// versionInfo is the JSON envelope for version output.
type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := versionInfo{
			Version: Version,
			Commit:  Commit,
			Date:    Date,
		}

		if outputFormat == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "fm %s\n", info.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", info.Commit)
		fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", info.Date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
