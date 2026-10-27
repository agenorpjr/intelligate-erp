package domain

import (
	"time"

	"github.com/google/uuid"
)

type Insumo struct {
	ID                 uuid.UUID  `json:"id"`
	EmpresaID          uuid.UUID  `json:"empresa_id"`
	Nome               string     `json:"nome"`
	UnidadeMedida      string     `json:"unidade_medida"`
	CustoUnitarioAtual float64    `json:"custo_unitario_atual"`
	Fornecedor         *string    `json:"fornecedor,omitempty"`
	Ativo              bool       `json:"ativo"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type InsumoCusto struct {
	ID            uuid.UUID `json:"id"`
	InsumoID      uuid.UUID `json:"insumo_id"`
	DataVigencia  time.Time `json:"data_vigencia"`
	CustoUnitario float64   `json:"custo_unitario"`
	Observacao    *string   `json:"observacao,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateInsumoInput struct {
	EmpresaID          uuid.UUID `json:"empresa_id" validate:"required"`
	Nome               string    `json:"nome" validate:"required"`
	UnidadeMedida      string    `json:"unidade_medida" validate:"required"`
	CustoUnitarioAtual float64   `json:"custo_unitario_atual"`
	Fornecedor         *string   `json:"fornecedor,omitempty"`
}

type UpdateInsumoInput struct {
	Nome               *string  `json:"nome,omitempty"`
	UnidadeMedida      *string  `json:"unidade_medida,omitempty"`
	CustoUnitarioAtual *float64 `json:"custo_unitario_atual,omitempty"`
	Fornecedor         *string  `json:"fornecedor,omitempty"`
	Ativo              *bool    `json:"ativo,omitempty"`
}

type CreateInsumoCustoInput struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	DataVigencia  time.Time `json:"data_vigencia" validate:"required"`
	CustoUnitario float64   `json:"custo_unitario" validate:"required"`
	Observacao    *string   `json:"observacao,omitempty"`
}

type InsumoFilter struct {
	EmpresaID *uuid.UUID
	Ativo     *bool
	Limit     int
	Offset    int
}
