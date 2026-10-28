package ai

import (
	"fmt"
	"strings"
)

// Config agrupa os parâmetros de configuração do provedor de IA.
type Config struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// NewLLMClient cria o cliente de LLM adequado ao provedor configurado.
//
// Provedores suportados: "openai", "anthropic" e "mock". Quando o provedor é
// vazio, assume-se "mock" (sem chamada externa). A API key nunca é registrada
// em logs nem incluída em mensagens de erro.
func NewLLMClient(cfg Config) (LLMClient, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "mock":
		return NewMockClient(), nil
	case "openai":
		if strings.TrimSpace(cfg.APIKey) == "" {
			return nil, ErrAPIKeyAusente
		}
		return NewOpenAIClient(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	case "anthropic":
		if strings.TrimSpace(cfg.APIKey) == "" {
			return nil, ErrAPIKeyAusente
		}
		return NewAnthropicClient(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrProviderInvalido, cfg.Provider)
	}
}
