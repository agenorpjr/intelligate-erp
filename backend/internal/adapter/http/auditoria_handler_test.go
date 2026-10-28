package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type handlerMockAuditoriaRepo struct {
	lastFilter domain.AuditoriaFilter
	items      []*domain.Auditoria
}

func (m *handlerMockAuditoriaRepo) Create(ctx context.Context, a *domain.Auditoria) error {
	return nil
}

func (m *handlerMockAuditoriaRepo) List(ctx context.Context, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	m.lastFilter = filter
	return m.items, nil
}

func TestAuditoriaHandler_Listar_SemEmpresa(t *testing.T) {
	handler := NewAuditoriaHandler(usecase.NewAuditoriaUsecase(&handlerMockAuditoriaRepo{}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria", nil)
	rr := httptest.NewRecorder()
	handler.Listar(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuditoriaHandler_Listar_ForcaEmpresaDoContexto(t *testing.T) {
	empresaID := uuid.New()
	repo := &handlerMockAuditoriaRepo{items: []*domain.Auditoria{{ID: 1, Entidade: "papel", Acao: "criar"}}}
	handler := NewAuditoriaHandler(usecase.NewAuditoriaUsecase(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=5&entidade=papel", nil)
	ctx := context.WithValue(req.Context(), EmpresaContextKey, empresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Listar(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, repo.lastFilter.EmpresaID)
	assert.Equal(t, empresaID, *repo.lastFilter.EmpresaID)
	assert.Equal(t, 5, repo.lastFilter.Limit)
	require.NotNil(t, repo.lastFilter.Entidade)
	assert.Equal(t, "papel", *repo.lastFilter.Entidade)

	var body []map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.Len(t, body, 1)
}

func TestAuditoriaHandler_Listar_UsuarioIDInvalido(t *testing.T) {
	empresaID := uuid.New()
	handler := NewAuditoriaHandler(usecase.NewAuditoriaUsecase(&handlerMockAuditoriaRepo{}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?usuario_id=abc", nil)
	ctx := context.WithValue(req.Context(), EmpresaContextKey, empresaID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Listar(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
