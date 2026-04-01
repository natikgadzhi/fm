package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	cliauth "github.com/natikgadzhi/cli-kit/auth"
	"github.com/natikgadzhi/fm/internal/jmap"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage API token authentication",
	Long:  "Commands for logging in, checking authentication, and logging out of Fastmail.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store a Fastmail API token in the OS keychain",
	Long: `Prompts for a Fastmail API token and stores it in the OS keychain.

Create a token at https://app.fastmail.com/settings/security/tokens/new
The token should start with "fmu1-".`,
	RunE: runAuthLogin,
}

var authCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify that your API token is valid",
	Long: `Resolves the API token and verifies it by making a JMAP session request.
On success, displays token source, masked token, account ID, and username.
On failure, displays a clear error message.`,
	RunE: runAuthCheck,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the API token from the OS keychain",
	Long:  "Deletes the stored API token from the OS keychain.",
	RunE:  runAuthLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authCheckCmd)
	authCmd.AddCommand(authLogoutCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	fmt.Print("Enter Fastmail API token: ")

	var tokenInput string
	// If stdin is a terminal, use secure (no-echo) reading.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		tokenInput = string(raw)
		fmt.Println() // newline after hidden input
	} else {
		// Non-terminal (piped input) — read a line.
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			tokenInput = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
	}

	tokenInput = strings.TrimSpace(tokenInput)
	if tokenInput == "" {
		return fmt.Errorf("no token provided")
	}

	// Soft format validation.
	if !strings.HasPrefix(tokenInput, TokenPrefix) {
		fmt.Fprintf(os.Stderr, "Warning: token does not start with %q — it may not be a valid Fastmail API token.\n", TokenPrefix)
	}

	if err := cliauth.StoreToken(keychainService, keychainKey, tokenInput); err != nil {
		return fmt.Errorf("storing token in keychain: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Token stored in OS keychain. Masked: %s\n", cliauth.MaskToken(tokenInput))
	return nil
}

func runAuthCheck(cmd *cobra.Command, args []string) error {
	tok, source, err := cliauth.ResolveToken(tokenSource())
	if err != nil {
		return err
	}

	client := jmap.NewClient(tok, clientOpts()...)
	if err := client.Discover(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Token source: %s\n", source)
	fmt.Fprintf(w, "Token:        %s\n", cliauth.MaskToken(tok))

	accountID, err := client.PrimaryAccountID()
	if err == nil {
		fmt.Fprintf(w, "Account ID:   %s\n", accountID)
	}

	if session := client.Session(); session != nil && session.Username != "" {
		fmt.Fprintf(w, "Username:     %s\n", session.Username)
	}

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if err := cliauth.DeleteToken(keychainService, keychainKey); err != nil {
		return fmt.Errorf("removing token from keychain: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Token removed from OS keychain.")
	return nil
}
