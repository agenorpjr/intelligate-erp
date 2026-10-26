package domain

import (
	"time"

	"github.com/google/uuid"
)

type Usuario struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	Email       string     `json:"email"`
	Nome        string     `json:"nome"`
	SenhaHash   string     `json:"-"` // nunca serializar
	Ativo       bool       `json:"ativo"`
	UltimoLogin *time.Time `json:"ultimo_login,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateUsuarioInput struct {
	EmpresaID uuid.UUID `json:"empresa_id" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Nome      string    `json:"nome" validate:"required"`
	Senha     string    `json:"senha" validate:"required,min=8"`
}

type UpdateUsuarioInput struct {
	Nome  *string `json:"nome,omitempty"`
	Email *string `json:"email,omitempty"`
	Ativo *bool   `json:"ativo,omitempty"`
}

type LoginInput struct {
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required"`
}

type LoginResponse struct {
	Token     string  `json:"token"`
	Usuario   Usuario `json:"usuario"`
	ExpiresIn int64   `json:"expires_in"`
}
