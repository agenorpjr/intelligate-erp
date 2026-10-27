package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrInsumoNaoEncontrado      = errors.New("insumo não encontrado")
	ErrInsumoJaExiste           = errors.New("insumo com este nome já existe")
	ErrInsumoCustoNaoEncontrado = errors.New("custo de insumo não encontrado")
)

type InsumoUsecase struct {
	insumoRepo    domain.InsumoRepository
	auditoriaRepo domain.AuditoriaRepository
}

func NewInsumoUsecase(insumoRepo domain.InsumoRepository, auditoriaRepo domain.AuditoriaRepository) *InsumoUsecase {
	return &InsumoUsecase{
		insumoRepo:    insumoRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

type CreateInsumoInput struct {
	EmpresaID          uuid.UUID `json:"empresa_id" validate:"required"`
	Nome               string    `json:"nome" validate:"required"`
	UnidadeMedida      string    `json:"unidade_medida" validate:"required"`
	CustoUnitarioAtual float64   `json:"custo_unitario_atual"`
	Fornecedor         *string   `json:"fornecedor,omitempty"`
}

type UpdateInsumoInput struct {
	Nome               *string  `json:"nome,omitempty"`
	UnidadeMedida      *string  `json:"unidade_medida,omitempty"`
	CustoUnitarioAtual *float64 `json:"custo_unitario_atual,omitempty"`
	Fornecedor         *string  `json:"fornecedor,omitempty"`
	Ativo              *bool    `json:"ativo,omitempty"`
}

type CreateInsumoCustoInput struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	DataVigencia  time.Time `json:"data_vigencia" validate:"required"`
	CustoUnitario float64   `json:"custo_unitario" validate:"required"`
	Observacao    *string   `json:"observacao,omitempty"`
}

func (uc *InsumoUsecase) Create(ctx context.Context, input CreateInsumoInput, usuarioID uuid.UUID) (*domain.Insumo, error) {
	// Check if insumo already exists
	existing, err := uc.insumoRepo.GetByNome(ctx, input.EmpresaID, input.Nome)
	if err == nil && existing != nil {
		return nil, ErrInsumoJaExiste
	}

	insumo := &domain.Insumo{
		ID:                 uuid.New(),
		EmpresaID:          input.EmpresaID,
		Nome:               input.Nome,
		UnidadeMedida:      input.UnidadeMedida,
		CustoUnitarioAtual: input.CustoUnitarioAtual,
		Fornecedor:         input.Fornecedor,
		Ativo:              true,
	}

	if err := uc.insumoRepo.Create(ctx, insumo); err != nil {
		return nil, err
	}

	// Create initial cost history entry
	if input.CustoUnitarioAtual > 0 {
		custo := &domain.InsumoCusto{
			ID:            uuid.New(),
			InsumoID:      insumo.ID,
			DataVigencia:  time.Now(),
			CustoUnitario: input.CustoUnitarioAtual,
			Observacao:    nil,
		}
		if err := uc.insumoRepo.CreateCusto(ctx, custo); err != nil {
			return nil, err
		}
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "insumo",
		EntidadeID: insumo.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(insumo)),
	})

	return insumo, nil
}

func (uc *InsumoUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Insumo, error) {
	insumo, err := uc.insumoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInsumoNaoEncontrado
	}
	return insumo, nil
}

func (uc *InsumoUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Insumo, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.insumoRepo.ListByEmpresa(ctx, empresaID, ativo, limit, offset)
}

func (uc *InsumoUsecase) Update(ctx context.Context, id uuid.UUID, input UpdateInsumoInput, usuarioID uuid.UUID) (*domain.Insumo, error) {
	insumo, err := uc.insumoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInsumoNaoEncontrado
	}

	dadosAntigos := toJSON(insumo)

	if input.Nome != nil {
		insumo.Nome = *input.Nome
	}
	if input.UnidadeMedida != nil {
		insumo.UnidadeMedida = *input.UnidadeMedida
	}
	if input.CustoUnitarioAtual != nil {
		insumo.CustoUnitarioAtual = *input.CustoUnitarioAtual
	}
	if input.Fornecedor != nil {
		insumo.Fornecedor = input.Fornecedor
	}
	if input.Ativo != nil {
		insumo.Ativo = *input.Ativo
	}

	if err := uc.insumoRepo.Update(ctx, insumo); err != nil {
		return nil, err
	}

	// If cost changed, create new cost history entry
	if input.CustoUnitarioAtual != nil && *input.CustoUnitarioAtual != insumo.CustoUnitarioAtual {
		custo := &domain.InsumoCusto{
			ID:            uuid.New(),
			InsumoID:      insumo.ID,
			DataVigencia:  time.Now(),
			CustoUnitario: *input.CustoUnitarioAtual,
			Observacao:    nil,
		}
		if err := uc.insumoRepo.CreateCusto(ctx, custo); err != nil {
			return nil, err
		}
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &insumo.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "insumo",
		EntidadeID: insumo.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(insumo)),
	})

	return insumo, nil
}

func (uc *InsumoUsecase) Delete(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	insumo, err := uc.insumoRepo.GetByID(ctx, id)
	if err != nil {
		return ErrInsumoNaoEncontrado
	}

	dadosAntigos := toJSON(insumo)

	if err := uc.insumoRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &insumo.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "insumo",
		EntidadeID: id,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}

func (uc *InsumoUsecase) CreateCusto(ctx context.Context, input CreateInsumoCustoInput, usuarioID uuid.UUID) (*domain.InsumoCusto, error) {
	// Verify insumo exists
	insumo, err := uc.insumoRepo.GetByID(ctx, input.InsumoID)
	if err != nil {
		return nil, ErrInsumoNaoEncontrado
	}

	custo := &domain.InsumoCusto{
		ID:            uuid.New(),
		InsumoID:      input.InsumoID,
		DataVigencia:  input.DataVigencia,
		CustoUnitario: input.CustoUnitario,
		Observacao:    input.Observacao,
	}

	if err := uc.insumoRepo.CreateCusto(ctx, custo); err != nil {
		return nil, err
	}

	// Update insumo's current cost
	if insumo != nil {
		insumo.CustoUnitarioAtual = input.CustoUnitario
		uc.insumoRepo.Update(ctx, insumo)
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &insumo.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "insumo_custo",
		EntidadeID: custo.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(custo)),
	})

	return custo, nil
}

func (uc *InsumoUsecase) GetCustoVigente(ctx context.Context, insumoID uuid.UUID, dataVigencia time.Time) (*domain.InsumoCusto, error) {
	custo, err := uc.insumoRepo.GetCustoVigente(ctx, insumoID, dataVigencia)
	if err != nil {
		return nil, ErrInsumoCustoNaoEncontrado
	}
	return custo, nil
}

func (uc *InsumoUsecase) GetCustoAtual(ctx context.Context, insumoID uuid.UUID) (*domain.InsumoCusto, error) {
	custo, err := uc.insumoRepo.GetCustoAtual(ctx, insumoID)
	if err != nil {
		return nil, ErrInsumoCustoNaoEncontrado
	}
	return custo, nil
}

func (uc *InsumoUsecase) ListCustosByInsumo(ctx context.Context, insumoID uuid.UUID, limit, offset int) ([]*domain.InsumoCusto, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.insumoRepo.ListCustosByInsumo(ctx, insumoID, limit, offset)
}
