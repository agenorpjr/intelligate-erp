package domain

import (
	"time"

	"github.com/google/uuid"
)

// SenhaRedefinicao representa um token de redefinicao de senha (esqueci-senha).
type SenhaRedefinicao struct {
	ID        uuid.UUID  `json:"id"`
	UsuarioID uuid.UUID  `json:"usuario_id"`
	TokenHash string     `json:"-"` // nunca serializar o hash
	Usado     bool       `json:"usado"`
	ExpiraEm  time.Time  `json:"expira_em"`
	CriadoEm  time.Time  `json:"criado_em"`
	UsadoEm   *time.Time `json:"usado_em,omitempty"`
}

// SessaoRevogada representa um token JWT revogado (logout / invalidez de sessao).
type SessaoRevogada struct {
	ID         uuid.UUID `json:"id"`
	UsuarioID  uuid.UUID `json:"usuario_id"`
	JTI        string    `json:"jti"`
	ExpiraEm   time.Time `json:"expira_em"`
	RevogadoEm time.Time `json:"revogado_em"`
	Motivo     string    `json:"motivo"`
}

// Inputs

type EsqueciSenhaInput struct {
	Email string `json:"email" validate:"required,email"`
}

type RedefinirSenhaInput struct {
	Token     string `json:"token" validate:"required"`
	NovaSenha string `json:"nova_senha" validate:"required,min=8"`
}

type AlterarSenhaInput struct {
	SenhaAtual string `json:"senha_atual" validate:"required"`
	NovaSenha  string `json:"nova_senha" validate:"required,min=8"`
}

type LogoutInput struct {
	// Nenhum campo obrigatorio no corpo; o token vem do header Authorization.
}
