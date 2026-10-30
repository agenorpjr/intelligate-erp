package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrUsuarioNaoEncontrado     = errors.New("usuário não encontrado")
	ErrEmailJaExiste            = errors.New("e-mail já cadastrado para esta empresa")
	ErrNaoPodeDesativarSiMesmo  = errors.New("não é possível desativar a si mesmo")
	ErrNaoPodeRemoverPapelAdmin = errors.New("não é possível remover o último administrador ativo da empresa")
	ErrPapelDeOutraEmpresa      = errors.New("papel pertence a outra empresa")
	ErrLojaDeOutraEmpresa       = errors.New("loja pertence a outra empresa")
	ErrUsuarioJaTemPapel        = errors.New("usuário já possui este papel")
	ErrUsuarioNaoTemPapel       = errors.New("usuário não possui este papel")
	ErrUsuarioJaTemAcessoLoja   = errors.New("usuário já tem acesso a esta loja")
	ErrUsuarioNaoTemAcessoLoja  = errors.New("usuário não tem acesso a esta loja")
)

type UsuarioUsecase struct {
	usuarioRepo        domain.UsuarioRepository
	usuarioPapelRepo   domain.UsuarioPapelRepository
	papelRepo          domain.PapelRepository
	papelPermissaoRepo domain.PapelPermissaoRepository
	lojaRepo           domain.LojaRepository
	auditoriaRepo      domain.AuditoriaRepository
}

func NewUsuarioUsecase(
	usuarioRepo domain.UsuarioRepository,
	usuarioPapelRepo domain.UsuarioPapelRepository,
	papelRepo domain.PapelRepository,
	papelPermissaoRepo domain.PapelPermissaoRepository,
	lojaRepo domain.LojaRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *UsuarioUsecase {
	return &UsuarioUsecase{
		usuarioRepo:        usuarioRepo,
		usuarioPapelRepo:   usuarioPapelRepo,
		papelRepo:          papelRepo,
		papelPermissaoRepo: papelPermissaoRepo,
		lojaRepo:           lojaRepo,
		auditoriaRepo:      auditoriaRepo,
	}
}

// ========== CRUD Usuário ==========

func (uc *UsuarioUsecase) Criar(ctx context.Context, input domain.CreateUsuarioInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Usuario, error) {
	// Verificar se e-mail já existe na empresa
	existente, err := uc.usuarioRepo.GetByEmailAndEmpresa(ctx, input.Email, input.EmpresaID)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, ErrEmailJaExiste
	}

	// Hash da senha
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	usuario := &domain.Usuario{
		EmpresaID: input.EmpresaID,
		Email:     input.Email,
		Nome:      input.Nome,
		SenhaHash: string(senhaHash),
		Ativo:     true,
	}

	if err := uc.usuarioRepo.Create(ctx, usuario); err != nil {
		return nil, err
	}

	// Auditoria (sem senha)
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario",
		EntidadeID: usuario.ID,
		Acao:       "criar",
		Depois:     toJSONPtr(usuarioSemSenha(usuario)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	usuario.SenhaHash = ""
	return usuario, nil
}

func (uc *UsuarioUsecase) BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	usuario, err := uc.usuarioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usuario == nil {
		return nil, ErrUsuarioNaoEncontrado
	}
	usuario.SenhaHash = ""
	return usuario, nil
}

func (uc *UsuarioUsecase) Listar(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Usuario, error) {
	usuarios, err := uc.usuarioRepo.ListByEmpresa(ctx, empresaID, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, u := range usuarios {
		u.SenhaHash = ""
	}
	return usuarios, nil
}

func (uc *UsuarioUsecase) Atualizar(ctx context.Context, id uuid.UUID, input domain.UpdateUsuarioInput, usuarioID uuid.UUID, ip, userAgent string) (*domain.Usuario, error) {
	usuario, err := uc.usuarioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usuario == nil {
		return nil, ErrUsuarioNaoEncontrado
	}

	// Verificar se está tentando desativar a si mesmo
	if input.Ativo != nil && !*input.Ativo && id == usuarioID {
		return nil, ErrNaoPodeDesativarSiMesmo
	}

	// Verificar e-mail único se estiver alterando
	if input.Email != nil && *input.Email != usuario.Email {
		existente, err := uc.usuarioRepo.GetByEmailAndEmpresa(ctx, *input.Email, usuario.EmpresaID)
		if err != nil {
			return nil, err
		}
		if existente != nil && existente.ID != id {
			return nil, ErrEmailJaExiste
		}
	}

	antes := toJSONPtr(usuarioSemSenha(usuario))

	if input.Nome != nil {
		usuario.Nome = *input.Nome
	}
	if input.Email != nil {
		usuario.Email = *input.Email
	}
	if input.Ativo != nil {
		usuario.Ativo = *input.Ativo
	}

	if err := uc.usuarioRepo.Update(ctx, usuario); err != nil {
		return nil, err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario",
		EntidadeID: usuario.ID,
		Acao:       "atualizar",
		Antes:      antes,
		Depois:     toJSONPtr(usuarioSemSenha(usuario)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	usuario.SenhaHash = ""
	return usuario, nil
}

func (uc *UsuarioUsecase) Desativar(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	// Não pode desativar a si mesmo
	if id == usuarioID {
		return ErrNaoPodeDesativarSiMesmo
	}

	usuario, err := uc.usuarioRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}

	if !usuario.Ativo {
		return nil // Já inativo
	}

	// Verificar se é o último admin ativo da empresa
	if err := uc.verificarUltimoAdminAtivo(ctx, usuario.EmpresaID, id); err != nil {
		return err
	}

	antes := toJSONPtr(usuarioSemSenha(usuario))
	usuario.Ativo = false

	if err := uc.usuarioRepo.Update(ctx, usuario); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario",
		EntidadeID: usuario.ID,
		Acao:       "desativar",
		Antes:      antes,
		Depois:     toJSONPtr(usuarioSemSenha(usuario)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *UsuarioUsecase) Reativar(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID, ip, userAgent string) error {
	usuario, err := uc.usuarioRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}

	if usuario.Ativo {
		return nil // Já ativo
	}

	antes := toJSONPtr(usuarioSemSenha(usuario))
	usuario.Ativo = true

	if err := uc.usuarioRepo.Update(ctx, usuario); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario",
		EntidadeID: usuario.ID,
		Acao:       "reativar",
		Antes:      antes,
		Depois:     toJSONPtr(usuarioSemSenha(usuario)),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

// ========== Gestão de Papéis ==========

func (uc *UsuarioUsecase) AssociarPapel(ctx context.Context, usuarioID, papelID uuid.UUID, lojaID *uuid.UUID, usuarioLogadoID uuid.UUID, ip, userAgent string) error {
	// Verificar usuário
	usuario, err := uc.usuarioRepo.GetByID(ctx, usuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}

	// Verificar papel
	papel, err := uc.papelRepo.GetByID(ctx, papelID)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	// Verificar se papel pertence à mesma empresa
	if papel.EmpresaID != usuario.EmpresaID {
		return ErrPapelDeOutraEmpresa
	}

	// Verificar loja se informada
	if lojaID != nil {
		loja, err := uc.lojaRepo.GetByID(ctx, *lojaID)
		if err != nil {
			return err
		}
		if loja == nil {
			return ErrLojaNaoEncontrada
		}
		if loja.EmpresaID != usuario.EmpresaID {
			return ErrLojaDeOutraEmpresa
		}
	}

	// Verificar se já tem o papel
	papeisAtuais, err := uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
	if err != nil {
		return err
	}
	for _, p := range papeisAtuais {
		if p.PapelID == papelID && ((lojaID == nil && p.LojaID == nil) || (lojaID != nil && p.LojaID != nil && *p.LojaID == *lojaID)) {
			return ErrUsuarioJaTemPapel
		}
	}

	// Criar associação
	associacao := &domain.UsuarioPapel{
		UsuarioID: usuarioID,
		PapelID:   papelID,
		LojaID:    lojaID,
	}

	if err := uc.usuarioPapelRepo.Create(ctx, associacao); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioLogadoID,
		Entidade:   "usuario_papel",
		EntidadeID: associacao.ID,
		Acao:       "associar",
		Depois:     toJSONPtr(associacao),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *UsuarioUsecase) RemoverPapel(ctx context.Context, usuarioPapelID uuid.UUID, usuarioLogadoID uuid.UUID, ip, userAgent string) error {
	associacao, err := uc.usuarioPapelRepo.GetByID(ctx, usuarioPapelID)
	if err != nil {
		return err
	}
	if associacao == nil {
		return ErrUsuarioNaoTemPapel
	}

	// Verificar papel
	papel, err := uc.papelRepo.GetByID(ctx, associacao.PapelID)
	if err != nil {
		return err
	}
	if papel == nil {
		return ErrPapelNaoEncontrado
	}

	// Verificar se é papel do sistema
	if papel.Sistema {
		return errors.New("não é possível remover papel do sistema")
	}

	// Verificar se é o último admin ativo
	if err := uc.verificarUltimoAdminAtivoAoRemoverPapel(ctx, associacao.UsuarioID, papel); err != nil {
		return err
	}

	antes := toJSONPtr(associacao)

	if err := uc.usuarioPapelRepo.Delete(ctx, usuarioPapelID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &papel.EmpresaID,
		UsuarioID:  &usuarioLogadoID,
		Entidade:   "usuario_papel",
		EntidadeID: usuarioPapelID,
		Acao:       "remover",
		Antes:      antes,
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *UsuarioUsecase) ListarPapeis(ctx context.Context, usuarioID uuid.UUID) ([]*domain.UsuarioPapelComDetalhes, error) {
	return uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
}

func (uc *UsuarioUsecase) ListarPermissoesEfetivas(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]string, error) {
	papeisUsuario, err := uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	permissoesSet := make(map[string]struct{})
	for _, up := range papeisUsuario {
		papel, err := uc.papelRepo.GetByID(ctx, up.PapelID)
		if err != nil || papel == nil {
			continue
		}
		if papel.EmpresaID != empresaID {
			continue
		}

		permissoes, err := uc.papelPermissaoRepo.ListPermissoesByPapel(ctx, up.PapelID)
		if err != nil {
			return nil, err
		}
		for _, p := range permissoes {
			permissoesSet[p.Chave] = struct{}{}
		}
	}

	permissoes := make([]string, 0, len(permissoesSet))
	for p := range permissoesSet {
		permissoes = append(permissoes, p)
	}
	return permissoes, nil
}

// ========== Gestão de Lojas ==========

func (uc *UsuarioUsecase) AdicionarLoja(ctx context.Context, usuarioID, lojaID uuid.UUID, usuarioLogadoID uuid.UUID, ip, userAgent string) error {
	// Verificar usuário
	usuario, err := uc.usuarioRepo.GetByID(ctx, usuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}

	// Verificar loja
	loja, err := uc.lojaRepo.GetByID(ctx, lojaID)
	if err != nil {
		return err
	}
	if loja == nil {
		return ErrLojaNaoEncontrada
	}

	// Verificar se loja pertence à mesma empresa
	if loja.EmpresaID != usuario.EmpresaID {
		return ErrLojaDeOutraEmpresa
	}

	// Verificar se já tem acesso
	temAcesso, err := uc.usuarioRepo.CheckLojaAccess(ctx, usuarioID, lojaID)
	if err != nil {
		return err
	}
	if temAcesso {
		return ErrUsuarioJaTemAcessoLoja
	}

	if err := uc.usuarioRepo.AddLoja(ctx, usuarioID, lojaID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioLogadoID,
		Entidade:   "usuario_loja",
		EntidadeID: uuid.New(), // ID da associação não retornado, gerar para auditoria
		Acao:       "adicionar_loja",
		Depois:     toJSONPtr(map[string]uuid.UUID{"usuario_id": usuarioID, "loja_id": lojaID}),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *UsuarioUsecase) RemoverLoja(ctx context.Context, usuarioID, lojaID uuid.UUID, usuarioLogadoID uuid.UUID, ip, userAgent string) error {
	// Verificar usuário
	usuario, err := uc.usuarioRepo.GetByID(ctx, usuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}

	// Verificar se tem acesso
	temAcesso, err := uc.usuarioRepo.CheckLojaAccess(ctx, usuarioID, lojaID)
	if err != nil {
		return err
	}
	if !temAcesso {
		return ErrUsuarioNaoTemAcessoLoja
	}

	if err := uc.usuarioRepo.RemoveLoja(ctx, usuarioID, lojaID); err != nil {
		return err
	}

	// Auditoria
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioLogadoID,
		Entidade:   "usuario_loja",
		EntidadeID: uuid.New(),
		Acao:       "remover_loja",
		Antes:      toJSONPtr(map[string]uuid.UUID{"usuario_id": usuarioID, "loja_id": lojaID}),
		IP:         &ip,
		UserAgent:  &userAgent,
	})

	return nil
}

func (uc *UsuarioUsecase) ListarLojas(ctx context.Context, usuarioID uuid.UUID) ([]*domain.Loja, error) {
	return uc.usuarioRepo.ListLojas(ctx, usuarioID)
}

// ========== Helpers ==========

func (uc *UsuarioUsecase) verificarUltimoAdminAtivo(ctx context.Context, empresaID uuid.UUID, usuarioIDExcluir uuid.UUID) error {
	// Buscar papel ADMIN da empresa
	papeis, err := uc.papelRepo.ListByEmpresa(ctx, empresaID, 100, 0)
	if err != nil {
		return err
	}

	var papelAdminID *uuid.UUID
	for _, p := range papeis {
		if p.Codigo == domain.CodigoPadraoAdmin {
			papelAdminID = &p.ID
			break
		}
	}

	if papelAdminID == nil {
		return nil // Não há papel ADMIN, não precisa verificar
	}

	// Buscar usuários com papel ADMIN ativos
	usuarios, err := uc.usuarioRepo.ListByEmpresa(ctx, empresaID, 1000, 0)
	if err != nil {
		return err
	}

	adminsAtivos := 0
	for _, u := range usuarios {
		if u.ID == usuarioIDExcluir {
			continue
		}
		if !u.Ativo {
			continue
		}
		papeisU, err := uc.usuarioPapelRepo.ListByUsuario(ctx, u.ID)
		if err != nil {
			continue
		}
		for _, up := range papeisU {
			if up.PapelID == *papelAdminID {
				adminsAtivos++
				break
			}
		}
	}

	if adminsAtivos == 0 {
		return ErrNaoPodeRemoverPapelAdmin
	}

	return nil
}

func (uc *UsuarioUsecase) verificarUltimoAdminAtivoAoRemoverPapel(ctx context.Context, usuarioID uuid.UUID, papel *domain.Papel) error {
	// Se não é papel ADMIN, não precisa verificar
	if papel.Codigo != domain.CodigoPadraoAdmin {
		return nil
	}

	// Verificar se o usuário tem outro papel ADMIN
	papeisUsuario, err := uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
	if err != nil {
		return err
	}

	temOutroAdmin := false
	for _, up := range papeisUsuario {
		if up.PapelID == papel.ID {
			continue // Este é o papel que está sendo removido
		}
		p, err := uc.papelRepo.GetByID(ctx, up.PapelID)
		if err != nil || p == nil {
			continue
		}
		if p.Nome == "ADMIN" || p.Nome == "Admin" || p.Nome == "admin" {
			temOutroAdmin = true
			break
		}
	}

	if temOutroAdmin {
		return nil
	}

	// Verificar se há outros admins ativos na empresa
	return uc.verificarUltimoAdminAtivo(ctx, papel.EmpresaID, usuarioID)
}

func (uc *UsuarioUsecase) auditar(ctx context.Context, a domain.Auditoria) {
	_ = uc.auditoriaRepo.Create(ctx, &a)
}

func usuarioSemSenha(u *domain.Usuario) *domain.Usuario {
	copy := *u
	copy.SenhaHash = ""
	return &copy
}
