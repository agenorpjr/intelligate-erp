package domain

import (
	"time"

	"github.com/google/uuid"
)

type PdvOrigem struct {
	ID           uuid.UUID  `json:"id"`
	EmpresaID    uuid.UUID  `json:"empresa_id"`
	Nome         string     `json:"nome"`
	Codigo       string     `json:"codigo"`
	Tipo         string     `json:"tipo"`
	Configuracao *string    `json:"configuracao,omitempty"`
	Ativo        bool       `json:"ativo"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreatePdvOrigemInput struct {
	EmpresaID    uuid.UUID `json:"empresa_id" validate:"required"`
	Nome         string    `json:"nome" validate:"required"`
	Codigo       string    `json:"codigo" validate:"required"`
	Tipo         string    `json:"tipo" validate:"required"`
	Configuracao *string   `json:"configuracao,omitempty"`
}

type UpdatePdvOrigemInput struct {
	Nome         *string `json:"nome,omitempty"`
	Codigo       *string `json:"codigo,omitempty"`
	Tipo         *string `json:"tipo,omitempty"`
	Configuracao *string `json:"configuracao,omitempty"`
	Ativo        *bool   `json:"ativo,omitempty"`
}
