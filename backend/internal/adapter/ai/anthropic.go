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
	defaultAnthropicBaseURL = "https://api.anthropic.com"
	defaultAnthropicModel   = "claude-3-5-sonnet-latest"
	anthropicVersion        = "2023-06-01"
	anthropicMaxTokens      = 2048
)

// AnthropicClient implementa LLMClient para a API de Messages da Anthropic.
type AnthropicClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewAnthropicClient cria um cliente Anthropic. A API key nunca é registrada em logs.
func NewAnthropicClient(apiKey, model, baseURL string) *AnthropicClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultAnthropicBaseURL
	}
	if strings.TrimSpace(model) == "" {
		model = defaultAnthropicModel
	}
	return &AnthropicClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// Provider identifica o provedor Anthropic.
func (c *AnthropicClient) Provider() string { return "anthropic" }

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate envia as mensagens ao endpoint de Messages e retorna o texto
// concatenado dos blocos de conteúdo.
func (c *AnthropicClient) Generate(ctx context.Context, messages []Message) (string, error) {
	// A Anthropic recebe o prompt de sistema em campo separado.
	var system string
	conversation := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == RoleSystem {
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
			continue
		}
		conversation = append(conversation, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	payload, err := json.Marshal(anthropicRequest{
		Model:     c.model,
		MaxTokens: anthropicMaxTokens,
		System:    system,
		Messages:  conversation,
	})
	if err != nil {
		return "", fmt.Errorf("falha ao serializar requisição Anthropic: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("falha ao criar requisição Anthropic: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha na chamada ao provedor Anthropic: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("falha ao ler resposta Anthropic: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("provedor Anthropic retornou status %d: %s", resp.StatusCode, extractAnthropicError(body))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("resposta inválida do provedor Anthropic: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("provedor Anthropic retornou erro: %s", parsed.Error.Message)
	}

	var sb strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if strings.TrimSpace(sb.String()) == "" {
		return "", ErrRespostaVazia
	}
	return sb.String(), nil
}

// extractAnthropicError tenta extrair a mensagem de erro do corpo da resposta.
func extractAnthropicError(body []byte) string {
	var parsed anthropicResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil {
		return parsed.Error.Message
	}
	msg := strings.TrimSpace(string(body))
	if len(msg) > 300 {
		msg = msg[:300]
	}
	return msg
}
