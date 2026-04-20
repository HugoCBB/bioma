package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
)

func (a *Agent) newMCPClient(ctx context.Context, accessToken string) (*client.Client, error) {
	mcpURL := os.Getenv("MCP_SERVER_URL")
	if mcpURL == "" {
		mcpURL = "http://localhost:8000/mcp"
	}
	mcpClient, err := client.NewStreamableHttpClient(
		mcpURL,
		transport.WithHTTPHeaders(map[string]string{
			"X-Access-Token": accessToken,
		}),
	)
	if err != nil {
		return nil, err
	}
	if _, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		return nil, err
	}
	return mcpClient, nil
}

func (a *Agent) fetchTools(ctx context.Context, mcpClient *client.Client) ([]llms.Tool, error) {
	result, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	tools := make([]llms.Tool, 0, len(result.Tools))
	for _, t := range result.Tools {
		schema, _ := json.Marshal(t.InputSchema)

		var params map[string]any
		if err := json.Unmarshal(schema, &params); err != nil {
			params = map[string]any{"type": "object"}
		}

		cleanSchema(params)

		tools = append(tools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return tools, nil
}

func (a *Agent) executeTool(ctx context.Context, mcpClient *client.Client, tc llms.ToolCall, accessToken string) (string, error) {
	args := make(map[string]any)
	if raw := tc.FunctionCall.Arguments; raw != "" && raw != "null" {
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "", fmt.Errorf("argumentos inválidos: %w", err)
		}
	}

	args["access_token"] = accessToken

	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      tc.FunctionCall.Name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "Tool executada sem retorno.", nil
	}

	return result.Content[0].(mcp.TextContent).Text, nil
}

func cleanSchema(schema map[string]any) {
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return
	}

	delete(props, "access_token")

	if required, ok := schema["required"].([]any); ok {
		filtered := make([]any, 0, len(required))
		for _, r := range required {
			if r.(string) != "access_token" {
				filtered = append(filtered, r)
			}
		}
		schema["required"] = filtered
	}

	for _, v := range props {
		prop, ok := v.(map[string]any)
		if !ok {
			continue
		}
		delete(prop, "title")
		delete(prop, "default")
	}
}
