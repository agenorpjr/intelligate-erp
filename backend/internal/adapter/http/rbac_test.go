package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthUsecase implementa AuthUsecaseInterface para testes
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) GetPermissoesUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, usuarioID, empresaID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthUsecase) GetLojasUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, usuarioID, empresaID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockAuthUsecase) GetUsuarioFromToken(ctx context.Context, tokenString string) (*domain.Usuario, error) {
	args := m.Called(ctx, tokenString)
	return args.Get(0).(*domain.Usuario), args.Error(1)
}

func TestRBACMiddleware_RequirePermission(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	permissoes := []string{domain.PermLojasCriar, domain.PermLojasListar}

	mockAuth.On("GetPermissoesUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(permissoes, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequirePermission(domain.PermLojasCriar)(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lojas", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequirePermission_Forbidden(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	permissoes := []string{domain.PermLojasListar}

	mockAuth.On("GetPermissoesUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(permissoes, nil).Once()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequirePermission(domain.PermLojasCriar)(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lojas", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireAnyPermission(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	permissoes := []string{domain.PermLojasListar}

	mockAuth.On("GetPermissoesUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(permissoes, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequireAnyPermission(domain.PermLojasCriar, domain.PermLojasListar)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/lojas", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccess(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	lojaID := uuid.New()
	lojasPermitidas := []uuid.UUID{lojaID}

	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(lojasPermitidas, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequireLojaAccess("id")(testHandler)

	// Criar request com chi route context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lojaID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lojas/"+lojaID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccess_Forbidden(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	lojaID := uuid.New()
	lojasPermitidas := []uuid.UUID{uuid.New()}

	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(lojasPermitidas, nil).Once()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequireLojaAccess("id")(testHandler)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lojaID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lojas/"+lojaID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccess_GlobalAccess(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@test.com",
		Nome:      "Test User",
		Ativo:     true,
	}

	lojaID := uuid.New()
	lojasPermitidas := []uuid.UUID{}

	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return(lojasPermitidas, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrappedHandler := rbac.RequireLojaAccess("id")(testHandler)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lojaID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lojas/"+lojaID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}
