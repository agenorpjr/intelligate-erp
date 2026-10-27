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

type ProdutoCanonicoRepository struct {
	querier *Queries
}

func NewProdutoCanonicoRepository(pool *pgxpool.Pool) *ProdutoCanonicoRepository {
	return &ProdutoCanonicoRepository{querier: New(pool)}
}

func (r *ProdutoCanonicoRepository) Create(ctx context.Context, p *domain.ProdutoCanonico) error {
	var categoriaID pgtype.UUID
	if p.CategoriaID != nil {
		categoriaID = pgtype.UUID{Bytes: *p.CategoriaID, Valid: true}
	}
	var descricao pgtype.Text
	if p.Descricao != nil {
		descricao = pgtype.Text{String: *p.Descricao, Valid: true}
	}
	var ncm, cest, tributacao pgtype.Text
	if p.NCM != nil {
		ncm = pgtype.Text{String: *p.NCM, Valid: true}
	}
	if p.CEST != nil {
		cest = pgtype.Text{String: *p.CEST, Valid: true}
	}
	if p.Tributacao != nil {
		tributacao = pgtype.Text{String: *p.Tributacao, Valid: true}
	}

	precoBase := big.NewInt(int64(p.PrecoBase * 10000))

	params := CreateProdutoCanonicoParams{
		EmpresaID:   p.EmpresaID,
		CategoriaID: categoriaID,
		Sku:         p.SKU,
		Nome:        p.Nome,
		Descricao:   descricao,
		Unidade:     p.Unidade,
		PrecoBase:   pgtype.Numeric{Int: precoBase, Exp: -4, Valid: true},
		Ncm:         ncm,
		Cest:        cest,
		Tributacao:  tributacao,
	}
	result, err := r.querier.CreateProdutoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ProdutoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProdutoCanonico, error) {
	result, err := r.querier.GetProdutoCanonico(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainProdutoCanonico(result), nil
}

func (r *ProdutoCanonicoRepository) GetBySKU(ctx context.Context, empresaID uuid.UUID, sku string) (*domain.ProdutoCanonico, error) {
	result, err := r.querier.GetProdutoCanonicoBySKU(ctx, GetProdutoCanonicoBySKUParams{
		EmpresaID: empresaID,
		Sku:       sku,
	})
	if err != nil {
		return nil, err
	}
	return toDomainProdutoCanonico(result), nil
}

func (r *ProdutoCanonicoRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.ProdutoCanonico, error) {
	results, err := r.querier.ListProdutosCanonicosByEmpresa(ctx, ListProdutosCanonicosByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	produtos := make([]*domain.ProdutoCanonico, len(results))
	for i, r := range results {
		produtos[i] = toDomainProdutoCanonico(r)
	}
	return produtos, nil
}

func (r *ProdutoCanonicoRepository) Update(ctx context.Context, p *domain.ProdutoCanonico) error {
	var categoriaID pgtype.UUID
	if p.CategoriaID != nil {
		categoriaID = pgtype.UUID{Bytes: *p.CategoriaID, Valid: true}
	}
	var descricao pgtype.Text
	if p.Descricao != nil {
		descricao = pgtype.Text{String: *p.Descricao, Valid: true}
	}
	var ncm, cest, tributacao pgtype.Text
	if p.NCM != nil {
		ncm = pgtype.Text{String: *p.NCM, Valid: true}
	}
	if p.CEST != nil {
		cest = pgtype.Text{String: *p.CEST, Valid: true}
	}
	if p.Tributacao != nil {
		tributacao = pgtype.Text{String: *p.Tributacao, Valid: true}
	}

	precoBase := big.NewInt(int64(p.PrecoBase * 10000))

	params := UpdateProdutoCanonicoParams{
		ID:          p.ID,
		CategoriaID: categoriaID,
		Sku:         p.SKU,
		Nome:        p.Nome,
		Descricao:   descricao,
		Unidade:     p.Unidade,
		PrecoBase:   pgtype.Numeric{Int: precoBase, Exp: -4, Valid: true},
		Ncm:         ncm,
		Cest:        cest,
		Tributacao:  tributacao,
		Ativo:       p.Ativo,
	}
	result, err := r.querier.UpdateProdutoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ProdutoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteProdutoCanonico(ctx, id)
}

func toDomainProdutoCanonico(p ProdutosCanonico) *domain.ProdutoCanonico {
	var categoriaID *uuid.UUID
	if p.CategoriaID.Valid {
		var u uuid.UUID
		copy(u[:], p.CategoriaID.Bytes[:])
		categoriaID = &u
	}
	var descricao, ncm, cest, tributacao *string
	if p.Descricao.Valid {
		descricao = &p.Descricao.String
	}
	if p.Ncm.Valid {
		ncm = &p.Ncm.String
	}
	if p.Cest.Valid {
		cest = &p.Cest.String
	}
	if p.Tributacao.Valid {
		tributacao = &p.Tributacao.String
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}

	// Convert pgtype.Numeric to float64
	precoBase := 0.0
	if p.PrecoBase.Valid && p.PrecoBase.Int != nil {
		precoBase = float64(p.PrecoBase.Int.Int64()) / 10000
	}

	return &domain.ProdutoCanonico{
		ID:          p.ID,
		EmpresaID:   p.EmpresaID,
		CategoriaID: categoriaID,
		SKU:         p.Sku,
		Nome:        p.Nome,
		Descricao:   descricao,
		Unidade:     p.Unidade,
		PrecoBase:   precoBase,
		NCM:         ncm,
		CEST:        cest,
		Tributacao:  tributacao,
		Ativo:       p.Ativo,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
