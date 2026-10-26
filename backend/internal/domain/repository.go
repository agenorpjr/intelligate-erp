package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EmpresaRepository interface {
	Create(ctx context.Context, e *Empresa) error
	GetByID(ctx context.Context, id uuid.UUID) (*Empresa, error)
	GetByCNPJ(ctx context.Context, cnpj string) (*Empresa, error)
	List(ctx context.Context, limit, offset int) ([]*Empresa, error)
	Update(ctx context.Context, e *Empresa) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LojaRepository interface {
	Create(ctx context.Context, l *Loja) error
	GetByID(ctx context.Context, id uuid.UUID) (*Loja, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Loja, error)
	Update(ctx context.Context, l *Loja) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsuarioRepository interface {
	Create(ctx context.Context, u *Usuario) error
	GetByID(ctx context.Context, id uuid.UUID) (*Usuario, error)
	GetByEmail(ctx context.Context, email string) (*Usuario, error)
	GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*Usuario, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Usuario, error)
	Update(ctx context.Context, u *Usuario) error
	UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error
	UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PapelRepository interface {
	Create(ctx context.Context, p *Papel) error
	GetByID(ctx context.Context, id uuid.UUID) (*Papel, error)
	ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*Papel, error)
	Update(ctx context.Context, p *Papel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PermissaoRepository interface {
	Create(ctx context.Context, p *Permissao) error
	GetByID(ctx context.Context, id uuid.UUID) (*Permissao, error)
	GetByChave(ctx context.Context, chave string) (*Permissao, error)
	List(ctx context.Context, limit, offset int) ([]*Permissao, error)
}

type UsuarioPapelRepository interface {
	Create(ctx context.Context, up *UsuarioPapel) error
	GetByID(ctx context.Context, id uuid.UUID) (*UsuarioPapel, error)
	ListByUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*UsuarioPapelComDetalhes, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UsuarioPapelComDetalhes struct {
	ID        uuid.UUID  `json:"id"`
	UsuarioID uuid.UUID  `json:"usuario_id"`
	PapelID   uuid.UUID  `json:"papel_id"`
	LojaID    *uuid.UUID `json:"loja_id,omitempty"`
	PapelNome string     `json:"papel_nome"`
	LojaNome  *string    `json:"loja_nome,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type PapelPermissaoRepository interface {
	Create(ctx context.Context, papelID, permissaoID uuid.UUID) error
	Delete(ctx context.Context, papelID, permissaoID uuid.UUID) error
	ListPermissoesByPapel(ctx context.Context, papelID uuid.UUID) ([]*Permissao, error)
	ListPermissoesByUsuario(ctx context.Context, usuarioID, lojaID uuid.UUID) ([]*Permissao, error)
}

type AuditoriaRepository interface {
	Create(ctx context.Context, a *Auditoria) error
	List(ctx context.Context, filter AuditoriaFilter) ([]*Auditoria, error)
}
