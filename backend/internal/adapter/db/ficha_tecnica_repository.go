package db

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FichaTecnicaRepository struct {
	querier *Queries
}

func NewFichaTecnicaRepository(pool *pgxpool.Pool) *FichaTecnicaRepository {
	return &FichaTecnicaRepository{querier: New(pool)}
}

func (r *FichaTecnicaRepository) Create(ctx context.Context, ft *domain.FichaTecnica) error {
	rendimento := big.NewInt(int64(ft.Rendimento * 10000))
	params := CreateFichaTecnicaParams{
		ProdutoID:         ft.ProdutoID,
		LojaID:            ft.LojaID,
		Rendimento:        pgtype.Numeric{Int: rendimento, Exp: -4, Valid: true},
		UnidadeRendimento: ft.UnidadeRendimento,
	}
	result, err := r.querier.CreateFichaTecnica(ctx, params)
	if err != nil {
		return err
	}
	ft.ID = result.ID
	ft.CreatedAt = result.CreatedAt
	ft.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *FichaTecnicaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.FichaTecnica, error) {
	result, err := r.querier.GetFichaTecnica(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainFichaTecnica(result), nil
}

func (r *FichaTecnicaRepository) GetByProdutoLoja(ctx context.Context, produtoID, lojaID uuid.UUID) (*domain.FichaTecnica, error) {
	result, err := r.querier.GetFichaTecnicaByProdutoLoja(ctx, GetFichaTecnicaByProdutoLojaParams{
		ProdutoID: produtoID,
		LojaID:    lojaID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainFichaTecnica(result), nil
}

func (r *FichaTecnicaRepository) ListByProduto(ctx context.Context, produtoID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListFichasTecnicasByProduto(ctx, ListFichasTecnicasByProdutoParams{
		ProdutoID: produtoID,
		Column2:   ativoParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	fichas := make([]*domain.FichaTecnica, len(results))
	for i, r := range results {
		fichas[i] = toDomainFichaTecnica(r)
	}
	return fichas, nil
}

func (r *FichaTecnicaRepository) ListByLoja(ctx context.Context, lojaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListFichasTecnicasByLoja(ctx, ListFichasTecnicasByLojaParams{
		LojaID:  lojaID,
		Column2: ativoParam,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	fichas := make([]*domain.FichaTecnica, len(results))
	for i, r := range results {
		fichas[i] = toDomainFichaTecnica(r)
	}
	return fichas, nil
}

func (r *FichaTecnicaRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.FichaTecnica, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListFichasTecnicasByEmpresa(ctx, ListFichasTecnicasByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   ativoParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	fichas := make([]*domain.FichaTecnica, len(results))
	for i, r := range results {
		fichas[i] = toDomainFichaTecnica(r)
	}
	return fichas, nil
}

func (r *FichaTecnicaRepository) Update(ctx context.Context, ft *domain.FichaTecnica) error {
	var rendimento *pgtype.Numeric
	if ft.Rendimento > 0 {
		rend := big.NewInt(int64(ft.Rendimento * 10000))
		rendimento = &pgtype.Numeric{Int: rend, Exp: -4, Valid: true}
	}
	params := UpdateFichaTecnicaParams{
		ID:                ft.ID,
		Rendimento:        *rendimento,
		UnidadeRendimento: ft.UnidadeRendimento,
		Ativo:             ft.Ativo,
	}
	result, err := r.querier.UpdateFichaTecnica(ctx, params)
	if err != nil {
		return err
	}
	ft.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *FichaTecnicaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteFichaTecnica(ctx, id)
}

func (r *FichaTecnicaRepository) CreateComponente(ctx context.Context, c *domain.FichaTecnicaComponente) error {
	quantidade := big.NewInt(int64(c.Quantidade * 10000))
	params := CreateFichaTecnicaComponenteParams{
		FichaTecnicaID: c.FichaTecnicaID,
		InsumoID:       c.InsumoID,
		Quantidade:     pgtype.Numeric{Int: quantidade, Exp: -4, Valid: true},
		UnidadeMedida:  c.UnidadeMedida,
	}
	result, err := r.querier.CreateFichaTecnicaComponente(ctx, params)
	if err != nil {
		return err
	}
	c.ID = result.ID
	c.CreatedAt = result.CreatedAt
	c.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *FichaTecnicaRepository) GetComponentes(ctx context.Context, fichaTecnicaID uuid.UUID) ([]*domain.FichaTecnicaComponente, error) {
	results, err := r.querier.ListFichaTecnicaComponentes(ctx, fichaTecnicaID)
	if err != nil {
		return nil, err
	}
	componentes := make([]*domain.FichaTecnicaComponente, len(results))
	for i, r := range results {
		componentes[i] = toDomainFichaTecnicaComponente(r)
	}
	return componentes, nil
}

func (r *FichaTecnicaRepository) UpdateComponente(ctx context.Context, c *domain.FichaTecnicaComponente) error {
	quantidade := big.NewInt(int64(c.Quantidade * 10000))
	params := UpdateFichaTecnicaComponenteParams{
		ID:            c.ID,
		InsumoID:      c.InsumoID,
		Quantidade:    pgtype.Numeric{Int: quantidade, Exp: -4, Valid: true},
		UnidadeMedida: c.UnidadeMedida,
	}
	result, err := r.querier.UpdateFichaTecnicaComponente(ctx, params)
	if err != nil {
		return err
	}
	c.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *FichaTecnicaRepository) DeleteComponente(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteFichaTecnicaComponente(ctx, id)
}

func (r *FichaTecnicaRepository) DeleteComponentesByFicha(ctx context.Context, fichaTecnicaID uuid.UUID) error {
	return r.querier.DeleteFichaTecnicaComponentesByFicha(ctx, fichaTecnicaID)
}

func (r *FichaTecnicaRepository) GetCMVResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (*domain.CMVResumo, error) {
	var lojaIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	result, err := r.querier.GetCMVResumo(ctx, GetCMVResumoParams{
		Column1: empresaID,
		Column2: lojaIDParam,
		Column3: dataInicio,
		Column4: dataFim,
	})
	if err != nil {
		return nil, err
	}
	return toDomainCMVResumo(result), nil
}

func (r *FichaTecnicaRepository) GetCMVProdutosDetalhe(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time, limit, offset int) ([]*domain.CMVProdutoDetalhe, error) {
	var lojaIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	results, err := r.querier.GetCMVProdutosDetalhe(ctx, GetCMVProdutosDetalheParams{
		EmpresaID:  empresaID,
		Column2:    lojaIDParam,
		DataHora:   dataInicio,
		DataHora_2: dataFim,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	detalhes := make([]*domain.CMVProdutoDetalhe, len(results))
	for i, r := range results {
		detalhes[i] = toDomainCMVProdutoDetalhe(r)
	}
	return detalhes, nil
}

func (r *FichaTecnicaRepository) GetProdutosSemFichaTecnica(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time, limit, offset int) ([]*domain.CMVProdutoDetalhe, error) {
	var lojaIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	results, err := r.querier.GetProdutosSemFichaTecnica(ctx, GetProdutosSemFichaTecnicaParams{
		EmpresaID:  empresaID,
		Column2:    lojaIDParam,
		DataHora:   dataInicio,
		DataHora_2: dataFim,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	detalhes := make([]*domain.CMVProdutoDetalhe, len(results))
	for i, r := range results {
		detalhes[i] = toDomainCMVProdutoDetalheSemFicha(r)
	}
	return detalhes, nil
}

func toDomainFichaTecnica(ft FichasTecnica) *domain.FichaTecnica {
	var deletedAt *time.Time
	if ft.DeletedAt.Valid {
		deletedAt = &ft.DeletedAt.Time
	}
	rendimento := 0.0
	if ft.Rendimento.Valid && ft.Rendimento.Int != nil {
		rendimento = float64(ft.Rendimento.Int.Int64()) / 10000
	}
	return &domain.FichaTecnica{
		ID:                ft.ID,
		ProdutoID:         ft.ProdutoID,
		LojaID:            ft.LojaID,
		Rendimento:        rendimento,
		UnidadeRendimento: ft.UnidadeRendimento,
		Ativo:             ft.Ativo,
		CreatedAt:         ft.CreatedAt,
		UpdatedAt:         ft.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}

func toDomainFichaTecnicaComponente(c ListFichaTecnicaComponentesRow) *domain.FichaTecnicaComponente {
	quantidade := 0.0
	if c.Quantidade.Valid && c.Quantidade.Int != nil {
		quantidade = float64(c.Quantidade.Int.Int64()) / 10000
	}
	return &domain.FichaTecnicaComponente{
		ID:             c.ID,
		FichaTecnicaID: c.FichaTecnicaID,
		InsumoID:       c.InsumoID,
		Quantidade:     quantidade,
		UnidadeMedida:  c.UnidadeMedida,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		InsumoNome:     c.InsumoNome,
		InsumoUnidade:  c.InsumoUnidade,
	}
}

func toDomainCMVResumo(r GetCMVResumoRow) *domain.CMVResumo {
	totalVendas := 0.0
	if r.TotalVendas != nil {
		if v, ok := r.TotalVendas.(float64); ok {
			totalVendas = v
		}
	}
	totalCMV := 0.0
	if r.TotalCmv != nil {
		if v, ok := r.TotalCmv.(float64); ok {
			totalCMV = v
		}
	}
	margemBruta := 0.0
	if r.MargemBruta != 0 {
		margemBruta = float64(r.MargemBruta)
	}
	margemBrutaPercent := 0.0
	if r.MargemBrutaPercent != 0 {
		margemBrutaPercent = float64(r.MargemBrutaPercent)
	}
	return &domain.CMVResumo{
		EmpresaID:          r.EmpresaID,
		LojaID:             r.LojaID,
		DataInicio:         r.DataInicio,
		DataFim:            r.DataFim,
		TotalVendas:        totalVendas,
		TotalCMV:           totalCMV,
		MargemBruta:        margemBruta,
		MargemBrutaPercent: margemBrutaPercent,
		ProdutosComFicha:   int(r.ProdutosComFicha),
		ProdutosSemFicha:   int(r.ProdutosSemFichaTecnica),
		ItensVendidos:      int(r.ItensVendidos),
	}
}

func toDomainCMVProdutoDetalhe(r GetCMVProdutosDetalheRow) *domain.CMVProdutoDetalhe {
	custoUnitarioTeorico := 0.0
	if r.CustoUnitarioTeorico != nil {
		if v, ok := r.CustoUnitarioTeorico.(float64); ok {
			custoUnitarioTeorico = v
		}
	}
	cmvTotal := 0.0
	if r.CmvTotal != nil {
		if v, ok := r.CmvTotal.(float64); ok {
			cmvTotal = v
		}
	}
	receitaTotal := 0.0
	if r.ReceitaTotal != 0 {
		receitaTotal = float64(r.ReceitaTotal)
	}
	margemBruta := 0.0
	if r.MargemBruta != 0 {
		margemBruta = float64(r.MargemBruta)
	}
	margemBrutaPercent := 0.0
	if r.MargemBrutaPercent != 0 {
		margemBrutaPercent = float64(r.MargemBrutaPercent)
	}
	return &domain.CMVProdutoDetalhe{
		ProdutoID:            r.ProdutoID,
		ProdutoNome:          r.ProdutoNome,
		ProdutoSKU:           nil,
		QuantidadeVendida:    float64(r.QuantidadeVendida),
		CustoUnitarioTeorico: custoUnitarioTeorico,
		CMVTotal:             cmvTotal,
		ReceitaTotal:         receitaTotal,
		MargemBruta:          margemBruta,
		MargemBrutaPercent:   margemBrutaPercent,
		TemFichaTecnica:      r.TemFichaTecnica,
	}
}

func toDomainCMVProdutoDetalheSemFicha(r GetProdutosSemFichaTecnicaRow) *domain.CMVProdutoDetalhe {
	var produtoID uuid.UUID
	if r.ProdutoID.Valid {
		copy(produtoID[:], r.ProdutoID.Bytes[:])
	}
	var produtoNome string
	if r.ProdutoNome.Valid {
		produtoNome = r.ProdutoNome.String
	}
	var produtoSKU *string
	if r.ProdutoSku.Valid {
		produtoSKU = &r.ProdutoSku.String
	}
	return &domain.CMVProdutoDetalhe{
		ProdutoID:            produtoID,
		ProdutoNome:          produtoNome,
		ProdutoSKU:           produtoSKU,
		QuantidadeVendida:    float64(r.QuantidadeVendida),
		CustoUnitarioTeorico: 0,
		CMVTotal:             0,
		ReceitaTotal:         float64(r.ReceitaTotal),
		MargemBruta:          float64(r.MargemBruta),
		MargemBrutaPercent:   float64(r.MargemBrutaPercent),
		TemFichaTecnica:      r.TemFichaTecnica,
	}
}
