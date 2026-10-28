package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// Erros esperados do dashboard.
var (
	ErrDashboardLojaObrigatoria = errors.New("loja_id é obrigatório")
	ErrDashboardPeriodoInvalido = errors.New("período inválido")
	ErrDashboardLojaNaoPertence = errors.New("loja não pertence à empresa ou está inativa")
)

const dashboardCMVProdutoLimit = 50

// DashboardUsecase orquestra a montagem do dashboard operacional.
type DashboardUsecase struct {
	dashboardRepo domain.DashboardRepository
	lojaRepo      domain.LojaRepository
}

// NewDashboardUsecase cria uma nova instância de DashboardUsecase.
func NewDashboardUsecase(
	dashboardRepo domain.DashboardRepository,
	lojaRepo domain.LojaRepository,
) *DashboardUsecase {
	return &DashboardUsecase{
		dashboardRepo: dashboardRepo,
		lojaRepo:      lojaRepo,
	}
}

// DashboardRequest representa os parâmetros validados para consulta do dashboard.
type DashboardRequest struct {
	EmpresaID  uuid.UUID `json:"empresa_id"`
	LojaID     uuid.UUID `json:"loja_id"`
	DataInicio time.Time `json:"data_inicio"`
	DataFim    time.Time `json:"data_fim"`
}

// GetDashboard valida os filtros, calcula o período de comparação e retorna os
// indicadores consolidados da loja no período solicitado.
func (u *DashboardUsecase) GetDashboard(ctx context.Context, req DashboardRequest) (*domain.DashboardResponse, error) {
	if req.EmpresaID == uuid.Nil {
		return nil, fmt.Errorf("empresa_id é obrigatório")
	}
	if req.LojaID == uuid.Nil {
		return nil, ErrDashboardLojaObrigatoria
	}
	if req.DataInicio.IsZero() || req.DataFim.IsZero() {
		return nil, fmt.Errorf("%w: data_inicio e data_fim são obrigatórios", ErrDashboardPeriodoInvalido)
	}
	if !req.DataInicio.Before(req.DataFim) {
		return nil, fmt.Errorf("%w: data_inicio deve ser anterior a data_fim", ErrDashboardPeriodoInvalido)
	}

	// Multi-loja: garante que a loja pertence à empresa do usuário e está ativa.
	loja, err := u.lojaRepo.GetByID(ctx, req.LojaID)
	if err != nil || loja == nil || loja.EmpresaID != req.EmpresaID || !loja.Ativo {
		return nil, ErrDashboardLojaNaoPertence
	}

	prevInicio, prevFim := previousPeriod(req.DataInicio, req.DataFim)

	resp, err := u.dashboardRepo.GetDashboard(ctx, domain.DashboardRequest{
		EmpresaID:       req.EmpresaID,
		LojaID:          req.LojaID,
		DataInicio:      req.DataInicio,
		DataFim:         req.DataFim,
		PrevDataInicio:  prevInicio,
		PrevDataFim:     prevFim,
		CmvProdutoLimit: dashboardCMVProdutoLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build dashboard: %w", err)
	}

	resp.Vendas.VariacaoPercentual = computeVariacao(resp.Vendas.ComparacaoPeriodo)

	return resp, nil
}

// previousPeriod calcula o período imediatamente anterior de mesma duração.
func previousPeriod(dataInicio, dataFim time.Time) (time.Time, time.Time) {
	duration := dataFim.Sub(dataInicio)
	prevFim := dataInicio.Add(-time.Nanosecond)
	prevInicio := prevFim.Add(-duration)
	return prevInicio, prevFim
}

// computeVariacao calcula as variações percentuais entre o período atual e o anterior.
func computeVariacao(c domain.ComparacaoPeriodo) domain.VariacaoPercentual {
	atual := c.PeriodoAtual
	anterior := c.PeriodoAnterior

	return domain.VariacaoPercentual{
		Vendas:      percentChange(anterior.TotalVendas, atual.TotalVendas),
		Pedidos:     percentChange(float64(anterior.TotalPedidos), float64(atual.TotalPedidos)),
		TicketMedio: percentChange(anterior.TicketMedio, atual.TicketMedio),
	}
}

// percentChange retorna a variação percentual; 0 quando não há base de comparação.
func percentChange(anterior, atual float64) float64 {
	if anterior == 0 {
		return 0
	}
	return ((atual - anterior) / anterior) * 100
}
