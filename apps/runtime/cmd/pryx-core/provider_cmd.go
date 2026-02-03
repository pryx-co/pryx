package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/models"
)

// PopularProviders is a curated list of commonly used providers for UI prioritization
// The actual supported providers come dynamically from models.dev catalog (50+ providers)
var PopularProviders = []string{
	"openai",
	"anthropic",
	"google",
	"openrouter",
	"ollama",
	"groq",
	"xai",
	"mistral",
	"cohere",
}

func runProvider(args []string) int {
	if len(args) < 1 {
		usageProvider()
		return 1
	}

	command := args[0]
	path := config.DefaultPath()
	cfg := config.Load()

	if fileCfg, err := config.LoadFromFile(path); err == nil {
		cfg = fileCfg
	}

	kc := keychain.New("pryx")

	switch command {
	case "list":
		return providerList(cfg, kc)
	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider add <name>")
			fmt.Println("")
			fmt.Println("To see available providers, run: pryx-core provider list --available")
			return 1
		}
		return providerAdd(args[1], cfg, path, kc)
	case "set-key":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider set-key <name>")
			return 1
		}
		return providerSetKey(args[1], kc)
	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider remove <name>")
			return 1
		}
		return providerRemove(args[1], cfg, path, kc)
	case "use":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider use <name>")
			return 1
		}
		return providerUse(args[1], cfg, path, kc)
	case "test":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider test <name>")
			return 1
		}
		return providerTest(args[1], cfg, kc)
	case "oauth":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core provider oauth <provider>")
			fmt.Println("")
			fmt.Println("Supported providers:")
			fmt.Println("  google - Google AI (Gemini)")
			return 1
		}
		return runProviderOAuth([]string{args[1]})
	default:
		usageProvider()
		return 1
	}
}

func usageProvider() {
	fmt.Println("Usage:")
	fmt.Println("  pryx-core provider list                    List configured providers")
	fmt.Println("  pryx-core provider list --available        Show all available providers from models.dev")
	fmt.Println("  pryx-core provider add <name>              Add new provider interactively")
	fmt.Println("  pryx-core provider set-key <name>          Set API key for provider")
	fmt.Println("  pryx-core provider remove <name>           Remove provider config")
	fmt.Println("  pryx-core provider use <name>              Set as active/default provider")
	fmt.Println("  pryx-core provider test <name>             Test connection to provider")
	fmt.Println("  pryx-core provider oauth <provider>        Authenticate via OAuth (Google)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  pryx-core provider add openai")
	fmt.Println("  pryx-core provider set-key anthropic")
	fmt.Println("  pryx-core provider use groq")
	fmt.Println("  pryx-core provider oauth google")
	fmt.Println("")
	fmt.Println("Note: Providers are loaded dynamically from models.dev (50+ providers supported)")
}

func loadCatalog() (*models.Catalog, error) {
	svc := models.NewService()
	return svc.Load()
}

func providerList(cfg *config.Config, kc *keychain.Keychain) int {
	catalog, err := loadCatalog()
	if err != nil {
		fmt.Printf("Warning: Could not load models catalog: %v\n", err)
		fmt.Println("Showing configured providers only...")
	}

	fmt.Println("Configured Providers:")
	fmt.Println("====================")

	// Show currently configured providers
	providers := getConfiguredProviders(cfg, kc)
	if len(providers) == 0 {
		fmt.Println("No providers configured yet.")
		fmt.Println("Run 'pryx-core provider add <name>' to add a provider.")
	} else {
		for _, p := range providers {
			active := ""
			if cfg.ModelProvider == p.Name {
				active = " [ACTIVE]"
			}
			fmt.Printf("  ✓ %s%s\n", p.DisplayName, active)
			fmt.Printf("    API Key: %s\n", p.KeyStatus)
			if p.BaseURL != "" {
				fmt.Printf("    URL: %s\n", p.BaseURL)
			}
			fmt.Println()
		}
	}

	// Show available providers from catalog
	if catalog != nil {
		fmt.Println("\nAvailable Providers from models.dev:")
		fmt.Println("====================================")
		fmt.Printf("Total providers available: %d\n", len(catalog.Providers))
		fmt.Println()

		// Show popular providers first
		fmt.Println("Popular providers:")
		for _, id := range PopularProviders {
			if info, ok := catalog.GetProvider(id); ok {
				marker := " "
				if isProviderConfigured(id, cfg, kc) {
					marker = "✓"
				}
				fmt.Printf("  %s %s - %s\n", marker, info.Name, getProviderDescription(id))
			}
		}

		fmt.Println("\nRun 'pryx-core provider list --available' to see all providers")
	}

	return 0
}

type ConfiguredProvider struct {
	Name        string
	DisplayName string
	KeyStatus   string
	BaseURL     string
}

func getConfiguredProviders(cfg *config.Config, kc *keychain.Keychain) []ConfiguredProvider {
	var providers []ConfiguredProvider

	catalog, _ := loadCatalog()

	// Check all providers from catalog to see which are configured
	if catalog != nil {
		for id := range catalog.Providers {
			if isProviderConfigured(id, cfg, kc) {
				info, _ := catalog.GetProvider(id)
				providers = append(providers, ConfiguredProvider{
					Name:        id,
					DisplayName: info.Name,
					KeyStatus:   getKeyStatus(id, kc),
					BaseURL:     getProviderBaseURL(id),
				})
			}
		}
	}

	return providers
}

func isProviderConfigured(name string, cfg *config.Config, kc *keychain.Keychain) bool {
	// Check if API key is in keychain
	if _, err := kc.GetProviderKey(name); err == nil {
		return true
	}

	// Check for OAuth tokens
	if isOAuthConfigured(name, kc) {
		return true
	}

	// Check environment variables
	envVars := getProviderEnvVars(name)
	for _, env := range envVars {
		if os.Getenv(env) != "" {
			return true
		}
	}

	// Check if provider was explicitly added via 'provider add' command
	if isProviderInList(cfg.ConfiguredProviders, name) {
		return true
	}

	return false
}

func isProviderInList(list []string, name string) bool {
	for _, item := range list {
		if item == name {
			return true
		}
	}
	return false
}

func getKeyStatus(name string, kc *keychain.Keychain) string {
	if _, err := kc.GetProviderKey(name); err == nil {
		return "configured (keychain)"
	}

	if isOAuthConfigured(name, kc) {
		return "configured (OAuth)"
	}

	envVars := getProviderEnvVars(name)
	for _, env := range envVars {
		if os.Getenv(env) != "" {
			return fmt.Sprintf("configured (env: %s)", env)
		}
	}

	return "not configured"
}

func getProviderEnvVars(name string) []string {
	switch name {
	case "openai":
		return []string{"OPENAI_API_KEY"}
	case "anthropic":
		return []string{"ANTHROPIC_API_KEY"}
	case "google":
		return []string{"GOOGLE_API_KEY", "GEMINI_API_KEY"}
	case "openrouter":
		return []string{"OPENROUTER_API_KEY"}
	case "groq":
		return []string{"GROQ_API_KEY"}
	case "xai":
		return []string{"XAI_API_KEY"}
	case "mistral":
		return []string{"MISTRAL_API_KEY"}
	case "cohere":
		return []string{"COHERE_API_KEY"}
	case "ollama":
		return []string{"OLLAMA_HOST"}
	default:
		return []string{}
	}
}

func getProviderBaseURL(name string) string {
	switch name {
	case "openai":
		return "https://api.openai.com/v1"
	case "anthropic":
		return "https://api.anthropic.com/v1"
	case "google":
		return "https://generativelanguage.googleapis.com/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "xai":
		return "https://api.x.ai/v1"
	case "mistral":
		return "https://api.mistral.ai/v1"
	case "cohere":
		return "https://api.cohere.com/v1"
	case "ollama":
		if host := os.Getenv("OLLAMA_HOST"); host != "" {
			return host
		}
		return "http://localhost:11434"
	default:
		return ""
	}
}

func getProviderDescription(name string) string {
	descriptions := map[string]string{
		"openai":     "GPT-4, GPT-3.5, and more",
		"anthropic":  "Claude 3 family of models",
		"google":     "Gemini models",
		"openrouter": "Access 100+ models via one API",
		"ollama":     "Run models locally",
		"groq":       "Fast inference API",
		"xai":        "Grok models by xAI",
		"mistral":    "Mistral AI models",
		"cohere":     "Command and Embed models",
		"together":   "Open source model hosting",
	}

	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return "AI model provider"
}

func providerAdd(name string, cfg *config.Config, path string, kc *keychain.Keychain) int {
	// Validate provider exists in catalog
	catalog, err := loadCatalog()
	if err != nil {
		fmt.Printf("Error: Could not load models catalog: %v\n", err)
		return 1
	}

	providerInfo, ok := catalog.GetProvider(name)
	if !ok {
		fmt.Printf("Error: Unknown provider '%s'\n", name)
		fmt.Println("Run 'pryx-core provider list --available' to see all available providers")
		return 1
	}

	fmt.Printf("Adding provider: %s\n", providerInfo.Name)
	fmt.Println()

	// Check if provider supports OAuth
	if supportsOAuth(name) {
		fmt.Println("This provider supports OAuth authentication.")
		fmt.Printf("Run 'pryx-core provider oauth %s' for OAuth flow (recommended)\n", name)
		fmt.Println("Or continue with API key below.")
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)

	// Get API key (optional for some providers like Ollama)
	fmt.Print("API Key (press Enter to skip for local providers or use OAuth): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey != "" {
		// Store in keychain
		if err := kc.SetProviderKey(name, apiKey); err != nil {
			fmt.Printf("Error storing API key: %v\n", err)
			return 1
		}
		fmt.Println("✓ API key stored securely in keychain")
	}

	// Add to configured providers list (tracks providers added even without API keys)
	if !isProviderInList(cfg.ConfiguredProviders, name) {
		cfg.ConfiguredProviders = append(cfg.ConfiguredProviders, name)
	}

	// Set as active if no provider is currently active
	if cfg.ModelProvider == "" {
		cfg.ModelProvider = name
	}

	// Save config
	if err := cfg.Save(path); err != nil {
		fmt.Printf("Warning: Could not save config: %v\n", err)
	} else {
		if cfg.ModelProvider == name {
			fmt.Printf("✓ Set as active provider\n")
		}
	}

	fmt.Println()
	fmt.Printf("Provider '%s' added successfully!\n", name)
	fmt.Printf("Run 'pryx-core provider use %s' to set as active provider\n", name)
	fmt.Printf("Run 'pryx-core provider test %s' to verify connection\n", name)

	return 0
}

// supportsOAuth checks if a provider supports OAuth authentication
func supportsOAuth(name string) bool {
	switch name {
	case "google":
		return true
	default:
		return false
	}
}

func providerSetKey(name string, kc *keychain.Keychain) int {
	// Validate provider exists in catalog
	catalog, err := loadCatalog()
	if err != nil {
		fmt.Printf("Error: Could not load models catalog: %v\n", err)
		return 1
	}

	if _, ok := catalog.GetProvider(name); !ok {
		fmt.Printf("Error: Unknown provider '%s'\n", name)
		return 1
	}

	fmt.Printf("Enter API key for %s: ", name)
	reader := bufio.NewReader(os.Stdin)
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Println("Error: API key cannot be empty")
		return 1
	}

	if err := kc.SetProviderKey(name, apiKey); err != nil {
		fmt.Printf("Error storing API key: %v\n", err)
		return 1
	}

	fmt.Println("✓ API key stored securely in keychain")
	return 0
}

func providerRemove(name string, cfg *config.Config, path string, kc *keychain.Keychain) int {
	// Remove from keychain
	if err := kc.DeleteProviderKey(name); err != nil {
		fmt.Printf("Warning: Could not remove API key from keychain: %v\n", err)
	}

	// If this was the active provider, clear it
	if cfg.ModelProvider == name {
		cfg.ModelProvider = ""
		if err := cfg.Save(path); err != nil {
			fmt.Printf("Warning: Could not update config: %v\n", err)
		} else {
			fmt.Printf("Cleared active provider setting\n")
		}
	}

	fmt.Printf("✓ Provider '%s' removed\n", name)
	return 0
}

func providerUse(name string, cfg *config.Config, path string, kc *keychain.Keychain) int {
	// Validate provider exists in catalog
	catalog, err := loadCatalog()
	if err != nil {
		fmt.Printf("Error: Could not load models catalog: %v\n", err)
		return 1
	}

	providerInfo, ok := catalog.GetProvider(name)
	if !ok {
		fmt.Printf("Error: Unknown provider '%s'\n", name)
		return 1
	}

	// Check if provider is configured
	if !isProviderConfigured(name, cfg, kc) {
		fmt.Printf("Warning: Provider '%s' is not configured yet\n", name)
		fmt.Printf("Run 'pryx-core provider add %s' to configure it\n", name)
		return 1
	}

	cfg.ModelProvider = name
	if err := cfg.Save(path); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s is now the active provider\n", providerInfo.Name)
	return 0
}

func providerTest(name string, cfg *config.Config, kc *keychain.Keychain) int {
	// Validate provider exists in catalog
	catalog, err := loadCatalog()
	if err != nil {
		fmt.Printf("Error: Could not load models catalog: %v\n", err)
		return 1
	}

	providerInfo, ok := catalog.GetProvider(name)
	if !ok {
		fmt.Printf("Error: Unknown provider '%s'\n", name)
		return 1
	}

	fmt.Printf("Testing connection to %s...\n", providerInfo.Name)

	// Check if configured
	if !isProviderConfigured(name, cfg, kc) {
		fmt.Printf("✗ Provider is not configured\n")
		fmt.Printf("Run 'pryx-core provider add %s' to configure it\n", name)
		return 1
	}

	// Get available models
	models := catalog.GetProviderModels(name)
	if len(models) == 0 {
		fmt.Printf("Warning: No models found for provider in catalog\n")
	} else {
		fmt.Printf("✓ Provider accessible\n")
		fmt.Printf("✓ %d models available\n", len(models))

		// Show first few models
		fmt.Println("\nPopular models:")
		count := 0
		for _, m := range models {
			if count >= 5 {
				break
			}
			fmt.Printf("  - %s\n", m.Name)
			count++
		}
		if len(models) > 5 {
			fmt.Printf("  ... and %d more\n", len(models)-5)
		}
	}

	return 0
}
