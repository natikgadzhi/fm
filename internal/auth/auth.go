// Package auth handles API token resolution and storage.
package auth

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	// keychainService is the service name used in the OS keychain.
	keychainService = "fm"
	// keychainKey is the key name used in the OS keychain.
	keychainKey = "api-token"

	// TokenPrefix is the expected prefix for Fastmail API tokens.
	TokenPrefix = "fmu1-"

	// SourceFlag indicates the token came from a CLI flag.
	SourceFlag = "flag"
	// SourceEnvironment indicates the token came from an environment variable.
	SourceEnvironment = "environment"
	// SourceKeychain indicates the token came from the OS keychain.
	SourceKeychain = "keychain"
)

// ErrNoToken is returned when no API token can be found from any source.
var ErrNoToken = errors.New(
	"No API token found. Run 'fm auth login' or set FM_API_TOKEN. " +
		"Create a token at https://app.fastmail.com/settings/security/tokens/new",
)

// ResolveToken finds the API token using a 3-source priority chain:
//  1. flagValue (from --token flag) — highest priority
//  2. FM_API_TOKEN environment variable
//  3. OS keychain (service: "fm", key: "api-token")
//
// Returns the token, its source identifier, and any error.
func ResolveToken(flagValue string) (token string, source string, err error) {
	// 1. Flag value has highest priority.
	if flagValue != "" {
		return flagValue, SourceFlag, nil
	}

	// 2. Environment variable.
	if v := os.Getenv("FM_API_TOKEN"); v != "" {
		return v, SourceEnvironment, nil
	}

	// 3. OS keychain.
	t, err := keyring.Get(keychainService, keychainKey)
	if err == nil && t != "" {
		return t, SourceKeychain, nil
	}

	return "", "", ErrNoToken
}

// StoreToken saves the API token to the OS keychain.
func StoreToken(token string) error {
	return keyring.Set(keychainService, keychainKey, token)
}

// DeleteToken removes the API token from the OS keychain.
func DeleteToken() error {
	err := keyring.Delete(keychainService, keychainKey)
	if err != nil {
		// If the item doesn't exist, treat it as success.
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// MaskToken returns a masked version of the token for display.
// Shows the first 8 characters followed by "...".
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:8] + "..."
}
