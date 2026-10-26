package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LojaRepository struct {
	querier *Queries
}

func NewLojaRepository(pool *pgxpool.Pool) *LojaRepository {
	return &LojaRepository{querier: New(pool)}
}

func (r *LojaRepository) Create(ctx context.Context, l *domain.Loja) error {
	var endereco []byte
	if l.Endereco != nil {
		var err error
		endereco, err = json.Marshal(l.Endereco)
		if err != nil {
			return err
		}
	}
	params := CreateLojaParams{
		EmpresaID:   l.EmpresaID,
		Codigo:      l.Codigo,
		Nome:        l.Nome,
		Endereco:    endereco,
		FusoHorario: l.FusoHorario,
	}
	result, err := r.querier.CreateLoja(ctx, params)
	if err != nil {
		return err
	}
	l.ID = result.ID
	l.CreatedAt = result.CreatedAt
	l.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *LojaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Loja, error) {
	result, err := r.querier.GetLoja(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainLoja(result), nil
}

func (r *LojaRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Loja, error) {
	results, err := r.querier.ListLojasByEmpresa(ctx, ListLojasByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	lojas := make([]*domain.Loja, len(results))
	for i, r := range results {
		lojas[i] = toDomainLoja(r)
	}
	return lojas, nil
}

func (r *LojaRepository) Update(ctx context.Context, l *domain.Loja) error {
	var endereco []byte
	if l.Endereco != nil {
		var err error
		endereco, err = json.Marshal(l.Endereco)
		if err != nil {
			return err
		}
	}
	params := UpdateLojaParams{
		ID:          l.ID,
		Codigo:      l.Codigo,
		Nome:        l.Nome,
		Endereco:    endereco,
		FusoHorario: l.FusoHorario,
		Ativo:       l.Ativo,
	}
	result, err := r.querier.UpdateLoja(ctx, params)
	if err != nil {
		return err
	}
	l.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *LojaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteLoja(ctx, id)
}

func toDomainLoja(l Loja) *domain.Loja {
	var endereco *string
	if len(l.Endereco) > 0 {
		var s string
		if err := json.Unmarshal(l.Endereco, &s); err == nil {
			endereco = &s
		}
	}
	var deletedAt *time.Time
	if l.DeletedAt.Valid {
		deletedAt = &l.DeletedAt.Time
	}
	return &domain.Loja{
		ID:          l.ID,
		EmpresaID:   l.EmpresaID,
		Codigo:      l.Codigo,
		Nome:        l.Nome,
		Endereco:    endereco,
		FusoHorario: l.FusoHorario,
		Ativo:       l.Ativo,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
