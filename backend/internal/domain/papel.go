package domain

import (
	"time"

	"github.com/google/uuid"
)

type Papel struct {
	ID        uuid.UUID  `json:"id"`
	EmpresaID uuid.UUID  `json:"empresa_id"`
	Nome      string     `json:"nome"`
	Descricao *string    `json:"descricao,omitempty"`
	Sistema   bool       `json:"sistema"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreatePapelInput struct {
	EmpresaID uuid.UUID `json:"empresa_id" validate:"required"`
	Nome      string    `json:"nome" validate:"required"`
	Descricao *string   `json:"descricao,omitempty"`
	Sistema   bool      `json:"sistema"`
}

type UpdatePapelInput struct {
	Nome      *string `json:"nome,omitempty"`
	Descricao *string `json:"descricao,omitempty"`
	Sistema   *bool   `json:"sistema,omitempty"`
}
