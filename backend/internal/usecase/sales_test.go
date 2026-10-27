package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/adapter/db"
	"github.com/intelligate/erp/backend/internal/domain"
)

// mockSalesTxRepository is a mock for the transactional sales repository
type mockSalesTxRepository struct {
	commitError   error
	rollbackError error
	// Hold references to the mock repositories used in tests
	pedidoRepo    *mockPedidoCanonicoRepository
	itemRepo      *mockPedidoItemCanonicoRepository
	pagamentoRepo *mockPedidoPagamentoCanonicoRepository
	eventoRepo    *mockEventoIngestaoRepository
}

func newMockSalesTxRepository() *mockSalesTxRepository {
	return &mockSalesTxRepository{}
}

func (m *mockSalesTxRepository) SetRepositories(
	pedidoRepo *mockPedidoCanonicoRepository,
	itemRepo *mockPedidoItemCanonicoRepository,
	pagamentoRepo *mockPedidoPagamentoCanonicoRepository,
	eventoRepo *mockEventoIngestaoRepository,
) {
	m.pedidoRepo = pedidoRepo
	m.itemRepo = itemRepo
	m.pagamentoRepo = pagamentoRepo
	m.eventoRepo = eventoRepo
}

func (m *mockSalesTxRepository) WithTx(ctx context.Context, fn func(*db.SalesTxQueries) error) error {
	// In tests, we just execute the function directly without a real transaction
	// Use the mock repositories that were set up in the test
	txQueries := &db.SalesTxQueries{
		PedidoRepo:    m.pedidoRepo,
		ItemRepo:      m.itemRepo,
		PagamentoRepo: m.pagamentoRepo,
		EventoRepo:    m.eventoRepo,
	}
	return fn(txQueries)
}

// Ensure mockSalesTxRepository implements db.SalesTxRepositoryInterface
var _ db.SalesTxRepositoryInterface = (*mockSalesTxRepository)(nil)

type mockPdvOrigemRepository struct {
	pdvs map[uuid.UUID]*domain.PdvOrigem
}

func newMockPdvOrigemRepository() *mockPdvOrigemRepository {
	return &mockPdvOrigemRepository{
		pdvs: make(map[uuid.UUID]*domain.PdvOrigem),
	}
}

func (m *mockPdvOrigemRepository) Create(ctx context.Context, p *domain.PdvOrigem) error {
	m.pdvs[p.ID] = p
	return nil
}

func (m *mockPdvOrigemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PdvOrigem, error) {
	if p, ok := m.pdvs[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPdvOrigemRepository) GetByCodigo(ctx context.Context, empresaID uuid.UUID, codigo string) (*domain.PdvOrigem, error) {
	for _, p := range m.pdvs {
		if p.EmpresaID == empresaID && p.Codigo == codigo {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockPdvOrigemRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.PdvOrigem, error) {
	var result []*domain.PdvOrigem
	for _, p := range m.pdvs {
		if p.EmpresaID == empresaID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockPdvOrigemRepository) Update(ctx context.Context, p *domain.PdvOrigem) error {
	m.pdvs[p.ID] = p
	return nil
}

func (m *mockPdvOrigemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.pdvs, id)
	return nil
}

type mockProdutoCanonicoRepository struct {
	produtos map[uuid.UUID]*domain.ProdutoCanonico
}

func newMockProdutoCanonicoRepository() *mockProdutoCanonicoRepository {
	return &mockProdutoCanonicoRepository{
		produtos: make(map[uuid.UUID]*domain.ProdutoCanonico),
	}
}

func (m *mockProdutoCanonicoRepository) Create(ctx context.Context, p *domain.ProdutoCanonico) error {
	m.produtos[p.ID] = p
	return nil
}

func (m *mockProdutoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProdutoCanonico, error) {
	if p, ok := m.produtos[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockProdutoCanonicoRepository) GetBySKU(ctx context.Context, empresaID uuid.UUID, sku string) (*domain.ProdutoCanonico, error) {
	for _, p := range m.produtos {
		if p.EmpresaID == empresaID && p.SKU == sku {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProdutoCanonicoRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.ProdutoCanonico, error) {
	var result []*domain.ProdutoCanonico
	for _, p := range m.produtos {
		if p.EmpresaID == empresaID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProdutoCanonicoRepository) Update(ctx context.Context, p *domain.ProdutoCanonico) error {
	m.produtos[p.ID] = p
	return nil
}

func (m *mockProdutoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.produtos, id)
	return nil
}

type mockPedidoCanonicoRepository struct {
	pedidos map[uuid.UUID]*domain.PedidoCanonico
}

func newMockPedidoCanonicoRepository() *mockPedidoCanonicoRepository {
	return &mockPedidoCanonicoRepository{
		pedidos: make(map[uuid.UUID]*domain.PedidoCanonico),
	}
}

func (m *mockPedidoCanonicoRepository) Create(ctx context.Context, p *domain.PedidoCanonico) error {
	m.pedidos[p.ID] = p
	return nil
}

func (m *mockPedidoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoCanonico, error) {
	if p, ok := m.pedidos[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPedidoCanonicoRepository) GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.PedidoCanonico, error) {
	for _, p := range m.pedidos {
		if p.EmpresaID == empresaID && p.IdempotencyKey == idempotencyKey {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockPedidoCanonicoRepository) List(ctx context.Context, filter domain.PedidoCanonicoFilter) ([]*domain.PedidoCanonico, error) {
	var result []*domain.PedidoCanonico
	for _, p := range m.pedidos {
		if p.EmpresaID == *filter.EmpresaID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockPedidoCanonicoRepository) Update(ctx context.Context, p *domain.PedidoCanonico) error {
	m.pedidos[p.ID] = p
	return nil
}

func (m *mockPedidoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.pedidos, id)
	return nil
}

type mockPedidoItemCanonicoRepository struct {
	itens map[uuid.UUID]*domain.PedidoItemCanonico
}

func newMockPedidoItemCanonicoRepository() *mockPedidoItemCanonicoRepository {
	return &mockPedidoItemCanonicoRepository{
		itens: make(map[uuid.UUID]*domain.PedidoItemCanonico),
	}
}

func (m *mockPedidoItemCanonicoRepository) Create(ctx context.Context, i *domain.PedidoItemCanonico) error {
	m.itens[i.ID] = i
	return nil
}

func (m *mockPedidoItemCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoItemCanonico, error) {
	if i, ok := m.itens[id]; ok {
		return i, nil
	}
	return nil, nil
}

func (m *mockPedidoItemCanonicoRepository) ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*domain.PedidoItemCanonico, error) {
	var result []*domain.PedidoItemCanonico
	for _, i := range m.itens {
		if i.PedidoID == pedidoID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *mockPedidoItemCanonicoRepository) Update(ctx context.Context, i *domain.PedidoItemCanonico) error {
	m.itens[i.ID] = i
	return nil
}

func (m *mockPedidoItemCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.itens, id)
	return nil
}

func (m *mockPedidoItemCanonicoRepository) DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error {
	for id, i := range m.itens {
		if i.PedidoID == pedidoID {
			delete(m.itens, id)
		}
	}
	return nil
}

type mockPedidoPagamentoCanonicoRepository struct {
	pagamentos map[uuid.UUID]*domain.PedidoPagamentoCanonico
}

func newMockPedidoPagamentoCanonicoRepository() *mockPedidoPagamentoCanonicoRepository {
	return &mockPedidoPagamentoCanonicoRepository{
		pagamentos: make(map[uuid.UUID]*domain.PedidoPagamentoCanonico),
	}
}

func (m *mockPedidoPagamentoCanonicoRepository) Create(ctx context.Context, p *domain.PedidoPagamentoCanonico) error {
	m.pagamentos[p.ID] = p
	return nil
}

func (m *mockPedidoPagamentoCanonicoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PedidoPagamentoCanonico, error) {
	if p, ok := m.pagamentos[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPedidoPagamentoCanonicoRepository) ListByPedido(ctx context.Context, pedidoID uuid.UUID) ([]*domain.PedidoPagamentoCanonico, error) {
	var result []*domain.PedidoPagamentoCanonico
	for _, p := range m.pagamentos {
		if p.PedidoID == pedidoID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockPedidoPagamentoCanonicoRepository) Update(ctx context.Context, p *domain.PedidoPagamentoCanonico) error {
	m.pagamentos[p.ID] = p
	return nil
}

func (m *mockPedidoPagamentoCanonicoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.pagamentos, id)
	return nil
}

func (m *mockPedidoPagamentoCanonicoRepository) DeleteByPedido(ctx context.Context, pedidoID uuid.UUID) error {
	for id, p := range m.pagamentos {
		if p.PedidoID == pedidoID {
			delete(m.pagamentos, id)
		}
	}
	return nil
}

type mockEventoIngestaoRepository struct {
	eventos map[uuid.UUID]*domain.EventoIngestao
}

func newMockEventoIngestaoRepository() *mockEventoIngestaoRepository {
	return &mockEventoIngestaoRepository{
		eventos: make(map[uuid.UUID]*domain.EventoIngestao),
	}
}

func (m *mockEventoIngestaoRepository) Create(ctx context.Context, e *domain.EventoIngestao) error {
	m.eventos[e.ID] = e
	return nil
}

func (m *mockEventoIngestaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.EventoIngestao, error) {
	if e, ok := m.eventos[id]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockEventoIngestaoRepository) GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.EventoIngestao, error) {
	for _, e := range m.eventos {
		if e.EmpresaID == empresaID && e.IdempotencyKey == idempotencyKey {
			return e, nil
		}
	}
	return nil, nil
}

func (m *mockEventoIngestaoRepository) List(ctx context.Context, filter domain.EventoIngestaoFilter) ([]*domain.EventoIngestao, error) {
	var result []*domain.EventoIngestao
	for _, e := range m.eventos {
		if e.EmpresaID == *filter.EmpresaID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEventoIngestaoRepository) Update(ctx context.Context, e *domain.EventoIngestao) error {
	m.eventos[e.ID] = e
	return nil
}

type mockVendasResumoRepository struct{}

func newMockVendasResumoRepository() *mockVendasResumoRepository {
	return &mockVendasResumoRepository{}
}

func (m *mockVendasResumoRepository) GetResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time) (*domain.VendasResumo, error) {
	return &domain.VendasResumo{}, nil
}

func (m *mockEventoIngestaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.eventos, id)
	return nil
}

func TestSalesUsecase_IngestSalesEvent_Success(t *testing.T) {
	pdvOrigemRepo := newMockPdvOrigemRepository()
	produtoRepo := newMockProdutoCanonicoRepository()
	pedidoRepo := newMockPedidoCanonicoRepository()
	itemRepo := newMockPedidoItemCanonicoRepository()
	pagamentoRepo := newMockPedidoPagamentoCanonicoRepository()
	eventoRepo := newMockEventoIngestaoRepository()
	lojaRepo := newMockLojaRepository()
	vendasResumoRepo := newMockVendasResumoRepository()
	salesTxRepo := newMockSalesTxRepository()
	salesTxRepo.SetRepositories(pedidoRepo, itemRepo, pagamentoRepo, eventoRepo)

	uc := NewSalesUsecase(pdvOrigemRepo, produtoRepo, pedidoRepo, itemRepo, pagamentoRepo, eventoRepo, lojaRepo, vendasResumoRepo, salesTxRepo)

	empresaID := uuid.New()
	lojaID := uuid.New()
	pdvOrigemID := uuid.New()

	// Setup
	pdvOrigem := &domain.PdvOrigem{
		ID:        pdvOrigemID,
		EmpresaID: empresaID,
		Nome:      "PDV Teste",
		Codigo:    "PDV001",
		Tipo:      "gula",
		Ativo:     true,
	}
	pdvOrigemRepo.Create(context.Background(), pdvOrigem)

	loja := &domain.Loja{
		ID:          lojaID,
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	input := IngestSalesEventInput{
		EmpresaID:      empresaID,
		LojaID:         lojaID,
		PdvOrigemID:    pdvOrigemID,
		TipoEvento:     "venda",
		IdempotencyKey: "idem-001",
		Payload: map[string]interface{}{
			"numero_externo":  "PED-001",
			"data_hora":       time.Now().Format(time.RFC3339),
			"status":          "concluido",
			"valor_total":     100.50,
			"valor_desconto":  5.00,
			"valor_acrescimo": 0.00,
			"cliente_nome":    "Cliente Teste",
			"itens": []interface{}{
				map[string]interface{}{
					"nome":           "Produto A",
					"quantidade":     2.0,
					"preco_unitario": 50.00,
					"valor_total":    100.00,
					"desconto":       0.00,
				},
			},
			"pagamentos": []interface{}{
				map[string]interface{}{
					"forma_pagamento": "dinheiro",
					"valor":           100.50,
					"taxa":            0.00,
					"troco":           0.00,
					"status":          "aprovado",
				},
			},
		},
	}

	output, err := uc.IngestSalesEvent(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.JaProcessado)
	assert.NotEqual(t, uuid.Nil, output.EventoID)
	assert.NotEqual(t, uuid.Nil, output.PedidoID)
}

func TestSalesUsecase_IngestSalesEvent_Idempotency(t *testing.T) {
	pdvOrigemRepo := newMockPdvOrigemRepository()
	produtoRepo := newMockProdutoCanonicoRepository()
	pedidoRepo := newMockPedidoCanonicoRepository()
	itemRepo := newMockPedidoItemCanonicoRepository()
	pagamentoRepo := newMockPedidoPagamentoCanonicoRepository()
	eventoRepo := newMockEventoIngestaoRepository()
	lojaRepo := newMockLojaRepository()
	vendasResumoRepo := newMockVendasResumoRepository()
	salesTxRepo := newMockSalesTxRepository()
	salesTxRepo.SetRepositories(pedidoRepo, itemRepo, pagamentoRepo, eventoRepo)

	uc := NewSalesUsecase(pdvOrigemRepo, produtoRepo, pedidoRepo, itemRepo, pagamentoRepo, eventoRepo, lojaRepo, vendasResumoRepo, salesTxRepo)

	empresaID := uuid.New()
	lojaID := uuid.New()
	pdvOrigemID := uuid.New()

	// Setup
	pdvOrigem := &domain.PdvOrigem{
		ID:        pdvOrigemID,
		EmpresaID: empresaID,
		Nome:      "PDV Teste",
		Codigo:    "PDV001",
		Tipo:      "gula",
		Ativo:     true,
	}
	pdvOrigemRepo.Create(context.Background(), pdvOrigem)

	loja := &domain.Loja{
		ID:          lojaID,
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	input := IngestSalesEventInput{
		EmpresaID:      empresaID,
		LojaID:         lojaID,
		PdvOrigemID:    pdvOrigemID,
		TipoEvento:     "venda",
		IdempotencyKey: "idem-002",
		Payload: map[string]interface{}{
			"numero_externo": "PED-002",
			"data_hora":      time.Now().Format(time.RFC3339),
			"status":         "concluido",
			"valor_total":    50.00,
			"itens": []interface{}{
				map[string]interface{}{
					"nome":           "Produto B",
					"quantidade":     1.0,
					"preco_unitario": 50.00,
					"valor_total":    50.00,
				},
			},
			"pagamentos": []interface{}{
				map[string]interface{}{
					"forma_pagamento": "cartao",
					"valor":           50.00,
					"status":          "aprovado",
				},
			},
		},
	}

	// First ingestion
	output1, err := uc.IngestSalesEvent(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, output1.JaProcessado)

	// Second ingestion with same idempotency key
	output2, err := uc.IngestSalesEvent(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, output2.JaProcessado)
	assert.Equal(t, output1.EventoID, output2.EventoID)
	assert.Equal(t, output1.PedidoID, output2.PedidoID)
}

func TestSalesUsecase_IngestSalesEvent_InvalidPdvOrigem(t *testing.T) {
	pdvOrigemRepo := newMockPdvOrigemRepository()
	produtoRepo := newMockProdutoCanonicoRepository()
	pedidoRepo := newMockPedidoCanonicoRepository()
	itemRepo := newMockPedidoItemCanonicoRepository()
	pagamentoRepo := newMockPedidoPagamentoCanonicoRepository()
	eventoRepo := newMockEventoIngestaoRepository()
	lojaRepo := newMockLojaRepository()
	vendasResumoRepo := newMockVendasResumoRepository()
	salesTxRepo := newMockSalesTxRepository()
	salesTxRepo.SetRepositories(pedidoRepo, itemRepo, pagamentoRepo, eventoRepo)

	uc := NewSalesUsecase(pdvOrigemRepo, produtoRepo, pedidoRepo, itemRepo, pagamentoRepo, eventoRepo, lojaRepo, vendasResumoRepo, salesTxRepo)

	empresaID := uuid.New()
	lojaID := uuid.New()
	pdvOrigemID := uuid.New() // non-existent

	loja := &domain.Loja{
		ID:          lojaID,
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	input := IngestSalesEventInput{
		EmpresaID:      empresaID,
		LojaID:         lojaID,
		PdvOrigemID:    pdvOrigemID,
		TipoEvento:     "venda",
		IdempotencyKey: "idem-003",
		Payload: map[string]interface{}{
			"numero_externo": "PED-003",
			"data_hora":      time.Now().Format(time.RFC3339),
			"valor_total":    100.00,
			"itens": []interface{}{
				map[string]interface{}{
					"nome":           "Produto C",
					"quantidade":     1.0,
					"preco_unitario": 100.00,
					"valor_total":    100.00,
				},
			},
			"pagamentos": []interface{}{
				map[string]interface{}{
					"forma_pagamento": "pix",
					"valor":           100.00,
					"status":          "aprovado",
				},
			},
		},
	}

	_, err := uc.IngestSalesEvent(context.Background(), input)
	assert.ErrorIs(t, err, ErrPdvOrigemNaoEncontrado)
}
