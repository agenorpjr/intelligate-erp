package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PapelRepository struct {
	querier *Queries
}

func NewPapelRepository(pool *pgxpool.Pool) *PapelRepository {
	return &PapelRepository{querier: New(pool)}
}

func (r *PapelRepository) Create(ctx context.Context, p *domain.Papel) error {
	var descricao pgtype.Text
	if p.Descricao != nil {
		descricao = pgtype.Text{String: *p.Descricao, Valid: true}
	}
	params := CreatePapelParams{
		EmpresaID: p.EmpresaID,
		Nome:      p.Nome,
		Descricao: descricao,
		Sistema:   p.Sistema,
	}
	result, err := r.querier.CreatePapel(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PapelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Papel, error) {
	result, err := r.querier.GetPapel(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPapel(result), nil
}

func (r *PapelRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Papel, error) {
	results, err := r.querier.ListPapeisByEmpresa(ctx, ListPapeisByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	papeis := make([]*domain.Papel, len(results))
	for i, r := range results {
		papeis[i] = toDomainPapel(r)
	}
	return papeis, nil
}

func (r *PapelRepository) Update(ctx context.Context, p *domain.Papel) error {
	var descricao pgtype.Text
	if p.Descricao != nil {
		descricao = pgtype.Text{String: *p.Descricao, Valid: true}
	}
	params := UpdatePapelParams{
		ID:        p.ID,
		Nome:      p.Nome,
		Descricao: descricao,
		Sistema:   p.Sistema,
	}
	result, err := r.querier.UpdatePapel(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PapelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePapel(ctx, id)
}

func toDomainPapel(p Papei) *domain.Papel {
	var descricao *string
	if p.Descricao.Valid {
		descricao = &p.Descricao.String
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	return &domain.Papel{
		ID:        p.ID,
		EmpresaID: p.EmpresaID,
		Nome:      p.Nome,
		Descricao: descricao,
		Sistema:   p.Sistema,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
