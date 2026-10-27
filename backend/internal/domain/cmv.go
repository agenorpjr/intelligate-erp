package domain

import (
	"time"

	"github.com/google/uuid"
)

type CMVResumo struct {
	EmpresaID          uuid.UUID `json:"empresa_id"`
	LojaID             uuid.UUID `json:"loja_id"`
	DataInicio         time.Time `json:"data_inicio"`
	DataFim            time.Time `json:"data_fim"`
	TotalVendas        float64   `json:"total_vendas"`
	TotalCMV           float64   `json:"total_cmv"`
	MargemBruta        float64   `json:"margem_bruta"`
	MargemBrutaPercent float64   `json:"margem_bruta_percent"`
	ProdutosComFicha   int       `json:"produtos_com_ficha"`
	ProdutosSemFicha   int       `json:"produtos_sem_ficha_tecnica"`
	ItensVendidos      int       `json:"itens_vendidos"`
}

type CMVProdutoDetalhe struct {
	ProdutoID            uuid.UUID `json:"produto_id"`
	ProdutoNome          string    `json:"produto_nome"`
	ProdutoSKU           *string   `json:"produto_sku,omitempty"`
	QuantidadeVendida    float64   `json:"quantidade_vendida"`
	CustoUnitarioTeorico float64   `json:"custo_unitario_teorico"`
	CMVTotal             float64   `json:"cmv_total"`
	ReceitaTotal         float64   `json:"receita_total"`
	MargemBruta          float64   `json:"margem_bruta"`
	MargemBrutaPercent   float64   `json:"margem_bruta_percent"`
	TemFichaTecnica      bool      `json:"tem_ficha_tecnica"`
}

type CMVFilter struct {
	EmpresaID  uuid.UUID
	LojaID     *uuid.UUID
	DataInicio time.Time
	DataFim    time.Time
	Limit      int
	Offset     int
}
