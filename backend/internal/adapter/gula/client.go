package gula

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PDVClient define a interface para clientes de PDV externos
type PDVClient interface {
	// BuscarVendas busca vendas no PDV externo por período
	BuscarVendas(ctx context.Context, dataInicio, dataFim time.Time) ([]*VendaGula, error)
	// TestarConexao verifica se a conexão com o PDV está funcionando
	TestarConexao(ctx context.Context) error
}

// GulaClient implementa PDVClient para o Gula Menu
type GulaClient struct {
	httpClient   *http.Client
	urlBase      string
	apiKey       string
	clientID     string
	clientSecret string
	ambiente     string
}

// ConfigGula contém as configurações para criar um GulaClient
type ConfigGula struct {
	URLBase      string
	APIKey       string
	ClientID     string
	ClientSecret string
	Ambiente     string // "sandbox" ou "production"
	Timeout      time.Duration
}

// NewGulaClient cria uma nova instância do cliente Gula
func NewGulaClient(config ConfigGula) *GulaClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GulaClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		urlBase:      config.URLBase,
		apiKey:       config.APIKey,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		ambiente:     config.Ambiente,
	}
}

// VendaGula representa uma venda retornada pela API do Gula Menu
// NOTA: Esta estrutura é baseada em especulação da API do Gula Menu.
// Deve ser ajustada quando a documentação oficial estiver disponível.
type VendaGula struct {
	ID               string          `json:"id"`
	NumeroExterno    string          `json:"numero_externo"`
	DataHora         string          `json:"data_hora"`
	Status           string          `json:"status"`
	ValorTotal       float64         `json:"valor_total"`
	ValorDesconto    float64         `json:"valor_desconto"`
	ValorAcrescimo   float64         `json:"valor_acrescimo"`
	ClienteNome      string          `json:"cliente_nome"`
	ClienteDocumento string          `json:"cliente_documento"`
	ClienteEmail     string          `json:"cliente_email"`
	Observacoes      string          `json:"observacoes"`
	Itens            []ItemVendaGula `json:"itens"`
	Pagamentos       []PagamentoGula `json:"pagamentos"`
	Taxas            []TaxaGula      `json:"taxas"`
	Troco            float64         `json:"troco"`
}

// ItemVendaGula representa um item de venda do Gula
type ItemVendaGula struct {
	ID            string  `json:"id"`
	SKUExterno    string  `json:"sku_externo"`
	Nome          string  `json:"nome"`
	Quantidade    float64 `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	Desconto      float64 `json:"desconto"`
	Modificadores string  `json:"modificadores"`
	Observacoes   string  `json:"observacoes"`
}

// PagamentoGula representa um pagamento do Gula
type PagamentoGula struct {
	ID                string  `json:"id"`
	FormaPagamento    string  `json:"forma_pagamento"`
	Valor             float64 `json:"valor"`
	Taxa              float64 `json:"taxa"`
	Troco             float64 `json:"troco"`
	ReferenciaExterna string  `json:"referencia_externa"`
	Status            string  `json:"status"`
}

// TaxaGula representa uma taxa aplicada na venda
type TaxaGula struct {
	ID    string  `json:"id"`
	Nome  string  `json:"nome"`
	Valor float64 `json:"valor"`
	Tipo  string  `json:"tipo"`
}

// RespostaBuscarVendas representa a resposta paginada da API do Gula
type RespostaBuscarVendas struct {
	Data      []*VendaGula  `json:"data"`
	Paginacao PaginacaoGula `json:"paginacao"`
}

// PaginacaoGula representa informações de paginação
type PaginacaoGula struct {
	PaginaAtual  int `json:"pagina_atual"`
	TotalPaginas int `json:"total_paginas"`
	TotalItens   int `json:"total_itens"`
	PorPagina    int `json:"por_pagina"`
}

// BuscarVendas busca vendas no Gula Menu por período
// NOTA: Endpoints e parâmetros são especulativos e devem ser ajustados
// quando a documentação oficial da API do Gula estiver disponível.
func (c *GulaClient) BuscarVendas(ctx context.Context, dataInicio, dataFim time.Time) ([]*VendaGula, error) {
	// Construir URL com query parameters
	url := fmt.Sprintf("%s/api/v1/vendas", c.urlBase)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Adicionar headers de autenticação
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.clientID != "" {
		req.Header.Set("X-Client-ID", c.clientID)
	}

	// Query parameters
	q := req.URL.Query()
	q.Add("data_inicio", dataInicio.Format(time.RFC3339))
	q.Add("data_fim", dataFim.Format(time.RFC3339))
	q.Add("ambiente", c.ambiente)
	req.URL.RawQuery = q.Encode()

	// Executar requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API do Gula: status %d", resp.StatusCode)
	}

	var resposta RespostaBuscarVendas
	if err := json.NewDecoder(resp.Body).Decode(&resposta); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return resposta.Data, nil
}

// TestarConexao verifica se a conexão com o Gula está funcionando
func (c *GulaClient) TestarConexao(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/health", c.urlBase)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao executar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check falhou: status %d", resp.StatusCode)
	}

	return nil
}

// MapearParaSalesEvent converte uma VendaGula para o payload do endpoint de ingestão
func (v *VendaGula) MapearParaSalesEvent(pdvOrigemID uuid.UUID) (map[string]interface{}, string, error) {
	// Gerar idempotency_key estável baseada no ID externo e data/hora
	idempotencyKey := fmt.Sprintf("gula-%s-%s", v.NumeroExterno, v.DataHora)

	// Parse data/hora
	dataHora, err := time.Parse(time.RFC3339, v.DataHora)
	if err != nil {
		// Tentar outros formatos comuns
		formatos := []string{
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		var parseErr error
		for _, fmt := range formatos {
			dataHora, parseErr = time.Parse(fmt, v.DataHora)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			return nil, "", fmt.Errorf("formato de data/hora inválido: %w", parseErr)
		}
	}

	// Mapear itens
	itens := make([]map[string]interface{}, len(v.Itens))
	for i, item := range v.Itens {
		itens[i] = map[string]interface{}{
			"sku_externo":    item.SKUExterno,
			"nome":           item.Nome,
			"quantidade":     item.Quantidade,
			"preco_unitario": item.PrecoUnitario,
			"valor_total":    item.ValorTotal,
			"desconto":       item.Desconto,
			"modificadores":  item.Modificadores,
			"observacoes":    item.Observacoes,
		}
	}

	// Mapear pagamentos
	pagamentos := make([]map[string]interface{}, len(v.Pagamentos))
	for i, pag := range v.Pagamentos {
		pagamentos[i] = map[string]interface{}{
			"forma_pagamento":    pag.FormaPagamento,
			"valor":              pag.Valor,
			"taxa":               pag.Taxa,
			"troco":              pag.Troco,
			"referencia_externa": pag.ReferenciaExterna,
			"status":             pag.Status,
		}
	}

	// Mapear taxas
	taxas := make([]map[string]interface{}, len(v.Taxas))
	for i, taxa := range v.Taxas {
		taxas[i] = map[string]interface{}{
			"nome":  taxa.Nome,
			"valor": taxa.Valor,
			"tipo":  taxa.Tipo,
		}
	}

	payload := map[string]interface{}{
		"numero_externo":    v.NumeroExterno,
		"data_hora":         dataHora.Format(time.RFC3339),
		"status":            v.Status,
		"valor_total":       v.ValorTotal,
		"valor_desconto":    v.ValorDesconto,
		"valor_acrescimo":   v.ValorAcrescimo,
		"cliente_nome":      v.ClienteNome,
		"cliente_documento": v.ClienteDocumento,
		"cliente_email":     v.ClienteEmail,
		"observacoes":       v.Observacoes,
		"itens":             itens,
		"pagamentos":        pagamentos,
		"taxas":             taxas,
		"troco":             v.Troco,
		"pdv_origem_id":     pdvOrigemID.String(),
	}

	return payload, idempotencyKey, nil
}

// Erros comuns
var (
	ErrAutenticacaoFalhou = errors.New("falha na autenticação com o Gula Menu")
	ErrRateLimit          = errors.New("rate limit excedido na API do Gula")
	ErrRespostaInvalida   = errors.New("resposta inválida da API do Gula")
)
