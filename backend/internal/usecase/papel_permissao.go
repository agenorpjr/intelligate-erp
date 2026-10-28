package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrPapelNaoEncontrado     = errors.New("papel não encontrado")
	ErrPermissaoNaoEncontrada = errors.New("permissão não encontrada")
	ErrPapelSistema           = errors.New("não é possível modificar papel do sistema")
)

type PapelPermissaoUsecase struct {
	papelRepo          domain.PapelRepository
	permissaoRepo      domain.PermissaoRepository
	usuarioRepo        domain.UsuarioRepository
	usuarioPapelRepo   domain.UsuarioPapelRepository
	papelPermissaoRepo domain.PapelPermissaoRepository
	auditoriaRepo      domain.AuditoriaRepository
}

func NewPapelPermissaoUsecase(
	papelRepo domain.PapelRepository,
	permissaoRepo domain.PermissaoRepository,
	usuarioRepo domain.UsuarioRepository,
	usuarioPapelRepo domain.UsuarioPapelRepository,
	papelPermissaoRepo domain.PapelPermissaoRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *PapelPermissaoUsecase {
	return &PapelPermissaoUsecase{
		papelRepo:          papelRepo,
		permissaoRepo:      permissaoRepo,
		usuarioRepo:        usuarioRepo,
		usuarioPapelRepo:   usuarioPapelRepo,
		papelPermissaoRepo: papelPermissaoRepo,
		auditoriaRepo:      auditoriaRepo,
	}
}

// ========== Papel Methods ==========

func (uc *PapelPermissaoUsecase) CriarPapel(ctx context.Context, input domain.CreatePapelInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Papel, error) {
	papel := &domain.Papel{
		EmpresaID: input.EmpresaID,
		Nome:      input.Nome,
		Descricao: input.Descricao,
		Sistema:   input.Sistema,
	}

	if err := uc.papelRepo.Create(ctx, papel); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "papel",
		EntidadeID: papel.ID,
		Acao:       "criar",
		Depois:     toJSONPtr(papel),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return papel, nil
}

func (uc *PapelPermissaoUsecase) BuscarPapel(ctx context.Context, id uuid.UUID) (*domain.Papel, error) {
	papel, err := uc.papelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if papel == nil {
		return nil, ErrPapelNaoEncontrado
	}
	return papel, nil
}

func (uc *PapelPermissaoUsecase) ListarPapeis(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Papel, error) {
	return uc.papelRepo.ListByEmpresa(ctx, empresaID, limit, offset)
}

func (uc *PapelPermissaoUsecase) AtualizarPapel(ctx context.Context, id uuid.UUID, input domain.UpdatePapelInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Papel, error) {
	papel, err := uc.papelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if papel == nil {
		return nil, ErrPapelNaoEncontrado
	}

	if papel.Sistema {
		return nil, ErrPapelSistema
	}

	antes := toJSONPtr(papel)

	if input.Nome != nil {
		papel.Nome = *input.Nome
	}
	if input.Descricao != nil {
		papel.Descricao = input.Descricao
	}
	if input.Sistema != nil {
		papel.Sistema = *input.Sistema
	}

	if err := uc.papelRepo.Update(ctx, papel); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "papel",
		EntidadeID: papel.ID,
		Acao:       "atualizar",
		Antes:      antes,
		Depois:     toJSONPtr(papel),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return papel, nil
}

func (uc *PapelPermissaoUsecase) DeletarPapel(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	papel, err := uc.papelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	if papel.Sistema {
		return ErrPapelSistema
	}

	antes := toJSONPtr(papel)

	if err := uc.papelRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "papel",
		EntidadeID: papel.ID,
		Acao:       "deletar",
		Antes:      antes,
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

// ========== Permissão Methods ==========

func (uc *PapelPermissaoUsecase) ListarPermissoes(ctx context.Context, limit, offset int) ([]*domain.Permissao, error) {
	return uc.permissaoRepo.List(ctx, limit, offset)
}

func (uc *PapelPermissaoUsecase) BuscarPermissao(ctx context.Context, id uuid.UUID) (*domain.Permissao, error) {
	permissao, err := uc.permissaoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permissao == nil {
		return nil, ErrPermissaoNaoEncontrada
	}
	return permissao, nil
}

func (uc *PapelPermissaoUsecase) BuscarPermissaoPorChave(ctx context.Context, chave string) (*domain.Permissao, error) {
	return uc.permissaoRepo.GetByChave(ctx, chave)
}

// ========== Papel-Permissão Association ==========

func (uc *PapelPermissaoUsecase) AdicionarPermissaoAoPapel(ctx context.Context, papelID, permissaoID uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	papel, err := uc.papelRepo.GetByID(ctx, papelID)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	permissao, err := uc.permissaoRepo.GetByID(ctx, permissaoID)
	if err != nil {
		return err
	}
	if permissao == nil {
		return ErrPermissaoNaoEncontrada
	}

	if papel.Sistema {
		return ErrPapelSistema
	}

	if err := uc.papelPermissaoRepo.Create(ctx, papelID, permissaoID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "papel_permissao",
		EntidadeID: papelID,
		Acao:       "adicionar_permissao",
		Depois:     toJSONPtr(map[string]uuid.UUID{"papel_id": papelID, "permissao_id": permissaoID}),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *PapelPermissaoUsecase) RemoverPermissaoDoPapel(ctx context.Context, papelID, permissaoID uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	papel, err := uc.papelRepo.GetByID(ctx, papelID)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	if papel.Sistema {
		return ErrPapelSistema
	}

	if err := uc.papelPermissaoRepo.Delete(ctx, papelID, permissaoID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "papel_permissao",
		EntidadeID: papelID,
		Acao:       "remover_permissao",
		Antes:      toJSONPtr(map[string]uuid.UUID{"papel_id": papelID, "permissao_id": permissaoID}),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *PapelPermissaoUsecase) ListarPermissoesDoPapel(ctx context.Context, papelID uuid.UUID) ([]*domain.Permissao, error) {
	return uc.papelPermissaoRepo.ListPermissoesByPapel(ctx, papelID)
}

// ========== Usuario-Papel Association ==========

func (uc *PapelPermissaoUsecase) AssociarUsuarioPapel(ctx context.Context, input domain.UsuarioPapel, usuarioID uuid.UUID, ip, userAgent string) error {
	// Verificar se papel existe
	papel, err := uc.papelRepo.GetByID(ctx, input.PapelID)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	// Verificar se usuário existe e pertence à mesma empresa
	usuario, err := uc.usuarioRepo.GetByID(ctx, input.UsuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return errors.New("usuário não encontrado")
	}
	if usuario.EmpresaID != papel.EmpresaID {
		return errors.New("usuário e papel devem pertencer à mesma empresa")
	}

	if err := uc.usuarioPapelRepo.Create(ctx, &input); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario_papel",
		EntidadeID: input.ID,
		Acao:       "associar",
		Depois:     toJSONPtr(input),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *PapelPermissaoUsecase) RemoverUsuarioPapel(ctx context.Context, usuarioPapelID uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	associacao, err := uc.usuarioPapelRepo.GetByID(ctx, usuarioPapelID)
	if err != nil {
		return err
	}
	if associacao == nil {
		return errors.New("associação usuário-papel não encontrada")
	}

	papel, err := uc.papelRepo.GetByID(ctx, associacao.PapelID)
	if err != nil {
		return err
	}

	if papel.Sistema {
		return ErrPapelSistema
	}

	antes := toJSONPtr(associacao)

	if err := uc.usuarioPapelRepo.Delete(ctx, usuarioPapelID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario_papel",
		EntidadeID: usuarioPapelID,
		Acao:       "remover",
		Antes:      antes,
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *PapelPermissaoUsecase) ListarPapeisDoUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.UsuarioPapelComDetalhes, error) {
	return uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
}

// ========== Helper ==========

func (uc *PapelPermissaoUsecase) auditar(ctx context.Context, a domain.Auditoria) {
	// Best effort - não bloquear operação principal se auditoria falhar
	_ = uc.auditoriaRepo.Create(ctx, &a)
}

func toJSONPtr(v interface{}) *string {
	data, _ := json.Marshal(v)
	s := string(data)
	return &s
}
