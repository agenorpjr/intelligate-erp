package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permissao struct {
	ID        uuid.UUID `json:"id"`
	Chave     string    `json:"chave"`
	Descricao *string   `json:"descricao,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePermissaoInput struct {
	Chave     string  `json:"chave" validate:"required"`
	Descricao *string `json:"descricao,omitempty"`
}
