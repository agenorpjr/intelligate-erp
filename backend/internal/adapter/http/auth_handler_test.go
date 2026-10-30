package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// --- Mocks mínimos de identidade para exercitar o AuthUsecase ---

type authMockUsuarioRepo struct {
	usuarios map[uuid.UUID]*domain.Usuario
}

func (m *authMockUsuarioRepo) Create(ctx context.Context, u *domain.Usuario) error { return nil }
func (m *authMockUsuarioRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	return m.usuarios[id], nil
}
func (m *authMockUsuarioRepo) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	return nil, nil
}
func (m *authMockUsuarioRepo) GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*domain.Usuario, error) {
	return nil, nil
}
func (m *authMockUsuarioRepo) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Usuario, error) {
	return nil, nil
}
func (m *authMockUsuarioRepo) Update(ctx context.Context, u *domain.Usuario) error { return nil }
func (m *authMockUsuarioRepo) UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error {
	return nil
}
func (m *authMockUsuarioRepo) UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error { return nil }
func (m *authMockUsuarioRepo) Delete(ctx context.Context, id uuid.UUID) error            { return nil }
func (m *authMockUsuarioRepo) AddLoja(ctx context.Context, usuarioID, lojaID uuid.UUID) error {
	return nil
}
func (m *authMockUsuarioRepo) RemoveLoja(ctx context.Context, usuarioID, lojaID uuid.UUID) error {
	return nil
}
func (m *authMockUsuarioRepo) ListLojas(ctx context.Context, usuarioID uuid.UUID) ([]*domain.Loja, error) {
	return nil, nil
}
func (m *authMockUsuarioRepo) ListUsuariosByLoja(ctx context.Context, lojaID uuid.UUID) ([]*domain.Usuario, error) {
	return nil, nil
}
func (m *authMockUsuarioRepo) CheckLojaAccess(ctx context.Context, usuarioID, lojaID uuid.UUID) (bool, error) {
	return false, nil
}

type authMockUsuarioPapelRepo struct {
	associacoes []*domain.UsuarioPapelComDetalhes
}

func (m *authMockUsuarioPapelRepo) Create(ctx context.Context, up *domain.UsuarioPapel) error {
	return nil
}
func (m *authMockUsuarioPapelRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UsuarioPapel, error) {
	return nil, nil
}
func (m *authMockUsuarioPapelRepo) ListByUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.UsuarioPapelComDetalhes, error) {
	var result []*domain.UsuarioPapelComDetalhes
	for _, a := range m.associacoes {
		if a.UsuarioID == usuarioID {
			result = append(result, a)
		}
	}
	return result, nil
}
func (m *authMockUsuarioPapelRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type authMockPapelRepo struct {
	papeis map[uuid.UUID]*domain.Papel
}

func (m *authMockPapelRepo) Create(ctx context.Context, p *domain.Papel) error { return nil }
func (m *authMockPapelRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Papel, error) {
	return m.papeis[id], nil
}
func (m *authMockPapelRepo) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Papel, error) {
	return nil, nil
}
func (m *authMockPapelRepo) Update(ctx context.Context, p *domain.Papel) error { return nil }
func (m *authMockPapelRepo) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

type authMockPapelPermissaoRepo struct {
	porPapel map[uuid.UUID][]*domain.Permissao
}

func (m *authMockPapelPermissaoRepo) Create(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return nil
}
func (m *authMockPapelPermissaoRepo) Delete(ctx context.Context, papelID, permissaoID uuid.UUID) error {
	return nil
}
func (m *authMockPapelPermissaoRepo) ListPermissoesByPapel(ctx context.Context, papelID uuid.UUID) ([]*domain.Permissao, error) {
	return m.porPapel[papelID], nil
}
func (m *authMockPapelPermissaoRepo) ListPermissoesByUsuario(ctx context.Context, usuarioID, lojaID uuid.UUID) ([]*domain.Permissao, error) {
	return nil, nil
}

func TestAuthHandler_MinhasPermissoes(t *testing.T) {
	empresaID := uuid.New()
	usuarioID := uuid.New()
	papelID := uuid.New()

	usuarioRepo := &authMockUsuarioRepo{usuarios: map[uuid.UUID]*domain.Usuario{
		usuarioID: {ID: usuarioID, EmpresaID: empresaID, Email: "u@test.com", Ativo: true},
	}}
	upRepo := &authMockUsuarioPapelRepo{associacoes: []*domain.UsuarioPapelComDetalhes{
		{ID: uuid.New(), UsuarioID: usuarioID, PapelID: papelID},
	}}
	papelRepo := &authMockPapelRepo{papeis: map[uuid.UUID]*domain.Papel{
		papelID: {ID: papelID, EmpresaID: empresaID, Nome: "Admin", Codigo: "ADMIN"},
	}}
	ppRepo := &authMockPapelPermissaoRepo{porPapel: map[uuid.UUID][]*domain.Permissao{
		papelID: {{Chave: domain.PermLojasListar}, {Chave: domain.PermDashboardVer}},
	}}

	uc := usecase.NewAuthUsecase(usuarioRepo, upRepo, papelRepo, ppRepo, "secret", time.Hour)
	senhaSessaoUC := usecase.NewSenhaSessaoUsecase(
		usuarioRepo,
		&mockSenhaRedefinicaoRepoForHandler{},
		&mockSessaoRevogadaRepoForHandler{},
		&mockAuditoriaRepoForHandler{},
		"development",
		true,
	)
	handler := NewAuthHandler(uc, senhaSessaoUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissoes", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, &domain.Usuario{ID: usuarioID, EmpresaID: empresaID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.MinhasPermissoes(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var body struct {
		Permissoes []string `json:"permissoes"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.Equal(t, []string{domain.PermDashboardVer, domain.PermLojasListar}, body.Permissoes)
}

func TestAuthHandler_MinhasPermissoes_SemUsuario(t *testing.T) {
	uc := usecase.NewAuthUsecase(&authMockUsuarioRepo{}, &authMockUsuarioPapelRepo{}, &authMockPapelRepo{}, &authMockPapelPermissaoRepo{}, "secret", time.Hour)
	senhaSessaoUC := usecase.NewSenhaSessaoUsecase(
		&authMockUsuarioRepo{},
		&mockSenhaRedefinicaoRepoForHandler{},
		&mockSessaoRevogadaRepoForHandler{},
		&mockAuditoriaRepoForHandler{},
		"development",
		true,
	)
	handler := NewAuthHandler(uc, senhaSessaoUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissoes", nil)
	rr := httptest.NewRecorder()
	handler.MinhasPermissoes(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- Mocks auxiliares para o handler ---

type mockSenhaRedefinicaoRepoForHandler struct{}

func (m *mockSenhaRedefinicaoRepoForHandler) Create(ctx context.Context, s *domain.SenhaRedefinicao) error {
	return nil
}
func (m *mockSenhaRedefinicaoRepoForHandler) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.SenhaRedefinicao, error) {
	return nil, nil
}
func (m *mockSenhaRedefinicaoRepoForHandler) InvalidarPorUsuario(ctx context.Context, usuarioID uuid.UUID) error {
	return nil
}
func (m *mockSenhaRedefinicaoRepoForHandler) MarcarUsado(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockSenhaRedefinicaoRepoForHandler) ListarValidasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SenhaRedefinicao, error) {
	return nil, nil
}

type mockSessaoRevogadaRepoForHandler struct{}

func (m *mockSessaoRevogadaRepoForHandler) Create(ctx context.Context, s *domain.SessaoRevogada) error {
	return nil
}
func (m *mockSessaoRevogadaRepoForHandler) ExistePorJTI(ctx context.Context, jti string) (bool, error) {
	return false, nil
}
func (m *mockSessaoRevogadaRepoForHandler) RevogarPorUsuario(ctx context.Context, usuarioID uuid.UUID, excecaoJTI *string, motivo string) error {
	return nil
}
func (m *mockSessaoRevogadaRepoForHandler) ListarAtivasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SessaoRevogada, error) {
	return nil, nil
}
func (m *mockSessaoRevogadaRepoForHandler) LimparExpiradas(ctx context.Context) error { return nil }

type mockAuditoriaRepoForHandler struct{}

func (m *mockAuditoriaRepoForHandler) Create(ctx context.Context, a *domain.Auditoria) error {
	return nil
}
func (m *mockAuditoriaRepoForHandler) List(ctx context.Context, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	return nil, nil
}
