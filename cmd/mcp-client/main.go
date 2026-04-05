package main

import (
	"context"
	"fmt"
	"log"

	adapter "github.com/i2y/langchaingo-mcp-adapter"
	"github.com/mark3labs/mcp-go/client"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
)

func newMcpClient(url string) (*client.Client, error) {
	mcpClient, err := client.NewStreamableHttpClient(
		url,
	)
	if err != nil {
		return nil, fmt.Errorf("Erro ao criar client: %v", err)
	}
	return mcpClient, nil
}

func loadMCPTools(ctx context.Context, mcpClient *client.Client) ([]tools.Tool, error) {
	if err := mcpClient.Start(ctx); err != nil {
		return nil, fmt.Errorf("erro ao iniciar cliente MCP: %w", err)
	}

	mcpAdapter, err := adapter.New(mcpClient)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar adapter: %w", err)
	}

	mcpTools, err := mcpAdapter.Tools()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter ferramentas: %w", err)
	}

	return mcpTools, nil
}

func newLLM(model, serverURL string) (*ollama.LLM, error) {
	llm, err := ollama.New(
		ollama.WithModel(model),
		ollama.WithServerURL(serverURL),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar LLM: %w", err)
	}
	return llm, nil
}

func newAgent(llm *ollama.LLM, mcpTools []tools.Tool) *agents.Executor {
	return agents.NewExecutor(
		agents.NewConversationalAgent(llm, mcpTools),
	)
}

func run(ctx context.Context, agentExecutor *agents.Executor, prompt string) (string, error) {
	result, err := chains.Run(ctx, agentExecutor, prompt)
	if err != nil {
		return "", fmt.Errorf("erro ao executar agente: %w", err)
	}
	return result, nil
}
func main() {
	ctx := context.Background()

	mcpClient, err := newMcpClient("http://localhost:8000/mcp")
	if err != nil {
		log.Fatal(err)
	}
	defer mcpClient.Close()

	mcpTools, err := loadMCPTools(ctx, mcpClient)
	if err != nil {
		log.Fatal(err)
	}

	llm, err := newLLM("llama3.1", "http://localhost:11434")
	if err != nil {
		log.Fatal(err)
	}

	agentExecutor := newAgent(llm, mcpTools)

	result, err := run(ctx, agentExecutor, "My name is Hugo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
