package db

import (
	"context"

	"github.com/google/uuid"
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
