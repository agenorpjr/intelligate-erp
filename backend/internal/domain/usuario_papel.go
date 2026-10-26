package domain

import (
	"time"

	"github.com/google/uuid"
)

type UsuarioPapel struct {
	ID        uuid.UUID  `json:"id"`
	UsuarioID uuid.UUID  `json:"usuario_id"`
	PapelID   uuid.UUID  `json:"papel_id"`
	LojaID    *uuid.UUID `json:"loja_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateUsuarioPapelInput struct {
	UsuarioID uuid.UUID  `json:"usuario_id" validate:"required"`
	PapelID   uuid.UUID  `json:"papel_id" validate:"required"`
	LojaID    *uuid.UUID `json:"loja_id,omitempty"`
}
