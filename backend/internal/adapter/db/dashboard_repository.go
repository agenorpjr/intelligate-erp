package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultCMVProdutoLimit = 50
const produtosMaiorCMVLimit = 10

// DashboardRepository monta o dashboard operacional consultando os
// indicadores agregados de vendas, CMV, financeiro e DRE.
type DashboardRepository struct {
	querier *Queries
}

// NewDashboardRepository cria uma instância de DashboardRepository.
func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{querier: New(pool)}
}

// GetDashboard consolida todos os indicadores do dashboard para a loja e o
// período informados, incluindo a comparação com o período anterior.
func (r *DashboardRepository) GetDashboard(ctx context.Context, req domain.DashboardRequest) (*domain.DashboardResponse, error) {
	prevInicio, prevFim := resolvePreviousPeriod(req)

	// --- Vendas ---
	resumo, err := r.querier.GetDashboardVendasResumo(ctx, GetDashboardVendasResumoParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	vendasPorDiaRows, err := r.querier.GetDashboardVendasPorDia(ctx, GetDashboardVendasPorDiaParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	vendasPorStatusRows, err := r.querier.GetDashboardVendasPorStatus(ctx, GetDashboardVendasPorStatusParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	prev, err := r.querier.GetDashboardVendasResumoPeriodoAnterior(ctx, GetDashboardVendasResumoPeriodoAnteriorParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: prevInicio,
		DataFim:    prevFim,
	})
	if err != nil {
		return nil, err
	}

	// --- CMV ---
	cmvLimit := req.CmvProdutoLimit
	if cmvLimit <= 0 {
		cmvLimit = defaultCMVProdutoLimit
	}

	cmvResumo, err := r.querier.GetDashboardCMVResumo(ctx, GetDashboardCMVResumoParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	cmvPorDiaRows, err := r.querier.GetDashboardCMVPorDia(ctx, GetDashboardCMVPorDiaParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	cmvProdutoRows, err := r.querier.GetDashboardCMVPorProduto(ctx, GetDashboardCMVPorProdutoParams{
		PageOffset: 0,
		PageLimit:  int32(cmvLimit),
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	produtosSemFichaRows, err := r.querier.GetDashboardProdutosSemFichaTecnica(ctx, GetDashboardProdutosSemFichaTecnicaParams{
		PageOffset: 0,
		PageLimit:  int32(cmvLimit),
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	// --- Financeiro ---
	fin, err := r.querier.GetDashboardFinanceiroResumo(ctx, GetDashboardFinanceiroResumoParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	despesasRows, err := r.querier.GetDashboardDespesasPorCategoria(ctx, GetDashboardDespesasPorCategoriaParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	contasPagar, err := r.querier.GetDashboardContasPagarResumo(ctx, GetDashboardContasPagarResumoParams{
		EmpresaID: req.EmpresaID,
		LojaID:    req.LojaID,
	})
	if err != nil {
		return nil, err
	}

	fluxoRows, err := r.querier.GetDashboardFluxoCaixaPorDia(ctx, GetDashboardFluxoCaixaPorDiaParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	totaisTipoRows, err := r.querier.GetDashboardTotaisPorTipoContaGerencial(ctx, GetDashboardTotaisPorTipoContaGerencialParams{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
	})
	if err != nil {
		return nil, err
	}

	// --- Montagem dos indicadores ---
	vendas := domain.DashboardVendas{
		TotalVendas:     resumo.TotalVendas,
		TotalPedidos:    resumo.TotalPedidos,
		TicketMedio:     resumo.TicketMedio,
		TotalDescontos:  resumo.TotalDescontos,
		VendasPorDia:    make([]domain.VendaPorDia, 0, len(vendasPorDiaRows)),
		VendasPorStatus: make([]domain.VendaPorStatus, 0, len(vendasPorStatusRows)),
		ComparacaoPeriodo: domain.ComparacaoPeriodo{
			PeriodoAtual: domain.PeriodoComparacao{
				DataInicio:     req.DataInicio,
				DataFim:        req.DataFim,
				TotalVendas:    resumo.TotalVendas,
				TotalPedidos:   resumo.TotalPedidos,
				TicketMedio:    resumo.TicketMedio,
				TotalDescontos: resumo.TotalDescontos,
			},
			PeriodoAnterior: domain.PeriodoComparacao{
				DataInicio:     prevInicio,
				DataFim:        prevFim,
				TotalVendas:    prev.TotalVendas,
				TotalPedidos:   prev.TotalPedidos,
				TicketMedio:    prev.TicketMedio,
				TotalDescontos: prev.TotalDescontos,
			},
		},
	}
	for _, row := range vendasPorDiaRows {
		vendas.VendasPorDia = append(vendas.VendasPorDia, domain.VendaPorDia{
			Data:         pgDateToTime(row.Data),
			TotalVendas:  row.TotalVendas,
			TotalPedidos: row.TotalPedidos,
			TicketMedio:  row.TicketMedio,
		})
	}
	for _, row := range vendasPorStatusRows {
		vendas.VendasPorStatus = append(vendas.VendasPorStatus, domain.VendaPorStatus{
			Status:       row.Status,
			TotalVendas:  row.TotalVendas,
			TotalPedidos: row.TotalPedidos,
		})
	}

	cmvPorDia := make([]domain.CMVPorDia, 0, len(cmvPorDiaRows))
	for _, row := range cmvPorDiaRows {
		cmvPorDia = append(cmvPorDia, domain.CMVPorDia{
			Data:        pgDateToTime(row.Data),
			TotalCMV:    row.TotalCmv,
			TotalVendas: row.TotalVendas,
		})
	}

	cmvPorProduto := make([]domain.CMVProdutoResumo, 0, len(cmvProdutoRows))
	for _, row := range cmvProdutoRows {
		cmvPorProduto = append(cmvPorProduto, cmvProdutoRowToDomain(
			row.ProdutoID, row.ProdutoNome, row.ProdutoSku,
			row.QuantidadeVendida, row.CustoUnitarioTeorico, row.CmvTotal,
			row.ReceitaTotal, row.MargemBruta, row.MargemBrutaPercent, row.TemFichaTecnica,
		))
	}

	produtosSemFicha := make([]domain.CMVProdutoResumo, 0, len(produtosSemFichaRows))
	for _, row := range produtosSemFichaRows {
		produtosSemFicha = append(produtosSemFicha, cmvProdutoRowToDomain(
			row.ProdutoID, row.ProdutoNome, row.ProdutoSku,
			row.QuantidadeVendida, 0, 0,
			row.ReceitaTotal, row.ReceitaTotal, 100, false,
		))
	}

	produtosMaiorCMV := cmvPorProduto
	if len(produtosMaiorCMV) > produtosMaiorCMVLimit {
		produtosMaiorCMV = produtosMaiorCMV[:produtosMaiorCMVLimit]
	}

	receitaLiquida := resumo.TotalVendas - resumo.TotalDescontos
	percentualCMVReceita := 0.0
	if receitaLiquida > 0 {
		percentualCMVReceita = (cmvResumo.TotalCmv / receitaLiquida) * 100
	}

	cmvIndicadores := domain.DashboardCMV{
		TotalCMV:             cmvResumo.TotalCmv,
		CMVPorDia:            cmvPorDia,
		CMVPorProduto:        cmvPorProduto,
		PercentualCMVReceita: percentualCMVReceita,
		ProdutosSemFicha:     int(cmvResumo.ProdutosSemFichaTecnica),
		ProdutosMaiorCMV:     produtosMaiorCMV,
	}

	despesasPorCategoria := make([]domain.DespesaPorCategoria, 0, len(despesasRows))
	for _, row := range despesasRows {
		despesasPorCategoria = append(despesasPorCategoria, domain.DespesaPorCategoria{
			CategoriaFinanceiraID: row.CategoriaFinanceiraID,
			CategoriaNome:         row.CategoriaNome,
			Total:                 row.Total,
		})
	}

	fluxoPorDia := make([]domain.FluxoCaixaPorDia, 0, len(fluxoRows))
	acumulado := 0.0
	for _, row := range fluxoRows {
		acumulado += row.SaldoDiario
		fluxoPorDia = append(fluxoPorDia, domain.FluxoCaixaPorDia{
			Data:           pgDateToTime(row.Data),
			Entradas:       row.Entradas,
			Saidas:         row.Saidas,
			SaldoDiario:    row.SaldoDiario,
			SaldoAcumulado: acumulado,
		})
	}

	financeiro := domain.DashboardFinanceiro{
		TotalEntradas:           fin.TotalEntradas,
		TotalSaidas:             fin.TotalSaidas,
		SaldoFinal:              fin.SaldoFinal,
		DespesasPorCategoria:    despesasPorCategoria,
		ContasPagarEmAberto:     contasPagar.ContasPagarEmAberto,
		ContasPagarVencidas:     contasPagar.ContasPagarVencidas,
		ValorContasPagarAberto:  contasPagar.ValorContasPagarAberto,
		ValorContasPagarVencido: contasPagar.ValorContasPagarVencido,
		FluxoCaixaPorDia:        fluxoPorDia,
	}

	dre := buildDRE(receitaLiquida, resumo.TotalVendas, resumo.TotalDescontos, cmvResumo.TotalCmv, totaisTipoRows)

	return &domain.DashboardResponse{
		EmpresaID:  req.EmpresaID,
		LojaID:     req.LojaID,
		DataInicio: req.DataInicio,
		DataFim:    req.DataFim,
		Vendas:     vendas,
		CMV:        cmvIndicadores,
		Financeiro: financeiro,
		DRE:        dre,
	}, nil
}

// buildDRE monta o DRE gerencial a partir dos totais já calculados de vendas,
// CMV e movimentações por tipo de conta gerencial.
func buildDRE(receitaLiquida, receitaBruta, descontos, cmv float64, totaisPorTipo []GetDashboardTotaisPorTipoContaGerencialRow) domain.DashboardDRE {
	var despesas, investimentos float64
	for _, row := range totaisPorTipo {
		switch row.Tipo {
		case string(domain.ContaGerencialTipoDespesa):
			despesas += row.Total
		case string(domain.ContaGerencialTipoInvestimento):
			investimentos += row.Total
		}
	}

	margemBruta := receitaLiquida - cmv
	margemBrutaPercent := 0.0
	if receitaLiquida > 0 {
		margemBrutaPercent = (margemBruta / receitaLiquida) * 100
	}

	resultadoOperacional := margemBruta - despesas
	resultadoLiquido := resultadoOperacional - investimentos

	return domain.DashboardDRE{
		ReceitaBruta:         receitaBruta,
		Descontos:            descontos,
		ReceitaLiquida:       receitaLiquida,
		CMV:                  cmv,
		MargemBruta:          margemBruta,
		MargemBrutaPercent:   margemBrutaPercent,
		DespesasOperacionais: despesas,
		ResultadoOperacional: resultadoOperacional,
		ResultadoLiquido:     resultadoLiquido,
	}
}

func resolvePreviousPeriod(req domain.DashboardRequest) (time.Time, time.Time) {
	if !req.PrevDataInicio.IsZero() && !req.PrevDataFim.IsZero() {
		return req.PrevDataInicio, req.PrevDataFim
	}
	duration := req.DataFim.Sub(req.DataInicio)
	prevFim := req.DataInicio.Add(-time.Nanosecond)
	prevInicio := prevFim.Add(-duration)
	return prevInicio, prevFim
}

func cmvProdutoRowToDomain(
	produtoID uuid.UUID,
	produtoNome string,
	produtoSku pgtype.Text,
	quantidadeVendida, custoUnitarioTeorico, cmvTotal, receitaTotal, margemBruta, margemBrutaPercent float64,
	temFichaTecnica bool,
) domain.CMVProdutoResumo {
	var sku *string
	if produtoSku.Valid {
		sku = &produtoSku.String
	}
	return domain.CMVProdutoResumo{
		ProdutoID:            produtoID,
		ProdutoNome:          produtoNome,
		ProdutoSKU:           sku,
		QuantidadeVendida:    quantidadeVendida,
		CustoUnitarioTeorico: custoUnitarioTeorico,
		CMVTotal:             cmvTotal,
		ReceitaTotal:         receitaTotal,
		MargemBruta:          margemBruta,
		MargemBrutaPercent:   margemBrutaPercent,
		TemFichaTecnica:      temFichaTecnica,
	}
}

func pgDateToTime(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}
