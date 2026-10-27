package domain

import (
	"time"

	"github.com/google/uuid"
)

type PedidoCanonico struct {
	ID               uuid.UUID  `json:"id"`
	EmpresaID        uuid.UUID  `json:"empresa_id"`
	LojaID           uuid.UUID  `json:"loja_id"`
	PdvOrigemID      uuid.UUID  `json:"pdv_origem_id"`
	NumeroExterno    string     `json:"numero_externo"`
	DataHora         time.Time  `json:"data_hora"`
	Status           string     `json:"status"`
	ValorTotal       float64    `json:"valor_total"`
	ValorDesconto    float64    `json:"valor_desconto"`
	ValorAcrescimo   float64    `json:"valor_acrescimo"`
	IdempotencyKey   string     `json:"idempotency_key"`
	ClienteNome      *string    `json:"cliente_nome,omitempty"`
	ClienteDocumento *string    `json:"cliente_documento,omitempty"`
	ClienteEmail     *string    `json:"cliente_email,omitempty"`
	Observacoes      *string    `json:"observacoes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreatePedidoCanonicoInput struct {
	EmpresaID        uuid.UUID `json:"empresa_id" validate:"required"`
	LojaID           uuid.UUID `json:"loja_id" validate:"required"`
	PdvOrigemID      uuid.UUID `json:"pdv_origem_id" validate:"required"`
	NumeroExterno    string    `json:"numero_externo" validate:"required"`
	DataHora         time.Time `json:"data_hora" validate:"required"`
	Status           string    `json:"status" validate:"required"`
	ValorTotal       float64   `json:"valor_total" validate:"required"`
	ValorDesconto    float64   `json:"valor_desconto"`
	ValorAcrescimo   float64   `json:"valor_acrescimo"`
	IdempotencyKey   string    `json:"idempotency_key" validate:"required"`
	ClienteNome      *string   `json:"cliente_nome,omitempty"`
	ClienteDocumento *string   `json:"cliente_documento,omitempty"`
	ClienteEmail     *string   `json:"cliente_email,omitempty"`
	Observacoes      *string   `json:"observacoes,omitempty"`
}

type UpdatePedidoCanonicoInput struct {
	LojaID           *uuid.UUID `json:"loja_id,omitempty"`
	PdvOrigemID      *uuid.UUID `json:"pdv_origem_id,omitempty"`
	NumeroExterno    *string    `json:"numero_externo,omitempty"`
	DataHora         *time.Time `json:"data_hora,omitempty"`
	Status           *string    `json:"status,omitempty"`
	ValorTotal       *float64   `json:"valor_total,omitempty"`
	ValorDesconto    *float64   `json:"valor_desconto,omitempty"`
	ValorAcrescimo   *float64   `json:"valor_acrescimo,omitempty"`
	ClienteNome      *string    `json:"cliente_nome,omitempty"`
	ClienteDocumento *string    `json:"cliente_documento,omitempty"`
	ClienteEmail     *string    `json:"cliente_email,omitempty"`
	Observacoes      *string    `json:"observacoes,omitempty"`
}

type PedidoCanonicoFilter struct {
	EmpresaID  *uuid.UUID
	LojaID     *uuid.UUID
	DataInicio *time.Time
	DataFim    *time.Time
	Status     *string
	Limit      int
	Offset     int
}
