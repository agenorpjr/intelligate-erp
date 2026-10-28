package domain

import (
	"time"

	"github.com/google/uuid"
)

// Fornecedor representa um fornecedor da empresa
type Fornecedor struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	Nome        string     `json:"nome"`
	CNPJCPF     string     `json:"cnpj_cpf"`
	Email       *string    `json:"email,omitempty"`
	Telefone    *string    `json:"telefone,omitempty"`
	Banco       *string    `json:"banco,omitempty"`
	Agencia     *string    `json:"agencia,omitempty"`
	Conta       *string    `json:"conta,omitempty"`
	TipoConta   *string    `json:"tipo_conta,omitempty"` // 'corrente', 'poupanca'
	PixChave    *string    `json:"pix_chave,omitempty"`
	Observacoes *string    `json:"observacoes,omitempty"`
	Ativo       bool       `json:"ativo"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateFornecedorInput struct {
	EmpresaID   uuid.UUID `json:"empresa_id" validate:"required"`
	Nome        string    `json:"nome" validate:"required"`
	CNPJCPF     string    `json:"cnpj_cpf" validate:"required"`
	Email       *string   `json:"email,omitempty"`
	Telefone    *string   `json:"telefone,omitempty"`
	Banco       *string   `json:"banco,omitempty"`
	Agencia     *string   `json:"agencia,omitempty"`
	Conta       *string   `json:"conta,omitempty"`
	TipoConta   *string   `json:"tipo_conta,omitempty"`
	PixChave    *string   `json:"pix_chave,omitempty"`
	Observacoes *string   `json:"observacoes,omitempty"`
}

type UpdateFornecedorInput struct {
	Nome        *string `json:"nome,omitempty"`
	CNPJCPF     *string `json:"cnpj_cpf,omitempty"`
	Email       *string `json:"email,omitempty"`
	Telefone    *string `json:"telefone,omitempty"`
	Banco       *string `json:"banco,omitempty"`
	Agencia     *string `json:"agencia,omitempty"`
	Conta       *string `json:"conta,omitempty"`
	TipoConta   *string `json:"tipo_conta,omitempty"`
	PixChave    *string `json:"pix_chave,omitempty"`
	Observacoes *string `json:"observacoes,omitempty"`
	Ativo       *bool   `json:"ativo,omitempty"`
}

type FornecedorFilter struct {
	EmpresaID *uuid.UUID
	Ativo     *bool
	Limit     int
	Offset    int
}

// OrdemCompraStatus define os status possíveis de uma ordem de compra
type OrdemCompraStatus string

const (
	OrdemCompraStatusRascunho        OrdemCompraStatus = "rascunho"
	OrdemCompraStatusEnviada         OrdemCompraStatus = "enviada"
	OrdemCompraStatusRecebidaParcial OrdemCompraStatus = "recebida_parcial"
	OrdemCompraStatusRecebida        OrdemCompraStatus = "recebida"
	OrdemCompraStatusCancelada       OrdemCompraStatus = "cancelada"
)

// OrdemCompra representa uma ordem de compra
type OrdemCompra struct {
	ID                  uuid.UUID          `json:"id"`
	EmpresaID           uuid.UUID          `json:"empresa_id"`
	LojaID              uuid.UUID          `json:"loja_id"`
	FornecedorID        uuid.UUID          `json:"fornecedor_id"`
	DataCompra          time.Time          `json:"data_compra"`
	DataPrevistaEntrega *time.Time         `json:"data_prevista_entrega,omitempty"`
	Status              OrdemCompraStatus  `json:"status"`
	Observacoes         *string            `json:"observacoes,omitempty"`
	ValorTotal          float64            `json:"valor_total"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	DeletedAt           *time.Time         `json:"deleted_at,omitempty"`
	Itens               []*OrdemCompraItem `json:"itens,omitempty"`
	Fornecedor          *Fornecedor        `json:"fornecedor,omitempty"`
	Loja                *Loja              `json:"loja,omitempty"`
}

type OrdemCompraItem struct {
	ID                 uuid.UUID `json:"id"`
	OrdemCompraID      uuid.UUID `json:"ordem_compra_id"`
	InsumoID           uuid.UUID `json:"insumo_id"`
	Quantidade         float64   `json:"quantidade"`
	UnidadeMedida      string    `json:"unidade_medida"`
	PrecoUnitario      float64   `json:"preco_unitario"`
	QuantidadeRecebida float64   `json:"quantidade_recebida"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Insumo             *Insumo   `json:"insumo,omitempty"`
}

type CreateOrdemCompraInput struct {
	EmpresaID           uuid.UUID                    `json:"empresa_id" validate:"required"`
	LojaID              uuid.UUID                    `json:"loja_id" validate:"required"`
	FornecedorID        uuid.UUID                    `json:"fornecedor_id" validate:"required"`
	DataCompra          time.Time                    `json:"data_compra" validate:"required"`
	DataPrevistaEntrega *time.Time                   `json:"data_prevista_entrega,omitempty"`
	Observacoes         *string                      `json:"observacoes,omitempty"`
	Itens               []CreateOrdemCompraItemInput `json:"itens" validate:"required,min=1"`
}

type CreateOrdemCompraItemInput struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string    `json:"unidade_medida" validate:"required"`
	PrecoUnitario float64   `json:"preco_unitario" validate:"required,gte=0"`
}

type UpdateOrdemCompraInput struct {
	LojaID              *uuid.UUID `json:"loja_id,omitempty"`
	FornecedorID        *uuid.UUID `json:"fornecedor_id,omitempty"`
	DataCompra          *time.Time `json:"data_compra,omitempty"`
	DataPrevistaEntrega *time.Time `json:"data_prevista_entrega,omitempty"`
	Status              *string    `json:"status,omitempty"`
	Observacoes         *string    `json:"observacoes,omitempty"`
}

type OrdemCompraFilter struct {
	EmpresaID    *uuid.UUID
	LojaID       *uuid.UUID
	FornecedorID *uuid.UUID
	Status       *OrdemCompraStatus
	DataInicio   *time.Time
	DataFim      *time.Time
	Limit        int
	Offset       int
}

// RecebimentoCompra representa o recebimento de uma ordem de compra
type RecebimentoCompra struct {
	ID                uuid.UUID                `json:"id"`
	OrdemCompraID     uuid.UUID                `json:"ordem_compra_id"`
	LojaID            uuid.UUID                `json:"loja_id"`
	DataRecebimento   time.Time                `json:"data_recebimento"`
	Observacoes       *string                  `json:"observacoes,omitempty"`
	PermitirExcedente bool                     `json:"permitir_excedente"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	Itens             []*RecebimentoCompraItem `json:"itens,omitempty"`
}

type RecebimentoCompraItem struct {
	ID                  uuid.UUID `json:"id"`
	RecebimentoCompraID uuid.UUID `json:"recebimento_compra_id"`
	OrdemCompraItemID   uuid.UUID `json:"ordem_compra_item_id"`
	InsumoID            uuid.UUID `json:"insumo_id"`
	QuantidadeRecebida  float64   `json:"quantidade_recebida"`
	PrecoUnitario       float64   `json:"preco_unitario"`
	CreatedAt           time.Time `json:"created_at"`
}

type CreateRecebimentoCompraInput struct {
	OrdemCompraID     uuid.UUID                          `json:"ordem_compra_id" validate:"required"`
	LojaID            uuid.UUID                          `json:"loja_id" validate:"required"`
	DataRecebimento   *time.Time                         `json:"data_recebimento,omitempty"`
	Observacoes       *string                            `json:"observacoes,omitempty"`
	PermitirExcedente bool                               `json:"permitir_excedente"`
	Itens             []CreateRecebimentoCompraItemInput `json:"itens" validate:"required,min=1"`
}

type CreateRecebimentoCompraItemInput struct {
	OrdemCompraItemID  uuid.UUID `json:"ordem_compra_item_id" validate:"required"`
	InsumoID           uuid.UUID `json:"insumo_id" validate:"required"`
	QuantidadeRecebida float64   `json:"quantidade_recebida" validate:"required,gt=0"`
	PrecoUnitario      float64   `json:"preco_unitario" validate:"required,gte=0"`
}

// ContaPagarStatus define os status possíveis de uma conta a pagar
type ContaPagarStatus string

const (
	ContaPagarStatusPendente  ContaPagarStatus = "pendente"
	ContaPagarStatusPago      ContaPagarStatus = "pago"
	ContaPagarStatusCancelado ContaPagarStatus = "cancelado"
	ContaPagarStatusVencido   ContaPagarStatus = "vencido"
)

// ContaPagar representa uma conta a pagar
type ContaPagar struct {
	ID                    uuid.UUID            `json:"id"`
	EmpresaID             uuid.UUID            `json:"empresa_id"`
	LojaID                uuid.UUID            `json:"loja_id"`
	FornecedorID          uuid.UUID            `json:"fornecedor_id"`
	OrdemCompraID         *uuid.UUID           `json:"ordem_compra_id,omitempty"`
	RecebimentoCompraID   *uuid.UUID           `json:"recebimento_compra_id,omitempty"`
	Descricao             string               `json:"descricao"`
	Valor                 float64              `json:"valor"`
	DataEmissao           time.Time            `json:"data_emissao"`
	DataVencimento        time.Time            `json:"data_vencimento"`
	DataPagamento         *time.Time           `json:"data_pagamento,omitempty"`
	Status                ContaPagarStatus     `json:"status"`
	FormaPagamento        *string              `json:"forma_pagamento,omitempty"`
	Observacoes           *string              `json:"observacoes,omitempty"`
	CategoriaFinanceiraID *uuid.UUID           `json:"categoria_financeira_id,omitempty"`
	CategoriaFinanceira   *CategoriaFinanceira `json:"categoria_financeira,omitempty"`
	ContaGerencialID      *uuid.UUID           `json:"conta_gerencial_id,omitempty"`
	ContaGerencial        *ContaGerencial      `json:"conta_gerencial,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	DeletedAt             *time.Time           `json:"deleted_at,omitempty"`
	Fornecedor            *Fornecedor          `json:"fornecedor,omitempty"`
	Loja                  *Loja                `json:"loja,omitempty"`
}

type CreateContaPagarInput struct {
	EmpresaID             uuid.UUID  `json:"empresa_id" validate:"required"`
	LojaID                uuid.UUID  `json:"loja_id" validate:"required"`
	FornecedorID          uuid.UUID  `json:"fornecedor_id" validate:"required"`
	OrdemCompraID         *uuid.UUID `json:"ordem_compra_id,omitempty"`
	RecebimentoCompraID   *uuid.UUID `json:"recebimento_compra_id,omitempty"`
	Descricao             string     `json:"descricao" validate:"required"`
	Valor                 float64    `json:"valor" validate:"required,gt=0"`
	DataEmissao           time.Time  `json:"data_emissao" validate:"required"`
	DataVencimento        time.Time  `json:"data_vencimento" validate:"required"`
	Status                *string    `json:"status,omitempty"`
	FormaPagamento        *string    `json:"forma_pagamento,omitempty"`
	Observacoes           *string    `json:"observacoes,omitempty"`
	CategoriaFinanceiraID *uuid.UUID `json:"categoria_financeira_id,omitempty"`
	ContaGerencialID      *uuid.UUID `json:"conta_gerencial_id,omitempty"`
}

type UpdateContaPagarInput struct {
	Descricao             *string    `json:"descricao,omitempty"`
	Valor                 *float64   `json:"valor,omitempty"`
	DataEmissao           *time.Time `json:"data_emissao,omitempty"`
	DataVencimento        *time.Time `json:"data_vencimento,omitempty"`
	DataPagamento         *time.Time `json:"data_pagamento,omitempty"`
	Status                *string    `json:"status,omitempty"`
	FormaPagamento        *string    `json:"forma_pagamento,omitempty"`
	Observacoes           *string    `json:"observacoes,omitempty"`
	CategoriaFinanceiraID *uuid.UUID `json:"categoria_financeira_id,omitempty"`
	ContaGerencialID      *uuid.UUID `json:"conta_gerencial_id,omitempty"`
}

type PagarContaInput struct {
	DataPagamento  time.Time `json:"data_pagamento" validate:"required"`
	FormaPagamento string    `json:"forma_pagamento" validate:"required"`
}

type ContaPagarFilter struct {
	EmpresaID             *uuid.UUID
	LojaID                *uuid.UUID
	FornecedorID          *uuid.UUID
	Status                *ContaPagarStatus
	DataVencimentoInicio  *time.Time
	DataVencimentoFim     *time.Time
	CategoriaFinanceiraID *uuid.UUID
	ContaGerencialID      *uuid.UUID
	Limit                 int
	Offset                int
}
