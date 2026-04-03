.PHONY: run-mcp run-llm

run-mcp:
	@echo "Iniciando o Servidor MCP..."
	$(MAKE) -C ./services/mcp-server run-mcp

run-api:
	go run ./cmd/api/main.go