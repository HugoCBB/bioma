package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Agent struct {
	llm llms.Model
}

func New() (*Agent, error) {
	llm, err := openai.New(
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithToken(os.Getenv("GROQ_API_KEY")),
		openai.WithModel("llama-3.1-8b-instant"),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar LLM: %w", err)
	}
	return &Agent{llm: llm}, nil
}

func (a *Agent) Run(ctx context.Context, message, accessToken string) (string, error) {
	mcpClient, err := a.newMCPClient(ctx, accessToken)
	if err != nil {
		return "", err
	}
	defer mcpClient.Close()

	tools, err := a.listTools(ctx, mcpClient)
	if err != nil {
		return "", err
	}

	intent, err := a.extractIntent(ctx, message)
	if err != nil {
		return "", err
	}

	resolvedStart, resolvedEnd := a.resolveDatetime(intent)

	response, err := a.think(ctx, message, tools, resolvedStart, resolvedEnd)
	if err != nil {
		return "", err
	}

	toolCall, err := extractJSON(response)
	if err != nil || toolCall.Tool == "" {
		return response, nil
	}

	return a.callTool(ctx, mcpClient, toolCall, accessToken, resolvedStart, resolvedEnd)
}

func (a *Agent) newMCPClient(ctx context.Context, accessToken string) (*client.Client, error) {
	mcpClient, err := client.NewStreamableHttpClient(
		"http://localhost:8000/mcp",
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

func (a *Agent) listTools(ctx context.Context, mcpClient *client.Client) (string, error) {
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("You have access to these tools:\n")
	for _, t := range toolsResult.Tools {
		schema, _ := json.Marshal(t.InputSchema)
		sb.WriteString(fmt.Sprintf("- %s: %s\n  Schema: %s\n", t.Name, t.Description, schema))
	}
	return sb.String(), nil
}

func (a *Agent) think(ctx context.Context, message, tools, resolvedStart, resolvedEnd string) (string, error) {
	systemPrompt := fmt.Sprintf(`%s
Today's date: %s (weekday: %s)
Resolved start datetime: %s
Resolved end datetime: %s

Rules:
- If you need to use a tool, respond ONLY with a JSON object like this:
{"tool": "tool_name", "args": {"field1": "value1"}}
- Use exactly the resolved datetimes above, do NOT recalculate them.
- If the message is not about scheduling, respond normally in the user's language.
- Never mix JSON with text. Either pure JSON or pure text.
`, tools, time.Now().Format("2006-01-02"), time.Now().Weekday(), resolvedStart, resolvedEnd)

	return a.llm.Call(ctx, systemPrompt+"\nUser: "+message, llms.WithTemperature(0.1))
}

func (a *Agent) callTool(ctx context.Context, mcpClient *client.Client, toolCall toolCallPayload, accessToken, start, end string) (string, error) {
	toolCall.Args["start_datetime"] = start
	toolCall.Args["end_datetime"] = end
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

func (a *Agent) extractIntent(ctx context.Context, message string) (map[string]any, error) {
	prompt := fmt.Sprintf(`Today is %s (%s).

Extract the scheduling intent from this message and respond ONLY with a JSON object:
{
  "date": "YYYY-MM-DD of the event, calculated from today",
  "summary": "event title",
  "start_time": "HH:MM in 24h format, or null if not mentioned",
  "end_time": "HH:MM in 24h format, max 23:59, or null if not mentioned"
}

Rules:
- "amanhã" or "tomorrow" = today + 1 day
- "hoje" or "today" = today
- "segunda", "monday" = next monday, and so on for any language
- end_time must never be "24:00", return null if not mentioned
- Always return a valid date, never null

Message: %s`,
		time.Now().Format("2006-01-02"),
		time.Now().Weekday(),
		message,
	)

	response, err := a.llm.Call(ctx, prompt, llms.WithTemperature(0.0))
	if err != nil {
		return nil, err
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("intent JSON not found: %s", response)
	}

	var intent map[string]any
	if err := json.Unmarshal([]byte(response[start:end+1]), &intent); err != nil {
		return nil, err
	}
	return intent, nil
}

func (a *Agent) resolveDatetime(intent map[string]any) (string, string) {
	startTime, endTime := "09:00", "10:00"

	if v, ok := intent["start_time"].(string); ok && v != "null" && v != "" {
		startTime = v
	}
	if v, ok := intent["end_time"].(string); ok && v != "null" && v != "" && v != "24:00" {
		endTime = v
	}
	if endTime == "10:00" && startTime != "09:00" {
		if start, err := time.Parse("15:04", startTime); err == nil {
			endTime = start.Add(time.Hour).Format("15:04")
		}
	}

	date := time.Now().Format("2006-01-02")
	if v, ok := intent["date"].(string); ok && v != "" && v != "null" {
		date = v
	}

	return date + "T" + startTime + ":00", date + "T" + endTime + ":00"
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
