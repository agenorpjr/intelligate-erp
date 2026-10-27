package db

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InsumoRepository struct {
	querier *Queries
}

func NewInsumoRepository(pool *pgxpool.Pool) *InsumoRepository {
	return &InsumoRepository{querier: New(pool)}
}

func (r *InsumoRepository) Create(ctx context.Context, i *domain.Insumo) error {
	var fornecedor pgtype.Text
	if i.Fornecedor != nil {
		fornecedor = pgtype.Text{String: *i.Fornecedor, Valid: true}
	}
	custoUnitarioAtual := big.NewInt(int64(i.CustoUnitarioAtual * 10000))
	params := CreateInsumoParams{
		EmpresaID:          i.EmpresaID,
		Nome:               i.Nome,
		UnidadeMedida:      i.UnidadeMedida,
		CustoUnitarioAtual: pgtype.Numeric{Int: custoUnitarioAtual, Exp: -4, Valid: true},
		Fornecedor:         fornecedor,
	}
	result, err := r.querier.CreateInsumo(ctx, params)
	if err != nil {
		return err
	}
	i.ID = result.ID
	i.CreatedAt = result.CreatedAt
	i.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *InsumoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Insumo, error) {
	result, err := r.querier.GetInsumo(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainInsumo(result), nil
}

func (r *InsumoRepository) GetByNome(ctx context.Context, empresaID uuid.UUID, nome string) (*domain.Insumo, error) {
	result, err := r.querier.GetInsumoByNome(ctx, GetInsumoByNomeParams{
		EmpresaID: empresaID,
		Nome:      nome,
	})
	if err != nil {
		return nil, err
	}
	return toDomainInsumo(result), nil
}

func (r *InsumoRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Insumo, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListInsumosByEmpresa(ctx, ListInsumosByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   ativoParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	insumos := make([]*domain.Insumo, len(results))
	for i, r := range results {
		insumos[i] = toDomainInsumo(r)
	}
	return insumos, nil
}

func (r *InsumoRepository) Update(ctx context.Context, i *domain.Insumo) error {
	var fornecedor pgtype.Text
	if i.Fornecedor != nil {
		fornecedor = pgtype.Text{String: *i.Fornecedor, Valid: true}
	}
	custoUnitarioAtual := big.NewInt(int64(i.CustoUnitarioAtual * 10000))
	params := UpdateInsumoParams{
		ID:                 i.ID,
		Nome:               i.Nome,
		UnidadeMedida:      i.UnidadeMedida,
		CustoUnitarioAtual: pgtype.Numeric{Int: custoUnitarioAtual, Exp: -4, Valid: true},
		Fornecedor:         fornecedor,
		Ativo:              i.Ativo,
	}
	result, err := r.querier.UpdateInsumo(ctx, params)
	if err != nil {
		return err
	}
	i.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *InsumoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteInsumo(ctx, id)
}

func (r *InsumoRepository) CreateCusto(ctx context.Context, c *domain.InsumoCusto) error {
	var observacao pgtype.Text
	if c.Observacao != nil {
		observacao = pgtype.Text{String: *c.Observacao, Valid: true}
	}
	custoUnitario := big.NewInt(int64(c.CustoUnitario * 10000))
	params := CreateInsumoCustoParams{
		InsumoID:      c.InsumoID,
		DataVigencia:  c.DataVigencia,
		CustoUnitario: pgtype.Numeric{Int: custoUnitario, Exp: -4, Valid: true},
		Observacao:    observacao,
	}
	result, err := r.querier.CreateInsumoCusto(ctx, params)
	if err != nil {
		return err
	}
	c.ID = result.ID
	c.CreatedAt = result.CreatedAt
	return nil
}

func (r *InsumoRepository) GetCustoVigente(ctx context.Context, insumoID uuid.UUID, dataVigencia time.Time) (*domain.InsumoCusto, error) {
	result, err := r.querier.GetInsumoCustoVigente(ctx, GetInsumoCustoVigenteParams{
		InsumoID:     insumoID,
		DataVigencia: dataVigencia,
	})
	if err != nil {
		return nil, err
	}
	return toDomainInsumoCusto(result), nil
}

func (r *InsumoRepository) GetCustoAtual(ctx context.Context, insumoID uuid.UUID) (*domain.InsumoCusto, error) {
	result, err := r.querier.GetInsumoCustoAtual(ctx, insumoID)
	if err != nil {
		return nil, err
	}
	return toDomainInsumoCusto(result), nil
}

func (r *InsumoRepository) ListCustosByInsumo(ctx context.Context, insumoID uuid.UUID, limit, offset int) ([]*domain.InsumoCusto, error) {
	results, err := r.querier.ListInsumoCustosByInsumo(ctx, ListInsumoCustosByInsumoParams{
		InsumoID: insumoID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	custos := make([]*domain.InsumoCusto, len(results))
	for i, r := range results {
		custos[i] = toDomainInsumoCusto(r)
	}
	return custos, nil
}

func toDomainInsumo(i Insumo) *domain.Insumo {
	var fornecedor *string
	if i.Fornecedor.Valid {
		fornecedor = &i.Fornecedor.String
	}
	var deletedAt *time.Time
	if i.DeletedAt.Valid {
		deletedAt = &i.DeletedAt.Time
	}
	custoUnitarioAtual := 0.0
	if i.CustoUnitarioAtual.Valid && i.CustoUnitarioAtual.Int != nil {
		custoUnitarioAtual = float64(i.CustoUnitarioAtual.Int.Int64()) / 10000
	}
	return &domain.Insumo{
		ID:                 i.ID,
		EmpresaID:          i.EmpresaID,
		Nome:               i.Nome,
		UnidadeMedida:      i.UnidadeMedida,
		CustoUnitarioAtual: custoUnitarioAtual,
		Fornecedor:         fornecedor,
		Ativo:              i.Ativo,
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
		DeletedAt:          deletedAt,
	}
}

func toDomainInsumoCusto(c InsumoCusto) *domain.InsumoCusto {
	var observacao *string
	if c.Observacao.Valid {
		observacao = &c.Observacao.String
	}
	custoUnitario := 0.0
	if c.CustoUnitario.Valid && c.CustoUnitario.Int != nil {
		custoUnitario = float64(c.CustoUnitario.Int.Int64()) / 10000
	}
	return &domain.InsumoCusto{
		ID:            c.ID,
		InsumoID:      c.InsumoID,
		DataVigencia:  c.DataVigencia,
		CustoUnitario: custoUnitario,
		Observacao:    observacao,
		CreatedAt:     c.CreatedAt,
	}
}
