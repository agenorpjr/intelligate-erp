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

type PedidoCanonicoRepository struct {
	querier *Queries
}

func NewPedidoCanonicoRepository(pool *pgxpool.Pool) *PedidoCanonicoRepository {
	return &PedidoCanonicoRepository{querier: New(pool)}
}

func (r *PedidoCanonicoRepository) Create(ctx context.Context, p *domain.PedidoCanonico) error {
	params := CreatePedidoCanonicoParams{
		EmpresaID:      p.EmpresaID,
		LojaID:         p.LojaID,
		PdvOrigemID:    p.PdvOrigemID,
		NumeroExterno:  p.NumeroExterno,
		DataHora:       p.DataHora,
		Status:         p.Status,
		ValorTotal:     pgtype.Numeric{Int: big.NewInt(int64(p.ValorTotal * 10000)), Exp: -4, Valid: true},
		ValorDesconto:  pgtype.Numeric{Int: big.NewInt(int64(p.ValorDesconto * 10000)), Exp: -4, Valid: true},
		ValorAcrescimo: pgtype.Numeric{Int: big.NewInt(int64(p.ValorAcrescimo * 10000)), Exp: -4, Valid: true},
		IdempotencyKey: p.IdempotencyKey,
		ClienteNome: func() pgtype.Text {
			if p.ClienteNome != nil {
				return pgtype.Text{String: *p.ClienteNome, Valid: true}
			}
			return pgtype.Text{}
		}(),
		ClienteDocumento: func() pgtype.Text {
			if p.ClienteDocumento != nil {
				return pgtype.Text{String: *p.ClienteDocumento, Valid: true}
			}
			return pgtype.Text{}
		}(),
		ClienteEmail: func() pgtype.Text {
			if p.ClienteEmail != nil {
				return pgtype.Text{String: *p.ClienteEmail, Valid: true}
			}
			return pgtype.Text{}
		}(),
		Observacoes: func() pgtype.Text {
			if p.Observacoes != nil {
				return pgtype.Text{String: *p.Observacoes, Valid: true}
			}
			return pgtype.Text{}
		}(),
	}
	result, err := r.querier.CreatePedidoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoCanonico, error) {
	result, err := r.querier.GetPedidoCanonico(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPedidoCanonico(result), nil
}

func (r *PedidoCanonicoRepository) GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.PedidoCanonico, error) {
	result, err := r.querier.GetPedidoCanonicoByIdempotencyKey(ctx, GetPedidoCanonicoByIdempotencyKeyParams{
		EmpresaID:      empresaID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, err
	}
	return toDomainPedidoCanonico(result), nil
}

func (r *PedidoCanonicoRepository) List(ctx context.Context, filter domain.PedidoCanonicoFilter) ([]*domain.PedidoCanonico, error) {
	var lojaID uuid.UUID
	if filter.LojaID != nil {
		lojaID = *filter.LojaID
	}
	var dataInicio, dataFim time.Time
	if filter.DataInicio != nil {
		dataInicio = *filter.DataInicio
	}
	if filter.DataFim != nil {
		dataFim = *filter.DataFim
	}
	var status string
	if filter.Status != nil {
		status = *filter.Status
	}

	results, err := r.querier.ListPedidosCanonicos(ctx, ListPedidosCanonicosParams{
		EmpresaID: *filter.EmpresaID,
		Column2:   lojaID,
		Column3:   dataInicio,
		Column4:   dataFim,
		Column5:   status,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	pedidos := make([]*domain.PedidoCanonico, len(results))
	for i, r := range results {
		pedidos[i] = toDomainPedidoCanonico(r)
	}
	return pedidos, nil
}

func (r *PedidoCanonicoRepository) Update(ctx context.Context, p *domain.PedidoCanonico) error {
	params := UpdatePedidoCanonicoParams{
		ID:             p.ID,
		LojaID:         p.LojaID,
		PdvOrigemID:    p.PdvOrigemID,
		NumeroExterno:  p.NumeroExterno,
		DataHora:       p.DataHora,
		Status:         p.Status,
		ValorTotal:     pgtype.Numeric{Int: big.NewInt(int64(p.ValorTotal * 10000)), Exp: -4, Valid: true},
		ValorDesconto:  pgtype.Numeric{Int: big.NewInt(int64(p.ValorDesconto * 10000)), Exp: -4, Valid: true},
		ValorAcrescimo: pgtype.Numeric{Int: big.NewInt(int64(p.ValorAcrescimo * 10000)), Exp: -4, Valid: true},
		ClienteNome: func() pgtype.Text {
			if p.ClienteNome != nil {
				return pgtype.Text{String: *p.ClienteNome, Valid: true}
			}
			return pgtype.Text{}
		}(),
		ClienteDocumento: func() pgtype.Text {
			if p.ClienteDocumento != nil {
				return pgtype.Text{String: *p.ClienteDocumento, Valid: true}
			}
			return pgtype.Text{}
		}(),
		ClienteEmail: func() pgtype.Text {
			if p.ClienteEmail != nil {
				return pgtype.Text{String: *p.ClienteEmail, Valid: true}
			}
			return pgtype.Text{}
		}(),
		Observacoes: func() pgtype.Text {
			if p.Observacoes != nil {
				return pgtype.Text{String: *p.Observacoes, Valid: true}
			}
			return pgtype.Text{}
		}(),
	}
	result, err := r.querier.UpdatePedidoCanonico(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PedidoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePedidoCanonico(ctx, id)
}

func toDomainPedidoCanonico(p PedidosCanonico) *domain.PedidoCanonico {
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}

	valorTotal := 0.0
	if p.ValorTotal.Valid && p.ValorTotal.Int != nil {
		valorTotal = float64(p.ValorTotal.Int.Int64()) / 10000
	}
	valorDesconto := 0.0
	if p.ValorDesconto.Valid && p.ValorDesconto.Int != nil {
		valorDesconto = float64(p.ValorDesconto.Int.Int64()) / 10000
	}
	valorAcrescimo := 0.0
	if p.ValorAcrescimo.Valid && p.ValorAcrescimo.Int != nil {
		valorAcrescimo = float64(p.ValorAcrescimo.Int.Int64()) / 10000
	}

	return &domain.PedidoCanonico{
		ID:             p.ID,
		EmpresaID:      p.EmpresaID,
		LojaID:         p.LojaID,
		PdvOrigemID:    p.PdvOrigemID,
		NumeroExterno:  p.NumeroExterno,
		DataHora:       p.DataHora,
		Status:         p.Status,
		ValorTotal:     valorTotal,
		ValorDesconto:  valorDesconto,
		ValorAcrescimo: valorAcrescimo,
		IdempotencyKey: p.IdempotencyKey,
		ClienteNome: func() *string {
			if p.ClienteNome.Valid {
				return &p.ClienteNome.String
			}
			return nil
		}(),
		ClienteDocumento: func() *string {
			if p.ClienteDocumento.Valid {
				return &p.ClienteDocumento.String
			}
			return nil
		}(),
		ClienteEmail: func() *string {
			if p.ClienteEmail.Valid {
				return &p.ClienteEmail.String
			}
			return nil
		}(),
		Observacoes: func() *string {
			if p.Observacoes.Valid {
				return &p.Observacoes.String
			}
			return nil
		}(),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
