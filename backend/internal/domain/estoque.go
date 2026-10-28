package domain

import (
	"time"

	"github.com/google/uuid"
)

// TipoMovimentacaoEstoque define os tipos de movimentação de estoque suportados.
type TipoMovimentacaoEstoque string

const (
	MovimentacaoEntradaCompra        TipoMovimentacaoEstoque = "entrada_compra"
	MovimentacaoSaidaConsumo         TipoMovimentacaoEstoque = "saida_consumo"
	MovimentacaoAjustePositivo       TipoMovimentacaoEstoque = "ajuste_positivo"
	MovimentacaoAjusteNegativo       TipoMovimentacaoEstoque = "ajuste_negativo"
	MovimentacaoPerda                TipoMovimentacaoEstoque = "perda"
	MovimentacaoDesperdicio          TipoMovimentacaoEstoque = "desperdicio"
	MovimentacaoTransferenciaSaida   TipoMovimentacaoEstoque = "transferencia_saida"
	MovimentacaoTransferenciaEntrada TipoMovimentacaoEstoque = "transferencia_entrada"
)

// Origens conhecidas de uma movimentação de estoque.
const (
	OrigemManual            = "manual"
	OrigemRecebimentoCompra = "recebimento_compra"
	OrigemTransferencia     = "transferencia"
)

var movimentacoesValidas = map[TipoMovimentacaoEstoque]bool{
	MovimentacaoEntradaCompra:        true,
	MovimentacaoSaidaConsumo:         true,
	MovimentacaoAjustePositivo:       true,
	MovimentacaoAjusteNegativo:       true,
	MovimentacaoPerda:                true,
	MovimentacaoDesperdicio:          true,
	MovimentacaoTransferenciaSaida:   true,
	MovimentacaoTransferenciaEntrada: true,
}

// IsTipoMovimentacaoValido indica se o tipo informado é suportado.
func IsTipoMovimentacaoValido(tipo TipoMovimentacaoEstoque) bool {
	return movimentacoesValidas[tipo]
}

// SaldoEstoque representa o saldo atual de um insumo em uma loja.
type SaldoEstoque struct {
	ID                 uuid.UUID `json:"id"`
	EmpresaID          uuid.UUID `json:"empresa_id"`
	LojaID             uuid.UUID `json:"loja_id"`
	InsumoID           uuid.UUID `json:"insumo_id"`
	Quantidade         float64   `json:"quantidade"`
	UnidadeMedida      string    `json:"unidade_medida"`
	CustoMedioUnitario float64   `json:"custo_medio_unitario"`
	PermiteNegativo    bool      `json:"permite_negativo"`
	CreatedAt          time.Time `json:"created_at"`
	AtualizadoEm       time.Time `json:"atualizado_em"`
}

// MovimentacaoEstoque representa o registro imutável de uma alteração de saldo.
type MovimentacaoEstoque struct {
	ID                uuid.UUID               `json:"id"`
	EmpresaID         uuid.UUID               `json:"empresa_id"`
	LojaID            uuid.UUID               `json:"loja_id"`
	InsumoID          uuid.UUID               `json:"insumo_id"`
	Tipo              TipoMovimentacaoEstoque `json:"tipo"`
	Quantidade        float64                 `json:"quantidade"`
	UnidadeMedida     string                  `json:"unidade_medida"`
	CustoUnitario     float64                 `json:"custo_unitario"`
	CustoTotal        float64                 `json:"custo_total"`
	SaldoAnterior     float64                 `json:"saldo_anterior"`
	SaldoPosterior    float64                 `json:"saldo_posterior"`
	Motivo            *string                 `json:"motivo,omitempty"`
	Observacao        *string                 `json:"observacao,omitempty"`
	Origem            string                  `json:"origem"`
	ReferenciaExterna *string                 `json:"referencia_externa,omitempty"`
	UsuarioID         *uuid.UUID              `json:"usuario_id,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
}

// RegistrarSaidaEstoqueInput registra saída por consumo, perda, desperdício ou ajuste negativo.
type RegistrarSaidaEstoqueInput struct {
	EmpresaID         uuid.UUID `json:"-"`
	LojaID            uuid.UUID `json:"loja_id" validate:"required"`
	InsumoID          uuid.UUID `json:"insumo_id" validate:"required"`
	Tipo              string    `json:"tipo" validate:"required"`
	Quantidade        float64   `json:"quantidade" validate:"required,gt=0"`
	Motivo            *string   `json:"motivo,omitempty"`
	Observacao        *string   `json:"observacao,omitempty"`
	ReferenciaExterna *string   `json:"referencia_externa,omitempty"`
}

// RegistrarAjusteEstoqueInput registra ajuste positivo ou negativo (motivo obrigatório).
type RegistrarAjusteEstoqueInput struct {
	EmpresaID  uuid.UUID `json:"-"`
	LojaID     uuid.UUID `json:"loja_id" validate:"required"`
	InsumoID   uuid.UUID `json:"insumo_id" validate:"required"`
	Tipo       string    `json:"tipo" validate:"required"`
	Quantidade float64   `json:"quantidade" validate:"required,gt=0"`
	Motivo     string    `json:"motivo" validate:"required"`
	Observacao *string   `json:"observacao,omitempty"`
}

// TransferirEstoqueInput transfere insumo entre lojas da mesma empresa.
type TransferirEstoqueInput struct {
	EmpresaID     uuid.UUID `json:"-"`
	LojaOrigemID  uuid.UUID `json:"loja_origem_id" validate:"required"`
	LojaDestinoID uuid.UUID `json:"loja_destino_id" validate:"required"`
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	Motivo        *string   `json:"motivo,omitempty"`
	Observacao    *string   `json:"observacao,omitempty"`
}

// RegistrarEntradaCompraInput cria entrada de estoque a partir de recebimento de compra.
type RegistrarEntradaCompraInput struct {
	EmpresaID         uuid.UUID
	LojaID            uuid.UUID
	InsumoID          uuid.UUID
	Quantidade        float64
	CustoUnitario     float64
	UnidadeMedida     string
	ReferenciaExterna *string
	Observacao        *string
	UsuarioID         *uuid.UUID
}

// SaldoEstoqueFilter filtra a listagem de saldos.
type SaldoEstoqueFilter struct {
	EmpresaID uuid.UUID
	LojaID    *uuid.UUID
	InsumoID  *uuid.UUID
	Limit     int
	Offset    int
}

// MovimentacaoEstoqueFilter filtra a listagem de movimentações.
type MovimentacaoEstoqueFilter struct {
	EmpresaID  uuid.UUID
	LojaID     *uuid.UUID
	InsumoID   *uuid.UUID
	Tipo       *TipoMovimentacaoEstoque
	Origem     *string
	DataInicio *time.Time
	DataFim    *time.Time
	Limit      int
	Offset     int
}

// ResumoPerdasFilter filtra o resumo de perdas e desperdícios.
type ResumoPerdasFilter struct {
	EmpresaID  uuid.UUID
	LojaID     uuid.UUID
	DataInicio time.Time
	DataFim    time.Time
}

// ResumoPerdas consolida perdas e desperdícios por período.
type ResumoPerdas struct {
	EmpresaID             uuid.UUID          `json:"empresa_id"`
	LojaID                uuid.UUID          `json:"loja_id"`
	DataInicio            time.Time          `json:"data_inicio"`
	DataFim               time.Time          `json:"data_fim"`
	QuantidadeTotal       float64            `json:"quantidade_total"`
	CustoTotalPerdas      float64            `json:"custo_total_perdas"`
	CustoTotalDesperdicio float64            `json:"custo_total_desperdicio"`
	CustoTotal            float64            `json:"custo_total"`
	Itens                 []ResumoPerdasItem `json:"itens"`
}

// ResumoPerdasItem detalha perdas/desperdícios de um insumo.
type ResumoPerdasItem struct {
	Tipo          TipoMovimentacaoEstoque `json:"tipo"`
	InsumoID      uuid.UUID               `json:"insumo_id"`
	InsumoNome    string                  `json:"insumo_nome"`
	UnidadeMedida string                  `json:"unidade_medida"`
	Quantidade    float64                 `json:"quantidade"`
	CustoTotal    float64                 `json:"custo_total"`
}
