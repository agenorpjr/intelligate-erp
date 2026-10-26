package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrLojaNaoEncontrada = errors.New("loja não encontrada")
	ErrCodigoDuplicado   = errors.New("código de loja já existe para esta empresa")
)

type LojaUsecase struct {
	lojaRepo      domain.LojaRepository
	empresaRepo   domain.EmpresaRepository
	auditoriaRepo domain.AuditoriaRepository
}

func NewLojaUsecase(lojaRepo domain.LojaRepository, empresaRepo domain.EmpresaRepository, auditoriaRepo domain.AuditoriaRepository) *LojaUsecase {
	return &LojaUsecase{
		lojaRepo:      lojaRepo,
		empresaRepo:   empresaRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

func (uc *LojaUsecase) Criar(ctx context.Context, input domain.CreateLojaInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Loja, error) {
	// Verificar se empresa existe
	empresa, err := uc.empresaRepo.GetByID(ctx, input.EmpresaID)
	if err != nil {
		return nil, err
	}
	if empresa == nil {
		return nil, errors.New("empresa não encontrada")
	}

	// Verificar se código já existe
	lojas, err := uc.lojaRepo.ListByEmpresa(ctx, input.EmpresaID, 1000, 0)
	if err != nil {
		return nil, err
	}
	for _, l := range lojas {
		if l.Codigo == input.Codigo {
			return nil, ErrCodigoDuplicado
		}
	}

	loja := &domain.Loja{
		ID:          uuid.New(),
		EmpresaID:   input.EmpresaID,
		Codigo:      input.Codigo,
		Nome:        input.Nome,
		Endereco:    input.Endereco,
		FusoHorario: input.FusoHorario,
		Ativo:       true,
	}

	if err := uc.lojaRepo.Create(ctx, loja); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "loja",
		EntidadeID: loja.ID,
		Acao:       "create",
		Depois:     stringPtr(toJSON(loja)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return loja, nil
}

func (uc *LojaUsecase) Listar(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Loja, error) {
	return uc.lojaRepo.ListByEmpresa(ctx, empresaID, limit, offset)
}

func (uc *LojaUsecase) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Loja, error) {
	loja, err := uc.lojaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loja == nil {
		return nil, ErrLojaNaoEncontrada
	}
	return loja, nil
}

func (uc *LojaUsecase) Atualizar(ctx context.Context, id uuid.UUID, input domain.UpdateLojaInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Loja, error) {
	loja, err := uc.lojaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loja == nil {
		return nil, ErrLojaNaoEncontrada
	}

	antes := toJSON(loja)

	if input.Codigo != nil {
		loja.Codigo = *input.Codigo
	}
	if input.Nome != nil {
		loja.Nome = *input.Nome
	}
	if input.Endereco != nil {
		loja.Endereco = input.Endereco
	}
	if input.FusoHorario != nil {
		loja.FusoHorario = *input.FusoHorario
	}
	if input.Ativo != nil {
		loja.Ativo = *input.Ativo
	}

	if err := uc.lojaRepo.Update(ctx, loja); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &loja.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "loja",
		EntidadeID: loja.ID,
		Acao:       "update",
		Antes:      stringPtr(antes),
		Depois:     stringPtr(toJSON(loja)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return loja, nil
}

func (uc *LojaUsecase) Deletar(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	loja, err := uc.lojaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if loja == nil {
		return ErrLojaNaoEncontrada
	}

	antes := toJSON(loja)

	if err := uc.lojaRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Auditoria
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &loja.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "loja",
		EntidadeID: loja.ID,
		Acao:       "delete",
		Antes:      stringPtr(antes),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func toJSON(v interface{}) string {
	// Simplified - in production use json.Marshal
	return "{}"
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
