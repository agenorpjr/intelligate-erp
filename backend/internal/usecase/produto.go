package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrProdutoCMVNaoEncontrado = errors.New("produto não encontrado")
	ErrProdutoJaExiste         = errors.New("produto com este SKU já existe")
)

type ProdutoUsecase struct {
	produtoRepo   domain.ProdutoRepository
	auditoriaRepo domain.AuditoriaRepository
}

func NewProdutoUsecase(produtoRepo domain.ProdutoRepository, auditoriaRepo domain.AuditoriaRepository) *ProdutoUsecase {
	return &ProdutoUsecase{
		produtoRepo:   produtoRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

type CreateProdutoInput struct {
	EmpresaID  uuid.UUID `json:"empresa_id" validate:"required"`
	Nome       string    `json:"nome" validate:"required"`
	SKUExterno *string   `json:"sku_externo,omitempty"`
	Categoria  *string   `json:"categoria,omitempty"`
}

type UpdateProdutoInput struct {
	Nome       *string `json:"nome,omitempty"`
	SKUExterno *string `json:"sku_externo,omitempty"`
	Categoria  *string `json:"categoria,omitempty"`
	Ativo      *bool   `json:"ativo,omitempty"`
}

func (uc *ProdutoUsecase) Create(ctx context.Context, input CreateProdutoInput, usuarioID uuid.UUID) (*domain.Produto, error) {
	// Check if produto already exists by SKU
	if input.SKUExterno != nil {
		existing, err := uc.produtoRepo.GetBySKUExterno(ctx, input.EmpresaID, *input.SKUExterno)
		if err == nil && existing != nil {
			return nil, ErrProdutoJaExiste
		}
	}

	produto := &domain.Produto{
		ID:         uuid.New(),
		EmpresaID:  input.EmpresaID,
		Nome:       input.Nome,
		SKUExterno: input.SKUExterno,
		Categoria:  input.Categoria,
		Ativo:      true,
	}

	if err := uc.produtoRepo.Create(ctx, produto); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "produto",
		EntidadeID: produto.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(produto)),
	})

	return produto, nil
}

func (uc *ProdutoUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	produto, err := uc.produtoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProdutoCMVNaoEncontrado
	}
	return produto, nil
}

func (uc *ProdutoUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Produto, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.produtoRepo.ListByEmpresa(ctx, empresaID, ativo, limit, offset)
}

func (uc *ProdutoUsecase) Update(ctx context.Context, id uuid.UUID, input UpdateProdutoInput, usuarioID uuid.UUID) (*domain.Produto, error) {
	produto, err := uc.produtoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProdutoNaoEncontrado
	}

	dadosAntigos := toJSON(produto)

	if input.Nome != nil {
		produto.Nome = *input.Nome
	}
	if input.SKUExterno != nil {
		produto.SKUExterno = input.SKUExterno
	}
	if input.Categoria != nil {
		produto.Categoria = input.Categoria
	}
	if input.Ativo != nil {
		produto.Ativo = *input.Ativo
	}

	if err := uc.produtoRepo.Update(ctx, produto); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &produto.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "produto",
		EntidadeID: produto.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(produto)),
	})

	return produto, nil
}

func (uc *ProdutoUsecase) Delete(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	produto, err := uc.produtoRepo.GetByID(ctx, id)
	if err != nil {
		return ErrProdutoNaoEncontrado
	}

	dadosAntigos := toJSON(produto)

	if err := uc.produtoRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &produto.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "produto",
		EntidadeID: id,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}
