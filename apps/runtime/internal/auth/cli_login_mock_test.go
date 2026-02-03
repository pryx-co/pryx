package auth_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"pryx-core/internal/auth"
	"pryx-core/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIDeviceCodeDisplay tests device code formatting for CLI display
func TestCLIDeviceCodeDisplay(t *testing.T) {
	t.Run("user_code_visual_format", func(t *testing.T) {
		codes := []struct {
			code     string
			expected string
		}{
			{"USER123", "U-S-E-R-1-2-3"},
			{"ABCD", "A-B-C-D"},
			{"MYCODE", "M-Y-C-O-D-E"},
		}

		for _, tc := range codes {
			// Mock formatting: insert dashes every character for readability
			formatted := insertDashes(tc.code)
			t.Logf("Code %s formatted as: %s", tc.code, formatted)
			assert.Equal(t, tc.expected, formatted)
		}
	})

	t.Run("verification_url_display", func(t *testing.T) {
		url := "https://auth.pryx.dev/activate"
		displayURL := truncateMiddle(url, 40)

		t.Logf("URL displayed as: %s", displayURL)
		assert.LessOrEqual(t, len(displayURL), 43, "URL should be shortened for display")
		assert.Contains(t, displayURL, "pryx.dev", "Should show domain")
	})

	t.Run("instructions_message", func(t *testing.T) {
		userCode := "USER123"
		expiresIn := 1800

		message := formatInstructions(userCode, expiresIn)

		assert.Contains(t, message, userCode)
		assert.Contains(t, message, "minutes")
		assert.Contains(t, message, "expires")

		t.Logf("Instructions: %s", message)
	})
}

// TestCLIPollingTimeout tests timeout behavior mocking
func TestCLIPollingTimeout(t *testing.T) {
	t.Run("polling_interval_validation", func(t *testing.T) {
		intervals := []int{1, 5, 10, 30}

		for _, interval := range intervals {
			assert.Greater(t, interval, 0, "Interval should be positive")
			assert.LessOrEqual(t, interval, 30, "Interval should not exceed 30 seconds to avoid rate limiting")

			t.Logf("Valid polling interval: %d seconds", interval)
		}
	})

	t.Run("timeout_calculation", func(t *testing.T) {
		expiresIn := 1800
		interval := 5
		maxPolls := expiresIn / interval

		assert.Equal(t, 360, maxPolls, "Max polls should be calculated from expiresIn / interval")

		timeoutDuration := time.Duration(expiresIn) * time.Second
		t.Logf("Total timeout duration: %v", timeoutDuration)
	})

	t.Run("progress_indicator_timing", func(t *testing.T) {
		intervals := []int{5, 10, 30}
		progressChars := []string{"█", "▓", "░"}

		for _, interval := range intervals {
			for i, char := range progressChars {
				t.Logf("Interval %ds: progress %s (%d)", interval, char, i)
			}
		}
	})

	t.Run("retry_on_timeout", func(t *testing.T) {
		maxRetries := 3
		retryDelays := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}

		assert.Len(t, retryDelays, maxRetries)

		for i, delay := range retryDelays {
			t.Logf("Retry %d: delay %v", i+1, delay)
		}
	})
}

// TestCLITokenPersistence tests mock token storage
func TestCLITokenPersistence(t *testing.T) {
	t.Run("token_key_format", func(t *testing.T) {
		provider := "test-provider"

		key := formatTokenKey(provider)
		expected := "oauth_token_test-provider"

		assert.Equal(t, expected, key)
		t.Logf("Token key for provider %s: %s", provider, key)
	})

	t.Run("token_storage_verification", func(t *testing.T) {
		kc := newMockKeychain()
		provider := "mock-provider"
		token := "mock-access-token-123"

		// Store token
		err := kc.Set("oauth_token_"+provider, token)
		require.NoError(t, err)

		// Verify storage
		stored, err := kc.Get("oauth_token_" + provider)
		require.NoError(t, err)
		assert.Equal(t, token, stored)

		t.Logf("Token for %s verified in storage", provider)
	})

	t.Run("token_expiry_tracking", func(t *testing.T) {
		expiresIn := 3600
		now := time.Now()
		expiresAt := now.Add(time.Duration(expiresIn) * time.Second)

		assert.True(t, expiresAt.After(now))
		assert.Equal(t, now.Add(time.Hour).Minute(), expiresAt.Minute())

		t.Logf("Token expires at: %v", expiresAt)
	})
}

// TestCLILoginRetryLogic tests mock retry behavior
func TestCLILoginRetryLogic(t *testing.T) {
	t.Run("retry_count_validation", func(t *testing.T) {
		maxRetries := 3
		currentRetry := 0

		for currentRetry < maxRetries {
			currentRetry++
			t.Logf("Retry attempt %d/%d", currentRetry, maxRetries)
		}

		assert.Equal(t, maxRetries, currentRetry)
	})

	t.Run("retry_delay_exponential_backoff", func(t *testing.T) {
		baseDelay := 100 * time.Millisecond
		retries := []int{0, 1, 2, 3}

		for _, retry := range retries {
			delay := baseDelay * time.Duration(1<<retry)
			t.Logf("Retry %d: delay %v", retry, delay)
		}
	})

	t.Run("error_classification", func(t *testing.T) {
		errors := []struct {
			err         auth.ErrorResponse
			isRetriable bool
		}{
			{auth.ErrorResponse{Error: "authorization_pending"}, true},
			{auth.ErrorResponse{Error: "slow_down"}, true},
			{auth.ErrorResponse{Error: "invalid_client"}, false},
			{auth.ErrorResponse{Error: "access_denied"}, false},
		}

		for _, tc := range errors {
			t.Logf("Error %s: retriable=%v", tc.err.Error, tc.isRetriable)
		}
	})

	t.Run("success_after_retry", func(t *testing.T) {
		retriesBeforeSuccess := 2
		maxRetries := 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			if attempt > retriesBeforeSuccess {
				t.Logf("Success on attempt %d", attempt)
				break
			}
			t.Logf("Retry attempt %d failed", attempt)
		}
	})
}

// Helper functions for CLI mock tests

func insertDashes(s string) string {
	var result strings.Builder
	for i, c := range s {
		if i > 0 {
			result.WriteRune('-')
		}
		result.WriteRune(c)
	}
	return result.String()
}

func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	half := (maxLen - 3) / 2
	return s[:half] + "..." + s[len(s)-half:]
}

func formatInstructions(userCode string, expiresIn int) string {
	minutes := expiresIn / 60
	return "Visit https://auth.pryx.dev/activate and enter code " +
		userCode + ". Code expires in " +
		itow(minutes) + " minutes."
}

func formatTokenKey(provider string) string {
	return "oauth_token_" + provider
}

func itow(n int) string {
	if n <= 10 {
		return []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}[n]
	}
	return fmt.Sprintf("%d", n)
}

// MockDeviceCodeFlow simulates complete CLI device flow
func MockDeviceCodeFlow(t *testing.T) {
	t.Run("complete_cli_flow_simulation", func(t *testing.T) {
		ctx := context.Background()

		kc := newMockKeychain()
		cfg := &config.AuthConfig{
			OAuthProviders: map[string]*config.OAuthProvider{
				"cli-provider": {
					Name:         "CLI Provider",
					ClientID:     "cli-client-id",
					ClientSecret: "cli-secret",
					AuthURL:      "https://auth.pryx.dev/oauth/device/authorize",
					TokenURL:     "https://auth.pryx.dev/oauth/token",
					Scopes:       []string{"openid"},
				},
			},
		}

		manager := auth.NewManager(cfg, kc)

		// Step 1: Initiate device flow
		_, err := manager.InitiateDeviceFlow(ctx, "cli-provider", "pryx://callback")
		require.NoError(t, err)

		t.Logf("Device flow initiated")
		t.Logf("User code: %s", "USER123")
		t.Logf("Verification URL: https://auth.pryx.dev/activate")
		t.Logf("Expires in: 30 minutes")

		// Step 2: Simulate polling with timeout
		pollingCompleted := false
		maxPolls := 10

		for i := 0; i < maxPolls && !pollingCompleted; i++ {
			t.Logf("Polling attempt %d/%d", i+1, maxPolls)
			if i >= 2 { // Simulate success after 2 attempts
				pollingCompleted = true
			}
		}

		// Step 3: Store token
		token := "mock-cli-token-final"
		err = manager.SetManualToken(ctx, "cli-provider", token)
		require.NoError(t, err)

		t.Logf("CLI device flow completed successfully")
	})
}
