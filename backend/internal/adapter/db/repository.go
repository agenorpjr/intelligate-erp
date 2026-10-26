package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmpresaRepository struct {
	querier *Queries
}

func NewEmpresaRepository(pool *pgxpool.Pool) *EmpresaRepository {
	return &EmpresaRepository{querier: New(pool)}
}

func (r *EmpresaRepository) Create(ctx context.Context, e *domain.Empresa) error {
	var nomeFantasia pgtype.Text
	if e.NomeFantasia != nil {
		nomeFantasia = pgtype.Text{String: *e.NomeFantasia, Valid: true}
	}
	params := CreateEmpresaParams{
		Cnpj:         e.CNPJ,
		RazaoSocial:  e.RazaoSocial,
		NomeFantasia: nomeFantasia,
	}
	result, err := r.querier.CreateEmpresa(ctx, params)
	if err != nil {
		return err
	}
	e.ID = result.ID
	e.CreatedAt = result.CreatedAt
	e.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *EmpresaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Empresa, error) {
	result, err := r.querier.GetEmpresa(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainEmpresa(result), nil
}

func (r *EmpresaRepository) GetByCNPJ(ctx context.Context, cnpj string) (*domain.Empresa, error) {
	result, err := r.querier.GetEmpresaByCNPJ(ctx, cnpj)
	if err != nil {
		return nil, err
	}
	return toDomainEmpresa(result), nil
}

func (r *EmpresaRepository) List(ctx context.Context, limit, offset int) ([]*domain.Empresa, error) {
	results, err := r.querier.ListEmpresas(ctx, ListEmpresasParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	empresas := make([]*domain.Empresa, len(results))
	for i, r := range results {
		empresas[i] = toDomainEmpresa(r)
	}
	return empresas, nil
}

func (r *EmpresaRepository) Update(ctx context.Context, e *domain.Empresa) error {
	var nomeFantasia pgtype.Text
	if e.NomeFantasia != nil {
		nomeFantasia = pgtype.Text{String: *e.NomeFantasia, Valid: true}
	}
	params := UpdateEmpresaParams{
		ID:           e.ID,
		RazaoSocial:  e.RazaoSocial,
		NomeFantasia: nomeFantasia,
		Ativo:        e.Ativo,
	}
	result, err := r.querier.UpdateEmpresa(ctx, params)
	if err != nil {
		return err
	}
	e.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *EmpresaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteEmpresa(ctx, id)
}

func toDomainEmpresa(e Empresa) *domain.Empresa {
	var nomeFantasia *string
	if e.NomeFantasia.Valid {
		nomeFantasia = &e.NomeFantasia.String
	}
	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}
	return &domain.Empresa{
		ID:           e.ID,
		CNPJ:         e.Cnpj,
		RazaoSocial:  e.RazaoSocial,
		NomeFantasia: nomeFantasia,
		Ativo:        e.Ativo,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}
