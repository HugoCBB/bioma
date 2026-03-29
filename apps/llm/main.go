package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	llm, err := ollama.New(
		ollama.WithModel("llama3"),
		ollama.WithServerURL("http://localhost:11435"),
	)
	if err != nil {
		log.Fatal("Erro ao conectar ao Ollama: ", err)
	}

	ctx := context.Background()

	fmt.Println("A processar resposta...")
	response, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Explica o que é o LangChain em três pontos curtos.",
	)
	if err != nil {
		log.Fatal("Erro ao gerar resposta: ", err)
	}

	fmt.Println("\n🤖 Resposta do Llama 3 (via Ollama):")
	fmt.Println(response)
}
