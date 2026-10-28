package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newLojaBodyRequest(usuario *domain.Usuario, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/estoque/saidas", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), UserContextKey, usuario)
	ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
	return req.WithContext(ctx)
}

func TestRBACMiddleware_RequireLojaAccessBody_Allowed(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{ID: uuid.New(), EmpresaID: uuid.New(), Ativo: true}
	lojaID := uuid.New()
	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return([]uuid.UUID{lojaID}, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// O corpo deve continuar disponível para o handler.
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), lojaID.String())
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrapped := rbac.RequireLojaAccessBody("loja_id")(testHandler)

	body := `{"loja_id":"` + lojaID.String() + `","insumo_id":"` + uuid.New().String() + `"}`
	req := newLojaBodyRequest(usuario, body)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccessBody_Forbidden(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{ID: uuid.New(), EmpresaID: uuid.New(), Ativo: true}
	lojaID := uuid.New()
	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return([]uuid.UUID{uuid.New()}, nil).Once()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrapped := rbac.RequireLojaAccessBody("loja_id")(testHandler)

	body := `{"loja_id":"` + lojaID.String() + `"}`
	req := newLojaBodyRequest(usuario, body)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccessBody_GlobalAccess(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{ID: uuid.New(), EmpresaID: uuid.New(), Ativo: true}
	lojaID := uuid.New()
	mockAuth.On("GetLojasUsuario", mock.Anything, usuario.ID, usuario.EmpresaID).Return([]uuid.UUID{}, nil).Once()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrapped := rbac.RequireLojaAccessBody("loja_id")(testHandler)

	body := `{"loja_id":"` + lojaID.String() + `"}`
	req := newLojaBodyRequest(usuario, body)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertExpectations(t)
}

func TestRBACMiddleware_RequireLojaAccessBody_CampoAusente(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{ID: uuid.New(), EmpresaID: uuid.New(), Ativo: true}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrapped := rbac.RequireLojaAccessBody("loja_id")(testHandler)

	// Sem loja_id: o middleware não bloqueia (handler valida obrigatoriedade).
	req := newLojaBodyRequest(usuario, `{"insumo_id":"`+uuid.New().String()+`"}`)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	mockAuth.AssertNotCalled(t, "GetLojasUsuario", mock.Anything, mock.Anything, mock.Anything)
}

func TestRBACMiddleware_RequireLojaAccessBody_CorpoInvalido(t *testing.T) {
	mockAuth := new(MockAuthUsecase)
	rbac := NewRBACMiddleware(mockAuth)

	usuario := &domain.Usuario{ID: uuid.New(), EmpresaID: uuid.New(), Ativo: true}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	wrapped := rbac.RequireLojaAccessBody("loja_id")(testHandler)

	req := newLojaBodyRequest(usuario, `{invalid-json`)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
