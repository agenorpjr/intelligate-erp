package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// DREGerencialUsecase handles business logic for managerial income statement (DRE)
type DREGerencialUsecase struct {
	movimentacaoRepo domain.MovimentacaoCaixaRepository
	contaRepo        domain.ContaGerencialRepository
}

// NewDREGerencialUsecase creates a new DREGerencialUsecase instance
func NewDREGerencialUsecase(
	movimentacaoRepo domain.MovimentacaoCaixaRepository,
	contaRepo domain.ContaGerencialRepository,
) *DREGerencialUsecase {
	return &DREGerencialUsecase{
		movimentacaoRepo: movimentacaoRepo,
		contaRepo:        contaRepo,
	}
}

// DRERequest represents the request parameters for DRE query
type DRERequest struct {
	EmpresaID  uuid.UUID  `json:"empresa_id"`
	LojaID     *uuid.UUID `json:"loja_id,omitempty"`
	DataInicio time.Time  `json:"data_inicio"`
	DataFim    time.Time  `json:"data_fim"`
}

// GetDRE generates the managerial income statement (DRE) for a given period
func (u *DREGerencialUsecase) GetDRE(ctx context.Context, req DRERequest) (*domain.DREGerencial, error) {
	// Validate dates
	if !req.DataInicio.Before(req.DataFim) {
		return nil, fmt.Errorf("data_inicio must be before data_fim")
	}

	// Get all accounts by type
	allContas, err := u.contaRepo.ListByEmpresa(ctx, req.EmpresaID, domain.ContaGerencialFilter{
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list contas gerenciais: %w", err)
	}

	// Group accounts by type
	receitas := make([]domain.DRELinha, 0)
	custos := make([]domain.DRELinha, 0)
	despesas := make([]domain.DRELinha, 0)
	investimentos := make([]domain.DRELinha, 0)

	for _, cg := range allContas {
		drel := domain.DRELinha{
			ContaGerencialID: cg.ID,
			Codigo:           cg.Codigo,
			Nome:             cg.Nome,
			Tipo:             cg.Tipo,
			Valor:            0, // Will be calculated below
		}

		switch cg.Tipo {
		case domain.ContaGerencialTipoReceita:
			receitas = append(receitas, drel)
		case domain.ContaGerencialTipoCusto:
			custos = append(custos, drel)
		case domain.ContaGerencialTipoDespesa:
			despesas = append(despesas, drel)
		case domain.ContaGerencialTipoInvestimento:
			investimentos = append(investimentos, drel)
		}
	}

	// Get totals by account
	totals, err := u.movimentacaoRepo.GetTotaisPorContaGerencial(ctx, req.EmpresaID, req.LojaID, req.DataInicio, req.DataFim)
	if err != nil {
		return nil, fmt.Errorf("failed to get totais por conta gerencial: %w", err)
	}

	// Calculate values for each line
	for i, r := range receitas {
		receitas[i].Valor = totals[r.ContaGerencialID]
	}
	for i, c := range custos {
		custos[i].Valor = totals[c.ContaGerencialID]
	}
	for i, d := range despesas {
		despesas[i].Valor = totals[d.ContaGerencialID]
	}
	for i, inv := range investimentos {
		investimentos[i].Valor = totals[inv.ContaGerencialID]
	}

	// Calculate totals
	totalReceitas := sumValues(receitas)
	totalCustos := sumValues(custos)
	totalDespesas := sumValues(despesas)
	totalInvestimentos := sumValues(investimentos)

	resultadoBruto := totalReceitas - totalCustos
	resultadoOperacional := resultadoBruto - totalDespesas
	resultadoLiquido := resultadoOperacional - totalInvestimentos

	return &domain.DREGerencial{
		EmpresaID:            req.EmpresaID,
		LojaID:               req.LojaID,
		DataInicio:           req.DataInicio,
		DataFim:              req.DataFim,
		Receitas:             receitas,
		Custos:               custos,
		Despesas:             despesas,
		Investimentos:        investimentos,
		TotalReceitas:        totalReceitas,
		TotalCustos:          totalCustos,
		TotalDespesas:        totalDespesas,
		TotalInvestimentos:   totalInvestimentos,
		ResultadoBruto:       resultadoBruto,
		ResultadoOperacional: resultadoOperacional,
		ResultadoLiquido:     resultadoLiquido,
	}, nil
}

// GetDREByType gets DRE lines for a specific type
func (u *DREGerencialUsecase) GetDREByType(ctx context.Context, req DRERequest, tipo domain.ContaGerencialTipo) ([]domain.DRELinha, error) {
	allContas, err := u.contaRepo.ListByEmpresa(ctx, req.EmpresaID, domain.ContaGerencialFilter{
		Tipo:  &tipo,
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list contas gerenciais: %w", err)
	}

	totals, err := u.movimentacaoRepo.GetTotaisPorContaGerencial(ctx, req.EmpresaID, req.LojaID, req.DataInicio, req.DataFim)
	if err != nil {
		return nil, fmt.Errorf("failed to get totais por conta gerencial: %w", err)
	}

	lines := make([]domain.DRELinha, 0, len(allContas))
	for _, cg := range allContas {
		lines = append(lines, domain.DRELinha{
			ContaGerencialID: cg.ID,
			Codigo:           cg.Codigo,
			Nome:             cg.Nome,
			Tipo:             cg.Tipo,
			Valor:            totals[cg.ID],
		})
	}

	return lines, nil
}

// sumValues calculates the sum of all values in a slice of DRELinha
func sumValues(lines []domain.DRELinha) float64 {
	total := 0.0
	for _, l := range lines {
		total += l.Valor
	}
	return total
}
