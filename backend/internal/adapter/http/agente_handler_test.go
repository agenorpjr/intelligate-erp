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

	"github.com/intelligate/erp/backend/internal/adapter/ai"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Fakes do agente (pacote http)
// ============================================================

type fakeAgenteRepoHandler struct {
	sessoes   map[uuid.UUID]*domain.AgenteSessao
	mensagens map[uuid.UUID][]*domain.AgenteMensagem
}

func newFakeAgenteRepoHandler() *fakeAgenteRepoHandler {
	return &fakeAgenteRepoHandler{
		sessoes:   make(map[uuid.UUID]*domain.AgenteSessao),
		mensagens: make(map[uuid.UUID][]*domain.AgenteMensagem),
	}
}

func (r *fakeAgenteRepoHandler) CreateSessao(_ context.Context, s *domain.AgenteSessao) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	cp := *s
	r.sessoes[s.ID] = &cp
	return nil
}

func (r *fakeAgenteRepoHandler) GetSessao(_ context.Context, id uuid.UUID) (*domain.AgenteSessao, error) {
	s, ok := r.sessoes[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeAgenteRepoHandler) ListSessoes(_ context.Context, filter domain.AgenteSessaoFilter) ([]*domain.AgenteSessao, error) {
	out := make([]*domain.AgenteSessao, 0)
	for _, s := range r.sessoes {
		if s.EmpresaID == filter.EmpresaID {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeAgenteRepoHandler) UpdateSessaoStatus(_ context.Context, id uuid.UUID, status string) error {
	if s, ok := r.sessoes[id]; ok {
		s.Status = status
	}
	return nil
}

func (r *fakeAgenteRepoHandler) CreateMensagem(_ context.Context, m *domain.AgenteMensagem) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.CreatedAt = time.Now()
	cp := *m
	r.mensagens[m.SessaoID] = append(r.mensagens[m.SessaoID], &cp)
	return nil
}

func (r *fakeAgenteRepoHandler) ListMensagensBySessao(_ context.Context, sessaoID uuid.UUID) ([]*domain.AgenteMensagem, error) {
	return r.mensagens[sessaoID], nil
}

type fakeDashboardRepoHandler struct {
	response *domain.DashboardResponse
}

func (r *fakeDashboardRepoHandler) GetDashboard(_ context.Context, req domain.DashboardRequest) (*domain.DashboardResponse, error) {
	if r.response != nil {
		return r.response, nil
	}
	return &domain.DashboardResponse{EmpresaID: req.EmpresaID, LojaID: req.LojaID}, nil
}

type fakeFichaRepoHandler struct{}

func (r *fakeFichaRepoHandler) Create(context.Context, *domain.FichaTecnica) error { return nil }
func (r *fakeFichaRepoHandler) GetByID(context.Context, uuid.UUID) (*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) GetByProdutoLoja(context.Context, uuid.UUID, uuid.UUID) (*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) ListByProduto(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) ListByLoja(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) ListByEmpresa(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) Update(context.Context, *domain.FichaTecnica) error { return nil }
func (r *fakeFichaRepoHandler) Delete(context.Context, uuid.UUID) error            { return nil }
func (r *fakeFichaRepoHandler) CreateComponente(context.Context, *domain.FichaTecnicaComponente) error {
	return nil
}
func (r *fakeFichaRepoHandler) GetComponentes(context.Context, uuid.UUID) ([]*domain.FichaTecnicaComponente, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) UpdateComponente(context.Context, *domain.FichaTecnicaComponente) error {
	return nil
}
func (r *fakeFichaRepoHandler) DeleteComponente(context.Context, uuid.UUID) error { return nil }
func (r *fakeFichaRepoHandler) DeleteComponentesByFicha(context.Context, uuid.UUID) error {
	return nil
}
func (r *fakeFichaRepoHandler) GetCMVResumo(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time) (*domain.CMVResumo, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) GetCMVProdutosDetalhe(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time, int, int) ([]*domain.CMVProdutoDetalhe, error) {
	return nil, nil
}
func (r *fakeFichaRepoHandler) GetProdutosSemFichaTecnica(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time, int, int) ([]*domain.CMVProdutoDetalhe, error) {
	return nil, nil
}

type fakeEstoqueRepoAgenteHandler struct{}

func (r *fakeEstoqueRepoAgenteHandler) GetSaldo(context.Context, uuid.UUID, uuid.UUID) (*domain.SaldoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepoAgenteHandler) ListSaldos(context.Context, domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepoAgenteHandler) ListMovimentacoes(context.Context, domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepoAgenteHandler) GetResumoPerdas(_ context.Context, filter domain.ResumoPerdasFilter) (*domain.ResumoPerdas, error) {
	return &domain.ResumoPerdas{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Itens:      []domain.ResumoPerdasItem{},
	}, nil
}
func (r *fakeEstoqueRepoAgenteHandler) WithTx(context.Context, func(domain.EstoqueTxRepository) error) error {
	return nil
}

// ============================================================
// Fixture
// ============================================================

func newAgenteHandlerFixture(t *testing.T) (*AgenteHandler, uuid.UUID, uuid.UUID, *fakeAgenteRepoHandler) {
	t.Helper()
	ctx := context.Background()

	empresaID := uuid.New()
	lojaID := uuid.New()

	lojaRepo := newHandlerMockLojaRepo()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Nome:      "Loja Centro",
		Ativo:     true,
	}))

	agenteRepo := newFakeAgenteRepoHandler()
	uc := usecase.NewAgenteUsecase(
		agenteRepo,
		&fakeDashboardRepoHandler{},
		&fakeFichaRepoHandler{},
		&fakeEstoqueRepoAgenteHandler{},
		lojaRepo,
		&fakeAuditoriaRepoHandler{},
		ai.NewMockClient(),
	)
	return NewAgenteHandler(uc), empresaID, lojaID, agenteRepo
}

// ============================================================
// Testes
// ============================================================

func TestAgenteHandler_CriarSessao_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, _ := newAgenteHandlerFixture(t)

	body, _ := json.Marshal(map[string]any{"loja_id": lojaID.String(), "titulo": "Análise"})
	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", empresaID, body)
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var sessao domain.AgenteSessao
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessao))
	assert.Equal(t, lojaID, sessao.LojaID)
	assert.Equal(t, domain.AgenteSessaoStatusAtiva, sessao.Status)
}

func TestAgenteHandler_CriarSessao_SemEmpresa(t *testing.T) {
	handler, _, _, _ := newAgenteHandlerFixture(t)

	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", uuid.Nil, []byte(`{}`))
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAgenteHandler_CriarSessao_LojaInvalida(t *testing.T) {
	handler, empresaID, _, _ := newAgenteHandlerFixture(t)

	body, _ := json.Marshal(map[string]any{"loja_id": uuid.New().String()})
	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", empresaID, body)
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAgenteHandler_ListarSessoes_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, _ := newAgenteHandlerFixture(t)

	body, _ := json.Marshal(map[string]any{"loja_id": lojaID.String()})
	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", empresaID, body)
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	reqList := newEstoqueRequest(http.MethodGet, "/api/v1/agente/sessoes", empresaID, nil)
	wList := httptest.NewRecorder()
	handler.ListarSessoes(wList, reqList)

	require.Equal(t, http.StatusOK, wList.Code)
	var sessoes []domain.AgenteSessao
	require.NoError(t, json.Unmarshal(wList.Body.Bytes(), &sessoes))
	assert.Len(t, sessoes, 1)
}

func TestAgenteHandler_ObterSessao_IdInvalido(t *testing.T) {
	handler, empresaID, _, _ := newAgenteHandlerFixture(t)

	req := newEstoqueRequest(http.MethodGet, "/api/v1/agente/sessoes/abc", empresaID, nil)
	req = withURLParam(req, "id", "abc")
	w := httptest.NewRecorder()
	handler.ObterSessao(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAgenteHandler_ObterSessao_NaoEncontrada(t *testing.T) {
	handler, empresaID, _, _ := newAgenteHandlerFixture(t)

	id := uuid.New()
	req := newEstoqueRequest(http.MethodGet, "/api/v1/agente/sessoes/"+id.String(), empresaID, nil)
	req = withURLParam(req, "id", id.String())
	w := httptest.NewRecorder()
	handler.ObterSessao(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAgenteHandler_Perguntar_Sucesso(t *testing.T) {
	handler, empresaID, lojaID, _ := newAgenteHandlerFixture(t)

	// Cria a sessão.
	body, _ := json.Marshal(map[string]any{"loja_id": lojaID.String()})
	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", empresaID, body)
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var sessao domain.AgenteSessao
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessao))

	// Pergunta.
	perguntaBody, _ := json.Marshal(map[string]any{"pergunta": "Como está o CMV?"})
	reqP := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes/"+sessao.ID.String()+"/perguntas", empresaID, perguntaBody)
	reqP = withURLParam(reqP, "id", sessao.ID.String())
	wP := httptest.NewRecorder()
	handler.Perguntar(wP, reqP)

	require.Equal(t, http.StatusOK, wP.Code)

	var resultado usecase.PerguntarResult
	require.NoError(t, json.Unmarshal(wP.Body.Bytes(), &resultado))
	require.NotNil(t, resultado.Resposta)
	assert.NotEmpty(t, resultado.Resposta.Resumo)
}

func TestAgenteHandler_Perguntar_PerguntaVazia(t *testing.T) {
	handler, empresaID, lojaID, _ := newAgenteHandlerFixture(t)

	body, _ := json.Marshal(map[string]any{"loja_id": lojaID.String()})
	req := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes", empresaID, body)
	w := httptest.NewRecorder()
	handler.CriarSessao(w, req)
	var sessao domain.AgenteSessao
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessao))

	perguntaBody, _ := json.Marshal(map[string]any{"pergunta": ""})
	reqP := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes/"+sessao.ID.String()+"/perguntas", empresaID, perguntaBody)
	reqP = withURLParam(reqP, "id", sessao.ID.String())
	wP := httptest.NewRecorder()
	handler.Perguntar(wP, reqP)

	assert.Equal(t, http.StatusBadRequest, wP.Code)
}

func TestAgenteHandler_Perguntar_SessaoNaoEncontrada(t *testing.T) {
	handler, empresaID, _, _ := newAgenteHandlerFixture(t)

	id := uuid.New()
	perguntaBody, _ := json.Marshal(map[string]any{"pergunta": "Pergunta"})
	reqP := newEstoqueRequest(http.MethodPost, "/api/v1/agente/sessoes/"+id.String()+"/perguntas", empresaID, perguntaBody)
	reqP = withURLParam(reqP, "id", id.String())
	wP := httptest.NewRecorder()
	handler.Perguntar(wP, reqP)

	assert.Equal(t, http.StatusNotFound, wP.Code)
}
