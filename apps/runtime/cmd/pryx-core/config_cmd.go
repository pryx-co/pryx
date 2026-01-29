package main

import (
	"fmt"
	"reflect"
	"strings"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
)

func runConfig(args []string) int {
	if len(args) < 1 {
		usageConfig()
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
		printConfig(cfg)
		return 0
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: pryx-core config get <key>")
			return 1
		}
		key := strings.ReplaceAll(args[1], ".", "_")
		val, ok := getConfigValue(cfg, key)
		if !ok {
			fmt.Printf("Unknown config key: %s\n", args[1])
			return 1
		}
		fmt.Println(val)
		return 0
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: pryx-core config set <key> <value>")
			return 1
		}
		key := strings.ReplaceAll(args[1], ".", "_")
		value := args[2]

		if isProviderKeyField(key) {
			provider := extractProviderFromKeyField(key)
			if err := kc.SetProviderKey(provider, value); err != nil {
				fmt.Printf("Error storing key in keychain: %v\n", err)
				return 1
			}
			fmt.Printf("Stored API key for provider '%s' in keychain\n", provider)
			return 0
		}

		if err := setConfigValue(cfg, key, value); err != nil {
			fmt.Printf("Error setting value: %v\n", err)
			return 1
		}

		if err := cfg.Save(path); err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
			return 1
		}
		fmt.Printf("Updated %s = %s\n", key, value)
		return 0
	default:
		usageConfig()
		return 1
	}
}

func usageConfig() {
	fmt.Println("Usage:")
	fmt.Println("  pryx-core config list")
	fmt.Println("  pryx-core config get <key>")
	fmt.Println("  pryx-core config set <key> <value>")
}

func printConfig(cfg *config.Config) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	fmt.Println("Current Configuration:")
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		// Mask keys
		val := fmt.Sprintf("%v", v.Field(i).Interface())
		if strings.Contains(strings.ToLower(field.Name), "key") && len(val) > 4 {
			val = val[:4] + "***"
		}
		fmt.Printf("  %-20s %s\n", tag, val)
	}
}

func getConfigValue(cfg *config.Config, key string) (string, bool) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == key {
			return fmt.Sprintf("%v", v.Field(i).Interface()), true
		}
	}
	return "", false
}

func setConfigValue(cfg *config.Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == key {
			f := v.Field(i)
			if !f.CanSet() {
				return fmt.Errorf("field %s is not settable", key)
			}
			if f.Kind() == reflect.String {
				f.SetString(value)
				return nil
			}
			return fmt.Errorf("unsupported type for key %s", key)
		}
	}
	return fmt.Errorf("unknown config key: %s", key)
}

func isProviderKeyField(key string) bool {
	key = strings.ToLower(key)
	return strings.HasSuffix(key, "_key") ||
		key == "openai_key" ||
		key == "anthropic_key" ||
		key == "glm_key" ||
		key == "openrouter_key" ||
		key == "together_key" ||
		key == "groq_key" ||
		key == "xai_key" ||
		key == "mistral_key" ||
		key == "cohere_key" ||
		key == "google_key"
}

func extractProviderFromKeyField(key string) string {
	key = strings.ToLower(key)
	key = strings.TrimSuffix(key, "_key")
	return key
}
