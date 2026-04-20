package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type Provider string

const (
	ProviderGroq   Provider = "groq"
	ProviderGemini Provider = "gemini"
	ProviderOllama Provider = "ollama"
)

func providerFromEnv() Provider {
	switch strings.ToLower(os.Getenv("LLM_PROVIDER")) {
	case "gemini":
		return ProviderGemini
	case "ollama":
		return ProviderOllama
	default:
		return ProviderGroq
	}
}

func newLLM(ctx context.Context, provider Provider) (llms.Model, error) {
	switch provider {

	case ProviderGroq:
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GROQ_API_KEY não definida")
		}
		model := os.Getenv("GROQ_MODEL")
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		return openai.New(
			openai.WithBaseURL("https://api.groq.com/openai/v1"),
			openai.WithToken(apiKey),
			openai.WithModel(model),
		)

	case ProviderGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY não definida")
		}
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = "gemini-2.0-flash"
		}
		return googleai.New(ctx,
			googleai.WithAPIKey(apiKey),
			googleai.WithDefaultModel(model),
		)

	case ProviderOllama:
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "llama3.1"
		}
		baseURL := os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return ollama.New(
			ollama.WithModel(model),
			ollama.WithServerURL(baseURL),
		)

	default:
		return nil, fmt.Errorf("provider desconhecido: %q — use groq, gemini ou ollama", provider)
	}
}
