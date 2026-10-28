package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLLMClient_MockPadrao(t *testing.T) {
	client, err := NewLLMClient(Config{})
	require.NoError(t, err)
	assert.Equal(t, "mock", client.Provider())
}

func TestNewLLMClient_OpenAISemAPIKey(t *testing.T) {
	_, err := NewLLMClient(Config{Provider: "openai"})
	assert.ErrorIs(t, err, ErrAPIKeyAusente)
}

func TestNewLLMClient_AnthropicSemAPIKey(t *testing.T) {
	_, err := NewLLMClient(Config{Provider: "anthropic"})
	assert.ErrorIs(t, err, ErrAPIKeyAusente)
}

func TestNewLLMClient_ProviderInvalido(t *testing.T) {
	_, err := NewLLMClient(Config{Provider: "gemini"})
	assert.ErrorIs(t, err, ErrProviderInvalido)
}

func TestMockClient_RespostaJSONValida(t *testing.T) {
	client := NewMockClient()
	out, err := client.Generate(context.Background(), []Message{{Role: RoleUser, Content: "oi"}})
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	assert.Contains(t, parsed, "resumo")
	assert.Contains(t, parsed, "diagnostico")
	assert.Contains(t, parsed, "evidencias")
	assert.Contains(t, parsed, "acoes_recomendadas")
	assert.Contains(t, parsed, "limitacoes")
}

func TestMockClient_ErroConfigurado(t *testing.T) {
	client := &MockClient{Err: ErrRespostaVazia}
	_, err := client.Generate(context.Background(), nil)
	assert.ErrorIs(t, err, ErrRespostaVazia)
}

func TestOpenAIClient_Generate(t *testing.T) {
	var recebido openAIChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&recebido))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "resposta openai"}},
			},
		})
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", "gpt-4o-mini", server.URL)
	out, err := client.Generate(context.Background(), []Message{{Role: RoleUser, Content: "pergunta"}})
	require.NoError(t, err)
	assert.Equal(t, "resposta openai", out)
	assert.Equal(t, "gpt-4o-mini", recebido.Model)
	require.Len(t, recebido.Messages, 1)
}

func TestOpenAIClient_ErroStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid api key", "type": "auth"},
		})
	}))
	defer server.Close()

	client := NewOpenAIClient("bad-key", "", server.URL)
	_, err := client.Generate(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	// A API key nunca deve aparecer na mensagem de erro.
	assert.NotContains(t, err.Error(), "bad-key")
}

func TestAnthropicClient_Generate(t *testing.T) {
	var recebido anthropicRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, anthropicVersion, r.Header.Get("anthropic-version"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&recebido))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "resposta anthropic"},
			},
		})
	}))
	defer server.Close()

	client := NewAnthropicClient("test-key", "claude-3-5-sonnet-latest", server.URL)
	out, err := client.Generate(context.Background(), []Message{
		{Role: RoleSystem, Content: "sistema"},
		{Role: RoleUser, Content: "pergunta"},
	})
	require.NoError(t, err)
	assert.Equal(t, "resposta anthropic", out)
	// O prompt de sistema é separado das mensagens.
	assert.Equal(t, "sistema", recebido.System)
	require.Len(t, recebido.Messages, 1)
	assert.Equal(t, RoleUser, recebido.Messages[0].Role)
}

func TestAnthropicClient_RespostaVazia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []map[string]any{}})
	}))
	defer server.Close()

	client := NewAnthropicClient("test-key", "", server.URL)
	_, err := client.Generate(context.Background(), nil)
	assert.ErrorIs(t, err, ErrRespostaVazia)
}
