package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProdutoRepository struct {
	querier *Queries
}

func NewProdutoRepository(pool *pgxpool.Pool) *ProdutoRepository {
	return &ProdutoRepository{querier: New(pool)}
}

func (r *ProdutoRepository) Create(ctx context.Context, p *domain.Produto) error {
	var skuExterno, categoria pgtype.Text
	if p.SKUExterno != nil {
		skuExterno = pgtype.Text{String: *p.SKUExterno, Valid: true}
	}
	if p.Categoria != nil {
		categoria = pgtype.Text{String: *p.Categoria, Valid: true}
	}
	params := CreateProdutoParams{
		EmpresaID:  p.EmpresaID,
		Nome:       p.Nome,
		SkuExterno: skuExterno,
		Categoria:  categoria,
	}
	result, err := r.querier.CreateProduto(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ProdutoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	result, err := r.querier.GetProduto(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainProduto(result), nil
}

func (r *ProdutoRepository) GetBySKUExterno(ctx context.Context, empresaID uuid.UUID, skuExterno string) (*domain.Produto, error) {
	result, err := r.querier.GetProdutoBySKUExterno(ctx, GetProdutoBySKUExternoParams{
		EmpresaID:  empresaID,
		SkuExterno: pgtype.Text{String: skuExterno, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return toDomainProduto(result), nil
}

func (r *ProdutoRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Produto, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListProdutosByEmpresa(ctx, ListProdutosByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   ativoParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	produtos := make([]*domain.Produto, len(results))
	for i, r := range results {
		produtos[i] = toDomainProduto(r)
	}
	return produtos, nil
}

func (r *ProdutoRepository) Update(ctx context.Context, p *domain.Produto) error {
	var skuExterno, categoria pgtype.Text
	if p.SKUExterno != nil {
		skuExterno = pgtype.Text{String: *p.SKUExterno, Valid: true}
	}
	if p.Categoria != nil {
		categoria = pgtype.Text{String: *p.Categoria, Valid: true}
	}
	params := UpdateProdutoParams{
		ID:         p.ID,
		Nome:       p.Nome,
		SkuExterno: skuExterno,
		Categoria:  categoria,
		Ativo:      p.Ativo,
	}
	result, err := r.querier.UpdateProduto(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ProdutoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteProduto(ctx, id)
}

func toDomainProduto(p Produto) *domain.Produto {
	var skuExterno, categoria *string
	if p.SkuExterno.Valid {
		skuExterno = &p.SkuExterno.String
	}
	if p.Categoria.Valid {
		categoria = &p.Categoria.String
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	return &domain.Produto{
		ID:         p.ID,
		EmpresaID:  p.EmpresaID,
		Nome:       p.Nome,
		SKUExterno: skuExterno,
		Categoria:  categoria,
		Ativo:      p.Ativo,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}
