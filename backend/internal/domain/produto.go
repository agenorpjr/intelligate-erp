package domain

import (
	"time"

	"github.com/google/uuid"
)

type Produto struct {
	ID         uuid.UUID  `json:"id"`
	EmpresaID  uuid.UUID  `json:"empresa_id"`
	Nome       string     `json:"nome"`
	SKUExterno *string    `json:"sku_externo,omitempty"`
	Categoria  *string    `json:"categoria,omitempty"`
	Ativo      bool       `json:"ativo"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateProdutoInput struct {
	EmpresaID  uuid.UUID `json:"empresa_id" validate:"required"`
	Nome       string    `json:"nome" validate:"required"`
	SKUExterno *string   `json:"sku_externo,omitempty"`
	Categoria  *string   `json:"categoria,omitempty"`
}

type UpdateProdutoInput struct {
	Nome       *string `json:"nome,omitempty"`
	SKUExterno *string `json:"sku_externo,omitempty"`
	Categoria  *string `json:"categoria,omitempty"`
	Ativo      *bool   `json:"ativo,omitempty"`
}

type ProdutoFilter struct {
	EmpresaID *uuid.UUID
	Ativo     *bool
	Limit     int
	Offset    int
}
