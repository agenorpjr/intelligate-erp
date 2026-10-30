package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrSenhaRedefinicaoNaoEncontrada = errors.New("token de redefinicao invalido ou expirado")
	ErrSenhaRedefinicaoJaUsada       = errors.New("token de redefinicao ja foi utilizado")
	ErrSenhaAtualInvalida            = errors.New("senha atual invalida")
	ErrNovaSenhaIgualAtual           = errors.New("nova senha deve ser diferente da senha atual")
	ErrSessaoNaoEncontrada           = errors.New("sessao nao encontrada")
)

const (
	// MinSenhaLength define o tamanho minimo da senha.
	MinSenhaLength = 8
	// RedefinicaoExpiracao define o tempo de expiracao do token de redefinicao (30 minutos).
	RedefinicaoExpiracao = 30 * time.Minute
)

// SenhaSessaoUsecase gerencia recuperacao de senha, alteracao de senha e revocacao de sessoes.
type SenhaSessaoUsecase struct {
	usuarioRepo          domain.UsuarioRepository
	senhaRedefinicaoRepo domain.SenhaRedefinicaoRepository
	sessaoRevogadaRepo   domain.SessaoRevogadaRepository
	auditoriaRepo        domain.AuditoriaRepository
	hashSenha            func(senha string) (string, error)
	gerarTokenAleatorio  func() (string, error)
	appEnv               string
	exibirTokenDev       bool
}

func NewSenhaSessaoUsecase(
	usuarioRepo domain.UsuarioRepository,
	senhaRedefinicaoRepo domain.SenhaRedefinicaoRepository,
	sessaoRevogadaRepo domain.SessaoRevogadaRepository,
	auditoriaRepo domain.AuditoriaRepository,
	appEnv string,
	exibirTokenDev bool,
) *SenhaSessaoUsecase {
	return &SenhaSessaoUsecase{
		usuarioRepo:          usuarioRepo,
		senhaRedefinicaoRepo: senhaRedefinicaoRepo,
		sessaoRevogadaRepo:   sessaoRevogadaRepo,
		auditoriaRepo:        auditoriaRepo,
		hashSenha:            bcryptHash,
		gerarTokenAleatorio:  gerarTokenHex,
		appEnv:               appEnv,
		exibirTokenDev:       exibirTokenDev,
	}
}

// EsqueciSenha gera um token de redefinicao de senha para o email informado.
// Responde sempre com sucesso generico para evitar enumeração de usuarios.
func (uc *SenhaSessaoUsecase) EsqueciSenha(ctx context.Context, input domain.EsqueciSenhaInput) (tokenDev string, err error) {
	// Buscar usuario por email (sem revelar se existe).
	usuario, err := uc.usuarioRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		// Erro de banco: retornar generico sem expor detalhes.
		return "", nil
	}
	if usuario == nil || !usuario.Ativo {
		// Usuario nao encontrado ou inativo: retornar sucesso generico.
		return "", nil
	}

	// Invalidar tokens anteriores nao utilizados para este usuario.
	if err := uc.senhaRedefinicaoRepo.InvalidarPorUsuario(ctx, usuario.ID); err != nil {
		return "", err
	}

	// Gerar token aleatorio criptograficamente seguro.
	token, err := uc.gerarTokenAleatorio()
	if err != nil {
		return "", err
	}

	// Hash do token para armazenamento (nunca o token puro).
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	redefinicao := &domain.SenhaRedefinicao{
		UsuarioID: usuario.ID,
		TokenHash: string(tokenHash),
		ExpiraEm:  time.Now().Add(RedefinicaoExpiracao),
	}

	if err := uc.senhaRedefinicaoRepo.Create(ctx, redefinicao); err != nil {
		return "", err
	}

	// Auditoria (sem expor o token).
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuario.ID,
		Entidade:   "senha_redefinicao",
		EntidadeID: redefinicao.ID,
		Acao:       "solicitar",
		Depois:     toJSONPtr(map[string]interface{}{"usuario_id": usuario.ID, "expira_em": redefinicao.ExpiraEm}),
	})

	// Em desenvolvimento, retornar o token para testes locais.
	// Em producao, o token seria enviado por email; aqui apenas logamos sem expor.
	if uc.appEnv == "development" && uc.exibirTokenDev {
		return token, nil
	}

	return "", nil
}

// RedefinirSenha valida o token e define uma nova senha para o usuario.
func (uc *SenhaSessaoUsecase) RedefinirSenha(ctx context.Context, input domain.RedefinirSenhaInput) error {
	// Validar forca minima da nova senha.
	if err := validarForcaSenha(input.NovaSenha); err != nil {
		return err
	}

	// Buscar token valido (nao usado e nao expirado).
	redefinicao, err := uc.senhaRedefinicaoRepo.GetByTokenHash(ctx, input.Token)
	if err != nil {
		return ErrSenhaRedefinicaoNaoEncontrada
	}
	if redefinicao == nil {
		return ErrSenhaRedefinicaoNaoEncontrada
	}
	if redefinicao.Usado {
		return ErrSenhaRedefinicaoJaUsada
	}

	// Comparar o token puro com o hash armazenado.
	if err := bcrypt.CompareHashAndPassword([]byte(redefinicao.TokenHash), []byte(input.Token)); err != nil {
		return ErrSenhaRedefinicaoNaoEncontrada
	}

	// Buscar usuario.
	usuario, err := uc.usuarioRepo.GetByID(ctx, redefinicao.UsuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrSenhaRedefinicaoNaoEncontrada
	}
	if !usuario.Ativo {
		return ErrUsuarioInativo
	}

	// Gerar hash da nova senha.
	novoHash, err := uc.hashSenha(input.NovaSenha)
	if err != nil {
		return err
	}

	// Atualizar senha e marcar token como usado.
	if err := uc.usuarioRepo.UpdateSenha(ctx, usuario.ID, novoHash); err != nil {
		return err
	}
	if err := uc.senhaRedefinicaoRepo.MarcarUsado(ctx, redefinicao.ID); err != nil {
		return err
	}

	// Invalidar quaisquer tokens de redefinicao restantes para este usuario.
	if err := uc.senhaRedefinicaoRepo.InvalidarPorUsuario(ctx, usuario.ID); err != nil {
		return err
	}

	// Revogar todas as sessoes ativas do usuario apos redefinicao de senha.
	if err := uc.sessaoRevogadaRepo.RevogarPorUsuario(ctx, usuario.ID, nil, "redefinicao_senha"); err != nil {
		return err
	}

	// Auditoria (sem senha).
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuario.ID,
		Entidade:   "usuario",
		EntidadeID: usuario.ID,
		Acao:       "redefinir_senha",
		Depois:     toJSONPtr(map[string]interface{}{"usuario_id": usuario.ID}),
	})

	return nil
}

// AlterarSenha altera a senha de um usuario autenticado.
func (uc *SenhaSessaoUsecase) AlterarSenha(ctx context.Context, usuarioID uuid.UUID, input domain.AlterarSenhaInput) error {
	// Validar forca minima da nova senha.
	if err := validarForcaSenha(input.NovaSenha); err != nil {
		return err
	}

	// Nova senha nao pode ser igual à atual.
	if input.NovaSenha == input.SenhaAtual {
		return ErrNovaSenhaIgualAtual
	}

	// Buscar usuario.
	usuario, err := uc.usuarioRepo.GetByID(ctx, usuarioID)
	if err != nil {
		return err
	}
	if usuario == nil {
		return ErrUsuarioNaoEncontrado
	}
	if !usuario.Ativo {
		return ErrUsuarioInativo
	}

	// Validar senha atual.
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(input.SenhaAtual)); err != nil {
		return ErrSenhaAtualInvalida
	}

	// Gerar hash da nova senha.
	novoHash, err := uc.hashSenha(input.NovaSenha)
	if err != nil {
		return err
	}

	// Atualizar senha.
	if err := uc.usuarioRepo.UpdateSenha(ctx, usuario.ID, novoHash); err != nil {
		return err
	}

	// Revogar todas as sessoes ativas, exceto a atual (se o JTI for fornecido via contexto).
	jtiAtual := getJTIFromContext(ctx)
	if err := uc.sessaoRevogadaRepo.RevogarPorUsuario(ctx, usuarioID, jtiAtual, "alteracao_senha"); err != nil {
		return err
	}

	// Auditoria (sem senha).
	uc.auditar(ctx, domain.Auditoria{
		EmpresaID:  &usuario.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "usuario",
		EntidadeID: usuarioID,
		Acao:       "alterar_senha",
		Depois:     toJSONPtr(map[string]interface{}{"usuario_id": usuarioID}),
	})

	return nil
}

// Logout revoga o token JWT atual do usuario.
func (uc *SenhaSessaoUsecase) Logout(ctx context.Context, usuarioID uuid.UUID, jti string, expiraEm time.Time) error {
	if jti == "" {
		return ErrSessaoNaoEncontrada
	}

	sessao := &domain.SessaoRevogada{
		UsuarioID:  usuarioID,
		JTI:        jti,
		ExpiraEm:   expiraEm,
		RevogadoEm: time.Now(),
		Motivo:     "logout",
	}

	if err := uc.sessaoRevogadaRepo.Create(ctx, sessao); err != nil {
		return err
	}

	// Auditoria.
	usuario, err := uc.usuarioRepo.GetByID(ctx, usuarioID)
	if err != nil {
		return err
	}
	if usuario != nil {
		uc.auditar(ctx, domain.Auditoria{
			EmpresaID:  &usuario.EmpresaID,
			UsuarioID:  &usuarioID,
			Entidade:   "sessao",
			EntidadeID: usuarioID,
			Acao:       "logout",
			Depois:     toJSONPtr(map[string]interface{}{"usuario_id": usuarioID, "jti": jti}),
		})
	}

	return nil
}

// RevogarTodasSessoesUsuario revoga todas as sessoes ativas de um usuario (ex: desativacao).
func (uc *SenhaSessaoUsecase) RevogarTodasSessoesUsuario(ctx context.Context, usuarioID uuid.UUID, motivo string) error {
	return uc.sessaoRevogadaRepo.RevogarPorUsuario(ctx, usuarioID, nil, motivo)
}

// ExisteSessaoRevogadaPorJTI verifica se um token JTI foi revogado.
func (uc *SenhaSessaoUsecase) ExisteSessaoRevogadaPorJTI(ctx context.Context, jti string) (bool, error) {
	return uc.sessaoRevogadaRepo.ExistePorJTI(ctx, jti)
}

// validarForcaSenha verifica se a senha atende ao requisito minimo.
func validarForcaSenha(senha string) error {
	if len(senha) < MinSenhaLength {
		return errors.New("senha deve ter pelo menos 8 caracteres")
	}
	return nil
}

// bcryptHash gera o hash bcrypt de uma senha.
func bcryptHash(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// gerarTokenHex gera um token aleatorio criptograficamente seguro (hex).
func gerarTokenHex() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hexEncode(b), nil
}

// hexEncode converte bytes para string hexadecimal.
func hexEncode(b []byte) string {
	const chars = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = chars[v>>4]
		out[i*2+1] = chars[v&0x0f]
	}
	return string(out)
}

// getJTIFromContext extrai o JTI do contexto (definido pelo middleware de auth).
func getJTIFromContext(ctx context.Context) *string {
	if v := ctx.Value(jtiContextKey); v != nil {
		if s, ok := v.(string); ok {
			return &s
		}
	}
	return nil
}

var jtiContextKey = struct{}{}

func (uc *SenhaSessaoUsecase) auditar(ctx context.Context, a domain.Auditoria) {
	_ = uc.auditoriaRepo.Create(ctx, &a)
}
