// Package ai fornece um adapter isolado para provedores de LLM.
//
// O pacote não conhece o domínio do ERP: recebe mensagens estruturadas e
// devolve texto. Isso mantém o acoplamento baixo e permite trocar de provedor
// (ou usar o mock) sem alterar os casos de uso.
package ai

import (
	"context"
	"errors"
)

// Papéis suportados em uma mensagem estruturada.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Erros esperados do adapter de IA.
var (
	ErrProviderInvalido = errors.New("provedor de IA inválido")
	ErrAPIKeyAusente    = errors.New("INTELLIGATE_AI_API_KEY é obrigatória para o provedor selecionado")
	ErrRespostaVazia    = errors.New("resposta vazia do provedor de IA")
)

// Message representa uma mensagem estruturada enviada ao modelo.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMClient abstrai um provedor de LLM.
type LLMClient interface {
	// Generate retorna a resposta textual do modelo para as mensagens dadas.
	Generate(ctx context.Context, messages []Message) (string, error)
	// Provider retorna o identificador do provedor (openai, anthropic, mock).
	Provider() string
}
