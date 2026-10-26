package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissaoRepository struct {
	querier *Queries
}

func NewPermissaoRepository(pool *pgxpool.Pool) *PermissaoRepository {
	return &PermissaoRepository{querier: New(pool)}
}

func (r *PermissaoRepository) Create(ctx context.Context, p *domain.Permissao) error {
	var descricao pgtype.Text
	if p.Descricao != nil {
		descricao = pgtype.Text{String: *p.Descricao, Valid: true}
	}
	params := CreatePermissaoParams{
		Chave:     p.Chave,
		Descricao: descricao,
	}
	result, err := r.querier.CreatePermissao(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	return nil
}

func (r *PermissaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permissao, error) {
	result, err := r.querier.GetPermissao(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPermissao(result), nil
}

func (r *PermissaoRepository) GetByChave(ctx context.Context, chave string) (*domain.Permissao, error) {
	result, err := r.querier.GetPermissaoByChave(ctx, chave)
	if err != nil {
		return nil, err
	}
	return toDomainPermissao(result), nil
}

func (r *PermissaoRepository) List(ctx context.Context, limit, offset int) ([]*domain.Permissao, error) {
	results, err := r.querier.ListPermissoes(ctx, ListPermissoesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	permissoes := make([]*domain.Permissao, len(results))
	for i, r := range results {
		permissoes[i] = toDomainPermissao(r)
	}
	return permissoes, nil
}

func toDomainPermissao(p Permisso) *domain.Permissao {
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
