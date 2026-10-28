package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// ContaGerencialUsecase handles business logic for managerial accounts
type ContaGerencialUsecase struct {
	repo          domain.ContaGerencialRepository
	auditoriaRepo domain.AuditoriaRepository
}

// NewContaGerencialUsecase creates a new ContaGerencialUsecase instance
func NewContaGerencialUsecase(
	repo domain.ContaGerencialRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *ContaGerencialUsecase {
	return &ContaGerencialUsecase{
		repo:          repo,
		auditoriaRepo: auditoriaRepo,
	}
}

// Create creates a new managerial account
func (u *ContaGerencialUsecase) Create(ctx context.Context, empresaID uuid.UUID, arg domain.CreateContaGerencialInput) (*domain.ContaGerencial, error) {
	// Validate empresa_id matches
	if arg.EmpresaID != empresaID {
		return nil, errors.New("empresa_id mismatch")
	}

	// Validate code uniqueness
	existing, err := u.repo.GetByCodigo(ctx, empresaID, arg.Codigo)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("código '%s' já existe para esta empresa", arg.Codigo)
	}

	cg := &domain.ContaGerencial{
		EmpresaID:  empresaID,
		Codigo:     arg.Codigo,
		Nome:       arg.Nome,
		Tipo:       arg.Tipo,
		Grupo:      arg.Grupo,
		ContaPaiID: arg.ContaPaiID,
		Ativo:      true,
	}

	if err := u.repo.Create(ctx, cg); err != nil {
		return nil, fmt.Errorf("failed to create conta gerencial: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "ContaGerencial",
		EntidadeID: cg.ID,
		Acao:       "create",
	})

	return cg, nil
}

// GetByID retrieves a managerial account by ID
func (u *ContaGerencialUsecase) GetByID(ctx context.Context, empresaID, id uuid.UUID) (*domain.ContaGerencial, error) {
	cg, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conta gerencial: %w", err)
	}

	if cg.EmpresaID != empresaID {
		return nil, errors.New("account not found")
	}

	return cg, nil
}

// GetByCodigo retrieves a managerial account by code
func (u *ContaGerencialUsecase) GetByCodigo(ctx context.Context, empresaID uuid.UUID, codigo string) (*domain.ContaGerencial, error) {
	cg, err := u.repo.GetByCodigo(ctx, empresaID, codigo)
	if err != nil {
		return nil, fmt.Errorf("failed to get conta gerencial by code: %w", err)
	}

	return cg, nil
}

// List lists managerial accounts for an enterprise with optional filters
func (u *ContaGerencialUsecase) List(ctx context.Context, empresaID uuid.UUID, filter domain.ContaGerencialFilter) ([]*domain.ContaGerencial, error) {
	filter.EmpresaID = &empresaID

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	items, err := u.repo.ListByEmpresa(ctx, empresaID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list contas gerenciais: %w", err)
	}

	return items, nil
}

// ListWithHierarchy lists all accounts for building hierarchical tree
func (u *ContaGerencialUsecase) ListWithHierarchy(ctx context.Context, empresaID uuid.UUID) ([]*domain.ContaGerencial, error) {
	items, err := u.repo.ListByEmpresaHierarquia(ctx, empresaID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contas gerenciais hierarchy: %w", err)
	}

	// Build hierarchy tree
	rootNodes := buildContaHierarchy(items)

	return rootNodes, nil
}

// Update updates an existing managerial account
func (u *ContaGerencialUsecase) Update(ctx context.Context, empresaID, id uuid.UUID, arg domain.UpdateContaGerencialInput) (*domain.ContaGerencial, error) {
	cg, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conta gerencial: %w", err)
	}

	if cg.EmpresaID != empresaID {
		return nil, errors.New("account not found")
	}

	// Check if code is being changed and if it's unique
	if arg.Codigo != nil && *arg.Codigo != cg.Codigo {
		existing, err := u.repo.GetByCodigo(ctx, empresaID, *arg.Codigo)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("código '%s' já existe para esta empresa", *arg.Codigo)
		}
	}

	// Apply updates
	if arg.Codigo != nil {
		cg.Codigo = *arg.Codigo
	}
	if arg.Nome != nil {
		cg.Nome = *arg.Nome
	}
	if arg.Tipo != nil {
		cg.Tipo = *arg.Tipo
	}
	if arg.Grupo != nil {
		cg.Grupo = arg.Grupo
	}
	if arg.ContaPaiID != nil {
		cg.ContaPaiID = arg.ContaPaiID
	}
	if arg.Ativo != nil {
		cg.Ativo = *arg.Ativo
	}

	if err := u.repo.Update(ctx, cg); err != nil {
		return nil, fmt.Errorf("failed to update conta gerencial: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "ContaGerencial",
		EntidadeID: id,
		Acao:       "update",
	})

	return cg, nil
}

// Delete soft-deletes a managerial account (only if no movements exist)
func (u *ContaGerencialUsecase) Delete(ctx context.Context, empresaID, id uuid.UUID) error {
	cg, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get conta gerencial: %w", err)
	}

	if cg.EmpresaID != empresaID {
		return errors.New("account not found")
	}

	// Check if account has movements
	hasMov, err := u.repo.HasMovimentacoes(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check movements: %w", err)
	}
	if hasMov {
		return errors.New("não é possível excluir conta gerencial com movimentações existentes")
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete conta gerencial: %w", err)
	}

	// Audit log
	u.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		Entidade:   "ContaGerencial",
		EntidadeID: id,
		Acao:       "delete",
	})

	return nil
}

// buildContaHierarchy builds a hierarchical tree from flat list of accounts
func buildContaHierarchy(items []*domain.ContaGerencial) []*domain.ContaGerencial {
	// Index by ID
	index := make(map[uuid.UUID]*domain.ContaGerencial)
	for _, item := range items {
		index[item.ID] = item
	}

	var roots []*domain.ContaGerencial

	for _, item := range items {
		if item.ContaPaiID == nil {
			roots = append(roots, item)
		} else {
			if parent, ok := index[*item.ContaPaiID]; ok {
				parent.Filhas = append(parent.Filhas, item)
			}
		}
	}

	return roots
}
