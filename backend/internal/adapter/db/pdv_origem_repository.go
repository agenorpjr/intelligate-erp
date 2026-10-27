package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PdvOrigemRepository struct {
	querier *Queries
}

func NewPdvOrigemRepository(pool *pgxpool.Pool) *PdvOrigemRepository {
	return &PdvOrigemRepository{querier: New(pool)}
}

func (r *PdvOrigemRepository) Create(ctx context.Context, p *domain.PdvOrigem) error {
	var configuracao []byte
	if p.Configuracao != nil {
		configuracao = []byte(*p.Configuracao)
	}
	params := CreatePdvOrigemParams{
		EmpresaID:    p.EmpresaID,
		Nome:         p.Nome,
		Codigo:       p.Codigo,
		Tipo:         p.Tipo,
		Configuracao: configuracao,
	}
	result, err := r.querier.CreatePdvOrigem(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PdvOrigemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PdvOrigem, error) {
	result, err := r.querier.GetPdvOrigem(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPdvOrigem(result), nil
}

func (r *PdvOrigemRepository) GetByCodigo(ctx context.Context, empresaID uuid.UUID, codigo string) (*domain.PdvOrigem, error) {
	result, err := r.querier.GetPdvOrigemByCodigo(ctx, GetPdvOrigemByCodigoParams{
		EmpresaID: empresaID,
		Codigo:    codigo,
	})
	if err != nil {
		return nil, err
	}
	return toDomainPdvOrigem(result), nil
}

func (r *PdvOrigemRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.PdvOrigem, error) {
	results, err := r.querier.ListPdvsOrigemByEmpresa(ctx, ListPdvsOrigemByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	pdvs := make([]*domain.PdvOrigem, len(results))
	for i, r := range results {
		pdvs[i] = toDomainPdvOrigem(r)
	}
	return pdvs, nil
}

func (r *PdvOrigemRepository) Update(ctx context.Context, p *domain.PdvOrigem) error {
	var configuracao []byte
	if p.Configuracao != nil {
		configuracao = []byte(*p.Configuracao)
	}
	params := UpdatePdvOrigemParams{
		ID:           p.ID,
		Nome:         p.Nome,
		Codigo:       p.Codigo,
		Tipo:         p.Tipo,
		Configuracao: configuracao,
		Ativo:        p.Ativo,
	}
	result, err := r.querier.UpdatePdvOrigem(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PdvOrigemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePdvOrigem(ctx, id)
}

func toDomainPdvOrigem(p PdvsOrigem) *domain.PdvOrigem {
	var configuracao *string
	if len(p.Configuracao) > 0 {
		s := string(p.Configuracao)
		configuracao = &s
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	return &domain.PdvOrigem{
		ID:           p.ID,
		EmpresaID:    p.EmpresaID,
		Nome:         p.Nome,
		Codigo:       p.Codigo,
		Tipo:         p.Tipo,
		Configuracao: configuracao,
		Ativo:        p.Ativo,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}
