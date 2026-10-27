package domain

import (
	"time"

	"github.com/google/uuid"
)

type PedidoPagamentoCanonico struct {
	ID                uuid.UUID  `json:"id"`
	PedidoID          uuid.UUID  `json:"pedido_id"`
	FormaPagamento    string     `json:"forma_pagamento"`
	Valor             float64    `json:"valor"`
	Taxa              float64    `json:"taxa"`
	Troco             float64    `json:"troco"`
	ReferenciaExterna *string    `json:"referencia_externa,omitempty"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

type CreatePedidoPagamentoCanonicoInput struct {
	PedidoID          uuid.UUID `json:"pedido_id" validate:"required"`
	FormaPagamento    string    `json:"forma_pagamento" validate:"required"`
	Valor             float64   `json:"valor" validate:"required,gt=0"`
	Taxa              float64   `json:"taxa"`
	Troco             float64   `json:"troco"`
	ReferenciaExterna *string   `json:"referencia_externa,omitempty"`
	Status            string    `json:"status" validate:"required"`
}

type UpdatePedidoPagamentoCanonicoInput struct {
	FormaPagamento    *string  `json:"forma_pagamento,omitempty"`
	Valor             *float64 `json:"valor,omitempty"`
	Taxa              *float64 `json:"taxa,omitempty"`
	Troco             *float64 `json:"troco,omitempty"`
	ReferenciaExterna *string  `json:"referencia_externa,omitempty"`
	Status            *string  `json:"status,omitempty"`
}
