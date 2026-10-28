package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4o-mini"
)

// OpenAIClient implementa LLMClient para a API de Chat Completions da OpenAI
// (e gateways compatíveis, via BaseURL customizada).
type OpenAIClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewOpenAIClient cria um cliente OpenAI. A API key nunca é registrada em logs.
func NewOpenAIClient(apiKey, model, baseURL string) *OpenAIClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOpenAIBaseURL
	}
	if strings.TrimSpace(model) == "" {
		model = defaultOpenAIModel
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// Provider identifica o provedor OpenAI.
func (c *OpenAIClient) Provider() string { return "openai" }

type openAIChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Generate envia as mensagens ao endpoint de chat completions e retorna o
// conteúdo textual da primeira escolha.
func (c *OpenAIClient) Generate(ctx context.Context, messages []Message) (string, error) {
	payload, err := json.Marshal(openAIChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("falha ao serializar requisição OpenAI: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("falha ao criar requisição OpenAI: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha na chamada ao provedor OpenAI: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("falha ao ler resposta OpenAI: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Não inclui a API key; apenas status e mensagem do provedor.
		return "", fmt.Errorf("provedor OpenAI retornou status %d: %s", resp.StatusCode, extractProviderError(body))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("resposta inválida do provedor OpenAI: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("provedor OpenAI retornou erro: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", ErrRespostaVazia
	}
	return parsed.Choices[0].Message.Content, nil
}

// extractProviderError tenta extrair a mensagem de erro do corpo da resposta.
func extractProviderError(body []byte) string {
	var parsed openAIChatResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil {
		return parsed.Error.Message
	}
	msg := strings.TrimSpace(string(body))
	if len(msg) > 300 {
		msg = msg[:300]
	}
	return msg
}
