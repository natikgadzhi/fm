package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/natikgadzhi/fm/internal/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/natikgadzhi/fm/internal/verbose"
	"github.com/spf13/cobra"
)

// Version, Commit, and Date are set at build time via ldflags.
var (
	Version = "dev"
	Commit  = "dev"
	Date    = "unknown"
)

// Exit codes.
const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitAuthError = 2
)

// validOutputFormats lists the allowed values for --output.
var validOutputFormats = map[string]bool{
	"text":     true,
	"json":     true,
	"markdown": true,
}

var (
	outputFormat string
	cacheDir     string
	timeout      time.Duration
	token        string
	endpoint     string
	verboseFlag  bool
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
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Enable verbose logging if --verbose is set.
		if verboseFlag {
			verbose.Enable()
			verbose.Log("fm version %s", Version)
		}

		// Validate output format.
		if !validOutputFormats[outputFormat] {
			return fmt.Errorf("invalid output format %q: must be one of text, json, markdown", outputFormat)
		}

		// Set up graceful Ctrl+C handling with context cancellation.
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
		// Store the cancel function so it can be cleaned up.
		// The context will be cancelled when the signal is received.
		cmd.SetContext(ctx)

		// Ensure cancel is called when the command finishes.
		go func() {
			<-ctx.Done()
			cancel()
		}()

		return nil
	},
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
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false,
		"Enable debug logging to stderr")
}

// clientOpts returns the common JMAP client options derived from global flags.
func clientOpts() []jmap.Option {
	opts := []jmap.Option{jmap.WithTimeout(timeout)}
	if endpoint != "" {
		opts = append(opts, jmap.WithBaseURL(endpoint))
	}
	return opts
}

// exitCode determines the process exit code based on the error type.
func exitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var authErr *auth.AuthError
	if errors.As(err, &authErr) {
		return ExitAuthError
	}

	return ExitError
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}
