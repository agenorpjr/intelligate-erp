package domain

import (
	"time"

	"github.com/google/uuid"
)

type Auditoria struct {
	ID         int64      `json:"id"`
	EmpresaID  *uuid.UUID `json:"empresa_id,omitempty"`
	UsuarioID  *uuid.UUID `json:"usuario_id,omitempty"`
	Entidade   string     `json:"entidade"`
	EntidadeID uuid.UUID  `json:"entidade_id"`
	Acao       string     `json:"acao"`
	Antes      *string    `json:"antes,omitempty"`
	Depois     *string    `json:"depois,omitempty"`
	IP         *string    `json:"ip,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateAuditoriaInput struct {
	EmpresaID  *uuid.UUID `json:"empresa_id,omitempty"`
	UsuarioID  *uuid.UUID `json:"usuario_id,omitempty"`
	Entidade   string     `json:"entidade" validate:"required"`
	EntidadeID uuid.UUID  `json:"entidade_id" validate:"required"`
	Acao       string     `json:"acao" validate:"required"`
	Antes      *string    `json:"antes,omitempty"`
	Depois     *string    `json:"depois,omitempty"`
	IP         *string    `json:"ip,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
}

type AuditoriaFilter struct {
	EmpresaID *uuid.UUID
	UsuarioID *uuid.UUID
	Entidade  *string
	Acao      *string
	Limit     int
	Offset    int
}
