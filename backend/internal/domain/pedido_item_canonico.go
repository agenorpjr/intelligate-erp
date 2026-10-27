package domain

import (
	"time"

	"github.com/google/uuid"
)

type PedidoItemCanonico struct {
	ID                uuid.UUID  `json:"id"`
	PedidoID          uuid.UUID  `json:"pedido_id"`
	ProdutoCanonicoID *uuid.UUID `json:"produto_canonico_id,omitempty"`
	SKUExterno        *string    `json:"sku_externo,omitempty"`
	Nome              string     `json:"nome"`
	Quantidade        float64    `json:"quantidade"`
	PrecoUnitario     float64    `json:"preco_unitario"`
	ValorTotal        float64    `json:"valor_total"`
	Desconto          float64    `json:"desconto"`
	Modificadores     *string    `json:"modificadores,omitempty"`
	Observacoes       *string    `json:"observacoes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

type CreatePedidoItemCanonicoInput struct {
	PedidoID          uuid.UUID  `json:"pedido_id" validate:"required"`
	ProdutoCanonicoID *uuid.UUID `json:"produto_canonico_id,omitempty"`
	SKUExterno        *string    `json:"sku_externo,omitempty"`
	Nome              string     `json:"nome" validate:"required"`
	Quantidade        float64    `json:"quantidade" validate:"required,gt=0"`
	PrecoUnitario     float64    `json:"preco_unitario" validate:"required,gte=0"`
	ValorTotal        float64    `json:"valor_total" validate:"required,gte=0"`
	Desconto          float64    `json:"desconto"`
	Modificadores     *string    `json:"modificadores,omitempty"`
	Observacoes       *string    `json:"observacoes,omitempty"`
}

type UpdatePedidoItemCanonicoInput struct {
	ProdutoCanonicoID *uuid.UUID `json:"produto_canonico_id,omitempty"`
	SKUExterno        *string    `json:"sku_externo,omitempty"`
	Nome              *string    `json:"nome,omitempty"`
	Quantidade        *float64   `json:"quantidade,omitempty"`
	PrecoUnitario     *float64   `json:"preco_unitario,omitempty"`
	ValorTotal        *float64   `json:"valor_total,omitempty"`
	Desconto          *float64   `json:"desconto,omitempty"`
	Modificadores     *string    `json:"modificadores,omitempty"`
	Observacoes       *string    `json:"observacoes,omitempty"`
}
