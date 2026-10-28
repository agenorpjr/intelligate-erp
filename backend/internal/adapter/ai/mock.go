package ai

import (
	"context"
	"encoding/json"
)

// MockClient é um cliente de LLM determinístico, usado em testes e no
// ambiente de desenvolvimento. Nunca realiza chamadas externas.
type MockClient struct {
	// Response, quando definido, é retornado verbatim.
	Response string
	// Err, quando definido, é retornado no lugar de uma resposta.
	Err error
}

// NewMockClient cria um MockClient com resposta padrão.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Provider identifica o cliente mock.
func (c *MockClient) Provider() string { return "mock" }

// Generate retorna a resposta configurada ou uma resposta JSON padrão.
func (c *MockClient) Generate(_ context.Context, _ []Message) (string, error) {
	if c.Err != nil {
		return "", c.Err
	}
	if c.Response != "" {
		return c.Response, nil
	}
	return defaultMockResponse(), nil
}

// defaultMockResponse devolve um JSON válido com os cinco campos esperados
// pelo agente, permitindo exercitar o parsing sem chamada externa.
func defaultMockResponse() string {
	resp := map[string]any{
		"resumo":      "Análise gerencial gerada pelo provedor mock (sem chamada externa).",
		"diagnostico": "Resposta de exemplo do cliente mock. Configure INTELLIGATE_AI_PROVIDER para usar um provedor real.",
		"evidencias": []string{
			"Indicadores fornecidos no contexto estruturado da loja e período.",
		},
		"acoes_recomendadas": []string{
			"Revisar os indicadores do dashboard para o período analisado.",
		},
		"limitacoes": []string{
			"Provedor mock: nenhuma inferência real foi executada.",
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}
