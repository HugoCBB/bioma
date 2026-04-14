<div align="center">

# 🌿 Bioma

**Um agente de IA pessoal conectado à sua conta Google.**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![Python](https://img.shields.io/badge/Python-3.10+-3776AB?style=for-the-badge&logo=python&logoColor=white)](https://python.org)
[![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://gin-gonic.com)
[![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io)
[![MCP](https://img.shields.io/badge/MCP-Protocol-6C3483?style=for-the-badge)](https://modelcontextprotocol.io)
[![Ollama](https://img.shields.io/badge/Ollama-LLaMA3.1-black?style=for-the-badge)](https://ollama.com)

</div>

---

## ✨ Sobre o Projeto

O **Bioma** é um assistente de IA pessoal que se conecta aos serviços Google do usuário (como Google Calendar e Google Drive) via autenticação OAuth2. Usando [LangChainGo](https://github.com/tmc/langchaingo) com Ollama (LLaMA 3.1), o agente é capaz de interpretar mensagens em linguagem natural e executar ações reais — como agendar eventos no Google Calendar — por meio do protocolo **MCP (Model Context Protocol)**.

### 🔑 Principais Funcionalidades

- 🔐 **Autenticação Google OAuth2** — Login seguro e armazenamento de tokens no Redis
- 🤖 **Agente LLM Local** — Usa Ollama (LLaMA 3.1) para processar linguagem natural
- 📅 **Google Calendar** — Cria eventos via linguagem natural
- 💾 **Google Drive** — Integração planejada para salvar arquivos
- 🔌 **MCP Server** — Servidor de ferramentas em Python via FastMCP

---

## 🏗️ Arquitetura

```
bioma/
├── cmd/
│   └── server/
│       └── main.go          # Entrypoint da API principal (Go)
│
├── internal/
│   ├── agent/
│   │   └── agent.go         # Lógica do agente LLM (LangChainGo + MCP)
│   ├── config/
│   │   └── ...              # Configurações de ambiente e OAuth2
│   ├── domain/
│   │   └── chat.go          # Modelos de domínio
│   ├── handler/
│   │   ├── auth.go          # Handlers de autenticação Google
│   │   ├── chat.go          # Handler do chat com o agente
│   │   └── route.go         # Definição das rotas (Gin)
│   ├── infra/
│   │   └── redis.go         # Inicialização do cliente Redis
│   └── repository/
│       └── ...              # Persistência de tokens no Redis
│
└── services/
    └── mcp-server/
        ├── main.py          # Servidor MCP em Python (FastMCP)
        ├── domain/          # Modelos e lógica das ferramentas
        └── requirements.txt # Dependências Python
```

### 🔄 Fluxo de Dados

```
Usuário → POST /chat → ChatHandler → Agent (LLaMA 3.1)
                                         ↓
                              Precisa de uma tool?
                           ↙ Não          ↘ Sim
                    Resposta          MCP Server (Python)
                    em texto               ↓
                                   Google Calendar API
                                          ↓
                                    Resposta final
```

---

## 🚀 Como Rodar

### Pré-requisitos

| Ferramenta | Versão | Descrição |
|---|---|---|
| [Go](https://go.dev) | 1.26+ | Runtime da API principal |
| [Python](https://python.org) | 3.10+ | Runtime do MCP Server |
| [Redis](https://redis.io) | qualquer | Armazenamento de tokens |
| [Ollama](https://ollama.com) | qualquer | Servidor LLM local |
| [Docker](https://docker.com) | qualquer | Para subir o Redis facilmente |

---

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/bioma.git
cd bioma
```

### 2. Configure as variáveis de ambiente

```bash
cp .env.example .env
```

Edite o `.env` com suas credenciais do Google Cloud:

```env
GOOGLE_CLIENT_ID=seu_client_id_aqui
GOOGLE_CLIENT_SECRET=seu_client_secret_aqui
```

> 💡 Para obter as credenciais, acesse o [Google Cloud Console](https://console.cloud.google.com), crie um projeto, habilite a **Google Calendar API** e crie credenciais OAuth2 do tipo _Web Application_.
> Adicione `http://localhost:8080/auth/callback` como URI de redirecionamento autorizado.

### 3. Suba o Redis com Docker

```bash
docker-compose up -d
```

### 4. Baixe o modelo LLaMA no Ollama

```bash
ollama pull llama3.1
```

### 5. Inicie o MCP Server (Python)

```bash
make run-mcp
```

Ou manualmente:

```bash
cd services/mcp-server
pip install -r requirements.txt
python main.py
```

> O servidor MCP ficará disponível em `http://localhost:8000/mcp`

### 6. Inicie a API principal (Go)

```bash
make run-api
```

Ou manualmente:

```bash
go run ./cmd/server/main.go
```

> A API ficará disponível em `http://localhost:8080`

---

## 📡 Endpoints da API

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/` | Health check |
| `GET` | `/auth/login?user_id=<id>` | Inicia o fluxo OAuth2 com o Google |
| `GET` | `/auth/callback` | Callback do OAuth2 (redirecionamento automático) |
| `POST` | `/chat/` | Envia uma mensagem para o agente |

### Exemplo de uso do chat

```bash
curl -X POST http://localhost:8080/chat/ \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "meu-usuario",
    "message": "Agende uma reunião amanhã às 10h com título Standup"
  }'
```

### Fluxo de autenticação

1. Acesse `http://localhost:8080/auth/login?user_id=meu-usuario`
2. Faça login com sua conta Google e conceda as permissões
3. O token é salvo automaticamente no Redis
4. Agora você pode usar o `/chat/` com o `user_id` configurado

---

## 🛠️ Desenvolvimento com Hot Reload

O projeto está configurado com [Air](https://github.com/air-verse/air) para hot reload durante o desenvolvimento:

```bash
air
```

---

## 🧩 Tecnologias

| Camada | Tecnologia |
|---|---|
| API HTTP | [Gin](https://gin-gonic.com) |
| LLM | [LangChainGo](https://github.com/tmc/langchaingo) + [Ollama](https://ollama.com) |
| Protocolo de Tools | [MCP (FastMCP)](https://gofastmcp.com) |
| Autenticação | Google OAuth2 |
| Cache / Tokens | Redis |
| MCP Server | Python + FastAPI |

---

<div align="center">
  <sub>Feito com 💚 e Go</sub>
</div>
