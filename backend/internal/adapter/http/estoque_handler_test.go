package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Fakes do pacote http
// ============================================================

type fakeEstoqueRepoHandler struct {
	saldos map[string]*domain.SaldoEstoque
	movs   []*domain.MovimentacaoEstoque
}

func newFakeEstoqueRepoHandler() *fakeEstoqueRepoHandler {
	return &fakeEstoqueRepoHandler{saldos: make(map[string]*domain.SaldoEstoque)}
}

func chaveSaldoHandler(insumoID, lojaID uuid.UUID) string {
	return insumoID.String() + ":" + lojaID.String()
}

func (r *fakeEstoqueRepoHandler) GetSaldo(_ context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	s, ok := r.saldos[chaveSaldoHandler(insumoID, lojaID)]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeEstoqueRepoHandler) ListSaldos(_ context.Context, filter domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	out := make([]*domain.SaldoEstoque, 0)
	for _, s := range r.saldos {
		if s.EmpresaID == filter.EmpresaID {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeEstoqueRepoHandler) ListMovimentacoes(_ context.Context, _ domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	return r.movs, nil
}

func (r *fakeEstoqueRepoHandler) GetResumoPerdas(_ context.Context, filter domain.ResumoPerdasFilter) (*domain.ResumoPerdas, error) {
	return &domain.ResumoPerdas{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Itens:      []domain.ResumoPerdasItem{},
	}, nil
}

func (r *fakeEstoqueRepoHandler) WithTx(ctx context.Context, fn func(tx domain.EstoqueTxRepository) error) error {
	return fn(&fakeEstoqueTxHandler{repo: r})
}

type fakeEstoqueTxHandler struct {
	repo *fakeEstoqueRepoHandler
}

func (t *fakeEstoqueTxHandler) GetSaldoForUpdate(ctx context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	return t.repo.GetSaldo(ctx, insumoID, lojaID)
}

func (t *fakeEstoqueTxHandler) UpsertSaldo(_ context.Context, saldo *domain.SaldoEstoque) (*domain.SaldoEstoque, error) {
	cp := *saldo
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	t.repo.saldos[chaveSaldoHandler(saldo.InsumoID, saldo.LojaID)] = &cp
	res := cp
	return &res, nil
}

func (t *fakeEstoqueTxHandler) CreateMovimentacao(_ context.Context, mov *domain.MovimentacaoEstoque) (*domain.MovimentacaoEstoque, error) {
	cp := *mov
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	cp.CreatedAt = time.Now()
	t.repo.movs = append(t.repo.movs, &cp)
	res := cp
	return &res, nil
}

type fakeInsumoRepoHandler struct {
	insumos map[uuid.UUID]*domain.Insumo
}

func newFakeInsumoRepoHandler() *fakeInsumoRepoHandler {
	return &fakeInsumoRepoHandler{insumos: make(map[uuid.UUID]*domain.Insumo)}
}

func (m *fakeInsumoRepoHandler) Create(_ context.Context, i *domain.Insumo) error {
	m.insumos[i.ID] = i
	return nil
}

func (m *fakeInsumoRepoHandler) GetByID(_ context.Context, id uuid.UUID) (*domain.Insumo, error) {
	if i, ok := m.insumos[id]; ok {
		return i, nil
	}
	return nil, errors.New("insumo não encontrado")
}

func (m *fakeInsumoRepoHandler) GetByNome(_ context.Context, _ uuid.UUID, _ string) (*domain.Insumo, error) {
	return nil, errors.New("não implementado")
}

func (m *fakeInsumoRepoHandler) ListByEmpresa(_ context.Context, _ uuid.UUID, _ *bool, _, _ int) ([]*domain.Insumo, error) {
	return nil, nil
}

func (m *fakeInsumoRepoHandler) Update(_ context.Context, _ *domain.Insumo) error { return nil }
func (m *fakeInsumoRepoHandler) Delete(_ context.Context, _ uuid.UUID) error      { return nil }

func (m *fakeInsumoRepoHandler) CreateCusto(_ context.Context, _ *domain.InsumoCusto) error {
	return nil
}

func (m *fakeInsumoRepoHandler) GetCustoVigente(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *fakeInsumoRepoHandler) GetCustoAtual(_ context.Context, _ uuid.UUID) (*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *fakeInsumoRepoHandler) ListCustosByInsumo(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *fakeInsumoRepoHandler) GetForUpdate(_ context.Context, id uuid.UUID) (*domain.Insumo, error) {
	return m.GetByID(context.Background(), id)
}

func (m *fakeInsumoRepoHandler) UpdateCustoUnitario(_ context.Context, _ uuid.UUID, _ float64) error {
	return nil
}

type fakeAuditoriaRepoHandler struct {
	criadas []*domain.Auditoria
}

func (m *fakeAuditoriaRepoHandler) Create(_ context.Context, a *domain.Auditoria) error {
	m.criadas = append(m.criadas, a)
	return nil
}

func (m *fakeAuditoriaRepoHandler) List(_ context.Context, _ domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	return m.criadas, nil
}

// ============================================================
// Helpers
// ============================================================

func newEstoqueHandlerFixture(t *testing.T) (*EstoqueHandler, uuid.UUID, uuid.UUID, uuid.UUID, *fakeEstoqueRepoHandler) {
	t.Helper()
	ctx := context.Background()

	empresaID := uuid.New()
	lojaID := uuid.New()
	insumoID := uuid.New()

	lojaRepo := newHandlerMockLojaRepo()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Ativo:     true,
	}))

	insumoRepo := newFakeInsumoRepoHandler()
	require.NoError(t, insumoRepo.Create(ctx, &domain.Insumo{
		ID:            insumoID,
		EmpresaID:     empresaID,
		Nome:          "Farinha",
		UnidadeMedida: "kg",
		Ativo:         true,
	}))

	estoqueRepo := newFakeEstoqueRepoHandler()
	estoqueRepo.saldos[chaveSaldoHandler(insumoID, lojaID)] = &domain.SaldoEstoque{
		ID:                 uuid.New(),
		EmpresaID:          empresaID,
		LojaID:             lojaID,
		InsumoID:           insumoID,
		Quantidade:         10,
		UnidadeMedida:      "kg",
		CustoMedioUnitario: 5,
	}

	uc := usecase.NewEstoqueUsecase(estoqueRepo, lojaRepo, insumoRepo, &fakeAuditoriaRepoHandler{})
	return NewEstoqueHandler(uc), empresaID, lojaID, insumoID, estoqueRepo
}

func newEstoqueRequest(method, target string, empresaID uuid.UUID, body []byte) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, target, reader)
	ctx := context.WithValue(req.Context(), EmpresaContextKey, empresaID)
	ctx = context.WithValue(ctx, UserContextKey, &domain.Usuario{ID: uuid.New(), EmpresaID: empresaID})
	if empresaID == uuid.Nil {
		ctx = context.WithValue(req.Context(), UserContextKey, &domain.Usuario{ID: uuid.New()})
	}
	return req.WithContext(ctx)
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ============================================================
// Testes
// ============================================================

func TestEstoqueHandler_RegistrarSaida_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, insumoID, _ := newEstoqueHandlerFixture(t)

	body, _ := json.Marshal(map[string]interface{}{
		"loja_id":    lojaID.String(),
		"insumo_id":  insumoID.String(),
		"tipo":       "saida_consumo",
		"quantidade": 3,
	})

	req := newEstoqueRequest(http.MethodPost, "/api/v1/estoque/saidas", empresaID, body)
	w := httptest.NewRecorder()
	handler.RegistrarSaida(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var mov domain.MovimentacaoEstoque
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &mov))
	assert.Equal(t, domain.MovimentacaoSaidaConsumo, mov.Tipo)
	assert.InDelta(t, 7, mov.SaldoPosterior, 1e-9)
}

func TestEstoqueHandler_RegistrarSaida_SaldoInsuficiente(t *testing.T) {
	handler, empresaID, lojaID, insumoID, _ := newEstoqueHandlerFixture(t)

	body, _ := json.Marshal(map[string]interface{}{
		"loja_id":    lojaID.String(),
		"insumo_id":  insumoID.String(),
		"tipo":       "saida_consumo",
		"quantidade": 100,
	})

	req := newEstoqueRequest(http.MethodPost, "/api/v1/estoque/saidas", empresaID, body)
	w := httptest.NewRecorder()
	handler.RegistrarSaida(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestEstoqueHandler_RegistrarSaida_SemEmpresa(t *testing.T) {
	handler, _, _, _, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodPost, "/api/v1/estoque/saidas", uuid.Nil, []byte(`{}`))
	w := httptest.NewRecorder()
	handler.RegistrarSaida(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_RegistrarAjuste_MotivoObrigatorio(t *testing.T) {
	handler, empresaID, lojaID, insumoID, _ := newEstoqueHandlerFixture(t)

	body, _ := json.Marshal(map[string]interface{}{
		"loja_id":    lojaID.String(),
		"insumo_id":  insumoID.String(),
		"tipo":       "ajuste_positivo",
		"quantidade": 5,
		"motivo":     "",
	})

	req := newEstoqueRequest(http.MethodPost, "/api/v1/estoque/ajustes", empresaID, body)
	w := httptest.NewRecorder()
	handler.RegistrarAjuste(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_Transferir_MesmaLoja(t *testing.T) {
	handler, empresaID, lojaID, insumoID, _ := newEstoqueHandlerFixture(t)

	body, _ := json.Marshal(map[string]interface{}{
		"loja_origem_id":  lojaID.String(),
		"loja_destino_id": lojaID.String(),
		"insumo_id":       insumoID.String(),
		"quantidade":      1,
	})

	req := newEstoqueRequest(http.MethodPost, "/api/v1/estoque/transferencias", empresaID, body)
	w := httptest.NewRecorder()
	handler.Transferir(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_ListSaldos_SemEmpresa(t *testing.T) {
	handler, _, _, _, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet, "/api/v1/estoque/saldos", uuid.Nil, nil)
	w := httptest.NewRecorder()
	handler.ListSaldos(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_ListSaldos_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, _, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet, "/api/v1/estoque/saldos?loja_id="+lojaID.String(), empresaID, nil)
	req = withURLParam(req, "insumo_id", "")
	w := httptest.NewRecorder()
	handler.ListSaldos(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var saldos []domain.SaldoEstoque
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &saldos))
	require.Len(t, saldos, 1)
	assert.InDelta(t, 10, saldos[0].Quantidade, 1e-9)
}

func TestEstoqueHandler_GetSaldo_SemLoja(t *testing.T) {
	handler, empresaID, _, insumoID, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet, "/api/v1/estoque/saldos/"+insumoID.String(), empresaID, nil)
	req = withURLParam(req, "insumo_id", insumoID.String())
	w := httptest.NewRecorder()
	handler.GetSaldo(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_GetSaldo_NaoEncontrado(t *testing.T) {
	handler, empresaID, lojaID, _, _ := newEstoqueHandlerFixture(t)

	outroInsumoID := uuid.New()
	req := newEstoqueRequest(http.MethodGet, "/api/v1/estoque/saldos/"+outroInsumoID.String()+"?loja_id="+lojaID.String(), empresaID, nil)
	req = withURLParam(req, "insumo_id", outroInsumoID.String())
	w := httptest.NewRecorder()
	handler.GetSaldo(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEstoqueHandler_GetResumoPerdas_SemLoja(t *testing.T) {
	handler, empresaID, _, _, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet, "/api/v1/estoque/perdas?data_inicio=2026-01-01&data_fim=2026-01-31", empresaID, nil)
	w := httptest.NewRecorder()
	handler.GetResumoPerdas(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEstoqueHandler_GetResumoPerdas_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, _, _ := newEstoqueHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet,
		"/api/v1/estoque/perdas?loja_id="+lojaID.String()+"&data_inicio=2026-01-01&data_fim=2026-01-31",
		empresaID, nil)
	w := httptest.NewRecorder()
	handler.GetResumoPerdas(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
