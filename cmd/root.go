package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var (
	outputFormat string
	cacheDir     string
	timeout      time.Duration
	token        string
	endpoint     string
)

var rootCmd = &cobra.Command{
	Use:   "fm",
	Short: "A read-only CLI for Fastmail via JMAP",
	Long: `fm is a command-line tool that connects to Fastmail using the JMAP protocol
to search, list, and fetch emails. It saves messages locally as Markdown files
with YAML frontmatter for metadata. The tool is read-only -- it never modifies,
sends, or deletes emails.`,
	Version:       Version,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text",
		"Output format: text, json, or markdown")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "",
		"Cache directory for fetched emails (default: ~/.local/share/fm/cache/)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second,
		"HTTP request timeout")
	rootCmd.PersistentFlags().StringVar(&token, "token", "",
		"Fastmail API token (overrides env and keychain)")
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "",
		"Override JMAP session endpoint URL (for testing)")
	rootCmd.PersistentFlags().MarkHidden("endpoint")
}

// clientOpts returns the common JMAP client options derived from global flags.
func clientOpts() []jmap.Option {
	opts := []jmap.Option{jmap.WithTimeout(timeout)}
	if endpoint != "" {
		opts = append(opts, jmap.WithBaseURL(endpoint))
	}
	return opts
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
