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

type handlerMockLojaRepo struct {
	lojas map[uuid.UUID]*domain.Loja
}

func newHandlerMockLojaRepo() *handlerMockLojaRepo {
	return &handlerMockLojaRepo{lojas: make(map[uuid.UUID]*domain.Loja)}
}

func (m *handlerMockLojaRepo) Create(ctx context.Context, l *domain.Loja) error {
	m.lojas[l.ID] = l
	return nil
}

func (m *handlerMockLojaRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Loja, error) {
	if l, ok := m.lojas[id]; ok {
		return l, nil
	}
	return nil, assert.AnError
}

func (m *handlerMockLojaRepo) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Loja, error) {
	result := make([]*domain.Loja, 0)
	for _, l := range m.lojas {
		if l.EmpresaID == empresaID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *handlerMockLojaRepo) Update(ctx context.Context, l *domain.Loja) error {
	m.lojas[l.ID] = l
	return nil
}

func (m *handlerMockLojaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.lojas, id)
	return nil
}

type handlerMockDashboardRepo struct {
	lastReq *domain.DashboardRequest
}

func (m *handlerMockDashboardRepo) GetDashboard(ctx context.Context, req domain.DashboardRequest) (*domain.DashboardResponse, error) {
	m.lastReq = &req
	return &domain.DashboardResponse{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	}, nil
}

func newDashboardTestRequest(empresaID uuid.UUID, target string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	ctx := context.WithValue(req.Context(), EmpresaContextKey, empresaID)
	return req.WithContext(ctx)
}

func TestDashboardHandler_GetDashboard_Validacao(t *testing.T) {
	empresaID := uuid.New()
	lojaID := uuid.New()

	tests := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{
			name:       "sem loja_id",
			target:     "/api/v1/dashboard?data_inicio=2026-01-01&data_fim=2026-01-31",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "loja_id inválido",
			target:     "/api/v1/dashboard?loja_id=abc&data_inicio=2026-01-01&data_fim=2026-01-31",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "sem datas",
			target:     "/api/v1/dashboard?loja_id=" + lojaID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "data inválida",
			target:     "/api/v1/dashboard?loja_id=" + lojaID.String() + "&data_inicio=invalida&data_fim=2026-01-31",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lojaRepo := newHandlerMockLojaRepo()
			require.NoError(t, lojaRepo.Create(context.Background(), &domain.Loja{
				ID:        lojaID,
				EmpresaID: empresaID,
				Ativo:     true,
			}))

			handler := NewDashboardHandler(usecase.NewDashboardUsecase(&handlerMockDashboardRepo{}, lojaRepo))

			w := httptest.NewRecorder()
			handler.GetDashboard(w, newDashboardTestRequest(empresaID, tt.target))

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDashboardHandler_GetDashboard_Success(t *testing.T) {
	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newHandlerMockLojaRepo()
	require.NoError(t, lojaRepo.Create(context.Background(), &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Ativo:     true,
	}))

	dashboardRepo := &handlerMockDashboardRepo{}
	handler := NewDashboardHandler(usecase.NewDashboardUsecase(dashboardRepo, lojaRepo))

	target := "/api/v1/dashboard?loja_id=" + lojaID.String() +
		"&data_inicio=2026-01-01&data_fim=2026-01-31"

	w := httptest.NewRecorder()
	handler.GetDashboard(w, newDashboardTestRequest(empresaID, target))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp domain.DashboardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, empresaID, resp.EmpresaID)
	assert.Equal(t, lojaID, resp.LojaID)

	require.NotNil(t, dashboardRepo.lastReq)
	assert.Equal(t, empresaID, dashboardRepo.lastReq.EmpresaID)
	assert.Equal(t, lojaID, dashboardRepo.lastReq.LojaID)

	// data_inicio (date-only) deve começar no início do dia.
	assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), dashboardRepo.lastReq.DataInicio)
	// data_fim (date-only) deve ser estendida até o fim do dia.
	assert.Equal(t, time.Date(2026, 1, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC), dashboardRepo.lastReq.DataFim)
}

func TestDashboardHandler_GetDashboard_LojaNaoPertence(t *testing.T) {
	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newHandlerMockLojaRepo()
	require.NoError(t, lojaRepo.Create(context.Background(), &domain.Loja{
		ID:        lojaID,
		EmpresaID: uuid.New(),
		Ativo:     true,
	}))

	handler := NewDashboardHandler(usecase.NewDashboardUsecase(&handlerMockDashboardRepo{}, lojaRepo))

	target := "/api/v1/dashboard?loja_id=" + lojaID.String() +
		"&data_inicio=2026-01-01&data_fim=2026-01-31"

	w := httptest.NewRecorder()
	handler.GetDashboard(w, newDashboardTestRequest(empresaID, target))

	assert.Equal(t, http.StatusNotFound, w.Code)
}
