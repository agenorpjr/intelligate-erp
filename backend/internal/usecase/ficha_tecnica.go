package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrFichaTecnicaNaoEncontrada = errors.New("ficha técnica não encontrada")
	ErrFichaTecnicaJaExiste      = errors.New("ficha técnica já existe para este produto e loja")
)

type FichaTecnicaUsecase struct {
	fichaRepo     domain.FichaTecnicaRepository
	produtoRepo   domain.ProdutoRepository
	lojaRepo      domain.LojaRepository
	insumoRepo    domain.InsumoRepository
	auditoriaRepo domain.AuditoriaRepository
}

func NewFichaTecnicaUsecase(
	fichaRepo domain.FichaTecnicaRepository,
	produtoRepo domain.ProdutoRepository,
	lojaRepo domain.LojaRepository,
	insumoRepo domain.InsumoRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *FichaTecnicaUsecase {
	return &FichaTecnicaUsecase{
		fichaRepo:     fichaRepo,
		produtoRepo:   produtoRepo,
		lojaRepo:      lojaRepo,
		insumoRepo:    insumoRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

type CreateFichaTecnicaInput struct {
	ProdutoID         uuid.UUID                           `json:"produto_id" validate:"required"`
	LojaID            uuid.UUID                           `json:"loja_id" validate:"required"`
	Rendimento        float64                             `json:"rendimento"`
	UnidadeRendimento string                              `json:"unidade_rendimento"`
	Componentes       []CreateFichaTecnicaComponenteInput `json:"componentes" validate:"required,min=1"`
}

type CreateFichaTecnicaComponenteInput struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string    `json:"unidade_medida" validate:"required"`
}

type UpdateFichaTecnicaInput struct {
	Rendimento        *float64                            `json:"rendimento,omitempty"`
	UnidadeRendimento *string                             `json:"unidade_rendimento,omitempty"`
	Ativo             *bool                               `json:"ativo,omitempty"`
	Componentes       []UpdateFichaTecnicaComponenteInput `json:"componentes,omitempty"`
}

type UpdateFichaTecnicaComponenteInput struct {
	ID            *uuid.UUID `json:"id,omitempty"`
	InsumoID      uuid.UUID  `json:"insumo_id" validate:"required"`
	Quantidade    float64    `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string     `json:"unidade_medida" validate:"required"`
}

type CMVFilter struct {
	EmpresaID  uuid.UUID
	LojaID     *uuid.UUID
	DataInicio time.Time
	DataFim    time.Time
	Limit      int
	Offset     int
}

func (uc *FichaTecnicaUsecase) Create(ctx context.Context, input CreateFichaTecnicaInput, usuarioID uuid.UUID, empresaID uuid.UUID) (*domain.FichaTecnica, error) {
	// Verify produto exists and belongs to the same empresa
	produto, err := uc.produtoRepo.GetByID(ctx, input.ProdutoID)
	if err != nil || produto == nil {
		return nil, ErrProdutoNaoEncontrado
	}
	if produto.EmpresaID != empresaID {
		return nil, ErrProdutoNaoEncontrado
	}

	// Verify loja exists
	_, err = uc.lojaRepo.GetByID(ctx, input.LojaID)
	if err != nil {
		return nil, ErrLojaNaoEncontrada
	}

	// Check if ficha tecnica already exists for this produto/loja
	existing, err := uc.fichaRepo.GetByProdutoLoja(ctx, input.ProdutoID, input.LojaID)
	if err == nil && existing != nil {
		return nil, ErrFichaTecnicaJaExiste
	}

	// Verify all insumos exist
	for _, comp := range input.Componentes {
		_, err := uc.insumoRepo.GetByID(ctx, comp.InsumoID)
		if err != nil {
			return nil, ErrInsumoNaoEncontrado
		}
	}

	ficha := &domain.FichaTecnica{
		ID:                uuid.New(),
		ProdutoID:         input.ProdutoID,
		LojaID:            input.LojaID,
		Rendimento:        input.Rendimento,
		UnidadeRendimento: input.UnidadeRendimento,
		Ativo:             true,
	}

	if err := uc.fichaRepo.Create(ctx, ficha); err != nil {
		return nil, err
	}

	// Create componentes
	for _, comp := range input.Componentes {
		componente := &domain.FichaTecnicaComponente{
			ID:             uuid.New(),
			FichaTecnicaID: ficha.ID,
			InsumoID:       comp.InsumoID,
			Quantidade:     comp.Quantidade,
			UnidadeMedida:  comp.UnidadeMedida,
		}
		if err := uc.fichaRepo.CreateComponente(ctx, componente); err != nil {
			return nil, err
		}
	}

	// Calculate theoretical cost
	custoTeorico, err := uc.calcularCustoTeorico(ctx, ficha.ID, time.Now())
	if err == nil {
		ficha.CustoTeorico = custoTeorico
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &produto.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ficha_tecnica",
		EntidadeID: ficha.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(ficha)),
	})

	return ficha, nil
}

func (uc *FichaTecnicaUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.FichaTecnica, error) {
	ficha, err := uc.fichaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFichaTecnicaNaoEncontrada
	}

	// Load componentes
	componentes, err := uc.fichaRepo.GetComponentes(ctx, id)
	if err == nil {
		ficha.Componentes = componentes
	}

	return ficha, nil
}

func (uc *FichaTecnicaUsecase) ListByProduto(ctx context.Context, produtoID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.fichaRepo.ListByProduto(ctx, produtoID, ativo, limit, offset)
}

func (uc *FichaTecnicaUsecase) ListByLoja(ctx context.Context, lojaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.fichaRepo.ListByLoja(ctx, lojaID, ativo, limit, offset)
}

func (uc *FichaTecnicaUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.fichaRepo.ListByEmpresa(ctx, empresaID, ativo, limit, offset)
}

func (uc *FichaTecnicaUsecase) Update(ctx context.Context, id uuid.UUID, input UpdateFichaTecnicaInput, usuarioID uuid.UUID) (*domain.FichaTecnica, error) {
	ficha, err := uc.fichaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFichaTecnicaNaoEncontrada
	}

	dadosAntigos := toJSON(ficha)

	if input.Rendimento != nil {
		ficha.Rendimento = *input.Rendimento
	}
	if input.UnidadeRendimento != nil {
		ficha.UnidadeRendimento = *input.UnidadeRendimento
	}
	if input.Ativo != nil {
		ficha.Ativo = *input.Ativo
	}

	if err := uc.fichaRepo.Update(ctx, ficha); err != nil {
		return nil, err
	}

	// Update componentes if provided
	if input.Componentes != nil {
		// Delete existing componentes
		if err := uc.fichaRepo.DeleteComponentesByFicha(ctx, ficha.ID); err != nil {
			return nil, err
		}

		// Create new componentes
		for _, comp := range input.Componentes {
			componente := &domain.FichaTecnicaComponente{
				ID:             uuid.New(),
				FichaTecnicaID: ficha.ID,
				InsumoID:       comp.InsumoID,
				Quantidade:     comp.Quantidade,
				UnidadeMedida:  comp.UnidadeMedida,
			}
			if err := uc.fichaRepo.CreateComponente(ctx, componente); err != nil {
				return nil, err
			}
		}
	}

	// Recalculate theoretical cost
	custoTeorico, err := uc.calcularCustoTeorico(ctx, ficha.ID, time.Now())
	if err == nil {
		ficha.CustoTeorico = custoTeorico
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ficha.ProdutoID, // We need to get empresa from produto
		UsuarioID:  &usuarioID,
		Entidade:   "ficha_tecnica",
		EntidadeID: ficha.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(ficha)),
	})

	return ficha, nil
}

func (uc *FichaTecnicaUsecase) Delete(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	ficha, err := uc.fichaRepo.GetByID(ctx, id)
	if err != nil {
		return ErrFichaTecnicaNaoEncontrada
	}

	dadosAntigos := toJSON(ficha)

	if err := uc.fichaRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ficha.ProdutoID,
		UsuarioID:  &usuarioID,
		Entidade:   "ficha_tecnica",
		EntidadeID: id,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}

func (uc *FichaTecnicaUsecase) GetCMVResumo(ctx context.Context, filter CMVFilter) (*domain.CMVResumo, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.fichaRepo.GetCMVResumo(ctx, filter.EmpresaID, filter.LojaID, filter.DataInicio, filter.DataFim)
}

func (uc *FichaTecnicaUsecase) GetCMVProdutosDetalhe(ctx context.Context, filter CMVFilter) ([]*domain.CMVProdutoDetalhe, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Get products with ficha tecnica
	comFicha, err := uc.fichaRepo.GetCMVProdutosDetalhe(ctx, filter.EmpresaID, filter.LojaID, filter.DataInicio, filter.DataFim, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}

	// Get products without ficha tecnica
	semFicha, err := uc.fichaRepo.GetProdutosSemFichaTecnica(ctx, filter.EmpresaID, filter.LojaID, filter.DataInicio, filter.DataFim, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}

	// Combine results
	result := append(comFicha, semFicha...)
	return result, nil
}

func (uc *FichaTecnicaUsecase) calcularCustoTeorico(ctx context.Context, fichaTecnicaID uuid.UUID, dataVigencia time.Time) (float64, error) {
	componentes, err := uc.fichaRepo.GetComponentes(ctx, fichaTecnicaID)
	if err != nil {
		return 0, err
	}

	var custoTotal float64
	for _, comp := range componentes {
		custo, err := uc.insumoRepo.GetCustoVigente(ctx, comp.InsumoID, dataVigencia)
		if err != nil {
			// Fallback to current cost
			custo, err = uc.insumoRepo.GetCustoAtual(ctx, comp.InsumoID)
			if err != nil {
				continue
			}
		}
		custoTotal += comp.Quantidade * custo.CustoUnitario
	}

	// Get ficha to apply rendimento
	ficha, err := uc.fichaRepo.GetByID(ctx, fichaTecnicaID)
	if err != nil {
		return custoTotal, nil
	}

	if ficha.Rendimento > 0 {
		return custoTotal / ficha.Rendimento, nil
	}

	return custoTotal, nil
}
