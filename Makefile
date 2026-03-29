.PHONY: run-mcp run-llm

run-mcp:
	@echo "Iniciando o Servidor MCP..."
	$(MAKE) -C ./services/mcp-server run-mcp

run-llm:
	@echo "Iniciando llm..."
	$(MAKE) -C ./apps/llm/ run-llm
# run-all:
# 	@echo "🚀 Subindo o Ecossistema inteiro..."
# 	$(MAKE) -C pasta_do_mcp run & \
# 	sleep 2; \
# 	$(MAKE) -C pasta_do_go run

# stop:
# 	@echo "Parando todos os serviços..."
# 	pkill -f "fastmcp" || true
# 	pkill -f "go run" || true