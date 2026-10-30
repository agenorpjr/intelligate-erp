package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
)

type mockUsuarioRepository struct {
	usuarios map[string]*domain.Usuario
	lojas    map[string][]*domain.Loja
}

func newMockUsuarioRepository() *mockUsuarioRepository {
	return &mockUsuarioRepository{
		usuarios: make(map[string]*domain.Usuario),
		lojas:    make(map[string][]*domain.Loja),
	}
}

func (m *mockUsuarioRepository) Create(ctx context.Context, u *domain.Usuario) error {
	m.usuarios[u.Email] = u
	return nil
}

func (m *mockUsuarioRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	for _, u := range m.usuarios {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUsuarioRepository) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	if u, ok := m.usuarios[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUsuarioRepository) GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*domain.Usuario, error) {
	if u, ok := m.usuarios[email]; ok && u.EmpresaID == empresaID {
		return u, nil
	}
	return nil, nil
}

func (m *mockUsuarioRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Usuario, error) {
	var result []*domain.Usuario
	for _, u := range m.usuarios {
		if u.EmpresaID == empresaID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUsuarioRepository) Update(ctx context.Context, u *domain.Usuario) error {
	m.usuarios[u.Email] = u
	return nil
}

func (m *mockUsuarioRepository) UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error {
	for _, u := range m.usuarios {
		if u.ID == id {
			u.SenhaHash = senhaHash
			return nil
		}
	}
	return nil
}

func (m *mockUsuarioRepository) UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error {
	for _, u := range m.usuarios {
		if u.ID == id {
			now := time.Now()
			u.UltimoLogin = &now
			return nil
		}
	}
	return nil
}

func (m *mockUsuarioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for email, u := range m.usuarios {
		if u.ID == id {
			delete(m.usuarios, email)
			return nil
		}
	}
	return nil
}

func (m *mockUsuarioRepository) AddLoja(ctx context.Context, usuarioID, lojaID uuid.UUID) error {
	return nil
}

func (m *mockUsuarioRepository) RemoveLoja(ctx context.Context, usuarioID, lojaID uuid.UUID) error {
	return nil
}

func (m *mockUsuarioRepository) ListLojas(ctx context.Context, usuarioID uuid.UUID) ([]*domain.Loja, error) {
	return nil, nil
}

func (m *mockUsuarioRepository) ListUsuariosByLoja(ctx context.Context, lojaID uuid.UUID) ([]*domain.Usuario, error) {
	return nil, nil
}

func (m *mockUsuarioRepository) CheckLojaAccess(ctx context.Context, usuarioID, lojaID uuid.UUID) (bool, error) {
	return false, nil
}

// mockUsuarioPapelRepository implementa domain.UsuarioPapelRepository para testes.
type mockUsuarioPapelRepository struct {
	associacoes []*domain.UsuarioPapelComDetalhes
}

func (m *mockUsuarioPapelRepository) Create(ctx context.Context, up *domain.UsuarioPapel) error {
	m.associacoes = append(m.associacoes, &domain.UsuarioPapelComDetalhes{
		ID:        up.ID,
		UsuarioID: up.UsuarioID,
		PapelID:   up.PapelID,
		LojaID:    up.LojaID,
	})
	return nil
}

func (m *mockUsuarioPapelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UsuarioPapel, error) {
	for _, a := range m.associacoes {
		if a.ID == id {
			return &domain.UsuarioPapel{
				ID:        a.ID,
				UsuarioID: a.UsuarioID,
				PapelID:   a.PapelID,
				LojaID:    a.LojaID,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockUsuarioPapelRepository) ListByUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.UsuarioPapelComDetalhes, error) {
	var result []*domain.UsuarioPapelComDetalhes
	for _, a := range m.associacoes {
		if a.UsuarioID == usuarioID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockUsuarioPapelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for i, a := range m.associacoes {
		if a.ID == id {
			m.associacoes = append(m.associacoes[:i], m.associacoes[i+1:]...)
			return nil
		}
	}
	return nil
}

// mockPapelRepository implementa domain.PapelRepository para testes.
type mockPapelRepository struct {
	papeis map[uuid.UUID]*domain.Papel
}

func newMockPapelRepository() *mockPapelRepository {
	return &mockPapelRepository{papeis: make(map[uuid.UUID]*domain.Papel)}
}

func (m *mockPapelRepository) Create(ctx context.Context, p *domain.Papel) error {
	m.papeis[p.ID] = p
	return nil
}

func (m *mockPapelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Papel, error) {
	if p, ok := m.papeis[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPapelRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Papel, error) {
	var result []*domain.Papel
	for _, p := range m.papeis {
		if p.EmpresaID == empresaID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockPapelRepository) Update(ctx context.Context, p *domain.Papel) error {
	m.papeis[p.ID] = p
	return nil
}

func (m *mockPapelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.papeis, id)
	return nil
}

// mockPapelPermissaoRepository implementa domain.PapelPermissaoRepository para testes.
type mockPapelPermissaoRepository struct {
	porPapel map[uuid.UUID][]*domain.Permissao
}

func newMockPapelPermissaoRepository() *mockPapelPermissaoRepository {
	return &mockPapelPermissaoRepository{porPapel: make(map[uuid.UUID][]*domain.Permissao)}
}

func (m *mockPapelPermissaoRepository) Create(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return nil
}

func (m *mockPapelPermissaoRepository) Delete(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return nil
}

func (m *mockPapelPermissaoRepository) ListPermissoesByPapel(ctx context.Context, papelID uuid.UUID) ([]*domain.Permissao, error) {
	return m.porPapel[papelID], nil
}

func (m *mockPapelPermissaoRepository) ListPermissoesByUsuario(ctx context.Context, usuarioID, lojaID uuid.UUID) ([]*domain.Permissao, error) {
	return nil, nil
}

// newTestAuthUsecase monta um AuthUsecase com repositórios mockados.
func newTestAuthUsecase(
	usuarioRepo domain.UsuarioRepository,
	usuarioPapelRepo domain.UsuarioPapelRepository,
	papelRepo domain.PapelRepository,
	papelPermissaoRepo domain.PapelPermissaoRepository,
) *AuthUsecase {
	return NewAuthUsecase(usuarioRepo, usuarioPapelRepo, papelRepo, papelPermissaoRepo, "test-secret", 24*time.Hour)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := newTestAuthUsecase(repo, &mockUsuarioPapelRepository{}, newMockPapelRepository(), newMockPapelPermissaoRepository())

	// Create user with hashed password
	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	// Test login
	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "test@example.com", resp.Usuario.Email)
	assert.Equal(t, "Test User", resp.Usuario.Nome)
	assert.Empty(t, resp.Usuario.SenhaHash) // password should not be returned
	assert.Equal(t, int64(24*time.Hour.Seconds()), resp.ExpiresIn)
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := newTestAuthUsecase(repo, &mockUsuarioPapelRepository{}, newMockPapelRepository(), newMockPapelPermissaoRepository())

	// Test with non-existent user
	_, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "nonexistent@example.com",
		Senha: "senha123",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// Test with wrong password
	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	_, err = uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "wrongpassword",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthUsecase_Login_InactiveUser(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := newTestAuthUsecase(repo, &mockUsuarioPapelRepository{}, newMockPapelRepository(), newMockPapelPermissaoRepository())

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     false, // inactive user
	}
	repo.Create(context.Background(), usuario)

	_, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	assert.ErrorIs(t, err, ErrUsuarioInativo)
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := newTestAuthUsecase(repo, &mockUsuarioPapelRepository{}, newMockPapelRepository(), newMockPapelPermissaoRepository())

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	// Login to get token
	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	require.NoError(t, err)

	// Validate token
	_, claims, err := uc.ValidateToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, usuario.ID.String(), claims["sub"])
	assert.Equal(t, usuario.EmpresaID.String(), claims["empresa_id"])
	assert.Equal(t, usuario.Email, claims["email"])
}

func TestAuthUsecase_GetUsuarioFromToken(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := newTestAuthUsecase(repo, &mockUsuarioPapelRepository{}, newMockPapelRepository(), newMockPapelPermissaoRepository())

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	require.NoError(t, err)

	// Get user from token
	user, err := uc.GetUsuarioFromToken(context.Background(), resp.Token)
	require.NoError(t, err)
	assert.Equal(t, usuario.ID, user.ID)
	assert.Equal(t, usuario.Email, user.Email)
	assert.Empty(t, user.SenhaHash)
}

func TestAuthUsecase_GetPermissoesUsuario_AggregatesRoles(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	usuarioID := uuid.New()

	usuarioRepo := newMockUsuarioRepository()
	usuarioRepo.usuarios["u@test.com"] = &domain.Usuario{ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true}

	papelA := uuid.New()
	papelB := uuid.New()
	papelRepo := newMockPapelRepository()
	papelRepo.papeis[papelA] = &domain.Papel{ID: papelA, EmpresaID: empresaID, Nome: "A"}
	papelRepo.papeis[papelB] = &domain.Papel{ID: papelB, EmpresaID: empresaID, Nome: "B"}

	upRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelA},
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelB},
	}}

	ppRepo := newMockPapelPermissaoRepository()
	ppRepo.porPapel[papelA] = []*domain.Permissao{{Chave: domain.PermLojasListar}, {Chave: domain.PermLojasCriar}}
	ppRepo.porPapel[papelB] = []*domain.Permissao{{Chave: domain.PermLojasListar}, {Chave: domain.PermDashboardVer}}

	uc := newTestAuthUsecase(usuarioRepo, upRepo, papelRepo, ppRepo)

	permissoes, err := uc.GetPermissoesUsuario(ctx, usuarioID, empresaID)
	require.NoError(t, err)
	// Deduplicado e ordenado
	assert.Equal(t, []string{domain.PermDashboardVer, domain.PermLojasCriar, domain.PermLojasListar}, permissoes)
}

func TestAuthUsecase_GetPermissoesUsuario_IsolaPorEmpresa(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	outraEmpresaID := uuid.New()
	usuarioID := uuid.New()

	usuarioRepo := newMockUsuarioRepository()
	usuarioRepo.usuarios["u@test.com"] = &domain.Usuario{ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true}

	papelProprio := uuid.New()
	papelOutraEmpresa := uuid.New()
	papelRepo := newMockPapelRepository()
	papelRepo.papeis[papelProprio] = &domain.Papel{ID: papelProprio, EmpresaID: empresaID, Nome: "Proprio"}
	papelRepo.papeis[papelOutraEmpresa] = &domain.Papel{ID: papelOutraEmpresa, EmpresaID: outraEmpresaID, Nome: "Outra"}

	upRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelProprio},
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelOutraEmpresa},
	}}

	ppRepo := newMockPapelPermissaoRepository()
	ppRepo.porPapel[papelProprio] = []*domain.Permissao{{Chave: domain.PermLojasListar}}
	ppRepo.porPapel[papelOutraEmpresa] = []*domain.Permissao{{Chave: domain.PermUsuariosDeletar}}

	uc := newTestAuthUsecase(usuarioRepo, upRepo, papelRepo, ppRepo)

	permissoes, err := uc.GetPermissoesUsuario(ctx, usuarioID, empresaID)
	require.NoError(t, err)
	assert.Equal(t, []string{domain.PermLojasListar}, permissoes)
	assert.NotContains(t, permissoes, domain.PermUsuariosDeletar)
}

func TestAuthUsecase_GetLojasUsuario_GlobalQuandoPapelSemLoja(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	usuarioID := uuid.New()

	usuarioRepo := newMockUsuarioRepository()
	usuarioRepo.usuarios["u@test.com"] = &domain.Usuario{ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true}

	papelID := uuid.New()
	papelRepo := newMockPapelRepository()
	papelRepo.papeis[papelID] = &domain.Papel{ID: papelID, EmpresaID: empresaID, Nome: "Admin"}

	upRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelID, LojaID: nil},
	}}

	uc := newTestAuthUsecase(usuarioRepo, upRepo, papelRepo, newMockPapelPermissaoRepository())

	lojas, err := uc.GetLojasUsuario(ctx, usuarioID, empresaID)
	require.NoError(t, err)
	assert.Empty(t, lojas) // vazio = acesso global
}

func TestAuthUsecase_GetLojasUsuario_RetornaLojasEspecificas(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	usuarioID := uuid.New()
	lojaID := uuid.New()

	usuarioRepo := newMockUsuarioRepository()
	usuarioRepo.usuarios["u@test.com"] = &domain.Usuario{ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true}

	papelID := uuid.New()
	papelRepo := newMockPapelRepository()
	papelRepo.papeis[papelID] = &domain.Papel{ID: papelID, EmpresaID: empresaID, Nome: "Gerente"}

	upRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelID, LojaID: &lojaID},
	}}

	uc := newTestAuthUsecase(usuarioRepo, upRepo, papelRepo, newMockPapelPermissaoRepository())

	lojas, err := uc.GetLojasUsuario(ctx, usuarioID, empresaID)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{lojaID}, lojas)
}

func TestAuthUsecase_GetLojasUsuario_IsolaPorEmpresa(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	outraEmpresaID := uuid.New()
	usuarioID := uuid.New()
	lojaID := uuid.New()

	usuarioRepo := newMockUsuarioRepository()
	usuarioRepo.usuarios["u@test.com"] = &domain.Usuario{ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true}

	papelOutraEmpresa := uuid.New()
	papelRepo := newMockPapelRepository()
	papelRepo.papeis[papelOutraEmpresa] = &domain.Papel{ID: papelOutraEmpresa, EmpresaID: outraEmpresaID, Nome: "Outra"}

	upRepo := &mockUsuarioPapelRepository{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelOutraEmpresa, LojaID: &lojaID},
	}}

	uc := newTestAuthUsecase(usuarioRepo, upRepo, papelRepo, newMockPapelPermissaoRepository())

	lojas, err := uc.GetLojasUsuario(ctx, usuarioID, empresaID)
	require.NoError(t, err)
	assert.Empty(t, lojas)
}
