package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// FluxoCaixaUsecase handles business logic for cash flow
type FluxoCaixaUsecase struct {
	movimentacaoRepo domain.MovimentacaoCaixaRepository
	contaRepo        domain.ContaGerencialRepository
	lojaRepo         domain.LojaRepository
}

// NewFluxoCaixaUsecase creates a new FluxoCaixaUsecase instance
func NewFluxoCaixaUsecase(
	movimentacaoRepo domain.MovimentacaoCaixaRepository,
	contaRepo domain.ContaGerencialRepository,
	lojaRepo domain.LojaRepository,
) *FluxoCaixaUsecase {
	return &FluxoCaixaUsecase{
		movimentacaoRepo: movimentacaoRepo,
		contaRepo:        contaRepo,
		lojaRepo:         lojaRepo,
	}
}

// FluxoCaixaRequest represents the request parameters for cash flow query
type FluxoCaixaRequest struct {
	EmpresaID  uuid.UUID  `json:"empresa_id"`
	LojaID     *uuid.UUID `json:"loja_id,omitempty"`
	DataInicio time.Time  `json:"data_inicio"`
	DataFim    time.Time  `json:"data_fim"`
}

// GetFluxoCaixa generates the cash flow statement for a given period
func (u *FluxoCaixaUsecase) GetFluxoCaixa(ctx context.Context, req FluxoCaixaRequest) (*domain.FluxoCaixa, error) {
	// Validate dates
	if !req.DataInicio.Before(req.DataFim) {
		return nil, fmt.Errorf("data_inicio must be before data_fim")
	}

	// Validate loja if provided
	if req.LojaID != nil {
		loja, err := u.lojaRepo.GetByID(ctx, *req.LojaID)
		if err != nil || loja.EmpresaID != req.EmpresaID || !loja.Ativo {
			return nil, fmt.Errorf("loja not found or inactive")
		}
	}

	// Get balance from before the period (initial balance)
	saldoInicial, err := u.movimentacaoRepo.GetSaldoPeriodo(ctx, req.EmpresaID, req.LojaID, time.Time{}, req.DataInicio.Add(-24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to get saldo inicial: %w", err)
	}

	// Get movements for the period
	filter := domain.MovimentacaoCaixaFilter{
		EmpresaID:  &req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: &req.DataInicio,
		DataFim:    &req.DataFim,
		Limit:      10000,
	}

	allMovements, err := u.movimentacaoRepo.ListByEmpresa(ctx, req.EmpresaID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list movements: %w", err)
	}

	// Separate entries and exits
	entradas := make([]domain.FluxoCaixaLinha, 0)
	saidas := make([]domain.FluxoCaixaLinha, 0)

	// Group by account using map with uuid.UUID key
	entradaByAccount := make(map[uuid.UUID]float64)
	saidaByAccount := make(map[uuid.UUID]float64)

	for _, m := range allMovements {
		linha := domain.FluxoCaixaLinha{
			Nome:  m.Descricao,
			Valor: m.Valor,
		}

		if m.ContaGerencialID != nil {
			cg, err := u.contaRepo.GetByID(ctx, *m.ContaGerencialID)
			if err == nil {
				linha.ContaGerencialID = cg.ID
				linha.Codigo = cg.Codigo
				linha.Nome = cg.Nome
			}
		}

		if m.Tipo == domain.MovimentacaoCaixaTipoEntrada {
			if m.ContaGerencialID != nil {
				entradaByAccount[*m.ContaGerencialID] += m.Valor
			} else {
				entradas = append(entradas, linha)
			}
		} else {
			if m.ContaGerencialID != nil {
				saidaByAccount[*m.ContaGerencialID] += m.Valor
			} else {
				saidas = append(saidas, linha)
			}
		}
	}

	// Add grouped entries
	for accountID, total := range entradaByAccount {
		cg, err := u.contaRepo.GetByID(ctx, accountID)
		if err == nil {
			entradas = append(entradas, domain.FluxoCaixaLinha{
				ContaGerencialID: cg.ID,
				Codigo:           cg.Codigo,
				Nome:             cg.Nome,
				Valor:            total,
			})
		} else {
			entradas = append(entradas, domain.FluxoCaixaLinha{
				ContaGerencialID: accountID,
				Valor:            total,
			})
		}
	}

	// Add grouped exits
	for accountID, total := range saidaByAccount {
		cg, err := u.contaRepo.GetByID(ctx, accountID)
		if err == nil {
			saidas = append(saidas, domain.FluxoCaixaLinha{
				ContaGerencialID: cg.ID,
				Codigo:           cg.Codigo,
				Nome:             cg.Nome,
				Valor:            total,
			})
		} else {
			saidas = append(saidas, domain.FluxoCaixaLinha{
				ContaGerencialID: accountID,
				Valor:            total,
			})
		}
	}

	// Calculate totals
	totalEntradas := 0.0
	for _, e := range entradas {
		totalEntradas += e.Valor
	}

	totalSaidas := 0.0
	for _, s := range saidas {
		totalSaidas += s.Valor
	}

	saldoFinal := saldoInicial + totalEntradas - totalSaidas

	return &domain.FluxoCaixa{
		EmpresaID:     req.EmpresaID,
		LojaID:        req.LojaID,
		DataInicio:    req.DataInicio,
		DataFim:       req.DataFim,
		SaldoInicial:  saldoInicial,
		Entradas:      entradas,
		Saidas:        saidas,
		TotalEntradas: totalEntradas,
		TotalSaidas:   totalSaidas,
		SaldoFinal:    saldoFinal,
	}, nil
}

// GetCashBalance gets the current cash balance up to a given date
func (u *FluxoCaixaUsecase) GetCashBalance(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, até time.Time) (float64, error) {
	balance, err := u.movimentacaoRepo.GetSaldoPeriodo(ctx, empresaID, lojaID, time.Time{}, até)
	if err != nil {
		return 0, fmt.Errorf("failed to get cash balance: %w", err)
	}

	return balance, nil
}
