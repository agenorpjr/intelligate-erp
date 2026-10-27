package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventoIngestao struct {
	ID              uuid.UUID  `json:"id"`
	EmpresaID       uuid.UUID  `json:"empresa_id"`
	LojaID          uuid.UUID  `json:"loja_id"`
	PdvOrigemID     uuid.UUID  `json:"pdv_origem_id"`
	TipoEvento      string     `json:"tipo_evento"`
	PayloadOriginal string     `json:"payload_original"`
	Status          string     `json:"status"`
	Erro            *string    `json:"erro,omitempty"`
	Tentativas      int        `json:"tentativas"`
	IdempotencyKey  string     `json:"idempotency_key"`
	ProcessadoEm    *time.Time `json:"processado_em,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CreateEventoIngestaoInput struct {
	EmpresaID       uuid.UUID `json:"empresa_id" validate:"required"`
	LojaID          uuid.UUID `json:"loja_id" validate:"required"`
	PdvOrigemID     uuid.UUID `json:"pdv_origem_id" validate:"required"`
	TipoEvento      string    `json:"tipo_evento" validate:"required"`
	PayloadOriginal string    `json:"payload_original" validate:"required"`
	IdempotencyKey  string    `json:"idempotency_key" validate:"required"`
}

type UpdateEventoIngestaoInput struct {
	Status       *string    `json:"status,omitempty"`
	Erro         *string    `json:"erro,omitempty"`
	Tentativas   *int       `json:"tentativas,omitempty"`
	ProcessadoEm *time.Time `json:"processado_em,omitempty"`
}

type EventoIngestaoFilter struct {
	EmpresaID   *uuid.UUID
	LojaID      *uuid.UUID
	PdvOrigemID *uuid.UUID
	TipoEvento  *string
	Status      *string
	Limit       int
	Offset      int
}
