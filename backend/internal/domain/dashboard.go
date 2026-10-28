package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Dashboard Operacional
// ============================================================

// DashboardRequest representa os parâmetros de filtro do dashboard
type DashboardRequest struct {
	EmpresaID  uuid.UUID `json:"empresa_id"`
	LojaID     uuid.UUID `json:"loja_id"`
	DataInicio time.Time `json:"data_inicio"`
	DataFim    time.Time `json:"data_fim"`

	// PrevDataInicio/PrevDataFim definem o período anterior de mesma duração,
	// usado para a comparação. Quando zero, o repositório calcula a partir do
	// período atual.
	PrevDataInicio time.Time `json:"-"`
	PrevDataFim    time.Time `json:"-"`

	// CmvProdutoLimit limita a quantidade de produtos retornados no ranking de
	// CMV. Quando <= 0, é aplicado o valor padrão.
	CmvProdutoLimit int `json:"-"`
}

// DashboardResponse representa a resposta consolidada do dashboard
type DashboardResponse struct {
	EmpresaID  uuid.UUID           `json:"empresa_id"`
	LojaID     uuid.UUID           `json:"loja_id"`
	DataInicio time.Time           `json:"data_inicio"`
	DataFim    time.Time           `json:"data_fim"`
	Vendas     DashboardVendas     `json:"vendas"`
	CMV        DashboardCMV        `json:"cmv"`
	Financeiro DashboardFinanceiro `json:"financeiro"`
	DRE        DashboardDRE        `json:"dre"`
}

// DashboardVendas contém indicadores de vendas
type DashboardVendas struct {
	TotalVendas        float64            `json:"total_vendas"`
	TotalPedidos       int64              `json:"total_pedidos"`
	TicketMedio        float64            `json:"ticket_medio"`
	TotalDescontos     float64            `json:"total_descontos"`
	VendasPorDia       []VendaPorDia      `json:"vendas_por_dia"`
	VendasPorStatus    []VendaPorStatus   `json:"vendas_por_status"`
	ComparacaoPeriodo  ComparacaoPeriodo  `json:"comparacao_periodo"`
	VariacaoPercentual VariacaoPercentual `json:"variacao_percentual"`
}

// VendaPorDia representa vendas agrupadas por dia
type VendaPorDia struct {
	Data         time.Time `json:"data"`
	TotalVendas  float64   `json:"total_vendas"`
	TotalPedidos int64     `json:"total_pedidos"`
	TicketMedio  float64   `json:"ticket_medio"`
}

// VendaPorStatus representa vendas agrupadas por status
type VendaPorStatus struct {
	Status       string  `json:"status"`
	TotalVendas  float64 `json:"total_vendas"`
	TotalPedidos int64   `json:"total_pedidos"`
}

// ComparacaoPeriodo compara o período atual com o anterior
type ComparacaoPeriodo struct {
	PeriodoAtual    PeriodoComparacao `json:"periodo_atual"`
	PeriodoAnterior PeriodoComparacao `json:"periodo_anterior"`
}

// PeriodoComparacao contém métricas de um período
type PeriodoComparacao struct {
	DataInicio     time.Time `json:"data_inicio"`
	DataFim        time.Time `json:"data_fim"`
	TotalVendas    float64   `json:"total_vendas"`
	TotalPedidos   int64     `json:"total_pedidos"`
	TicketMedio    float64   `json:"ticket_medio"`
	TotalDescontos float64   `json:"total_descontos"`
}

// VariacaoPercentual contém as variações percentuais
type VariacaoPercentual struct {
	Vendas      float64 `json:"vendas"`       // % variação total_vendas
	Pedidos     float64 `json:"pedidos"`      // % variação total_pedidos
	TicketMedio float64 `json:"ticket_medio"` // % variação ticket_medio
}

// DashboardCMV contém indicadores de CMV
type DashboardCMV struct {
	TotalCMV             float64            `json:"total_cmv"`
	CMVPorDia            []CMVPorDia        `json:"cmv_por_dia"`
	CMVPorProduto        []CMVProdutoResumo `json:"cmv_por_produto"`
	PercentualCMVReceita float64            `json:"percentual_cmv_receita"`
	ProdutosSemFicha     int                `json:"produtos_sem_ficha_tecnica"`
	ProdutosMaiorCMV     []CMVProdutoResumo `json:"produtos_maior_cmv"`
}

// CMVPorDia representa CMV agrupado por dia
type CMVPorDia struct {
	Data        time.Time `json:"data"`
	TotalCMV    float64   `json:"total_cmv"`
	TotalVendas float64   `json:"total_vendas"`
}

// CMVProdutoResumo representa resumo de CMV por produto
type CMVProdutoResumo struct {
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

// DashboardFinanceiro contém indicadores financeiros
type DashboardFinanceiro struct {
	TotalEntradas           float64               `json:"total_entradas"`
	TotalSaidas             float64               `json:"total_saidas"`
	SaldoFinal              float64               `json:"saldo_final"`
	DespesasPorCategoria    []DespesaPorCategoria `json:"despesas_por_categoria"`
	ContasPagarEmAberto     int64                 `json:"contas_pagar_em_aberto"`
	ContasPagarVencidas     int64                 `json:"contas_pagar_vencidas"`
	ValorContasPagarAberto  float64               `json:"valor_contas_pagar_aberto"`
	ValorContasPagarVencido float64               `json:"valor_contas_pagar_vencido"`
	FluxoCaixaPorDia        []FluxoCaixaPorDia    `json:"fluxo_caixa_por_dia"`
}

// DespesaPorCategoria representa despesas agrupadas por categoria financeira
type DespesaPorCategoria struct {
	CategoriaFinanceiraID uuid.UUID `json:"categoria_financeira_id"`
	CategoriaNome         string    `json:"categoria_nome"`
	Total                 float64   `json:"total"`
}

// FluxoCaixaPorDia representa fluxo de caixa agrupado por dia
type FluxoCaixaPorDia struct {
	Data           time.Time `json:"data"`
	Entradas       float64   `json:"entradas"`
	Saidas         float64   `json:"saidas"`
	SaldoDiario    float64   `json:"saldo_diario"`
	SaldoAcumulado float64   `json:"saldo_acumulado"`
}

// DashboardDRE contém indicadores do DRE
type DashboardDRE struct {
	ReceitaBruta         float64 `json:"receita_bruta"`
	Descontos            float64 `json:"descontos"`
	ReceitaLiquida       float64 `json:"receita_liquida"`
	CMV                  float64 `json:"cmv"`
	MargemBruta          float64 `json:"margem_bruta"`
	MargemBrutaPercent   float64 `json:"margem_bruta_percent"`
	DespesasOperacionais float64 `json:"despesas_operacionais"`
	ResultadoOperacional float64 `json:"resultado_operacional"`
	ResultadoLiquido     float64 `json:"resultado_liquido"`
}
