package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func RunAgent(ctx context.Context, message, accessToken string) (string, error) {
	mcpClient, err := client.NewStreamableHttpClient(
		"http://localhost:8000/mcp",
		transport.WithHTTPHeaders(map[string]string{
			"X-Access-Token": accessToken,
		}),
	)
	if err != nil {
		return "", err
	}
	defer mcpClient.Close()

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return "", err
	}

	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return "", err
	}

	var toolsDesc strings.Builder
	toolsDesc.WriteString("You have access to these tools:\n")
	for _, t := range toolsResult.Tools {
		schema, _ := json.Marshal(t.InputSchema)
		toolsDesc.WriteString(fmt.Sprintf("- %s: %s\n  Schema: %s\n", t.Name, t.Description, schema))
	}

	llm, err := ollama.New(
		ollama.WithModel("llama3.1"),
		ollama.WithServerURL("http://localhost:11434"),
	)
	if err != nil {
		return "", err
	}

	systemPrompt := fmt.Sprintf(`%s
Rules:
- If you need to use a tool, respond ONLY with a JSON object like this:
{"tool": "tool_name", "args": {"field1": "value1", "field2": "value2"}}
- If you don't need a tool, respond normally in the user's language.
- Never mix JSON with text. Either pure JSON or pure text.
`, toolsDesc.String())

	response, err := llm.Call(ctx, systemPrompt+"\nUser: "+message,
		llms.WithTemperature(0.1),
	)
	if err != nil {
		return "", err
	}

	toolCall, err := extractJSON(response)
	if err != nil || toolCall.Tool == "" {
		return response, nil
	}

	toolCall.Args["access_token"] = accessToken

	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolCall.Tool,
			Arguments: toolCall.Args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("erro ao chamar tool %s: %w", toolCall.Tool, err)
	}

	if len(result.Content) == 0 {
		return "Tool executada mas sem resposta.", nil
	}

	return result.Content[0].(mcp.TextContent).Text, nil
}

type toolCallPayload struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

func extractJSON(response string) (toolCallPayload, error) {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return toolCallPayload{}, fmt.Errorf("no json found")
	}

	jsonStr := response[start : end+1]
	var payload toolCallPayload
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		return toolCallPayload{}, err
	}
	return payload, nil
}
