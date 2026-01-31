package nlp

import (
	"regexp"
	"strings"
)

// Intent represents the user's intended action
type Intent string

const (
	IntentCreate   Intent = "create"
	IntentRead     Intent = "read"
	IntentUpdate   Intent = "update"
	IntentDelete   Intent = "delete"
	IntentSearch   Intent = "search"
	IntentRun      Intent = "run"
	IntentTest     Intent = "test"
	IntentExplain  Intent = "explain"
	IntentRefactor Intent = "refactor"
	IntentDebug    Intent = "debug"
	// Setup-related intents
	IntentSetup     Intent = "setup"
	IntentConnect   Intent = "connect"
	IntentConfigure Intent = "configure"
	IntentEnable    Intent = "enable"
	IntentDisable   Intent = "disable"
	IntentUnknown   Intent = "unknown"
)

// Entity represents an extracted entity from the text
type Entity struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// ParseResult contains the parsed intent and entities
type ParseResult struct {
	Intent     Intent   `json:"intent"`
	Entities   []Entity `json:"entities"`
	Confidence float64  `json:"confidence"`
	Original   string   `json:"original"`
}

// Parser handles natural language parsing
type Parser struct {
	intentPatterns map[Intent][]*regexp.Regexp
	entityPatterns map[string]*regexp.Regexp
}

// NewParser creates a new NLP parser
func NewParser() *Parser {
	p := &Parser{
		intentPatterns: make(map[Intent][]*regexp.Regexp),
		entityPatterns: make(map[string]*regexp.Regexp),
	}

	p.initializePatterns()
	return p
}

// initializePatterns sets up regex patterns for intent recognition
func (p *Parser) initializePatterns() {
	// Create patterns
	p.intentPatterns[IntentCreate] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(create|make|generate|new)\b`),
		regexp.MustCompile(`(?i)\b(write|build|implement)\s+(a|an|the)?\s*\w+\b`),
	}

	p.intentPatterns[IntentRead] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(show|display|read|view|list)\b`),
		regexp.MustCompile(`(?i)\b(what is|tell me about|describe)\b`),
	}

	p.intentPatterns[IntentUpdate] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(update|modify|edit|fix|improve)\b`),
		regexp.MustCompile(`(?i)\b(make it|should be|needs to be)\b`),
	}

	p.intentPatterns[IntentDelete] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(delete|remove|destroy|drop|clean up)\b`),
		regexp.MustCompile(`(?i)\b(get rid of|take out)\b`),
	}

	p.intentPatterns[IntentSearch] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(search|look for|find|locate|where is)\b`),
		regexp.MustCompile(`(?i)\b(find all|search for)\b`),
	}

	p.intentPatterns[IntentRun] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(run|execute|launch|perform)\b`),
		regexp.MustCompile(`(?i)\b(do|carry out)\b`),
	}

	p.intentPatterns[IntentTest] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(test|check|verify|validate|ensure)\b`),
		regexp.MustCompile(`(?i)\b(make sure|confirm)\b`),
	}

	p.intentPatterns[IntentExplain] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(explain|describe|tell me|how does|what does)\b`),
		regexp.MustCompile(`(?i)\b(why|what is the reason)\b`),
	}

	p.intentPatterns[IntentRefactor] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(refactor|restructure|reorganize|clean up|optimize)\b`),
		regexp.MustCompile(`(?i)\b(make it better|simplify)\b`),
	}

	p.intentPatterns[IntentDebug] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(debug|fix|solve|resolve|troubleshoot)\b`),
		regexp.MustCompile(`(?i)\b(there is a|there's a) (bug|error|problem|issue)\b`),
	}

	// Setup-related intent patterns
	// Note: Order matters - more specific patterns first
	p.intentPatterns[IntentDisable] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(disable|deactivate)\b`),
		regexp.MustCompile(`(?i)\bturn\s+(my|the)?\s*\w+\s+off\b`),
		regexp.MustCompile(`(?i)\bturn\s+off\b`),
		regexp.MustCompile(`(?i)\bturn\s+(my|the)?\s*\w+\s+off\b`),
		regexp.MustCompile(`(?i)\bstop\s+(my|the)?\s*\w+\b`),
	}

	p.intentPatterns[IntentEnable] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(enable|activate)\b`),
		regexp.MustCompile(`(?i)\bturn\s+(my|the)?\s*\w+\s+on\b`),
		regexp.MustCompile(`(?i)\bturn\s+on\b`),
		regexp.MustCompile(`(?i)\buse\s+(my|a|the)?\s*(tool|feature|plugin|integration)\b`),
	}

	p.intentPatterns[IntentSetup] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(setup|install|initialize|prepare)\b`),
		regexp.MustCompile(`(?i)\bget\s+ready\b`),
		regexp.MustCompile(`(?i)\b(help\s+me\s+(to\s+)?(set\s+up|get\s+started|configure))\b`),
		regexp.MustCompile(`(?i)\b(i\s+want\s+to\s+(use|set\s+up))\b`),
	}

	p.intentPatterns[IntentConnect] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(connect|link|integrate|attach)\b`),
		regexp.MustCompile(`(?i)\b(add\s+(my|a|the)?\s*(bot|channel|integration))\b`),
		regexp.MustCompile(`(?i)\b(join|hook\s+up)\b`),
	}

	p.intentPatterns[IntentConfigure] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(configure|adjust|customize|tweak)\b`),
		regexp.MustCompile(`(?i)\b(set\s+up\s+(my|the)?\s*(settings|config|configuration))\b`),
		regexp.MustCompile(`(?i)\b(change\s+(my|the)?\s*(settings|config|configuration))\b`),
		regexp.MustCompile(`(?i)\bchange\b`), // This will score lower than specific configure patterns
	}

	// Entity patterns
	p.entityPatterns["file"] = regexp.MustCompile(`(?i)\b(file|document)\s+(?:named?\s+)?["']?([\w\-\.\/]+)["']?\b`)
	p.entityPatterns["function"] = regexp.MustCompile(`(?i)\b(function|method|def|routine)\s+(?:named?\s+)?["']?([\w\-]+)["']?\b`)
	p.entityPatterns["class"] = regexp.MustCompile(`(?i)\b(class|struct|type)\s+(?:named?\s+)?["']?([\w\-]+)["']?\b`)
	p.entityPatterns["path"] = regexp.MustCompile(`(?i)\b(path|directory|folder|dir)\s+(?:at\s+)?["']?([\w\-\.\/]+)["']?\b`)
	p.entityPatterns["language"] = regexp.MustCompile(`(?i)\b(in|using|with)\s+(go|golang|python|javascript|typescript|rust|java|c\+\+|ruby)\b`)

	// Setup-related entity patterns
	p.entityPatterns["provider"] = regexp.MustCompile(`(?i)\b(openai|anthropic|google|claude|gpt|gemini|palm|mistral|llama|ollama|cohere|azure)\b`)
	p.entityPatterns["channel"] = regexp.MustCompile(`(?i)\b(telegram|discord|slack|teams|whatsapp|messenger|signal|matrix|irc)\b`)
	p.entityPatterns["integration"] = regexp.MustCompile(`(?i)\b(mcp|webhook|api|rest|graphql|grpc|websocket|skill|tool|plugin|filesystem)\b`)
	p.entityPatterns["token"] = regexp.MustCompile(`(?i)\b(?:token|key|api[- ]?key|secret|auth[- ]?token)[:\s]+([\w\-\.]+)\b`)
}

// Parse analyzes text and extracts intent and entities
func (p *Parser) Parse(text string) ParseResult {
	result := ParseResult{
		Original: text,
		Entities: []Entity{},
	}

	// Detect intent
	intent, confidence := p.detectIntent(text)
	result.Intent = intent
	result.Confidence = confidence

	// Extract entities
	result.Entities = p.extractEntities(text)

	return result
}

// detectIntent determines the user's intent from text
func (p *Parser) detectIntent(text string) (Intent, float64) {
	scores := make(map[Intent]int)

	for intent, patterns := range p.intentPatterns {
		for _, pattern := range patterns {
			matches := pattern.FindAllStringIndex(text, -1)
			scores[intent] += len(matches)
		}
	}

	// Find highest scoring intent
	var bestIntent Intent = IntentUnknown
	var maxScore int

	for intent, score := range scores {
		if score > maxScore {
			maxScore = score
			bestIntent = intent
		}
	}

	// Calculate confidence (simplified)
	confidence := 0.5
	if maxScore > 0 {
		confidence = 0.5 + float64(maxScore)*0.1
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	if maxScore == 0 {
		return IntentUnknown, 0.3
	}

	return bestIntent, confidence
}

// extractEntities finds entities in the text
func (p *Parser) extractEntities(text string) []Entity {
	var entities []Entity

	for entityType, pattern := range p.entityPatterns {
		matches := pattern.FindAllStringSubmatchIndex(text, -1)

		for _, match := range matches {
			if len(match) >= 4 {
				// match[2:4] contains the first capture group (the value)
				value := text[match[2]:match[3]]
				entities = append(entities, Entity{
					Type:  entityType,
					Value: strings.ToLower(value),
					Start: match[2],
					End:   match[3],
				})
			}
		}
	}

	return entities
}

// SuggestCommands suggests CLI commands based on the parse result
func (p *Parser) SuggestCommands(result ParseResult) []string {
	var suggestions []string

	switch result.Intent {
	case IntentCreate:
		suggestions = append(suggestions, "create", "new", "add")
	case IntentRead:
		suggestions = append(suggestions, "show", "list", "get")
	case IntentUpdate:
		suggestions = append(suggestions, "update", "edit", "set")
	case IntentDelete:
		suggestions = append(suggestions, "delete", "remove", "rm")
	case IntentSearch:
		suggestions = append(suggestions, "search", "find", "grep")
	case IntentRun:
		suggestions = append(suggestions, "run", "exec", "start")
	case IntentTest:
		suggestions = append(suggestions, "test", "validate", "check")
	case IntentExplain:
		suggestions = append(suggestions, "explain", "describe", "doc")
	case IntentRefactor:
		suggestions = append(suggestions, "refactor", "optimize", "cleanup")
	case IntentDebug:
		suggestions = append(suggestions, "debug", "trace", "analyze")
	}

	return suggestions
}

// SuggestSetupAction suggests setup actions based on the parse result
func (p *Parser) SuggestSetupAction(result ParseResult) []string {
	var suggestions []string

	switch result.Intent {
	case IntentSetup:
		suggestions = append(suggestions, "setup", "init", "install")
	case IntentConnect:
		suggestions = append(suggestions, "connect", "link", "integrate")
	case IntentConfigure:
		suggestions = append(suggestions, "config", "settings", "customize")
	case IntentEnable:
		suggestions = append(suggestions, "enable", "activate", "start")
	case IntentDisable:
		suggestions = append(suggestions, "disable", "deactivate", "stop")
	default:
		return suggestions
	}

	// Add context-specific suggestions based on entities
	for _, entity := range result.Entities {
		switch entity.Type {
		case "provider":
			suggestions = append(suggestions, "provider "+entity.Value)
		case "channel":
			suggestions = append(suggestions, "channel "+entity.Value)
		case "integration":
			suggestions = append(suggestions, "integration "+entity.Value)
		case "token":
			suggestions = append(suggestions, "set token")
		}
	}

	return suggestions
}

// IsAmbiguous returns true if the intent confidence is low
func (p *Parser) IsAmbiguous(result ParseResult) bool {
	return result.Confidence < 0.6 || result.Intent == IntentUnknown
}

// GetIntentDescription returns a human-readable description of an intent
func (p *Parser) GetIntentDescription(intent Intent) string {
	descriptions := map[Intent]string{
		IntentCreate:    "Creating something new",
		IntentRead:      "Reading or viewing information",
		IntentUpdate:    "Updating or modifying something",
		IntentDelete:    "Deleting or removing something",
		IntentSearch:    "Searching for something",
		IntentRun:       "Running or executing something",
		IntentTest:      "Testing or validating something",
		IntentExplain:   "Explaining or describing something",
		IntentRefactor:  "Refactoring or optimizing code",
		IntentDebug:     "Debugging or fixing issues",
		IntentSetup:     "Setting up or initializing something",
		IntentConnect:   "Connecting or integrating with a service",
		IntentConfigure: "Configuring settings or options",
		IntentEnable:    "Enabling a feature or service",
		IntentDisable:   "Disabling a feature or service",
		IntentUnknown:   "Unclear intent",
	}

	if desc, ok := descriptions[intent]; ok {
		return desc
	}
	return string(intent)
}
