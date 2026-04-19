package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

	intent, err := extractIntent(ctx, llm, message)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair intent: %w", err)
	}

	resolvedStart, resolvedEnd := resolveDatetime(intent)

	systemPrompt := fmt.Sprintf(`%s
Today's date: %s (weekday: %s)
Resolved start datetime: %s
Resolved end datetime: %s

Rules:
- If you need to use a tool, respond ONLY with a JSON object like this:
{"tool": "tool_name", "args": {"field1": "value1", "field2": "value2"}}
- Use exactly the "Resolved start datetime" and "Resolved end datetime" above for the event datetimes — do NOT recalculate them.
- If the user's message is not about scheduling, respond normally in the user's language.
- Never mix JSON with text. Either pure JSON or pure text.
`, toolsDesc.String(), time.Now().Format("2006-01-02"), time.Now().Weekday(), resolvedStart, resolvedEnd)

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

	toolCall.Args["start_datetime"] = resolvedStart
	toolCall.Args["end_datetime"] = resolvedEnd
	toolCall.Args["access_token"] = accessToken

	fmt.Printf("[DEBUG] Tool: %s\nArgs: %+v\n", toolCall.Tool, toolCall.Args)

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

func extractIntent(ctx context.Context, llm *ollama.LLM, message string) (map[string]any, error) {
	prompt := fmt.Sprintf(`Extract the scheduling intent from this message and respond ONLY with a JSON object, no extra text:
		{
		"weekday": "sunday|monday|tuesday|wednesday|thursday|friday|saturday or null",
		"summary": "event title",
		"start_time": "HH:MM in 24h format, or null if not mentioned",
		"end_time": "HH:MM in 24h format, max 23:59, or null if not mentioned"
		}

		Important: end_time must never be "24:00". If not mentioned, return null.

		Message: %s`, message)

	response, err := llm.Call(ctx, prompt, llms.WithTemperature(0.0))
	if err != nil {
		return nil, err
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("intent JSON not found in response: %s", response)
	}

	var intent map[string]any
	if err := json.Unmarshal([]byte(response[start:end+1]), &intent); err != nil {
		return nil, fmt.Errorf("erro ao parsear intent: %w", err)
	}

	return intent, nil
}
func resolveDatetime(intent map[string]any) (string, string) {
	startTime := "09:00"
	endTime := "10:00"

	if v, ok := intent["start_time"].(string); ok && v != "null" && v != "" {
		startTime = v
	}
	if v, ok := intent["end_time"].(string); ok && v != "null" && v != "" && v != "24:00" {
		endTime = v
	}

	if endTime == "10:00" && startTime != "09:00" {
		start, err := time.Parse("15:04", startTime)
		if err == nil {
			endTime = start.Add(time.Hour).Format("15:04")
		}
	}

	date := time.Now()
	if v, ok := intent["weekday"].(string); ok && v != "null" && v != "" {
		date = nextWeekday(parseWeekday(v))
	}

	dateStr := date.Format("2006-01-02")
	return dateStr + "T" + startTime + ":00", dateStr + "T" + endTime + ":00"
}
func nextWeekday(target time.Weekday) time.Time {
	now := time.Now()
	daysUntil := int(target) - int(now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return now.AddDate(0, 0, daysUntil)
}

func parseWeekday(s string) time.Weekday {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	return days[strings.ToLower(s)]
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
