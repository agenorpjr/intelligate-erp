package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SalesTxRepositoryInterface defines the interface for transactional sales operations
type SalesTxRepositoryInterface interface {
	WithTx(ctx context.Context, fn func(*SalesTxQueries) error) error
}

// SalesTxRepository provides transactional operations for sales ingestion
type SalesTxRepository struct {
	pool *pgxpool.Pool
}

func NewSalesTxRepository(pool *pgxpool.Pool) *SalesTxRepository {
	return &SalesTxRepository{pool: pool}
}

// WithTx executes a function within a database transaction
func (r *SalesTxRepository) WithTx(ctx context.Context, fn func(*SalesTxQueries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	q := New(tx)
	txQueries := &SalesTxQueries{
		PedidoRepo:    NewPedidoCanonicoRepositoryFromQueries(q),
		ItemRepo:      NewPedidoItemCanonicoRepositoryFromQueries(q),
		PagamentoRepo: NewPedidoPagamentoCanonicoRepositoryFromQueries(q),
		EventoRepo:    NewEventoIngestaoRepositoryFromQueries(q),
	}

	err = fn(txQueries)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}

// SalesTxQueries holds repositories that use a transaction
// Uses interfaces so that mocks can be used in tests
type SalesTxQueries struct {
	PedidoRepo    TxPedidoCanonicoRepository
	ItemRepo      TxPedidoItemCanonicoRepository
	PagamentoRepo TxPedidoPagamentoCanonicoRepository
	EventoRepo    TxEventoIngestaoRepository
}

// Constructor functions to create repositories from Queries (for transaction support)
func NewPedidoCanonicoRepositoryFromQueries(q *Queries) TxPedidoCanonicoRepository {
	return &PedidoCanonicoRepository{querier: q}
}

func NewPedidoItemCanonicoRepositoryFromQueries(q *Queries) TxPedidoItemCanonicoRepository {
	return &PedidoItemCanonicoRepository{querier: q}
}

func NewPedidoPagamentoCanonicoRepositoryFromQueries(q *Queries) TxPedidoPagamentoCanonicoRepository {
	return &PedidoPagamentoCanonicoRepository{querier: q}
}

func NewEventoIngestaoRepositoryFromQueries(q *Queries) TxEventoIngestaoRepository {
	return &EventoIngestaoRepository{querier: q}
}

// Domain repository interfaces for transaction support
type TxPedidoCanonicoRepository interface {
	Create(ctx context.Context, p *domain.PedidoCanonico) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoCanonico, error)
	GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.PedidoCanonico, error)
	Update(ctx context.Context, p *domain.PedidoCanonico) error
}

type TxPedidoItemCanonicoRepository interface {
	Create(ctx context.Context, i *domain.PedidoItemCanonico) error
	DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error
}

type TxPedidoPagamentoCanonicoRepository interface {
	Create(ctx context.Context, p *domain.PedidoPagamentoCanonico) error
	DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error
}

type TxEventoIngestaoRepository interface {
	Create(ctx context.Context, e *domain.EventoIngestao) error
	GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.EventoIngestao, error)
}
