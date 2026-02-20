package ai

import (
	"context"
)

// Message role types.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// Message represents a single turn in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client is the interface for all AI providers.
type Client interface {
	// Completion returns a response from the model.
	Completion(ctx context.Context, messages []Message) (string, error)
	// Name returns the provider name.
	Name() string
}

// Provider represents a supported AI provider with its metadata.
type Provider struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Models      []string `json:"models"`
	Description string   `json:"description"`
}

// GetProviders returns a list of recommended providers and models, inspired by catwalk.
func GetProviders() []Provider {
	return []Provider{
		{
			ID:   "anthropic",
			Name: "Anthropic",
			Models: []string{
				"claude-3-5-sonnet-20240620",
				"claude-3-opus-20240229",
				"claude-3-haiku-20240307",
			},
			Description: "Best-in-class reasoning and coding capabilities.",
		},
		{
			ID:   "google",
			Name: "Google Gemini",
			Models: []string{
				"gemini-1.5-pro",
				"gemini-1.5-flash",
			},
			Description: "Large context window and fast performance.",
		},
		{
			ID:   "openai",
			Name: "OpenAI",
			Models: []string{
				"gpt-4o",
				"gpt-4-turbo",
				"gpt-3.5-turbo",
			},
			Description: "Industry standard for performance and instruction following.",
		},
		{
			ID:   "ollama",
			Name: "Ollama (Local)",
			Models: []string{
				"llama3",
				"mistral",
				"phi3",
			},
			Description: "Run models locally on your homelab. No API key needed.",
		},
	}
}
