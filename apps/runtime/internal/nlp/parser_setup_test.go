package nlp

import (
	"testing"
)

// Test setup intent detection
func TestParser_Parse_SetupIntent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		expected     Intent
		wantEntities []string // entity types we expect to find
	}{
		{
			name:         "setup openai provider",
			input:        "setup openai provider",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "install anthropic",
			input:        "install anthropic",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "initialize google",
			input:        "initialize google provider",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "prepare openai",
			input:        "prepare my openai setup",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "get ready",
			input:        "get ready to use claude",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "help me set up",
			input:        "help me set up gpt",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
		{
			name:         "i want to use",
			input:        "i want to use gemini",
			expected:     IntentSetup,
			wantEntities: []string{"provider"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			if result.Confidence < 0.6 {
				t.Errorf("Parse(%q) Confidence = %v, want >= 0.6", tt.input, result.Confidence)
			}

			// Check for expected entity types
			foundEntityTypes := make(map[string]bool)
			for _, entity := range result.Entities {
				foundEntityTypes[entity.Type] = true
			}

			for _, wantType := range tt.wantEntities {
				if !foundEntityTypes[wantType] {
					t.Errorf("Parse(%q) missing entity type %q, found: %v", tt.input, wantType, result.Entities)
				}
			}
		})
	}
}

// Test connect intent detection
func TestParser_Parse_ConnectIntent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		expected     Intent
		wantEntities []string // entity types we expect to find
	}{
		{
			name:         "connect telegram",
			input:        "connect my telegram bot",
			expected:     IntentConnect,
			wantEntities: []string{"channel"},
		},
		{
			name:         "link slack",
			input:        "link slack workspace",
			expected:     IntentConnect,
			wantEntities: []string{"channel"},
		},
		{
			name:         "integrate teams",
			input:        "integrate microsoft teams",
			expected:     IntentConnect,
			wantEntities: []string{"channel"},
		},
		{
			name:         "attach webhook",
			input:        "attach a webhook to discord",
			expected:     IntentConnect,
			wantEntities: []string{"channel", "integration"},
		},
		{
			name:         "join channel",
			input:        "join my telegram channel",
			expected:     IntentConnect,
			wantEntities: []string{"channel"},
		},
		{
			name:         "hook up integration",
			input:        "hook up the api integration",
			expected:     IntentConnect,
			wantEntities: []string{"integration"},
		},
		{
			name:         "connect with whatsapp",
			input:        "connect with whatsapp",
			expected:     IntentConnect,
			wantEntities: []string{"channel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			if result.Confidence < 0.6 {
				t.Errorf("Parse(%q) Confidence = %v, want >= 0.6", tt.input, result.Confidence)
			}

			// Check for expected entity types
			foundEntityTypes := make(map[string]bool)
			for _, entity := range result.Entities {
				foundEntityTypes[entity.Type] = true
			}

			for _, wantType := range tt.wantEntities {
				if !foundEntityTypes[wantType] {
					t.Errorf("Parse(%q) missing entity type %q, found: %v", tt.input, wantType, result.Entities)
				}
			}
		})
	}
}

// Test configure intent detection
func TestParser_Parse_ConfigureIntent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		expected     Intent
		wantEntities []string
	}{
		{
			name:         "configure discord webhook",
			input:        "configure discord webhook",
			expected:     IntentConfigure,
			wantEntities: []string{"channel", "integration"},
		},
		{
			name:         "set up settings",
			input:        "set up my settings",
			expected:     IntentConfigure,
			wantEntities: nil,
		},
		{
			name:         "adjust config",
			input:        "adjust my config for telegram",
			expected:     IntentConfigure,
			wantEntities: []string{"channel"},
		},
		{
			name:         "customize discord",
			input:        "customize my discord bot",
			expected:     IntentConfigure,
			wantEntities: []string{"channel"},
		},
		{
			name:         "tweak settings",
			input:        "tweak the settings for slack",
			expected:     IntentConfigure,
			wantEntities: []string{"channel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			if result.Confidence < 0.6 {
				t.Errorf("Parse(%q) Confidence = %v, want >= 0.6", tt.input, result.Confidence)
			}

			if tt.wantEntities != nil {
				foundEntityTypes := make(map[string]bool)
				for _, entity := range result.Entities {
					foundEntityTypes[entity.Type] = true
				}

				for _, wantType := range tt.wantEntities {
					if !foundEntityTypes[wantType] {
						t.Errorf("Parse(%q) missing entity type %q, found: %v", tt.input, wantType, result.Entities)
					}
				}
			}
		})
	}
}

// Test enable intent detection
func TestParser_Parse_EnableIntent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		expected     Intent
		wantEntities []string
	}{
		{
			name:         "enable filesystem tool",
			input:        "enable the filesystem tool",
			expected:     IntentEnable,
			wantEntities: []string{"integration"},
		},
		{
			name:         "turn on mcp",
			input:        "turn on mcp integration",
			expected:     IntentEnable,
			wantEntities: []string{"integration"},
		},
		{
			name:         "activate slack",
			input:        "activate slack channel",
			expected:     IntentEnable,
			wantEntities: []string{"channel"},
		},
		{
			name:         "use tool",
			input:        "use the tool",
			expected:     IntentEnable,
			wantEntities: []string{"integration"},
		},
		{
			name:         "enable plugin",
			input:        "enable the webhook plugin",
			expected:     IntentEnable,
			wantEntities: []string{"integration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			if result.Confidence < 0.6 {
				t.Errorf("Parse(%q) Confidence = %v, want >= 0.6", tt.input, result.Confidence)
			}

			if tt.wantEntities != nil {
				foundEntityTypes := make(map[string]bool)
				for _, entity := range result.Entities {
					foundEntityTypes[entity.Type] = true
				}

				for _, wantType := range tt.wantEntities {
					if !foundEntityTypes[wantType] {
						t.Errorf("Parse(%q) missing entity type %q, found: %v", tt.input, wantType, result.Entities)
					}
				}
			}
		})
	}
}

// Test disable intent detection
func TestParser_Parse_DisableIntent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		expected     Intent
		wantEntities []string
	}{
		{
			name:         "disable filesystem tool",
			input:        "disable the filesystem tool",
			expected:     IntentDisable,
			wantEntities: []string{"integration"},
		},
		{
			name:         "turn off mcp",
			input:        "turn off mcp integration",
			expected:     IntentDisable,
			wantEntities: []string{"integration"},
		},
		{
			name:         "deactivate slack",
			input:        "deactivate slack channel",
			expected:     IntentDisable,
			wantEntities: []string{"channel"},
		},
		{
			name:         "stop discord",
			input:        "stop my discord bot",
			expected:     IntentDisable,
			wantEntities: []string{"channel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			if result.Confidence < 0.6 {
				t.Errorf("Parse(%q) Confidence = %v, want >= 0.6", tt.input, result.Confidence)
			}

			if tt.wantEntities != nil {
				foundEntityTypes := make(map[string]bool)
				for _, entity := range result.Entities {
					foundEntityTypes[entity.Type] = true
				}

				for _, wantType := range tt.wantEntities {
					if !foundEntityTypes[wantType] {
						t.Errorf("Parse(%q) missing entity type %q, found: %v", tt.input, wantType, result.Entities)
					}
				}
			}
		})
	}
}

// Test entity extraction for providers
func TestParser_Parse_ProviderEntity(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"openai", "setup openai provider", "openai"},
		{"anthropic", "install anthropic", "anthropic"},
		{"google", "setup google", "google"},
		{"claude", "configure claude", "claude"},
		{"gpt", "add gpt", "gpt"},
		{"gemini", "connect gemini", "gemini"},
		{"palm", "initialize palm", "palm"},
		{"mistral", "setup mistral", "mistral"},
		{"llama", "use llama", "llama"},
		{"ollama", "setup ollama", "ollama"},
		{"cohere", "install cohere", "cohere"},
		{"azure", "configure azure", "azure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			found := false
			for _, entity := range result.Entities {
				if entity.Type == "provider" && entity.Value == tt.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Parse(%q) did not find provider entity %q, found: %v", tt.input, tt.expected, result.Entities)
			}
		})
	}
}

// Test entity extraction for channels
func TestParser_Parse_ChannelEntity(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"telegram", "connect my telegram bot", "telegram"},
		{"discord", "setup discord", "discord"},
		{"slack", "link slack", "slack"},
		{"teams", "connect teams", "teams"},
		{"whatsapp", "integrate whatsapp", "whatsapp"},
		{"messenger", "attach messenger", "messenger"},
		{"signal", "setup signal", "signal"},
		{"matrix", "connect matrix", "matrix"},
		{"irc", "join irc", "irc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			found := false
			for _, entity := range result.Entities {
				if entity.Type == "channel" && entity.Value == tt.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Parse(%q) did not find channel entity %q, found: %v", tt.input, tt.expected, result.Entities)
			}
		})
	}
}

// Test entity extraction for integrations
func TestParser_Parse_IntegrationEntity(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"mcp", "enable mcp", "mcp"},
		{"webhook", "setup webhook", "webhook"},
		{"api", "connect api", "api"},
		{"rest", "configure rest", "rest"},
		{"graphql", "integrate graphql", "graphql"},
		{"grpc", "link grpc", "grpc"},
		{"websocket", "setup websocket", "websocket"},
		{"skill", "use skill", "skill"},
		{"tool", "enable tool", "tool"},
		{"plugin", "setup plugin", "plugin"},
		{"filesystem", "enable filesystem", "filesystem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			found := false
			for _, entity := range result.Entities {
				if entity.Type == "integration" && entity.Value == tt.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Parse(%q) did not find integration entity %q, found: %v", tt.input, tt.expected, result.Entities)
			}
		})
	}
}

// Test entity extraction for tokens
func TestParser_Parse_TokenEntity(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"token", "token: abc123", "abc123"},
		{"api key", "api key: xyz789", "xyz789"},
		{"secret", "secret: secret123", "secret123"},
		{"auth token", "auth token: token456", "token456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			found := false
			for _, entity := range result.Entities {
				if entity.Type == "token" && entity.Value == tt.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Parse(%q) did not find token entity %q, found: %v", tt.input, tt.expected, result.Entities)
			}
		})
	}
}

// Test SuggestSetupAction
func TestParser_SuggestSetupAction(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		input        string
		intent       Intent
		wantContains []string
	}{
		{
			name:         "setup with provider",
			input:        "setup openai",
			intent:       IntentSetup,
			wantContains: []string{"setup", "provider openai"},
		},
		{
			name:         "connect with channel",
			input:        "connect telegram",
			intent:       IntentConnect,
			wantContains: []string{"connect", "channel telegram"},
		},
		{
			name:         "configure with multiple entities",
			input:        "configure discord webhook",
			intent:       IntentConfigure,
			wantContains: []string{"config", "channel discord", "integration webhook"},
		},
		{
			name:         "enable with integration",
			input:        "enable filesystem",
			intent:       IntentEnable,
			wantContains: []string{"enable", "integration filesystem"},
		},
		{
			name:         "disable with channel",
			input:        "disable slack",
			intent:       IntentDisable,
			wantContains: []string{"disable", "channel slack"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResult{
				Intent:   tt.intent,
				Entities: []Entity{},
				Original: tt.input,
			}

			// Extract entities
			result.Entities = parser.extractEntities(tt.input)

			suggestions := parser.SuggestSetupAction(result)

			for _, want := range tt.wantContains {
				found := false
				for _, suggestion := range suggestions {
					if suggestion == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("SuggestSetupAction(%q) = %v, missing %q", tt.input, suggestions, want)
				}
			}
		})
	}
}

// Test case-insensitivity
func TestParser_Parse_CaseInsensitive(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		expected  Intent
		entityVal string
	}{
		{"uppercase", "SETUP OPENAI", IntentSetup, "openai"},
		{"lowercase", "setup openai", IntentSetup, "openai"},
		{"mixed case", "SeTuP OpEnAi", IntentSetup, "openai"},
		{"sentence", "Setup Openai Provider", IntentSetup, "openai"},
		{"uppercase connect", "CONNECT TELEGRAM", IntentConnect, "telegram"},
		{"mixed case enable", "EnAbLe FiLeSyStEm", IntentEnable, "filesystem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Intent != tt.expected {
				t.Errorf("Parse(%q) Intent = %v, want %v", tt.input, result.Intent, tt.expected)
			}

			found := false
			for _, entity := range result.Entities {
				if entity.Value == tt.entityVal {
					found = true
					break
				}
			}

			if !found && tt.entityVal != "" {
				t.Errorf("Parse(%q) did not find entity value %q, found: %v", tt.input, tt.entityVal, result.Entities)
			}
		})
	}
}

// Test GetIntentDescription for setup intents
func TestParser_GetIntentDescription_Setup(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		intent      Intent
		description string
	}{
		{IntentSetup, "Setting up or initializing something"},
		{IntentConnect, "Connecting or integrating with a service"},
		{IntentConfigure, "Configuring settings or options"},
		{IntentEnable, "Enabling a feature or service"},
		{IntentDisable, "Disabling a feature or service"},
	}

	for _, tt := range tests {
		t.Run(string(tt.intent), func(t *testing.T) {
			desc := parser.GetIntentDescription(tt.intent)
			if desc != tt.description {
				t.Errorf("GetIntentDescription(%v) = %v, want %v", tt.intent, desc, tt.description)
			}
		})
	}
}

// Test ambiguous detection for setup intents
func TestParser_IsAmbiguous_Setup(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		ambiguous bool
	}{
		{"clear setup", "setup openai provider", false},
		{"clear connect", "connect telegram bot", false},
		{"clear configure", "configure discord webhook", false},
		{"clear enable", "enable filesystem tool", false},
		{"unclear", "something something", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			ambiguous := parser.IsAmbiguous(result)

			if ambiguous != tt.ambiguous {
				t.Errorf("IsAmbiguous(Parse(%q)) = %v, want %v", tt.input, ambiguous, tt.ambiguous)
			}
		})
	}
}
