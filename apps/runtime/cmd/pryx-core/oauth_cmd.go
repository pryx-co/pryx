package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"pryx-core/internal/auth"
	"pryx-core/internal/keychain"
)

// runProviderOAuth initiates OAuth flow for a provider
func runProviderOAuth(args []string) int {
	if len(args) < 1 {
		fmt.Println("Usage: pryx-core provider oauth <provider>")
		fmt.Println("")
		fmt.Println("Supported providers:")
		fmt.Println("  google - Google AI (Gemini)")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  pryx-core provider oauth google")
		return 1
	}

	providerID := args[0]

	// Check if provider supports OAuth
	config, ok := auth.ProviderConfigs[providerID]
	if !ok {
		fmt.Printf("Error: Provider '%s' does not support OAuth\n", providerID)
		fmt.Println("Currently supported: google")
		return 1
	}

	kc := keychain.New("pryx")
	oauth := auth.NewProviderOAuth(kc)

	fmt.Printf("Starting OAuth flow for %s...\n", config.Name)
	fmt.Println("")
	fmt.Println("This will open your browser to authorize Pryx.")
	fmt.Println("Please complete the authorization in your browser.")
	fmt.Println("")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start OAuth flow
	tokens, err := oauth.StartOAuthFlow(ctx, providerID)
	if err != nil {
		fmt.Printf("✗ OAuth failed: %v\n", err)
		return 1
	}

	// Save tokens
	if err := oauth.SaveTokens(providerID, tokens); err != nil {
		fmt.Printf("✗ Failed to save tokens: %v\n", err)
		return 1
	}

	fmt.Println("✓ OAuth completed successfully!")
	fmt.Printf("✓ Tokens saved securely in keychain\n")
	fmt.Println("")
	fmt.Printf("You can now use %s as your AI provider.\n", config.Name)
	fmt.Printf("Run 'pryx-core provider use %s' to set as active.\n", providerID)

	return 0
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// isOAuthConfigured checks if OAuth tokens exist for a provider
func isOAuthConfigured(providerID string, kc *keychain.Keychain) bool {
	_, err := kc.Get("oauth_" + providerID + "_access")
	return err == nil
}

// getOAuthToken retrieves OAuth access token for API calls
func getOAuthToken(providerID string, kc *keychain.Keychain) (string, error) {
	oauth := auth.NewProviderOAuth(kc)

	// Check if token needs refresh
	needsRefresh, err := oauth.IsTokenExpired(providerID)
	if err != nil {
		return "", fmt.Errorf("failed to check token expiry: %w", err)
	}

	if needsRefresh {
		fmt.Println("Token expired, refreshing...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := oauth.RefreshToken(ctx, providerID); err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		fmt.Println("✓ Token refreshed")
	}

	return oauth.GetToken(providerID)
}
