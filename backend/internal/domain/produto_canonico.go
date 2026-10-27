package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProdutoCanonico struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	CategoriaID *uuid.UUID `json:"categoria_id,omitempty"`
	SKU         string     `json:"sku"`
	Nome        string     `json:"nome"`
	Descricao   *string    `json:"descricao,omitempty"`
	Unidade     string     `json:"unidade"`
	PrecoBase   float64    `json:"preco_base"`
	NCM         *string    `json:"ncm,omitempty"`
	CEST        *string    `json:"cest,omitempty"`
	Tributacao  *string    `json:"tributacao,omitempty"`
	Ativo       bool       `json:"ativo"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateProdutoCanonicoInput struct {
	EmpresaID   uuid.UUID  `json:"empresa_id" validate:"required"`
	CategoriaID *uuid.UUID `json:"categoria_id,omitempty"`
	SKU         string     `json:"sku" validate:"required"`
	Nome        string     `json:"nome" validate:"required"`
	Descricao   *string    `json:"descricao,omitempty"`
	Unidade     string     `json:"unidade" validate:"required"`
	PrecoBase   float64    `json:"preco_base"`
	NCM         *string    `json:"ncm,omitempty"`
	CEST        *string    `json:"cest,omitempty"`
	Tributacao  *string    `json:"tributacao,omitempty"`
}

type UpdateProdutoCanonicoInput struct {
	CategoriaID *uuid.UUID `json:"categoria_id,omitempty"`
	SKU         *string    `json:"sku,omitempty"`
	Nome        *string    `json:"nome,omitempty"`
	Descricao   *string    `json:"descricao,omitempty"`
	Unidade     *string    `json:"unidade,omitempty"`
	PrecoBase   *float64   `json:"preco_base,omitempty"`
	NCM         *string    `json:"ncm,omitempty"`
	CEST        *string    `json:"cest,omitempty"`
	Tributacao  *string    `json:"tributacao,omitempty"`
	Ativo       *bool      `json:"ativo,omitempty"`
}
