package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Plano de Contas Gerencial
// ============================================================

// ContaGerencialTipo define os tipos de contas gerenciais
type ContaGerencialTipo string

const (
	ContaGerencialTipoReceita      ContaGerencialTipo = "receita"
	ContaGerencialTipoCusto        ContaGerencialTipo = "custo"
	ContaGerencialTipoDespesa      ContaGerencialTipo = "despesa"
	ContaGerencialTipoInvestimento ContaGerencialTipo = "investimento"
)

// ContaGerencial representa uma conta do plano de contas gerencial
type ContaGerencial struct {
	ID         uuid.UUID          `json:"id"`
	EmpresaID  uuid.UUID          `json:"empresa_id"`
	Codigo     string             `json:"codigo"`
	Nome       string             `json:"nome"`
	Tipo       ContaGerencialTipo `json:"tipo"`
	Grupo      *string            `json:"grupo,omitempty"`
	ContaPaiID *uuid.UUID         `json:"conta_pai_id,omitempty"`
	ContaPai   *ContaGerencial    `json:"conta_pai,omitempty"`
	Filhas     []*ContaGerencial  `json:"filhas,omitempty"`
	Ativo      bool               `json:"ativo"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	DeletedAt  *time.Time         `json:"deleted_at,omitempty"`
}

type CreateContaGerencialInput struct {
	EmpresaID  uuid.UUID          `json:"empresa_id" validate:"required"`
	Codigo     string             `json:"codigo" validate:"required"`
	Nome       string             `json:"nome" validate:"required"`
	Tipo       ContaGerencialTipo `json:"tipo" validate:"required,oneof=receita custo despesa investimento"`
	Grupo      *string            `json:"grupo,omitempty"`
	ContaPaiID *uuid.UUID         `json:"conta_pai_id,omitempty"`
}

type UpdateContaGerencialInput struct {
	Codigo     *string             `json:"codigo,omitempty"`
	Nome       *string             `json:"nome,omitempty"`
	Tipo       *ContaGerencialTipo `json:"tipo,omitempty"`
	Grupo      *string             `json:"grupo,omitempty"`
	ContaPaiID *uuid.UUID          `json:"conta_pai_id,omitempty"`
	Ativo      *bool               `json:"ativo,omitempty"`
}

type ContaGerencialFilter struct {
	EmpresaID *uuid.UUID
	Tipo      *ContaGerencialTipo
	Grupo     *string
	Ativo     *bool
	Limit     int
	Offset    int
}

// ============================================================
// Categorias Financeiras
// ============================================================

// CategoriaFinanceira representa uma categoria financeira para classificação
type CategoriaFinanceira struct {
	ID               uuid.UUID       `json:"id"`
	EmpresaID        uuid.UUID       `json:"empresa_id"`
	Nome             string          `json:"nome"`
	Descricao        *string         `json:"descricao,omitempty"`
	ContaGerencialID *uuid.UUID      `json:"conta_gerencial_id,omitempty"`
	ContaGerencial   *ContaGerencial `json:"conta_gerencial,omitempty"`
	Ativo            bool            `json:"ativo"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"`
}

type CreateCategoriaFinanceiraInput struct {
	EmpresaID        uuid.UUID  `json:"empresa_id" validate:"required"`
	Nome             string     `json:"nome" validate:"required"`
	Descricao        *string    `json:"descricao,omitempty"`
	ContaGerencialID *uuid.UUID `json:"conta_gerencial_id,omitempty"`
}

type UpdateCategoriaFinanceiraInput struct {
	Nome             *string    `json:"nome,omitempty"`
	Descricao        *string    `json:"descricao,omitempty"`
	ContaGerencialID *uuid.UUID `json:"conta_gerencial_id,omitempty"`
	Ativo            *bool      `json:"ativo,omitempty"`
}

type CategoriaFinanceiraFilter struct {
	EmpresaID        *uuid.UUID
	ContaGerencialID *uuid.UUID
	Ativo            *bool
	Limit            int
	Offset           int
}

// ============================================================
// Movimentações de Caixa
// ============================================================

// MovimentacaoCaixaTipo define os tipos de movimentação de caixa
type MovimentacaoCaixaTipo string

const (
	MovimentacaoCaixaTipoEntrada MovimentacaoCaixaTipo = "entrada"
	MovimentacaoCaixaTipoSaida   MovimentacaoCaixaTipo = "saida"
)

// MovimentacaoCaixaOrigem define as origens possíveis de uma movimentação
type MovimentacaoCaixaOrigem string

const (
	MovimentacaoCaixaOrigemVenda      MovimentacaoCaixaOrigem = "venda"
	MovimentacaoCaixaOrigemContaPagar MovimentacaoCaixaOrigem = "conta_pagar"
	MovimentacaoCaixaOrigemManual     MovimentacaoCaixaOrigem = "manual"
	MovimentacaoCaixaOrigemOutros     MovimentacaoCaixaOrigem = "outros"
)

// MovimentacaoCaixa representa uma movimentação de caixa
type MovimentacaoCaixa struct {
	ID                    uuid.UUID             `json:"id"`
	EmpresaID             uuid.UUID             `json:"empresa_id"`
	LojaID                uuid.UUID             `json:"loja_id"`
	Data                  time.Time             `json:"data"`
	Tipo                  MovimentacaoCaixaTipo `json:"tipo"`
	Valor                 float64               `json:"valor"`
	Descricao             string                `json:"descricao"`
	CategoriaFinanceiraID *uuid.UUID            `json:"categoria_financeira_id,omitempty"`
	CategoriaFinanceira   *CategoriaFinanceira  `json:"categoria_financeira,omitempty"`
	ContaGerencialID      *uuid.UUID            `json:"conta_gerencial_id,omitempty"`
	ContaGerencial        *ContaGerencial       `json:"conta_gerencial,omitempty"`
	Origem                *string               `json:"origem,omitempty"`
	OrigemID              *uuid.UUID            `json:"origem_id,omitempty"`
	Observacoes           *string               `json:"observacoes,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	DeletedAt             *time.Time            `json:"deleted_at,omitempty"`
}

type CreateMovimentacaoCaixaInput struct {
	EmpresaID             uuid.UUID             `json:"empresa_id" validate:"required"`
	LojaID                uuid.UUID             `json:"loja_id" validate:"required"`
	Data                  time.Time             `json:"data" validate:"required"`
	Tipo                  MovimentacaoCaixaTipo `json:"tipo" validate:"required,oneof=entrada saida"`
	Valor                 float64               `json:"valor" validate:"required,gt=0"`
	Descricao             string                `json:"descricao" validate:"required"`
	CategoriaFinanceiraID *uuid.UUID            `json:"categoria_financeira_id,omitempty"`
	ContaGerencialID      *uuid.UUID            `json:"conta_gerencial_id,omitempty"`
	Origem                *string               `json:"origem,omitempty"`
	OrigemID              *uuid.UUID            `json:"origem_id,omitempty"`
	Observacoes           *string               `json:"observacoes,omitempty"`
}

type UpdateMovimentacaoCaixaInput struct {
	Data                  *time.Time             `json:"data,omitempty"`
	Tipo                  *MovimentacaoCaixaTipo `json:"tipo,omitempty"`
	Valor                 *float64               `json:"valor,omitempty"`
	Descricao             *string                `json:"descricao,omitempty"`
	CategoriaFinanceiraID *uuid.UUID             `json:"categoria_financeira_id,omitempty"`
	ContaGerencialID      *uuid.UUID             `json:"conta_gerencial_id,omitempty"`
	Origem                *string                `json:"origem,omitempty"`
	OrigemID              *uuid.UUID             `json:"origem_id,omitempty"`
	Observacoes           *string                `json:"observacoes,omitempty"`
}

type MovimentacaoCaixaFilter struct {
	EmpresaID             *uuid.UUID
	LojaID                *uuid.UUID
	Tipo                  *MovimentacaoCaixaTipo
	CategoriaFinanceiraID *uuid.UUID
	ContaGerencialID      *uuid.UUID
	Origem                *string
	DataInicio            *time.Time
	DataFim               *time.Time
	Limit                 int
	Offset                int
}

// ============================================================
// DRE Gerencial
// ============================================================

// DREGerencial representa o Demonstrativo de Resultado do Exercício gerencial
type DREGerencial struct {
	EmpresaID            uuid.UUID  `json:"empresa_id"`
	LojaID               *uuid.UUID `json:"loja_id,omitempty"`
	DataInicio           time.Time  `json:"data_inicio"`
	DataFim              time.Time  `json:"data_fim"`
	Receitas             []DRELinha `json:"receitas"`
	Custos               []DRELinha `json:"custos"`
	Despesas             []DRELinha `json:"despesas"`
	Investimentos        []DRELinha `json:"investimentos"`
	TotalReceitas        float64    `json:"total_receitas"`
	TotalCustos          float64    `json:"total_custos"`
	TotalDespesas        float64    `json:"total_despesas"`
	TotalInvestimentos   float64    `json:"total_investimentos"`
	ResultadoBruto       float64    `json:"resultado_bruto"`       // Receitas - Custos
	ResultadoOperacional float64    `json:"resultado_operacional"` // Resultado Bruto - Despesas
	ResultadoLiquido     float64    `json:"resultado_liquido"`     // Resultado Operacional - Investimentos
}

// DRELinha representa uma linha do DRE
type DRELinha struct {
	ContaGerencialID uuid.UUID          `json:"conta_gerencial_id"`
	Codigo           string             `json:"codigo"`
	Nome             string             `json:"nome"`
	Tipo             ContaGerencialTipo `json:"tipo"`
	Valor            float64            `json:"valor"`
	Filhas           []DRELinha         `json:"filhas,omitempty"`
}

// ============================================================
// Fluxo de Caixa
// ============================================================

// FluxoCaixa representa o fluxo de caixa por período
type FluxoCaixa struct {
	EmpresaID     uuid.UUID         `json:"empresa_id"`
	LojaID        *uuid.UUID        `json:"loja_id,omitempty"`
	DataInicio    time.Time         `json:"data_inicio"`
	DataFim       time.Time         `json:"data_fim"`
	SaldoInicial  float64           `json:"saldo_inicial"`
	Entradas      []FluxoCaixaLinha `json:"entradas"`
	Saidas        []FluxoCaixaLinha `json:"saidas"`
	TotalEntradas float64           `json:"total_entradas"`
	TotalSaidas   float64           `json:"total_saidas"`
	SaldoFinal    float64           `json:"saldo_final"`
}

// FluxoCaixaLinha representa uma linha do fluxo de caixa
type FluxoCaixaLinha struct {
	ContaGerencialID uuid.UUID         `json:"conta_gerencial_id,omitempty"`
	Codigo           string            `json:"codigo,omitempty"`
	Nome             string            `json:"nome"`
	Valor            float64           `json:"valor"`
	Filhas           []FluxoCaixaLinha `json:"filhas,omitempty"`
}
