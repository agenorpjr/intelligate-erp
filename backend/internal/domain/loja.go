package domain

import (
	"time"

	"github.com/google/uuid"
)

type Loja struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	Codigo      string     `json:"codigo"`
	Nome        string     `json:"nome"`
	Endereco    *string    `json:"endereco,omitempty"`
	FusoHorario string     `json:"fuso_horario"`
	Ativo       bool       `json:"ativo"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateLojaInput struct {
	EmpresaID   uuid.UUID `json:"empresa_id" validate:"required"`
	Codigo      string    `json:"codigo" validate:"required"`
	Nome        string    `json:"nome" validate:"required"`
	Endereco    *string   `json:"endereco,omitempty"`
	FusoHorario string    `json:"fuso_horario" validate:"required"`
}

type UpdateLojaInput struct {
	Codigo      *string `json:"codigo,omitempty"`
	Nome        *string `json:"nome,omitempty"`
	Endereco    *string `json:"endereco,omitempty"`
	FusoHorario *string `json:"fuso_horario,omitempty"`
	Ativo       *bool   `json:"ativo,omitempty"`
}
