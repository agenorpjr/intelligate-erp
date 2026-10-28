package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EstoqueRepository implementa domain.EstoqueRepository sobre PostgreSQL.
type EstoqueRepository struct {
	pool    *pgxpool.Pool
	querier *Queries
}

// NewEstoqueRepository cria uma nova instância de EstoqueRepository.
func NewEstoqueRepository(pool *pgxpool.Pool) *EstoqueRepository {
	return &EstoqueRepository{pool: pool, querier: New(pool)}
}

// GetSaldo retorna o saldo de um insumo em uma loja. Retorna (nil, nil) quando
// ainda não há saldo registrado para o par insumo/loja.
func (r *EstoqueRepository) GetSaldo(ctx context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	row, err := r.querier.GetEstoqueSaldo(ctx, GetEstoqueSaldoParams{
		InsumoID: insumoID,
		LojaID:   lojaID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.SaldoEstoque{
		ID:                 row.ID,
		EmpresaID:          row.EmpresaID,
		LojaID:             row.LojaID,
		InsumoID:           row.InsumoID,
		Quantidade:         row.Quantidade,
		UnidadeMedida:      row.UnidadeMedida,
		CustoMedioUnitario: row.CustoMedioUnitario,
		PermiteNegativo:    row.PermiteNegativo,
		CreatedAt:          row.CreatedAt,
		AtualizadoEm:       row.UpdatedAt,
	}, nil
}

// ListSaldos lista os saldos da empresa, opcionalmente filtrados por loja/insumo.
func (r *EstoqueRepository) ListSaldos(ctx context.Context, filter domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	rows, err := r.querier.ListEstoquesSaldos(ctx, ListEstoquesSaldosParams{
		EmpresaID:   filter.EmpresaID,
		LojaID:      pgUUIDFromPtr(filter.LojaID),
		InsumoID:    pgUUIDFromPtr(filter.InsumoID),
		LimitCount:  int32(filter.Limit),
		OffsetCount: int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	saldos := make([]*domain.SaldoEstoque, len(rows))
	for i, row := range rows {
		saldos[i] = &domain.SaldoEstoque{
			ID:                 row.ID,
			EmpresaID:          row.EmpresaID,
			LojaID:             row.LojaID,
			InsumoID:           row.InsumoID,
			Quantidade:         row.Quantidade,
			UnidadeMedida:      row.UnidadeMedida,
			CustoMedioUnitario: row.CustoMedioUnitario,
			PermiteNegativo:    row.PermiteNegativo,
			CreatedAt:          row.CreatedAt,
			AtualizadoEm:       row.UpdatedAt,
		}
	}
	return saldos, nil
}

// ListMovimentacoes lista as movimentações de estoque com filtros.
func (r *EstoqueRepository) ListMovimentacoes(ctx context.Context, filter domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	var tipo pgtype.Text
	if filter.Tipo != nil {
		tipo = pgtype.Text{String: string(*filter.Tipo), Valid: true}
	}

	rows, err := r.querier.ListMovimentacoesEstoque(ctx, ListMovimentacoesEstoqueParams{
		EmpresaID:   filter.EmpresaID,
		LojaID:      pgUUIDFromPtr(filter.LojaID),
		InsumoID:    pgUUIDFromPtr(filter.InsumoID),
		Tipo:        tipo,
		Origem:      pgTextFromPtr(filter.Origem),
		DataInicio:  timestamptzFromPtr(filter.DataInicio),
		DataFim:     timestamptzFromPtr(filter.DataFim),
		LimitCount:  int32(filter.Limit),
		OffsetCount: int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	movimentacoes := make([]*domain.MovimentacaoEstoque, len(rows))
	for i, row := range rows {
		movimentacoes[i] = &domain.MovimentacaoEstoque{
			ID:                row.ID,
			EmpresaID:         row.EmpresaID,
			LojaID:            row.LojaID,
			InsumoID:          row.InsumoID,
			Tipo:              domain.TipoMovimentacaoEstoque(row.Tipo),
			Quantidade:        row.Quantidade,
			UnidadeMedida:     row.UnidadeMedida,
			CustoUnitario:     row.CustoUnitario,
			CustoTotal:        row.CustoTotal,
			SaldoAnterior:     row.SaldoAnterior,
			SaldoPosterior:    row.SaldoPosterior,
			Motivo:            textPtrFromPg(row.Motivo),
			Observacao:        textPtrFromPg(row.Observacao),
			Origem:            row.Origem,
			ReferenciaExterna: textPtrFromPg(row.ReferenciaExterna),
			UsuarioID:         uuidPtrFromPg(row.UsuarioID),
			CreatedAt:         row.CreatedAt,
		}
	}
	return movimentacoes, nil
}

// GetResumoPerdas consolida perdas e desperdícios por período.
func (r *EstoqueRepository) GetResumoPerdas(ctx context.Context, filter domain.ResumoPerdasFilter) (*domain.ResumoPerdas, error) {
	rows, err := r.querier.GetResumoPerdasEstoque(ctx, GetResumoPerdasEstoqueParams{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
	})
	if err != nil {
		return nil, err
	}

	resumo := &domain.ResumoPerdas{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Itens:      make([]domain.ResumoPerdasItem, 0, len(rows)),
	}

	for _, row := range rows {
		tipo := domain.TipoMovimentacaoEstoque(row.Tipo)
		resumo.Itens = append(resumo.Itens, domain.ResumoPerdasItem{
			Tipo:          tipo,
			InsumoID:      row.InsumoID,
			InsumoNome:    row.InsumoNome,
			UnidadeMedida: row.UnidadeMedida,
			Quantidade:    row.QuantidadeTotal,
			CustoTotal:    row.CustoTotal,
		})
		resumo.QuantidadeTotal += row.QuantidadeTotal
		resumo.CustoTotal += row.CustoTotal
		switch tipo {
		case domain.MovimentacaoPerda:
			resumo.CustoTotalPerdas += row.CustoTotal
		case domain.MovimentacaoDesperdicio:
			resumo.CustoTotalDesperdicio += row.CustoTotal
		}
	}

	return resumo, nil
}

// WithTx executa a função informada dentro de uma transação, fornecendo as
// operações de estoque vinculadas à transação.
func (r *EstoqueRepository) WithTx(ctx context.Context, fn func(tx domain.EstoqueTxRepository) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(&estoqueTxRepository{querier: New(tx)}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// estoqueTxRepository agrupa operações de estoque dentro de uma transação.
type estoqueTxRepository struct {
	querier *Queries
}

// GetSaldoForUpdate bloqueia a linha do saldo para atualização concorrente.
// Retorna (nil, nil) quando não existe saldo registrado.
func (r *estoqueTxRepository) GetSaldoForUpdate(ctx context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	row, err := r.querier.GetEstoqueSaldoForUpdate(ctx, GetEstoqueSaldoForUpdateParams{
		InsumoID: insumoID,
		LojaID:   lojaID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.SaldoEstoque{
		ID:                 row.ID,
		EmpresaID:          row.EmpresaID,
		LojaID:             row.LojaID,
		InsumoID:           row.InsumoID,
		Quantidade:         row.Quantidade,
		UnidadeMedida:      row.UnidadeMedida,
		CustoMedioUnitario: row.CustoMedioUnitario,
		PermiteNegativo:    row.PermiteNegativo,
		CreatedAt:          row.CreatedAt,
		AtualizadoEm:       row.UpdatedAt,
	}, nil
}

// UpsertSaldo insere ou atualiza o saldo do par insumo/loja.
func (r *estoqueTxRepository) UpsertSaldo(ctx context.Context, saldo *domain.SaldoEstoque) (*domain.SaldoEstoque, error) {
	row, err := r.querier.UpsertEstoqueSaldo(ctx, UpsertEstoqueSaldoParams{
		EmpresaID:          saldo.EmpresaID,
		LojaID:             saldo.LojaID,
		InsumoID:           saldo.InsumoID,
		Quantidade:         numericFromFloat(saldo.Quantidade),
		UnidadeMedida:      saldo.UnidadeMedida,
		CustoMedioUnitario: numericFromFloat(saldo.CustoMedioUnitario),
		PermiteNegativo:    saldo.PermiteNegativo,
	})
	if err != nil {
		return nil, err
	}
	return &domain.SaldoEstoque{
		ID:                 row.ID,
		EmpresaID:          row.EmpresaID,
		LojaID:             row.LojaID,
		InsumoID:           row.InsumoID,
		Quantidade:         row.Quantidade,
		UnidadeMedida:      row.UnidadeMedida,
		CustoMedioUnitario: row.CustoMedioUnitario,
		PermiteNegativo:    row.PermiteNegativo,
		CreatedAt:          row.CreatedAt,
		AtualizadoEm:       row.UpdatedAt,
	}, nil
}

// CreateMovimentacao registra uma movimentação de estoque.
func (r *estoqueTxRepository) CreateMovimentacao(ctx context.Context, mov *domain.MovimentacaoEstoque) (*domain.MovimentacaoEstoque, error) {
	row, err := r.querier.CreateMovimentacaoEstoque(ctx, CreateMovimentacaoEstoqueParams{
		EmpresaID:         mov.EmpresaID,
		LojaID:            mov.LojaID,
		InsumoID:          mov.InsumoID,
		Tipo:              string(mov.Tipo),
		Quantidade:        numericFromFloat(mov.Quantidade),
		UnidadeMedida:     mov.UnidadeMedida,
		CustoUnitario:     numericFromFloat(mov.CustoUnitario),
		CustoTotal:        numericFromFloat(mov.CustoTotal),
		SaldoAnterior:     numericFromFloat(mov.SaldoAnterior),
		SaldoPosterior:    numericFromFloat(mov.SaldoPosterior),
		Motivo:            pgTextFromPtr(mov.Motivo),
		Observacao:        pgTextFromPtr(mov.Observacao),
		Origem:            mov.Origem,
		ReferenciaExterna: pgTextFromPtr(mov.ReferenciaExterna),
		UsuarioID:         pgUUIDFromPtr(mov.UsuarioID),
	})
	if err != nil {
		return nil, err
	}
	return &domain.MovimentacaoEstoque{
		ID:                row.ID,
		EmpresaID:         row.EmpresaID,
		LojaID:            row.LojaID,
		InsumoID:          row.InsumoID,
		Tipo:              domain.TipoMovimentacaoEstoque(row.Tipo),
		Quantidade:        row.Quantidade,
		UnidadeMedida:     row.UnidadeMedida,
		CustoUnitario:     row.CustoUnitario,
		CustoTotal:        row.CustoTotal,
		SaldoAnterior:     row.SaldoAnterior,
		SaldoPosterior:    row.SaldoPosterior,
		Motivo:            textPtrFromPg(row.Motivo),
		Observacao:        textPtrFromPg(row.Observacao),
		Origem:            row.Origem,
		ReferenciaExterna: textPtrFromPg(row.ReferenciaExterna),
		UsuarioID:         uuidPtrFromPg(row.UsuarioID),
		CreatedAt:         row.CreatedAt,
	}, nil
}

// ============================================================
// Helpers
// ============================================================

func pgUUIDFromPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func uuidPtrFromPg(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func pgTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func textPtrFromPg(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	v := t.String
	return &v
}

func timestamptzFromPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
