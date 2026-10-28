package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PapelPermissaoRepository struct {
	querier *Queries
}

func NewPapelPermissaoRepository(pool *pgxpool.Pool) *PapelPermissaoRepository {
	return &PapelPermissaoRepository{querier: New(pool)}
}

func (r *PapelPermissaoRepository) Create(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return r.querier.CreatePapelPermissao(ctx, CreatePapelPermissaoParams{
		PapelID:     papelID,
		PermissaoID: permissaoID,
	})
}

func (r *PapelPermissaoRepository) Delete(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return r.querier.DeletePapelPermissao(ctx, DeletePapelPermissaoParams{
		PapelID:     papelID,
		PermissaoID: permissaoID,
	})
}

func (r *PapelPermissaoRepository) ListPermissoesByPapel(ctx context.Context, papelID uuid.UUID) ([]*domain.Permissao, error) {
	results, err := r.querier.ListPermissoesByPapel(ctx, papelID)
	if err != nil {
		return nil, err
	}
	permissoes := make([]*domain.Permissao, len(results))
	for i, p := range results {
		permissoes[i] = toDomainPermissaoFromPermisso(p)
	}
	return permissoes, nil
}

func toDomainPermissaoFromPermisso(p Permisso) *domain.Permissao {
	var descricao *string
	if p.Descricao.Valid {
		descricao = &p.Descricao.String
	}
	return &domain.Permissao{
		ID:        p.ID,
		Chave:     p.Chave,
		Descricao: descricao,
		CreatedAt: p.CreatedAt,
	}
}

func (r *PapelPermissaoRepository) ListPermissoesByUsuario(ctx context.Context, usuarioID, lojaID uuid.UUID) ([]*domain.Permissao, error) {
	results, err := r.querier.ListPermissoesByUsuario(ctx, ListPermissoesByUsuarioParams{
		UsuarioID: usuarioID,
		LojaID:    pgtype.UUID{Bytes: lojaID, Valid: lojaID != uuid.Nil},
	})
	if err != nil {
		return nil, err
	}
	permissoes := make([]*domain.Permissao, len(results))
	for i, p := range results {
		permissoes[i] = toDomainPermissaoFromPermisso(p)
	}
	return permissoes, nil
}
