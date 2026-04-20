package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
)

type Agent struct {
	llm      llms.Model
	provider Provider
}

func New(ctx context.Context) (*Agent, error) {
	provider := providerFromEnv()
	llm, err := newLLM(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar LLM (%s): %w", provider, err)
	}
	return &Agent{llm: llm, provider: provider}, nil
}

func (a *Agent) Run(ctx context.Context, message, accessToken string) (string, error) {
	mcpClient, err := a.newMCPClient(ctx, accessToken)
	if err != nil {
		return "", fmt.Errorf("erro ao conectar no MCP: %w", err)
	}
	defer mcpClient.Close()

	tools, err := a.fetchTools(ctx, mcpClient)
	if err != nil {
		return "", fmt.Errorf("erro ao listar tools: %w", err)
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt())},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(message)},
		},
	}

	const maxIterations = 10
	for i := 0; i < maxIterations; i++ {
		resp, err := a.llm.GenerateContent(ctx, messages,
			llms.WithTools(tools),
			llms.WithTemperature(0.1),
		)
		if err != nil {
			return "", fmt.Errorf("erro ao chamar LLM: %w", err)
		}

		choice := resp.Choices[0]

		if len(choice.ToolCalls) == 0 {
			return choice.Content, nil
		}

		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(choice.Content)},
		})

		for _, tc := range choice.ToolCalls {
			result, err := a.executeTool(ctx, mcpClient, tc, accessToken)
			if err != nil {
				result = fmt.Sprintf("Erro ao executar tool %s: %v", tc.FunctionCall.Name, err)
			}

			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    result,
					},
				},
			})
		}
	}

	return "", fmt.Errorf("agente excedeu o limite de %d iterações", maxIterations)
}

func systemPrompt() string {
	now := time.Now()
	return fmt.Sprintf(`Você é o Bioma, um assistente pessoal inteligente conectado aos serviços Google do usuário.
	Data/hora atual: %s (%s), timezone: America/Sao_Paulo

	Você tem acesso a ferramentas reais. Use-as sempre que o usuário pedir algo que envolva Google Calendar, Google Drive ou outros serviços conectados.

	Regras:
	- Responda sempre no idioma do usuário.
	- Ao agendar eventos, infira datas relativas corretamente ("amanhã", "segunda", etc.) com base na data atual.
	- Se faltar informação para executar uma ação (ex: horário do evento), pergunte antes de chamar a tool.
	- Nunca invente resultados de tools — use sempre o retorno real.
	- Seja direto e conciso nas respostas.
	- Para eventos recorrentes em múltiplos dias da semana, chame schedule_recurring_google_calendar UMA ÚNICA VEZ passando todos os dias em 'weekdays'.Nunca chame a tool mais de uma vez para o mesmo evento.
	`,
		now.Format("02/01/2006 15:04"),
		now.Weekday(),
	)
}
