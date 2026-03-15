package auth

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

func init() {
	// Replace OS keychain with in-memory mock for all tests in this package.
	keyring.MockInit()
}

func TestResolveTokenFromFlag(t *testing.T) {
	// Setup: put tokens everywhere so we can verify flag wins.
	t.Setenv("FM_API_TOKEN", "fmu1-env-token")
	_ = keyring.Set(keychainService, keychainKey, "fmu1-keychain-token")

	token, source, err := ResolveToken("fmu1-flag-token")
	if err != nil {
		t.Fatalf("ResolveToken() error: %v", err)
	}
	if token != "fmu1-flag-token" {
		t.Errorf("token = %q, want %q", token, "fmu1-flag-token")
	}
	if source != SourceFlag {
		t.Errorf("source = %q, want %q", source, SourceFlag)
	}
}

func TestResolveTokenFromEnv(t *testing.T) {
	// Setup: put token in env and keychain, but not flag.
	t.Setenv("FM_API_TOKEN", "fmu1-env-token")
	_ = keyring.Set(keychainService, keychainKey, "fmu1-keychain-token")

	token, source, err := ResolveToken("")
	if err != nil {
		t.Fatalf("ResolveToken() error: %v", err)
	}
	if token != "fmu1-env-token" {
		t.Errorf("token = %q, want %q", token, "fmu1-env-token")
	}
	if source != SourceEnvironment {
		t.Errorf("source = %q, want %q", source, SourceEnvironment)
	}
}

func TestResolveTokenFromKeychain(t *testing.T) {
	// Setup: token only in keychain.
	t.Setenv("FM_API_TOKEN", "")
	_ = keyring.Set(keychainService, keychainKey, "fmu1-keychain-token")

	token, source, err := ResolveToken("")
	if err != nil {
		t.Fatalf("ResolveToken() error: %v", err)
	}
	if token != "fmu1-keychain-token" {
		t.Errorf("token = %q, want %q", token, "fmu1-keychain-token")
	}
	if source != SourceKeychain {
		t.Errorf("source = %q, want %q", source, SourceKeychain)
	}
}

func TestResolveTokenNotFound(t *testing.T) {
	// Setup: no token anywhere.
	t.Setenv("FM_API_TOKEN", "")
	// Clear keychain.
	_ = keyring.Delete(keychainService, keychainKey)

	_, _, err := ResolveToken("")
	if err == nil {
		t.Fatal("ResolveToken() expected error, got nil")
	}
	if err != ErrNoToken {
		t.Errorf("error = %v, want ErrNoToken", err)
	}
}

func TestStoreToken(t *testing.T) {
	err := StoreToken("fmu1-test-token")
	if err != nil {
		t.Fatalf("StoreToken() error: %v", err)
	}

	// Verify it can be retrieved.
	got, err := keyring.Get(keychainService, keychainKey)
	if err != nil {
		t.Fatalf("keyring.Get() error: %v", err)
	}
	if got != "fmu1-test-token" {
		t.Errorf("stored token = %q, want %q", got, "fmu1-test-token")
	}
}

func TestDeleteToken(t *testing.T) {
	// Setup: store a token first.
	_ = keyring.Set(keychainService, keychainKey, "fmu1-to-delete")

	err := DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken() error: %v", err)
	}

	// Verify it's gone.
	_, err = keyring.Get(keychainService, keychainKey)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestDeleteTokenNotFound(t *testing.T) {
	// Ensure no token exists.
	_ = keyring.Delete(keychainService, keychainKey)

	// Should not error when deleting a non-existent token.
	err := DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken() error on non-existent: %v", err)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal token", "fmu1-abcdef123456", "fmu1-abc..."},
		{"short token", "abc", "****"},
		{"exactly 8", "12345678", "****"},
		{"empty", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskToken(tt.input)
			if got != tt.want {
				t.Errorf("MaskToken(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// MockInit is called in init() above.
	os.Exit(m.Run())
}
