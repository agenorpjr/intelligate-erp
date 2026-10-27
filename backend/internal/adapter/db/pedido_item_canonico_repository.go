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

type PedidoItemCanonicoRepository struct {
	querier *Queries
}

func NewPedidoItemCanonicoRepository(pool *pgxpool.Pool) *PedidoItemCanonicoRepository {
	return &PedidoItemCanonicoRepository{querier: New(pool)}
}

func (r *PedidoItemCanonicoRepository) Create(ctx context.Context, i *domain.PedidoItemCanonico) error {
	var produtoCanonicoID pgtype.UUID
	if i.ProdutoCanonicoID != nil {
		produtoCanonicoID = pgtype.UUID{Bytes: *i.ProdutoCanonicoID, Valid: true}
	}
	var skuExterno pgtype.Text
	if i.SKUExterno != nil {
		skuExterno = pgtype.Text{String: *i.SKUExterno, Valid: true}
	}
	var modificadores []byte
	if i.Modificadores != nil {
		modificadores = []byte(*i.Modificadores)
	}
	var observacoes pgtype.Text
	if i.Observacoes != nil {
		observacoes = pgtype.Text{String: *i.Observacoes, Valid: true}
	}

	params := CreatePedidoItemCanonicoParams{
		PedidoID:          i.PedidoID,
		ProdutoCanonicoID: produtoCanonicoID,
		SkuExterno:        skuExterno,
		Nome:              i.Nome,
		Quantidade:        pgtype.Numeric{Int: big.NewInt(int64(i.Quantidade * 10000)), Exp: -4, Valid: true},
		PrecoUnitario:     pgtype.Numeric{Int: big.NewInt(int64(i.PrecoUnitario * 10000)), Exp: -4, Valid: true},
		ValorTotal:        pgtype.Numeric{Int: big.NewInt(int64(i.ValorTotal * 10000)), Exp: -4, Valid: true},
		Desconto:          pgtype.Numeric{Int: big.NewInt(int64(i.Desconto * 10000)), Exp: -4, Valid: true},
		Modificadores:     modificadores,
		Observacoes:       observacoes,
	}

	result, err := r.querier.CreatePedidoItemCanonico(ctx, params)
	if err != nil {
		return err
	}
	i.ID = result.ID
	i.CreatedAt = result.CreatedAt
	i.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoItemCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoItemCanonico, error) {
	result, err := r.querier.GetPedidoItemCanonico(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPedidoItemCanonico(result), nil
}

func (r *PedidoItemCanonicoRepository) ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*domain.PedidoItemCanonico, error) {
	results, err := r.querier.ListPedidoItensByPedido(ctx, pedidoID)
	if err != nil {
		return nil, err
	}
	itens := make([]*domain.PedidoItemCanonico, len(results))
	for i, r := range results {
		itens[i] = toDomainPedidoItemCanonico(r)
	}
	return itens, nil
}

func (r *PedidoItemCanonicoRepository) Update(ctx context.Context, i *domain.PedidoItemCanonico) error {
	var produtoCanonicoID pgtype.UUID
	if i.ProdutoCanonicoID != nil {
		produtoCanonicoID = pgtype.UUID{Bytes: *i.ProdutoCanonicoID, Valid: true}
	}
	var skuExterno pgtype.Text
	if i.SKUExterno != nil {
		skuExterno = pgtype.Text{String: *i.SKUExterno, Valid: true}
	}
	var modificadores []byte
	if i.Modificadores != nil {
		modificadores = []byte(*i.Modificadores)
	}
	var observacoes pgtype.Text
	if i.Observacoes != nil {
		observacoes = pgtype.Text{String: *i.Observacoes, Valid: true}
	}

	params := UpdatePedidoItemCanonicoParams{
		ID:                i.ID,
		ProdutoCanonicoID: produtoCanonicoID,
		SkuExterno:        skuExterno,
		Nome:              i.Nome,
		Quantidade:        pgtype.Numeric{Int: big.NewInt(int64(i.Quantidade * 10000)), Exp: -4, Valid: true},
		PrecoUnitario:     pgtype.Numeric{Int: big.NewInt(int64(i.PrecoUnitario * 10000)), Exp: -4, Valid: true},
		ValorTotal:        pgtype.Numeric{Int: big.NewInt(int64(i.ValorTotal * 10000)), Exp: -4, Valid: true},
		Desconto:          pgtype.Numeric{Int: big.NewInt(int64(i.Desconto * 10000)), Exp: -4, Valid: true},
		Modificadores:     modificadores,
		Observacoes:       observacoes,
	}
	result, err := r.querier.UpdatePedidoItemCanonico(ctx, params)
	if err != nil {
		return err
	}
	i.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoItemCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePedidoItemCanonico(ctx, id)
}

func (r *PedidoItemCanonicoRepository) DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error {
	return r.querier.DeletePedidoItensByPedido(ctx, pedidoID)
}

func toDomainPedidoItemCanonico(i PedidoItensCanonico) *domain.PedidoItemCanonico {
	var produtoCanonicoID *uuid.UUID
	if i.ProdutoCanonicoID.Valid {
		var u uuid.UUID
		copy(u[:], i.ProdutoCanonicoID.Bytes[:])
		produtoCanonicoID = &u
	}
	var skuExterno, modificadores, observacoes *string
	if i.SkuExterno.Valid {
		skuExterno = &i.SkuExterno.String
	}
	if len(i.Modificadores) > 0 {
		s := string(i.Modificadores)
		modificadores = &s
	}
	if i.Observacoes.Valid {
		observacoes = &i.Observacoes.String
	}
	var deletedAt *time.Time
	if i.DeletedAt.Valid {
		deletedAt = &i.DeletedAt.Time
	}

	quantidade := 0.0
	if i.Quantidade.Valid && i.Quantidade.Int != nil {
		quantidade = float64(i.Quantidade.Int.Int64()) / 10000
	}
	precoUnitario := 0.0
	if i.PrecoUnitario.Valid && i.PrecoUnitario.Int != nil {
		precoUnitario = float64(i.PrecoUnitario.Int.Int64()) / 10000
	}
	valorTotal := 0.0
	if i.ValorTotal.Valid && i.ValorTotal.Int != nil {
		valorTotal = float64(i.ValorTotal.Int.Int64()) / 10000
	}
	desconto := 0.0
	if i.Desconto.Valid && i.Desconto.Int != nil {
		desconto = float64(i.Desconto.Int.Int64()) / 10000
	}

	return &domain.PedidoItemCanonico{
		ID:                i.ID,
		PedidoID:          i.PedidoID,
		ProdutoCanonicoID: produtoCanonicoID,
		SKUExterno:        skuExterno,
		Nome:              i.Nome,
		Quantidade:        quantidade,
		PrecoUnitario:     precoUnitario,
		ValorTotal:        valorTotal,
		Desconto:          desconto,
		Modificadores:     modificadores,
		Observacoes:       observacoes,
		CreatedAt:         i.CreatedAt,
		UpdatedAt:         i.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}
