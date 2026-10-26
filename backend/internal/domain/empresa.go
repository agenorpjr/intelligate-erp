package domain

import (
	"time"

	"github.com/google/uuid"
)

type Empresa struct {
	ID           uuid.UUID  `json:"id"`
	CNPJ         string     `json:"cnpj"`
	RazaoSocial  string     `json:"razao_social"`
	NomeFantasia *string    `json:"nome_fantasia,omitempty"`
	Ativo        bool       `json:"ativo"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateEmpresaInput struct {
	CNPJ         string  `json:"cnpj" validate:"required"`
	RazaoSocial  string  `json:"razao_social" validate:"required"`
	NomeFantasia *string `json:"nome_fantasia,omitempty"`
}

type UpdateEmpresaInput struct {
	RazaoSocial  *string `json:"razao_social,omitempty"`
	NomeFantasia *string `json:"nome_fantasia,omitempty"`
	Ativo        *bool   `json:"ativo,omitempty"`
}
