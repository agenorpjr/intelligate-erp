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

type mockDashboardRepository struct {
	lastReq  domain.DashboardRequest
	response *domain.DashboardResponse
	err      error
}

func (m *mockDashboardRepository) GetDashboard(ctx context.Context, req domain.DashboardRequest) (*domain.DashboardResponse, error) {
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	if m.response == nil {
		return &domain.DashboardResponse{}, nil
	}
	return m.response, nil
}

func TestDashboardUsecase_GetDashboard_Success(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newMockLojaRepository()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Nome:      "Loja Centro",
		Ativo:     true,
	}))

	repo := &mockDashboardRepository{
		response: &domain.DashboardResponse{
			EmpresaID: empresaID,
			LojaID:    lojaID,
			Vendas: domain.DashboardVendas{
				ComparacaoPeriodo: domain.ComparacaoPeriodo{
					PeriodoAtual:    domain.PeriodoComparacao{TotalVendas: 120, TotalPedidos: 12, TicketMedio: 10},
					PeriodoAnterior: domain.PeriodoComparacao{TotalVendas: 100, TotalPedidos: 10, TicketMedio: 8},
				},
			},
		},
	}

	uc := NewDashboardUsecase(repo, lojaRepo)

	dataInicio := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	dataFim := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	resp, err := uc.GetDashboard(ctx, DashboardRequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Variação de vendas: (120-100)/100 = 20%
	assert.InDelta(t, 20, resp.Vendas.VariacaoPercentual.Vendas, 1e-9)
	// Variação de pedidos: (12-10)/10 = 20%
	assert.InDelta(t, 20, resp.Vendas.VariacaoPercentual.Pedidos, 1e-9)
	// Variação de ticket médio: (10-8)/8 = 25%
	assert.InDelta(t, 25, resp.Vendas.VariacaoPercentual.TicketMedio, 1e-9)

	// Período anterior: mesma duração, imediatamente antes do período atual.
	assert.Equal(t, empresaID, repo.lastReq.EmpresaID)
	assert.Equal(t, lojaID, repo.lastReq.LojaID)
	assert.Equal(t, dataInicio, repo.lastReq.DataInicio)
	assert.Equal(t, dataFim, repo.lastReq.DataFim)
	assert.Equal(t, dataInicio.Add(-time.Nanosecond), repo.lastReq.PrevDataFim)
	assert.Equal(t, dataInicio.Add(-time.Nanosecond).Add(-10*24*time.Hour), repo.lastReq.PrevDataInicio)
	assert.Equal(t, dashboardCMVProdutoLimit, repo.lastReq.CmvProdutoLimit)
}

func TestDashboardUsecase_GetDashboard_LojaObrigatoria(t *testing.T) {
	uc := NewDashboardUsecase(&mockDashboardRepository{}, newMockLojaRepository())

	_, err := uc.GetDashboard(context.Background(), DashboardRequest{
		EmpresaID:  uuid.New(),
		DataInicio: time.Now().Add(-24 * time.Hour),
		DataFim:    time.Now(),
	})

	assert.ErrorIs(t, err, ErrDashboardLojaObrigatoria)
}

func TestDashboardUsecase_GetDashboard_PeriodoInvalido(t *testing.T) {
	uc := NewDashboardUsecase(&mockDashboardRepository{}, newMockLojaRepository())

	agora := time.Now()
	_, err := uc.GetDashboard(context.Background(), DashboardRequest{
		EmpresaID:  uuid.New(),
		LojaID:     uuid.New(),
		DataInicio: agora,
		DataFim:    agora.Add(-time.Hour),
	})

	assert.ErrorIs(t, err, ErrDashboardPeriodoInvalido)
}

func TestDashboardUsecase_GetDashboard_LojaNaoPertence(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newMockLojaRepository()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: uuid.New(), // outra empresa
		Ativo:     true,
	}))

	uc := NewDashboardUsecase(&mockDashboardRepository{}, lojaRepo)

	_, err := uc.GetDashboard(ctx, DashboardRequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: time.Now().Add(-24 * time.Hour),
		DataFim:    time.Now(),
	})

	assert.ErrorIs(t, err, ErrDashboardLojaNaoPertence)
}

func TestDashboardUsecase_GetDashboard_LojaInativa(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newMockLojaRepository()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Ativo:     false,
	}))

	uc := NewDashboardUsecase(&mockDashboardRepository{}, lojaRepo)

	_, err := uc.GetDashboard(ctx, DashboardRequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: time.Now().Add(-24 * time.Hour),
		DataFim:    time.Now(),
	})

	assert.ErrorIs(t, err, ErrDashboardLojaNaoPertence)
}

func TestPercentChange(t *testing.T) {
	assert.Equal(t, 0.0, percentChange(0, 100))
	assert.InDelta(t, 50, percentChange(100, 150), 1e-9)
	assert.InDelta(t, -25, percentChange(200, 150), 1e-9)
}
