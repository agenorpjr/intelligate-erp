package db

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// ContaGerencialRepository implementation
// ============================================================

type contaGerencialRepository struct {
	querier *Queries
}

func NewContaGerencialRepository(pool *pgxpool.Pool) *contaGerencialRepository {
	return &contaGerencialRepository{querier: New(pool)}
}

func (r *contaGerencialRepository) Create(ctx context.Context, cg *domain.ContaGerencial) error {
	var grupo pgtype.Text
	if cg.Grupo != nil {
		grupo = pgtype.Text{String: *cg.Grupo, Valid: true}
	}
	var contaPaiID pgtype.UUID
	if cg.ContaPaiID != nil {
		contaPaiID = pgtype.UUID{Bytes: *cg.ContaPaiID, Valid: true}
	}

	result, err := r.querier.CreateContaGerencial(ctx, CreateContaGerencialParams{
		EmpresaID:  cg.EmpresaID,
		Codigo:     cg.Codigo,
		Nome:       cg.Nome,
		Tipo:       string(cg.Tipo),
		Grupo:      grupo,
		ContaPaiID: contaPaiID,
	})
	if err != nil {
		return err
	}

	cg.ID = result.ID
	cg.CreatedAt = result.CreatedAt
	cg.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *contaGerencialRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContaGerencial, error) {
	result, err := r.querier.GetContaGerencial(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainContaGerencial(result), nil
}

func (r *contaGerencialRepository) GetByCodigo(ctx context.Context, empresaID uuid.UUID, codigo string) (*domain.ContaGerencial, error) {
	result, err := r.querier.GetContaGerencialByCodigo(ctx, GetContaGerencialByCodigoParams{
		EmpresaID: empresaID,
		Codigo:    codigo,
	})
	if err != nil {
		return nil, err
	}
	return toDomainContaGerencial(result), nil
}

func (r *contaGerencialRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, filter domain.ContaGerencialFilter) ([]*domain.ContaGerencial, error) {
	tipo := ""
	if filter.Tipo != nil {
		tipo = string(*filter.Tipo)
	}
	grupo := ""
	if filter.Grupo != nil {
		grupo = *filter.Grupo
	}
	ativo := false
	if filter.Ativo != nil {
		ativo = *filter.Ativo
	}

	results, err := r.querier.ListContasGerenciaisByEmpresa(ctx, ListContasGerenciaisByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   tipo,
		Column3:   grupo,
		Column4:   ativo,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*domain.ContaGerencial, len(results))
	for i, item := range results {
		items[i] = toDomainContaGerencial(item)
	}
	return items, nil
}

func (r *contaGerencialRepository) ListByEmpresaHierarquia(ctx context.Context, empresaID uuid.UUID) ([]*domain.ContaGerencial, error) {
	results, err := r.querier.ListContasGerenciaisByEmpresaHierarquia(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.ContaGerencial, len(results))
	for i, item := range results {
		items[i] = toDomainContaGerencial(item)
	}
	return items, nil
}

func (r *contaGerencialRepository) Update(ctx context.Context, cg *domain.ContaGerencial) error {
	var grupo pgtype.Text
	if cg.Grupo != nil {
		grupo = pgtype.Text{String: *cg.Grupo, Valid: true}
	}
	var contaPaiID pgtype.UUID
	if cg.ContaPaiID != nil {
		contaPaiID = pgtype.UUID{Bytes: *cg.ContaPaiID, Valid: true}
	}

	result, err := r.querier.UpdateContaGerencial(ctx, UpdateContaGerencialParams{
		ID:         cg.ID,
		Codigo:     cg.Codigo,
		Nome:       cg.Nome,
		Tipo:       string(cg.Tipo),
		Grupo:      grupo,
		ContaPaiID: contaPaiID,
		Ativo:      cg.Ativo,
	})
	if err != nil {
		return err
	}

	cg.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *contaGerencialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if account has movements
	hasMov, err := r.querier.CheckContaGerencialHasMovimentacoes(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}
	if hasMov {
		return errors.New("cannot delete conta gerencial with existing movements")
	}

	return r.querier.DeleteContaGerencial(ctx, id)
}

func (r *contaGerencialRepository) HasMovimentacoes(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.querier.CheckContaGerencialHasMovimentacoes(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func toDomainContaGerencial(c ContasGerenciai) *domain.ContaGerencial {
	var grupo *string
	if c.Grupo.Valid {
		grupo = &c.Grupo.String
	}
	var contaPaiID *uuid.UUID
	if c.ContaPaiID.Valid {
		var u uuid.UUID
		copy(u[:], c.ContaPaiID.Bytes[:])
		contaPaiID = &u
	}

	return &domain.ContaGerencial{
		ID:         c.ID,
		EmpresaID:  c.EmpresaID,
		Codigo:     c.Codigo,
		Nome:       c.Nome,
		Tipo:       domain.ContaGerencialTipo(c.Tipo),
		Grupo:      grupo,
		ContaPaiID: contaPaiID,
		Ativo:      c.Ativo,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		DeletedAt:  timestamptzPtr(c.DeletedAt),
	}
}

// ============================================================
// CategoriaFinanceiraRepository implementation
// ============================================================

type categoriaFinanceiraRepository struct {
	querier *Queries
}

func NewCategoriaFinanceiraRepository(pool *pgxpool.Pool) *categoriaFinanceiraRepository {
	return &categoriaFinanceiraRepository{querier: New(pool)}
}

func (r *categoriaFinanceiraRepository) Create(ctx context.Context, cf *domain.CategoriaFinanceira) error {
	var descricao pgtype.Text
	if cf.Descricao != nil {
		descricao = pgtype.Text{String: *cf.Descricao, Valid: true}
	}
	var contaGerencialID pgtype.UUID
	if cf.ContaGerencialID != nil {
		contaGerencialID = pgtype.UUID{Bytes: *cf.ContaGerencialID, Valid: true}
	}

	result, err := r.querier.CreateCategoriaFinanceira(ctx, CreateCategoriaFinanceiraParams{
		EmpresaID:        cf.EmpresaID,
		Nome:             cf.Nome,
		Descricao:        descricao,
		ContaGerencialID: contaGerencialID,
	})
	if err != nil {
		return err
	}

	cf.ID = result.ID
	cf.CreatedAt = result.CreatedAt
	cf.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *categoriaFinanceiraRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CategoriaFinanceira, error) {
	result, err := r.querier.GetCategoriaFinanceira(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainCategoriaFinanceira(result), nil
}

func (r *categoriaFinanceiraRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, filter domain.CategoriaFinanceiraFilter) ([]*domain.CategoriaFinanceira, error) {
	contaGerencialID := uuid.Nil
	if filter.ContaGerencialID != nil {
		contaGerencialID = *filter.ContaGerencialID
	}
	ativo := false
	if filter.Ativo != nil {
		ativo = *filter.Ativo
	}

	results, err := r.querier.ListCategoriasFinanceirasByEmpresa(ctx, ListCategoriasFinanceirasByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   contaGerencialID,
		Column3:   ativo,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*domain.CategoriaFinanceira, len(results))
	for i, item := range results {
		items[i] = toDomainCategoriaFinanceira(item)
	}
	return items, nil
}

func (r *categoriaFinanceiraRepository) Update(ctx context.Context, cf *domain.CategoriaFinanceira) error {
	var descricao pgtype.Text
	if cf.Descricao != nil {
		descricao = pgtype.Text{String: *cf.Descricao, Valid: true}
	}
	var contaGerencialID pgtype.UUID
	if cf.ContaGerencialID != nil {
		contaGerencialID = pgtype.UUID{Bytes: *cf.ContaGerencialID, Valid: true}
	}

	result, err := r.querier.UpdateCategoriaFinanceira(ctx, UpdateCategoriaFinanceiraParams{
		ID:               cf.ID,
		Nome:             cf.Nome,
		Descricao:        descricao,
		ContaGerencialID: contaGerencialID,
		Ativo:            cf.Ativo,
	})
	if err != nil {
		return err
	}

	cf.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *categoriaFinanceiraRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteCategoriaFinanceira(ctx, id)
}

func toDomainCategoriaFinanceira(c CategoriasFinanceira) *domain.CategoriaFinanceira {
	var descricao *string
	if c.Descricao.Valid {
		descricao = &c.Descricao.String
	}
	var contaGerencialID *uuid.UUID
	if c.ContaGerencialID.Valid {
		var u uuid.UUID
		copy(u[:], c.ContaGerencialID.Bytes[:])
		contaGerencialID = &u
	}

	return &domain.CategoriaFinanceira{
		ID:               c.ID,
		EmpresaID:        c.EmpresaID,
		Nome:             c.Nome,
		Descricao:        descricao,
		ContaGerencialID: contaGerencialID,
		Ativo:            c.Ativo,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		DeletedAt:        timestamptzPtr(c.DeletedAt),
	}
}

// ============================================================
// MovimentacaoCaixaRepository implementation
// ============================================================

type movimentacaoCaixaRepository struct {
	querier *Queries
}

func NewMovimentacaoCaixaRepository(pool *pgxpool.Pool) *movimentacaoCaixaRepository {
	return &movimentacaoCaixaRepository{querier: New(pool)}
}

func (r *movimentacaoCaixaRepository) Create(ctx context.Context, mc *domain.MovimentacaoCaixa) error {
	var categoriaFinanceiraID pgtype.UUID
	if mc.CategoriaFinanceiraID != nil {
		categoriaFinanceiraID = pgtype.UUID{Bytes: *mc.CategoriaFinanceiraID, Valid: true}
	}
	var contaGerencialID pgtype.UUID
	if mc.ContaGerencialID != nil {
		contaGerencialID = pgtype.UUID{Bytes: *mc.ContaGerencialID, Valid: true}
	}
	var origem pgtype.Text
	if mc.Origem != nil {
		origem = pgtype.Text{String: *mc.Origem, Valid: true}
	}
	var origemID pgtype.UUID
	if mc.OrigemID != nil {
		origemID = pgtype.UUID{Bytes: *mc.OrigemID, Valid: true}
	}
	var observacoes pgtype.Text
	if mc.Observacoes != nil {
		observacoes = pgtype.Text{String: *mc.Observacoes, Valid: true}
	}

	valor := numericFromFloat(mc.Valor)

	result, err := r.querier.CreateMovimentacaoCaixa(ctx, CreateMovimentacaoCaixaParams{
		EmpresaID:             mc.EmpresaID,
		LojaID:                mc.LojaID,
		Data:                  mc.Data,
		Tipo:                  string(mc.Tipo),
		Valor:                 valor,
		Descricao:             mc.Descricao,
		CategoriaFinanceiraID: categoriaFinanceiraID,
		ContaGerencialID:      contaGerencialID,
		Origem:                origem,
		OrigemID:              origemID,
		Observacoes:           observacoes,
	})
	if err != nil {
		return err
	}

	mc.ID = result.ID
	mc.CreatedAt = result.CreatedAt
	mc.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *movimentacaoCaixaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MovimentacaoCaixa, error) {
	result, err := r.querier.GetMovimentacaoCaixa(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainMovimentacaoCaixa(result), nil
}

func (r *movimentacaoCaixaRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, filter domain.MovimentacaoCaixaFilter) ([]*domain.MovimentacaoCaixa, error) {
	tipo := ""
	if filter.Tipo != nil {
		tipo = string(*filter.Tipo)
	}
	categoriaFinanceiraID := uuid.Nil
	if filter.CategoriaFinanceiraID != nil {
		categoriaFinanceiraID = *filter.CategoriaFinanceiraID
	}
	contaGerencialID := uuid.Nil
	if filter.ContaGerencialID != nil {
		contaGerencialID = *filter.ContaGerencialID
	}
	origem := ""
	if filter.Origem != nil {
		origem = *filter.Origem
	}
	dataInicio := time.Time{}
	if filter.DataInicio != nil {
		dataInicio = *filter.DataInicio
	}
	dataFim := time.Time{}
	if filter.DataFim != nil {
		dataFim = *filter.DataFim
	}
	lojaID := uuid.Nil
	if filter.LojaID != nil {
		lojaID = *filter.LojaID
	}

	results, err := r.querier.ListMovimentacoesCaixaByEmpresa(ctx, ListMovimentacoesCaixaByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   lojaID,
		Column3:   tipo,
		Column4:   categoriaFinanceiraID,
		Column5:   contaGerencialID,
		Column6:   origem,
		Column7:   dataInicio,
		Column8:   dataFim,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*domain.MovimentacaoCaixa, len(results))
	for i, item := range results {
		items[i] = toDomainMovimentacaoCaixa(item)
	}
	return items, nil
}

func (r *movimentacaoCaixaRepository) ListByLoja(ctx context.Context, lojaID uuid.UUID, filter domain.MovimentacaoCaixaFilter) ([]*domain.MovimentacaoCaixa, error) {
	tipo := ""
	if filter.Tipo != nil {
		tipo = string(*filter.Tipo)
	}
	categoriaFinanceiraID := uuid.Nil
	if filter.CategoriaFinanceiraID != nil {
		categoriaFinanceiraID = *filter.CategoriaFinanceiraID
	}
	contaGerencialID := uuid.Nil
	if filter.ContaGerencialID != nil {
		contaGerencialID = *filter.ContaGerencialID
	}
	origem := ""
	if filter.Origem != nil {
		origem = *filter.Origem
	}
	dataInicio := time.Time{}
	if filter.DataInicio != nil {
		dataInicio = *filter.DataInicio
	}
	dataFim := time.Time{}
	if filter.DataFim != nil {
		dataFim = *filter.DataFim
	}

	results, err := r.querier.ListMovimentacoesCaixaByLoja(ctx, ListMovimentacoesCaixaByLojaParams{
		LojaID:  lojaID,
		Column2: tipo,
		Column3: categoriaFinanceiraID,
		Column4: contaGerencialID,
		Column5: origem,
		Column6: dataInicio,
		Column7: dataFim,
		Limit:   int32(filter.Limit),
		Offset:  int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*domain.MovimentacaoCaixa, len(results))
	for i, item := range results {
		items[i] = toDomainMovimentacaoCaixa(item)
	}
	return items, nil
}

func (r *movimentacaoCaixaRepository) Update(ctx context.Context, mc *domain.MovimentacaoCaixa) error {
	var categoriaFinanceiraID pgtype.UUID
	if mc.CategoriaFinanceiraID != nil {
		categoriaFinanceiraID = pgtype.UUID{Bytes: *mc.CategoriaFinanceiraID, Valid: true}
	}
	var contaGerencialID pgtype.UUID
	if mc.ContaGerencialID != nil {
		contaGerencialID = pgtype.UUID{Bytes: *mc.ContaGerencialID, Valid: true}
	}
	var origem pgtype.Text
	if mc.Origem != nil {
		origem = pgtype.Text{String: *mc.Origem, Valid: true}
	}
	var origemID pgtype.UUID
	if mc.OrigemID != nil {
		origemID = pgtype.UUID{Bytes: *mc.OrigemID, Valid: true}
	}
	var observacoes pgtype.Text
	if mc.Observacoes != nil {
		observacoes = pgtype.Text{String: *mc.Observacoes, Valid: true}
	}

	valor := numericFromFloat(mc.Valor)

	result, err := r.querier.UpdateMovimentacaoCaixa(ctx, UpdateMovimentacaoCaixaParams{
		ID:                    mc.ID,
		Data:                  mc.Data,
		Tipo:                  string(mc.Tipo),
		Valor:                 valor,
		Descricao:             mc.Descricao,
		CategoriaFinanceiraID: categoriaFinanceiraID,
		ContaGerencialID:      contaGerencialID,
		Origem:                origem,
		OrigemID:              origemID,
		Observacoes:           observacoes,
	})
	if err != nil {
		return err
	}

	mc.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *movimentacaoCaixaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteMovimentacaoCaixa(ctx, id)
}

func (r *movimentacaoCaixaRepository) GetSaldoPeriodo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (float64, error) {
	column2 := uuid.Nil
	if lojaID != nil {
		column2 = *lojaID
	}

	result, err := r.querier.GetSaldoPeriodo(ctx, GetSaldoPeriodoParams{
		EmpresaID: empresaID,
		Column2:   column2,
		Data:      dataInicio,
		Data_2:    dataFim,
	})
	if err != nil {
		return 0, err
	}

	return float64FromInterface(result), nil
}

func (r *movimentacaoCaixaRepository) GetTotaisPorContaGerencial(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (map[uuid.UUID]float64, error) {
	column2 := uuid.Nil
	if lojaID != nil {
		column2 = *lojaID
	}

	results, err := r.querier.GetTotaisPorContaGerencial(ctx, GetTotaisPorContaGerencialParams{
		EmpresaID: empresaID,
		Column2:   column2,
		Data:      dataInicio,
		Data_2:    dataFim,
	})
	if err != nil {
		return nil, err
	}

	totals := make(map[uuid.UUID]float64)
	for _, item := range results {
		if item.ContaGerencialID.Valid {
			var u uuid.UUID
			copy(u[:], item.ContaGerencialID.Bytes[:])
			totals[u] = float64FromInterface(item.Total)
		}
	}
	return totals, nil
}

func (r *movimentacaoCaixaRepository) GetTotaisPorCategoriaFinanceira(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim time.Time) (map[uuid.UUID]float64, error) {
	column2 := uuid.Nil
	if lojaID != nil {
		column2 = *lojaID
	}

	results, err := r.querier.GetTotaisPorCategoriaFinanceira(ctx, GetTotaisPorCategoriaFinanceiraParams{
		EmpresaID: empresaID,
		Column2:   column2,
		Data:      dataInicio,
		Data_2:    dataFim,
	})
	if err != nil {
		return nil, err
	}

	totals := make(map[uuid.UUID]float64)
	for _, item := range results {
		if item.CategoriaFinanceiraID.Valid {
			var u uuid.UUID
			copy(u[:], item.CategoriaFinanceiraID.Bytes[:])
			totals[u] = float64FromInterface(item.Total)
		}
	}
	return totals, nil
}

func toDomainMovimentacaoCaixa(m MovimentacoesCaixa) *domain.MovimentacaoCaixa {
	var categoriaFinanceiraID *uuid.UUID
	if m.CategoriaFinanceiraID.Valid {
		var u uuid.UUID
		copy(u[:], m.CategoriaFinanceiraID.Bytes[:])
		categoriaFinanceiraID = &u
	}
	var contaGerencialID *uuid.UUID
	if m.ContaGerencialID.Valid {
		var u uuid.UUID
		copy(u[:], m.ContaGerencialID.Bytes[:])
		contaGerencialID = &u
	}
	var origem *string
	if m.Origem.Valid {
		origem = &m.Origem.String
	}
	var origemID *uuid.UUID
	if m.OrigemID.Valid {
		var u uuid.UUID
		copy(u[:], m.OrigemID.Bytes[:])
		origemID = &u
	}
	var observacoes *string
	if m.Observacoes.Valid {
		observacoes = &m.Observacoes.String
	}

	return &domain.MovimentacaoCaixa{
		ID:                    m.ID,
		EmpresaID:             m.EmpresaID,
		LojaID:                m.LojaID,
		Data:                  m.Data,
		Tipo:                  domain.MovimentacaoCaixaTipo(m.Tipo),
		Valor:                 float64FromInterface(m.Valor),
		Descricao:             m.Descricao,
		CategoriaFinanceiraID: categoriaFinanceiraID,
		ContaGerencialID:      contaGerencialID,
		Origem:                origem,
		OrigemID:              origemID,
		Observacoes:           observacoes,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
		DeletedAt:             timestamptzPtr(m.DeletedAt),
	}
}

// ============================================================
// Helper functions
// ============================================================

func numericFromFloat(f float64) pgtype.Numeric {
	scaled := big.NewInt(int64(f * 10000))
	return pgtype.Numeric{Int: scaled, Exp: -4, Valid: true}
}

func float64FromInterface(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case *big.Float:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

func timestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
