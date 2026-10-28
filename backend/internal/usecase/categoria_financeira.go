package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// CategoriaFinanceiraUsecase handles business logic for financial categories
type CategoriaFinanceiraUsecase struct {
	repo          domain.CategoriaFinanceiraRepository
	contaRepo     domain.ContaGerencialRepository
	auditoriaRepo domain.AuditoriaRepository
}

// NewCategoriaFinanceiraUsecase creates a new CategoriaFinanceiraUsecase instance
func NewCategoriaFinanceiraUsecase(
	repo domain.CategoriaFinanceiraRepository,
	contaRepo domain.ContaGerencialRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *CategoriaFinanceiraUsecase {
	return &CategoriaFinanceiraUsecase{
		repo:          repo,
		contaRepo:     contaRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

// Create creates a new financial category
func (u *CategoriaFinanceiraUsecase) Create(ctx context.Context, empresaID uuid.UUID, arg domain.CreateCategoriaFinanceiraInput) (*domain.CategoriaFinanceira, error) {
	// Validate category name uniqueness
	existing, err := u.repo.ListByEmpresa(ctx, empresaID, domain.CategoriaFinanceiraFilter{
		EmpresaID: &empresaID,
		Limit:     1,
	})
	if err == nil {
		for _, cat := range existing {
			if cat.Nome == arg.Nome {
				return nil, fmt.Errorf("categoria '%s' já existe para esta empresa", arg.Nome)
			}
		}
	}

	cf := &domain.CategoriaFinanceira{
		EmpresaID:        empresaID,
		Nome:             arg.Nome,
		Descricao:        arg.Descricao,
		ContaGerencialID: arg.ContaGerencialID,
		Ativo:            true,
	}

	// Validate conta gerencial if provided
	if arg.ContaGerencialID != nil {
		cg, err := u.contaRepo.GetByID(ctx, *arg.ContaGerencialID)
		if err != nil || cg.EmpresaID != empresaID || !cg.Ativo {
			return nil, errors.New("conta gerencial inválida ou inativa")
		}
	}

	if err := u.repo.Create(ctx, cf); err != nil {
		return nil, fmt.Errorf("failed to create categoria financeira: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "CategoriaFinanceira",
		EntidadeID: cf.ID,
		Acao:       "create",
	})

	return cf, nil
}

// GetByID retrieves a financial category by ID
func (u *CategoriaFinanceiraUsecase) GetByID(ctx context.Context, empresaID, id uuid.UUID) (*domain.CategoriaFinanceira, error) {
	cf, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get categoria financeira: %w", err)
	}

	if cf.EmpresaID != empresaID {
		return nil, errors.New("category not found")
	}

	return cf, nil
}

// List lists financial categories for an enterprise with optional filters
func (u *CategoriaFinanceiraUsecase) List(ctx context.Context, empresaID uuid.UUID, filter domain.CategoriaFinanceiraFilter) ([]*domain.CategoriaFinanceira, error) {
	filter.EmpresaID = &empresaID

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	items, err := u.repo.ListByEmpresa(ctx, empresaID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list categorias financeiras: %w", err)
	}

	return items, nil
}

// Update updates an existing financial category
func (u *CategoriaFinanceiraUsecase) Update(ctx context.Context, empresaID, id uuid.UUID, arg domain.UpdateCategoriaFinanceiraInput) (*domain.CategoriaFinanceira, error) {
	cf, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get categoria financeira: %w", err)
	}

	if cf.EmpresaID != empresaID {
		return nil, errors.New("category not found")
	}

	// Check if name is being changed and if it's unique
	if arg.Nome != nil && *arg.Nome != cf.Nome {
		existing, err := u.repo.ListByEmpresa(ctx, empresaID, domain.CategoriaFinanceiraFilter{
			EmpresaID: &empresaID,
			Limit:     100,
		})
		if err == nil {
			for _, cat := range existing {
				if cat.Nome == *arg.Nome && cat.ID != id {
					return nil, fmt.Errorf("categoria '%s' já existe para esta empresa", *arg.Nome)
				}
			}
		}
	}

	// Validate conta gerencial if provided or changed
	if arg.ContaGerencialID != nil {
		cg, err := u.contaRepo.GetByID(ctx, *arg.ContaGerencialID)
		if err != nil || cg.EmpresaID != empresaID || !cg.Ativo {
			return nil, errors.New("conta gerencial inválida ou inativa")
		}
	}

	// Apply updates
	if arg.Nome != nil {
		cf.Nome = *arg.Nome
	}
	if arg.Descricao != nil {
		cf.Descricao = arg.Descricao
	}
	if arg.ContaGerencialID != nil {
		cf.ContaGerencialID = arg.ContaGerencialID
	}
	if arg.Ativo != nil {
		cf.Ativo = *arg.Ativo
	}

	if err := u.repo.Update(ctx, cf); err != nil {
		return nil, fmt.Errorf("failed to update categoria financeira: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "CategoriaFinanceira",
		EntidadeID: id,
		Acao:       "update",
	})

	return cf, nil
}

// Delete soft-deletes a financial category
func (u *CategoriaFinanceiraUsecase) Delete(ctx context.Context, empresaID, id uuid.UUID) error {
	cf, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get categoria financeira: %w", err)
	}

	if cf.EmpresaID != empresaID {
		return errors.New("category not found")
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete categoria financeira: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "CategoriaFinanceira",
		EntidadeID: id,
		Acao:       "delete",
	})

	return nil
}
