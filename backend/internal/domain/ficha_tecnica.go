package domain

import (
	"time"

	"github.com/google/uuid"
)

type FichaTecnica struct {
	ID                uuid.UUID  `json:"id"`
	ProdutoID         uuid.UUID  `json:"produto_id"`
	LojaID            uuid.UUID  `json:"loja_id"`
	Rendimento        float64    `json:"rendimento"`
	UnidadeRendimento string     `json:"unidade_rendimento"`
	Ativo             bool       `json:"ativo"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	// Computed fields
	CustoTeorico float64                   `json:"custo_teorico,omitempty"`
	Componentes  []*FichaTecnicaComponente `json:"componentes,omitempty"`
}

type FichaTecnicaComponente struct {
	ID             uuid.UUID `json:"id"`
	FichaTecnicaID uuid.UUID `json:"ficha_tecnica_id"`
	InsumoID       uuid.UUID `json:"insumo_id"`
	Quantidade     float64   `json:"quantidade"`
	UnidadeMedida  string    `json:"unidade_medida"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Joined fields
	InsumoNome    string  `json:"insumo_nome,omitempty"`
	InsumoUnidade string  `json:"insumo_unidade,omitempty"`
	CustoUnitario float64 `json:"custo_unitario,omitempty"`
	CustoTotal    float64 `json:"custo_total,omitempty"`
}

type CreateFichaTecnicaInput struct {
	ProdutoID         uuid.UUID                           `json:"produto_id" validate:"required"`
	LojaID            uuid.UUID                           `json:"loja_id" validate:"required"`
	Rendimento        float64                             `json:"rendimento"`
	UnidadeRendimento string                              `json:"unidade_rendimento"`
	Componentes       []CreateFichaTecnicaComponenteInput `json:"componentes" validate:"required,min=1"`
}

type CreateFichaTecnicaComponenteInput struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string    `json:"unidade_medida" validate:"required"`
}

type UpdateFichaTecnicaInput struct {
	Rendimento        *float64                            `json:"rendimento,omitempty"`
	UnidadeRendimento *string                             `json:"unidade_rendimento,omitempty"`
	Ativo             *bool                               `json:"ativo,omitempty"`
	Componentes       []UpdateFichaTecnicaComponenteInput `json:"componentes,omitempty"`
}

type UpdateFichaTecnicaComponenteInput struct {
	ID            *uuid.UUID `json:"id,omitempty"` // nil for new components
	InsumoID      uuid.UUID  `json:"insumo_id" validate:"required"`
	Quantidade    float64    `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string     `json:"unidade_medida" validate:"required"`
}

type FichaTecnicaFilter struct {
	EmpresaID *uuid.UUID
	ProdutoID *uuid.UUID
	LojaID    *uuid.UUID
	Ativo     *bool
	Limit     int
	Offset    int
}
