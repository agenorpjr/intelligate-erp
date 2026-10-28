package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// MovimentacaoCaixaUsecase handles business logic for cash movements
type MovimentacaoCaixaUsecase struct {
	repo          domain.MovimentacaoCaixaRepository
	lojaRepo      domain.LojaRepository
	categoriaRepo domain.CategoriaFinanceiraRepository
	contaRepo     domain.ContaGerencialRepository
	auditoriaRepo domain.AuditoriaRepository
}

// NewMovimentacaoCaixaUsecase creates a new MovimentacaoCaixaUsecase instance
func NewMovimentacaoCaixaUsecase(
	repo domain.MovimentacaoCaixaRepository,
	lojaRepo domain.LojaRepository,
	categoriaRepo domain.CategoriaFinanceiraRepository,
	contaRepo domain.ContaGerencialRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *MovimentacaoCaixaUsecase {
	return &MovimentacaoCaixaUsecase{
		repo:          repo,
		lojaRepo:      lojaRepo,
		categoriaRepo: categoriaRepo,
		contaRepo:     contaRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

// Create creates a new cash movement
func (u *MovimentacaoCaixaUsecase) Create(ctx context.Context, empresaID uuid.UUID, arg domain.CreateMovimentacaoCaixaInput) (*domain.MovimentacaoCaixa, error) {
	// Validate empresa_id matches
	if arg.EmpresaID != empresaID {
		return nil, errors.New("empresa_id mismatch")
	}

	// Validate loja exists and belongs to empresa
	loja, err := u.lojaRepo.GetByID(ctx, arg.LojaID)
	if err != nil || loja.EmpresaID != empresaID || !loja.Ativo {
		return nil, errors.New("loja inválida ou inativa")
	}

	// Validate valor is positive
	if arg.Valor <= 0 {
		return nil, errors.New("valor deve ser positivo")
	}

	// Validate data is not in the future
	if arg.Data.After(time.Now()) {
		return nil, errors.New("data não pode ser no futuro")
	}

	// Validate category if provided
	if arg.CategoriaFinanceiraID != nil {
		cat, err := u.categoriaRepo.GetByID(ctx, *arg.CategoriaFinanceiraID)
		if err != nil || cat.EmpresaID != empresaID || !cat.Ativo {
			return nil, errors.New("categoria financeira inválida ou inativa")
		}

		// If category has a linked managerial account, use it if not explicitly set
		if arg.ContaGerencialID == nil && cat.ContaGerencialID != nil {
			arg.ContaGerencialID = cat.ContaGerencialID
		}
	}

	// Validate conta gerencial if explicitly provided
	if arg.ContaGerencialID != nil {
		cg, err := u.contaRepo.GetByID(ctx, *arg.ContaGerencialID)
		if err != nil || cg.EmpresaID != empresaID || !cg.Ativo {
			return nil, errors.New("conta gerencial inválida ou inativa")
		}
	}

	mc := &domain.MovimentacaoCaixa{
		EmpresaID:             empresaID,
		LojaID:                arg.LojaID,
		Data:                  arg.Data,
		Tipo:                  arg.Tipo,
		Valor:                 arg.Valor,
		Descricao:             arg.Descricao,
		CategoriaFinanceiraID: arg.CategoriaFinanceiraID,
		ContaGerencialID:      arg.ContaGerencialID,
		Origem:                arg.Origem,
		OrigemID:              arg.OrigemID,
		Observacoes:           arg.Observacoes,
	}

	if err := u.repo.Create(ctx, mc); err != nil {
		return nil, fmt.Errorf("failed to create movimentacao caixa: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "MovimentacaoCaixa",
		EntidadeID: mc.ID,
		Acao:       "create",
	})

	return mc, nil
}

// GetByID retrieves a cash movement by ID
func (u *MovimentacaoCaixaUsecase) GetByID(ctx context.Context, empresaID, id uuid.UUID) (*domain.MovimentacaoCaixa, error) {
	mc, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get movimentacao caixa: %w", err)
	}

	if mc.EmpresaID != empresaID {
		return nil, errors.New("movement not found")
	}

	return mc, nil
}

// List lists cash movements for an enterprise with optional filters
func (u *MovimentacaoCaixaUsecase) List(ctx context.Context, empresaID uuid.UUID, filter domain.MovimentacaoCaixaFilter) ([]*domain.MovimentacaoCaixa, error) {
	filter.EmpresaID = &empresaID

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	items, err := u.repo.ListByEmpresa(ctx, empresaID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list movimentacoes caixa: %w", err)
	}

	return items, nil
}

// ListByLoja lists cash movements for a specific store
func (u *MovimentacaoCaixaUsecase) ListByLoja(ctx context.Context, empresaID, lojaID uuid.UUID, filter domain.MovimentacaoCaixaFilter) ([]*domain.MovimentacaoCaixa, error) {
	// Validate loja belongs to empresa
	loja, err := u.lojaRepo.GetByID(ctx, lojaID)
	if err != nil || loja.EmpresaID != empresaID {
		return nil, errors.New("loja not found")
	}

	filter.EmpresaID = &empresaID
	filter.LojaID = &lojaID

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	items, err := u.repo.ListByLoja(ctx, lojaID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list movimentacoes caixa by loja: %w", err)
	}

	return items, nil
}

// Update updates an existing cash movement
func (u *MovimentacaoCaixaUsecase) Update(ctx context.Context, empresaID, id uuid.UUID, arg domain.UpdateMovimentacaoCaixaInput) (*domain.MovimentacaoCaixa, error) {
	mc, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get movimentacao caixa: %w", err)
	}

	if mc.EmpresaID != empresaID {
		return nil, errors.New("movement not found")
	}

	// Apply updates
	if arg.Data != nil {
		mc.Data = *arg.Data
	}
	if arg.Tipo != nil {
		mc.Tipo = *arg.Tipo
	}
	if arg.Valor != nil {
		if *arg.Valor <= 0 {
			return nil, errors.New("valor deve ser positivo")
		}
		mc.Valor = *arg.Valor
	}
	if arg.Descricao != nil {
		mc.Descricao = *arg.Descricao
	}
	if arg.CategoriaFinanceiraID != nil {
		cat, err := u.categoriaRepo.GetByID(ctx, *arg.CategoriaFinanceiraID)
		if err != nil || cat.EmpresaID != empresaID || !cat.Ativo {
			return nil, errors.New("categoria financeira inválida ou inativa")
		}
		mc.CategoriaFinanceiraID = arg.CategoriaFinanceiraID

		// If category has a linked managerial account, use it if not explicitly set
		if mc.ContaGerencialID == nil && cat.ContaGerencialID != nil {
			mc.ContaGerencialID = cat.ContaGerencialID
		}
	}
	if arg.ContaGerencialID != nil {
		cg, err := u.contaRepo.GetByID(ctx, *arg.ContaGerencialID)
		if err != nil || cg.EmpresaID != empresaID || !cg.Ativo {
			return nil, errors.New("conta gerencial inválida ou inativa")
		}
		mc.ContaGerencialID = arg.ContaGerencialID
	}
	if arg.Origem != nil {
		mc.Origem = arg.Origem
	}
	if arg.OrigemID != nil {
		mc.OrigemID = arg.OrigemID
	}
	if arg.Observacoes != nil {
		mc.Observacoes = arg.Observacoes
	}

	if err := u.repo.Update(ctx, mc); err != nil {
		return nil, fmt.Errorf("failed to update movimentacao caixa: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "MovimentacaoCaixa",
		EntidadeID: id,
		Acao:       "update",
	})

	return mc, nil
}

// Delete soft-deletes a cash movement
func (u *MovimentacaoCaixaUsecase) Delete(ctx context.Context, empresaID, id uuid.UUID) error {
	mc, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get movimentacao caixa: %w", err)
	}

	if mc.EmpresaID != empresaID {
		return errors.New("movement not found")
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete movimentacao caixa: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "MovimentacaoCaixa",
		EntidadeID: id,
		Acao:       "delete",
	})

	return nil
}

// GetSaldoPeriodo gets the balance for a period
func (u *MovimentacaoCaixaUsecase) GetSaldoPeriodo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (float64, error) {
	saldo, err := u.repo.GetSaldoPeriodo(ctx, empresaID, lojaID, dataInicio, dataFim)
	if err != nil {
		return 0, fmt.Errorf("failed to get saldo periodo: %w", err)
	}

	return saldo, nil
}

// GetTotaisPorContaGerencial gets totals grouped by managerial account
func (u *MovimentacaoCaixaUsecase) GetTotaisPorContaGerencial(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (map[uuid.UUID]float64, error) {
	totals, err := u.repo.GetTotaisPorContaGerencial(ctx, empresaID, lojaID, dataInicio, dataFim)
	if err != nil {
		return nil, fmt.Errorf("failed to get totais por conta gerencial: %w", err)
	}

	return totals, nil
}

// GetTotaisPorCategoriaFinanceira gets totals grouped by financial category
func (u *MovimentacaoCaixaUsecase) GetTotaisPorCategoriaFinanceira(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (map[uuid.UUID]float64, error) {
	totals, err := u.repo.GetTotaisPorCategoriaFinanceira(ctx, empresaID, lojaID, dataInicio, dataFim)
	if err != nil {
		return nil, fmt.Errorf("failed to get totais por categoria financeira: %w", err)
	}

	return totals, nil
}
