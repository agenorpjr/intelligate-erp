package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/intelligate/erp/backend/internal/domain"
)

// --- Mocks para SenhaSessaoUsecase ---

type mockSenhaRedefinicaoRepo struct {
	tokens     map[string]*domain.SenhaRedefinicao
	porUsuario map[uuid.UUID][]*domain.SenhaRedefinicao
}

func newMockSenhaRedefinicaoRepo() *mockSenhaRedefinicaoRepo {
	return &mockSenhaRedefinicaoRepo{
		tokens:     make(map[string]*domain.SenhaRedefinicao),
		porUsuario: make(map[uuid.UUID][]*domain.SenhaRedefinicao),
	}
}

func (m *mockSenhaRedefinicaoRepo) Create(ctx context.Context, s *domain.SenhaRedefinicao) error {
	s.ID = uuid.New()
	m.tokens[s.TokenHash] = s
	m.porUsuario[s.UsuarioID] = append(m.porUsuario[s.UsuarioID], s)
	return nil
}

func (m *mockSenhaRedefinicaoRepo) GetByTokenHash(ctx context.Context, tokenPuro string) (*domain.SenhaRedefinicao, error) {
	for hash, t := range m.tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tokenPuro)); err == nil {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockSenhaRedefinicaoRepo) InvalidarPorUsuario(ctx context.Context, usuarioID uuid.UUID) error {
	for _, t := range m.porUsuario[usuarioID] {
		t.Usado = true
	}
	return nil
}

func (m *mockSenhaRedefinicaoRepo) MarcarUsado(ctx context.Context, id uuid.UUID) error {
	for _, t := range m.tokens {
		if t.ID == id {
			t.Usado = true
			return nil
		}
	}
	return nil
}

func (m *mockSenhaRedefinicaoRepo) ListarValidasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SenhaRedefinicao, error) {
	return m.porUsuario[usuarioID], nil
}

type mockSessaoRevogadaRepo struct {
	revogadas  map[string]*domain.SessaoRevogada
	porUsuario map[uuid.UUID][]*domain.SessaoRevogada
}

func newMockSessaoRevogadaRepo() *mockSessaoRevogadaRepo {
	return &mockSessaoRevogadaRepo{
		revogadas:  make(map[string]*domain.SessaoRevogada),
		porUsuario: make(map[uuid.UUID][]*domain.SessaoRevogada),
	}
}

func (m *mockSessaoRevogadaRepo) Create(ctx context.Context, s *domain.SessaoRevogada) error {
	s.ID = uuid.New()
	m.revogadas[s.JTI] = s
	m.porUsuario[s.UsuarioID] = append(m.porUsuario[s.UsuarioID], s)
	return nil
}

func (m *mockSessaoRevogadaRepo) ExistePorJTI(ctx context.Context, jti string) (bool, error) {
	_, ok := m.revogadas[jti]
	return ok, nil
}

func (m *mockSessaoRevogadaRepo) RevogarPorUsuario(ctx context.Context, usuarioID uuid.UUID, excecaoJTI *string, motivo string) error {
	for _, s := range m.porUsuario[usuarioID] {
		if excecaoJTI != nil && s.JTI == *excecaoJTI {
			continue
		}
		s.ExpiraEm = time.Now()
	}
	return nil
}

func (m *mockSessaoRevogadaRepo) ListarAtivasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SessaoRevogada, error) {
	var ativas []*domain.SessaoRevogada
	for _, s := range m.porUsuario[usuarioID] {
		if s.ExpiraEm.After(time.Now()) {
			ativas = append(ativas, s)
		}
	}
	return ativas, nil
}

func (m *mockSessaoRevogadaRepo) LimparExpiradas(ctx context.Context) error {
	return nil
}

type mockAuditoriaRepo struct {
	criadas []domain.Auditoria
}

func (m *mockAuditoriaRepo) Create(ctx context.Context, a *domain.Auditoria) error {
	m.criadas = append(m.criadas, *a)
	return nil
}

func (m *mockAuditoriaRepo) List(ctx context.Context, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	return nil, nil
}

// --- Testes ---

func TestSenhaSessaoUsecase_EsqueciSenha_UsuarioNaoExiste(t *testing.T) {
	uc := NewSenhaSessaoUsecase(
		&mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)},
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	token, err := uc.EsqueciSenha(context.Background(), domain.EsqueciSenhaInput{Email: "naoexiste@test.com"})
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestSenhaSessaoUsecase_EsqueciSenha_UsuarioInativo(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["inativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "inativo@test.com",
		Nome:      "Inativo",
		Ativo:     false,
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	token, err := uc.EsqueciSenha(context.Background(), domain.EsqueciSenhaInput{Email: "inativo@test.com"})
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestSenhaSessaoUsecase_EsqueciSenha_Sucesso(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	token, err := uc.EsqueciSenha(context.Background(), domain.EsqueciSenhaInput{Email: "ativo@test.com"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestSenhaSessaoUsecase_RedefinirSenha_TokenInvalido(t *testing.T) {
	uc := NewSenhaSessaoUsecase(
		&mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)},
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.RedefinirSenha(context.Background(), domain.RedefinirSenhaInput{
		Token:     "token-invalido",
		NovaSenha: "NovaSenha123",
	})
	assert.ErrorIs(t, err, ErrSenhaRedefinicaoNaoEncontrada)
}

func TestSenhaSessaoUsecase_RedefinirSenha_SenhaFraca(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
		SenhaHash: "$2a$10$hashfalso",
	}

	senhaRepo := newMockSenhaRedefinicaoRepo()
	tokenHash, _ := bcrypt.GenerateFromPassword([]byte("tokenvalido123"), bcrypt.DefaultCost)
	redefinicao := &domain.SenhaRedefinicao{
		ID:        uuid.New(),
		UsuarioID: usuarioID,
		TokenHash: string(tokenHash),
		Usado:     false,
		ExpiraEm:  time.Now().Add(30 * time.Minute),
	}
	_ = senhaRepo.Create(context.Background(), redefinicao)

	uc := NewSenhaSessaoUsecase(
		repo,
		senhaRepo,
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.RedefinirSenha(context.Background(), domain.RedefinirSenhaInput{
		Token:     "tokenvalido123",
		NovaSenha: "123",
	})
	assert.Error(t, err)
}

func TestSenhaSessaoUsecase_AlterarSenha_Sucesso(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	senhaAtualHash, _ := bcryptHash("senhaAtual123")
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
		SenhaHash: senhaAtualHash,
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.AlterarSenha(context.Background(), usuarioID, domain.AlterarSenhaInput{
		SenhaAtual: "senhaAtual123",
		NovaSenha:  "NovaSenha123!",
	})
	require.NoError(t, err)
}

func TestSenhaSessaoUsecase_AlterarSenha_SenhaAtualInvalida(t *testing.T) {
	usuarioID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: uuid.New(),
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
		SenhaHash: "$2a$10$hashfalso",
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.AlterarSenha(context.Background(), usuarioID, domain.AlterarSenhaInput{
		SenhaAtual: "senhaErrada",
		NovaSenha:  "NovaSenha123!",
	})
	assert.ErrorIs(t, err, ErrSenhaAtualInvalida)
}

func TestSenhaSessaoUsecase_AlterarSenha_NovaIgualAtual(t *testing.T) {
	usuarioID := uuid.New()
	senhaHash, _ := bcryptHash("senha123")
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: uuid.New(),
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
		SenhaHash: senhaHash,
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.AlterarSenha(context.Background(), usuarioID, domain.AlterarSenhaInput{
		SenhaAtual: "senha123",
		NovaSenha:  "senha123",
	})
	assert.ErrorIs(t, err, ErrNovaSenhaIgualAtual)
}

func TestSenhaSessaoUsecase_Logout_Sucesso(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
	}

	uc := NewSenhaSessaoUsecase(
		repo,
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.Logout(context.Background(), usuarioID, "jti-123", time.Now().Add(time.Hour))
	require.NoError(t, err)
}

func TestSenhaSessaoUsecase_Logout_JTIvazio(t *testing.T) {
	uc := NewSenhaSessaoUsecase(
		&mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)},
		newMockSenhaRedefinicaoRepo(),
		newMockSessaoRevogadaRepo(),
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.Logout(context.Background(), uuid.New(), "", time.Now())
	assert.ErrorIs(t, err, ErrSessaoNaoEncontrada)
}

func TestValidarForcaSenha(t *testing.T) {
	tests := []struct {
		senha string
		erro  bool
	}{
		{"senha123", false},
		{"12345678", false},
		{"1234567", true},
		{"", true},
	}

	for _, tt := range tests {
		err := validarForcaSenha(tt.senha)
		if tt.erro {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestSenhaSessaoUsecase_RedefinirSenha_RevogaSessoes(t *testing.T) {
	usuarioID := uuid.New()
	empresaID := uuid.New()
	repo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	repo.usuarios["ativo@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "ativo@test.com",
		Nome:      "Ativo",
		Ativo:     true,
	}

	senhaRepo := newMockSenhaRedefinicaoRepo()
	tokenHash, _ := bcrypt.GenerateFromPassword([]byte("tokenvalido123"), bcrypt.DefaultCost)
	redefinicao := &domain.SenhaRedefinicao{
		ID:        uuid.New(),
		UsuarioID: usuarioID,
		TokenHash: string(tokenHash),
		Usado:     false,
		ExpiraEm:  time.Now().Add(30 * time.Minute),
	}
	_ = senhaRepo.Create(context.Background(), redefinicao)

	sessaoRepo := newMockSessaoRevogadaRepo()
	_ = sessaoRepo.Create(context.Background(), &domain.SessaoRevogada{
		UsuarioID: usuarioID,
		JTI:       "jti-existente",
		ExpiraEm:  time.Now().Add(time.Hour),
		Motivo:    "login",
	})

	uc := NewSenhaSessaoUsecase(
		repo,
		senhaRepo,
		sessaoRepo,
		&mockAuditoriaRepo{},
		"development",
		true,
	)

	err := uc.RedefinirSenha(context.Background(), domain.RedefinirSenhaInput{
		Token:     "tokenvalido123",
		NovaSenha: "NovaSenha123!",
	})
	require.NoError(t, err)

	sessoes, err := sessaoRepo.ListarAtivasPorUsuario(context.Background(), usuarioID)
	require.NoError(t, err)
	assert.Len(t, sessoes, 0)
}

// --- Testes de ultimo ADMIN (usando codigo) ---

func TestUsuarioUsecase_UltimoAdmin_PorCodigo(t *testing.T) {
	empresaID := uuid.New()
	usuarioID := uuid.New()
	papelAdminID := uuid.New()

	papelRepo := &mockPapelRepository{papeis: make(map[uuid.UUID]*domain.Papel)}
	papelRepo.papeis[papelAdminID] = &domain.Papel{
		ID:        papelAdminID,
		EmpresaID: empresaID,
		Nome:      "ADMIN",
		Codigo:    domain.CodigoPadraoAdmin,
		Sistema:   true,
	}

	usuarioRepo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	usuarioRepo.usuarios["admin@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "admin@test.com",
		Nome:      "Admin",
		Ativo:     true,
	}

	usuarioPapelRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelAdminID},
	}}

	uc := &UsuarioUsecase{
		usuarioRepo:        usuarioRepo,
		usuarioPapelRepo:   usuarioPapelRepo,
		papelRepo:          papelRepo,
		papelPermissaoRepo: &mockPapelPermissaoRepository{},
		auditoriaRepo:      &mockAuditoriaRepo{},
	}

	err := uc.Desativar(context.Background(), usuarioID, uuid.New(), "127.0.0.1", "test")
	assert.ErrorIs(t, err, ErrNaoPodeRemoverPapelAdmin)
}

func TestUsuarioUsecase_UltimoAdmin_NaoEhAdminPorCodigo(t *testing.T) {
	empresaID := uuid.New()
	usuarioID := uuid.New()
	papelComumID := uuid.New()

	papelRepo := &mockPapelRepository{papeis: make(map[uuid.UUID]*domain.Papel)}
	papelRepo.papeis[papelComumID] = &domain.Papel{
		ID:        papelComumID,
		EmpresaID: empresaID,
		Nome:      "Comum",
		Codigo:    "COMUM",
		Sistema:   false,
	}

	usuarioRepo := &mockUsuarioRepository{usuarios: make(map[string]*domain.Usuario)}
	usuarioRepo.usuarios["comum@test.com"] = &domain.Usuario{
		ID:        usuarioID,
		EmpresaID: empresaID,
		Email:     "comum@test.com",
		Nome:      "Comum",
		Ativo:     true,
	}

	usuarioPapelRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelComumID},
	}}

	uc := &UsuarioUsecase{
		usuarioRepo:        usuarioRepo,
		usuarioPapelRepo:   usuarioPapelRepo,
		papelRepo:          papelRepo,
		papelPermissaoRepo: &mockPapelPermissaoRepository{},
		auditoriaRepo:      &mockAuditoriaRepo{},
	}

	err := uc.Desativar(context.Background(), usuarioID, uuid.New(), "127.0.0.1", "test")
	assert.NoError(t, err)
}

var _ = errors.Is
