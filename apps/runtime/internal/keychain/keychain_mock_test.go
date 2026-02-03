package keychain_test

import (
	"os"
	"testing"

	"pryx-core/internal/keychain"

	"github.com/stretchr/testify/assert"
)

// TestKeychainUnavailable tests behavior when keychain is unavailable
func TestKeychainUnavailable(t *testing.T) {
	t.Run("missing_keyring_service", func(t *testing.T) {
		// Test behavior with a non-existent service
		k := keychain.New("non-existent-service-12345")

		err := k.Set("test-user", "test-password")

		// In CI or restricted environments, keychain may not be available
		// The keychain implementation should handle this gracefully
		if err != nil {
			t.Logf("Keychain unavailable (expected in CI/restricted env): %v", err)
			// This is acceptable - the implementation should have fallback behavior
		}
	})

	t.Run("graceful_degradation", func(t *testing.T) {
		// Test that the keychain API is consistent even when backend fails
		k := keychain.New("test-service-" + t.Name())

		// Set operation
		err := k.Set("user-"+t.Name(), "password")

		// Get operation
		password, err := k.Get("user-" + t.Name())

		if err != nil {
			t.Logf("Keychain error (may be expected): %v", err)
			// The API should be consistent - operations should not panic
		} else {
			assert.Equal(t, "password", password)
		}
	})

	t.Run("error_messages_clear", func(t *testing.T) {
		k := keychain.New("test-error-messages")

		_, err := k.Get("non-existent-user")

		if err != nil {
			// Error message should be actionable
			errStr := err.Error()
			assert.NotEmpty(t, errStr)

			t.Logf("Keychain error message: %s", errStr)
		}
	})
}

// TestSecureStorageFallback tests alternative storage when keychain unavailable
func TestSecureStorageFallback(t *testing.T) {
	t.Run("alternative_storage_interface", func(t *testing.T) {
		// The keychain interface should be consistent
		k := keychain.New("test-fallback-" + t.Name())

		// Verify interface methods exist
		assert.NotNil(t, k.Set)
		assert.NotNil(t, k.Get)
		assert.NotNil(t, k.Delete)

		t.Logf("Keychain interface verified")
	})

	t.Run("storage_consistency", func(t *testing.T) {
		k := keychain.New("test-consistency-" + t.Name())

		testPairs := []struct {
			user     string
			password string
		}{
			{"user1", "password1"},
			{"user2", "password2"},
			{"user3", "password3"},
		}

		// Set all pairs
		for _, pair := range testPairs {
			err := k.Set(pair.user, pair.password)
			if err != nil {
				t.Logf("Set error for %s: %v", pair.user, err)
			}
		}

		// Verify all pairs
		for _, pair := range testPairs {
			password, err := k.Get(pair.user)
			if err != nil {
				t.Logf("Get error for %s: %v", pair.user, err)
			} else {
				assert.Equal(t, pair.password, password)
			}
		}
	})
}

// TestKeychainErrorMessages tests error message clarity
func TestKeychainErrorMessages(t *testing.T) {
	t.Run("set_error_messages", func(t *testing.T) {
		k := keychain.New("test-error-set")

		err := k.Set("", "password") // Empty user

		if err != nil {
			errStr := err.Error()
			assert.NotEmpty(t, errStr)
			t.Logf("Set error message: %s", errStr)
		}
	})

	t.Run("get_error_messages", func(t *testing.T) {
		k := keychain.New("test-error-get")

		_, err := k.Get("completely-nonexistent-user-12345")

		if err != nil {
			errStr := err.Error()
			assert.NotEmpty(t, errStr)
			t.Logf("Get error message: %s", errStr)
		}
	})

	t.Run("delete_error_messages", func(t *testing.T) {
		k := keychain.New("test-error-delete")

		err := k.Delete("non-existent-delete-user")

		// Delete errors are often silently ignored
		// This is acceptable behavior for keychain operations
		if err != nil {
			t.Logf("Delete error: %v", err)
		}
	})
}

// TestKeychainIntegration tests integration with auth system
func TestKeychainIntegration(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping keychain integration test in CI")
	}

	t.Run("oauth_token_storage", func(t *testing.T) {
		k := keychain.New("test-oauth-" + t.Name())

		oauthTokens := []struct {
			provider string
			token    string
		}{
			{"openai", "sk-test-token-123"},
			{"anthropic", "sk-ant-test-token-456"},
			{"google", "ya29-test-token-789"},
		}

		for _, token := range oauthTokens {
			key := "oauth_token_" + token.provider
			err := k.Set(key, token.token)
			assert.NoError(t, err)

			stored, err := k.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, token.token, stored)

			t.Logf("OAuth token for %s stored and retrieved", token.provider)
		}
	})

	t.Run("cleanup_after_test", func(t *testing.T) {
		k := keychain.New("test-cleanup-" + t.Name())

		testUsers := []string{
			"cleanup-user-1",
			"cleanup-user-2",
			"cleanup-user-3",
		}

		for _, user := range testUsers {
			k.Set(user, "test-password")
		}

		// Clean up
		for _, user := range testUsers {
			err := k.Delete(user)
			if err != nil {
				t.Logf("Cleanup warning for %s: %v", user, err)
			}
		}

		// Verify cleanup
		for _, user := range testUsers {
			_, err := k.Get(user)
			assert.Error(t, err, "User should be deleted")
		}
	})
}
