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

type PedidoPagamentoCanonicoRepository struct {
	querier *Queries
}

func NewPedidoPagamentoCanonicoRepository(pool *pgxpool.Pool) *PedidoPagamentoCanonicoRepository {
	return &PedidoPagamentoCanonicoRepository{querier: New(pool)}
}

func (r *PedidoPagamentoCanonicoRepository) Create(ctx context.Context, p *domain.PedidoPagamentoCanonico) error {
	var referenciaExterna pgtype.Text
	if p.ReferenciaExterna != nil {
		referenciaExterna = pgtype.Text{String: *p.ReferenciaExterna, Valid: true}
	}

	params := CreatePedidoPagamentoCanonicoParams{
		PedidoID:          p.PedidoID,
		FormaPagamento:    p.FormaPagamento,
		Valor:             pgtype.Numeric{Int: big.NewInt(int64(p.Valor * 10000)), Exp: -4, Valid: true},
		Taxa:              pgtype.Numeric{Int: big.NewInt(int64(p.Taxa * 10000)), Exp: -4, Valid: true},
		Troco:             pgtype.Numeric{Int: big.NewInt(int64(p.Troco * 10000)), Exp: -4, Valid: true},
		ReferenciaExterna: referenciaExterna,
		Status:            p.Status,
	}
	result, err := r.querier.CreatePedidoPagamentoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoPagamentoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoPagamentoCanonico, error) {
	result, err := r.querier.GetPedidoPagamentoCanonico(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPedidoPagamentoCanonico(result), nil
}

func (r *PedidoPagamentoCanonicoRepository) ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*domain.PedidoPagamentoCanonico, error) {
	results, err := r.querier.ListPedidoPagamentosByPedido(ctx, pedidoID)
	if err != nil {
		return nil, err
	}
	pagamentos := make([]*domain.PedidoPagamentoCanonico, len(results))
	for i, r := range results {
		pagamentos[i] = toDomainPedidoPagamentoCanonico(r)
	}
	return pagamentos, nil
}

func (r *PedidoPagamentoCanonicoRepository) Update(ctx context.Context, p *domain.PedidoPagamentoCanonico) error {
	var referenciaExterna pgtype.Text
	if p.ReferenciaExterna != nil {
		referenciaExterna = pgtype.Text{String: *p.ReferenciaExterna, Valid: true}
	}

	params := UpdatePedidoPagamentoCanonicoParams{
		ID:                p.ID,
		FormaPagamento:    p.FormaPagamento,
		Valor:             pgtype.Numeric{Int: big.NewInt(int64(p.Valor * 10000)), Exp: -4, Valid: true},
		Taxa:              pgtype.Numeric{Int: big.NewInt(int64(p.Taxa * 10000)), Exp: -4, Valid: true},
		Troco:             pgtype.Numeric{Int: big.NewInt(int64(p.Troco * 10000)), Exp: -4, Valid: true},
		ReferenciaExterna: referenciaExterna,
		Status:            p.Status,
	}
	result, err := r.querier.UpdatePedidoPagamentoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoPagamentoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePedidoPagamentoCanonico(ctx, id)
}

func (r *PedidoPagamentoCanonicoRepository) DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error {
	return r.querier.DeletePedidoPagamentosByPedido(ctx, pedidoID)
}

func toDomainPedidoPagamentoCanonico(p PedidoPagamentosCanonico) *domain.PedidoPagamentoCanonico {
	var referenciaExterna *string
	if p.ReferenciaExterna.Valid {
		referenciaExterna = &p.ReferenciaExterna.String
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}

	valor := 0.0
	if p.Valor.Valid && p.Valor.Int != nil {
		valor = float64(p.Valor.Int.Int64()) / 10000
	}
	taxa := 0.0
	if p.Taxa.Valid && p.Taxa.Int != nil {
		taxa = float64(p.Taxa.Int.Int64()) / 10000
	}
	troco := 0.0
	if p.Troco.Valid && p.Troco.Int != nil {
		troco = float64(p.Troco.Int.Int64()) / 10000
	}

	return &domain.PedidoPagamentoCanonico{
		ID:                p.ID,
		PedidoID:          p.PedidoID,
		FormaPagamento:    p.FormaPagamento,
		Valor:             valor,
		Taxa:              taxa,
		Troco:             troco,
		ReferenciaExterna: referenciaExterna,
		Status:            p.Status,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}
