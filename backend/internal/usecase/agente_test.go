package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/adapter/ai"
	"github.com/intelligate/erp/backend/internal/domain"
)

// ============================================================
// Fakes do agente
// ============================================================

type fakeAgenteRepository struct {
	sessoes   map[uuid.UUID]*domain.AgenteSessao
	mensagens map[uuid.UUID][]*domain.AgenteMensagem
}

func newFakeAgenteRepository() *fakeAgenteRepository {
	return &fakeAgenteRepository{
		sessoes:   make(map[uuid.UUID]*domain.AgenteSessao),
		mensagens: make(map[uuid.UUID][]*domain.AgenteMensagem),
	}
}

func (r *fakeAgenteRepository) CreateSessao(_ context.Context, s *domain.AgenteSessao) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	cp := *s
	r.sessoes[s.ID] = &cp
	return nil
}

func (r *fakeAgenteRepository) GetSessao(_ context.Context, id uuid.UUID) (*domain.AgenteSessao, error) {
	s, ok := r.sessoes[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeAgenteRepository) ListSessoes(_ context.Context, filter domain.AgenteSessaoFilter) ([]*domain.AgenteSessao, error) {
	out := make([]*domain.AgenteSessao, 0)
	for _, s := range r.sessoes {
		if s.EmpresaID != filter.EmpresaID {
			continue
		}
		if filter.LojaID != nil && s.LojaID != *filter.LojaID {
			continue
		}
		cp := *s
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeAgenteRepository) UpdateSessaoStatus(_ context.Context, id uuid.UUID, status string) error {
	if s, ok := r.sessoes[id]; ok {
		s.Status = status
	}
	return nil
}

func (r *fakeAgenteRepository) CreateMensagem(_ context.Context, m *domain.AgenteMensagem) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.CreatedAt = time.Now()
	cp := *m
	r.mensagens[m.SessaoID] = append(r.mensagens[m.SessaoID], &cp)
	return nil
}

func (r *fakeAgenteRepository) ListMensagensBySessao(_ context.Context, sessaoID uuid.UUID) ([]*domain.AgenteMensagem, error) {
	return r.mensagens[sessaoID], nil
}

// fakeFichaTecnicaRepository implementa apenas o método usado pelo agente.
type fakeFichaTecnicaRepository struct {
	semFicha []*domain.CMVProdutoDetalhe
	err      error
}

func (r *fakeFichaTecnicaRepository) Create(context.Context, *domain.FichaTecnica) error { return nil }
func (r *fakeFichaTecnicaRepository) GetByID(context.Context, uuid.UUID) (*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) GetByProdutoLoja(context.Context, uuid.UUID, uuid.UUID) (*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) ListByProduto(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) ListByLoja(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) ListByEmpresa(context.Context, uuid.UUID, *bool, int, int) ([]*domain.FichaTecnica, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) Update(context.Context, *domain.FichaTecnica) error { return nil }
func (r *fakeFichaTecnicaRepository) Delete(context.Context, uuid.UUID) error            { return nil }
func (r *fakeFichaTecnicaRepository) CreateComponente(context.Context, *domain.FichaTecnicaComponente) error {
	return nil
}
func (r *fakeFichaTecnicaRepository) GetComponentes(context.Context, uuid.UUID) ([]*domain.FichaTecnicaComponente, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) UpdateComponente(context.Context, *domain.FichaTecnicaComponente) error {
	return nil
}
func (r *fakeFichaTecnicaRepository) DeleteComponente(context.Context, uuid.UUID) error { return nil }
func (r *fakeFichaTecnicaRepository) DeleteComponentesByFicha(context.Context, uuid.UUID) error {
	return nil
}
func (r *fakeFichaTecnicaRepository) GetCMVResumo(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time) (*domain.CMVResumo, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) GetCMVProdutosDetalhe(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time, int, int) ([]*domain.CMVProdutoDetalhe, error) {
	return nil, nil
}
func (r *fakeFichaTecnicaRepository) GetProdutosSemFichaTecnica(context.Context, uuid.UUID, *uuid.UUID, time.Time, time.Time, int, int) ([]*domain.CMVProdutoDetalhe, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.semFicha, nil
}

// fakeEstoqueRepositoryAgente implementa o EstoqueRepository para o agente.
type fakeEstoqueRepositoryAgente struct {
	perdas *domain.ResumoPerdas
}

func (r *fakeEstoqueRepositoryAgente) GetSaldo(context.Context, uuid.UUID, uuid.UUID) (*domain.SaldoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepositoryAgente) ListSaldos(context.Context, domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepositoryAgente) ListMovimentacoes(context.Context, domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	return nil, nil
}
func (r *fakeEstoqueRepositoryAgente) GetResumoPerdas(_ context.Context, filter domain.ResumoPerdasFilter) (*domain.ResumoPerdas, error) {
	if r.perdas != nil {
		return r.perdas, nil
	}
	return &domain.ResumoPerdas{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Itens:      []domain.ResumoPerdasItem{},
	}, nil
}
func (r *fakeEstoqueRepositoryAgente) WithTx(context.Context, func(domain.EstoqueTxRepository) error) error {
	return nil
}

// fakeLLM captura as mensagens e devolve uma resposta configurada.
type fakeLLM struct {
	response string
	err      error
	lastMsgs []ai.Message
}

func (f *fakeLLM) Provider() string { return "fake" }
func (f *fakeLLM) Generate(_ context.Context, messages []ai.Message) (string, error) {
	f.lastMsgs = messages
	if f.err != nil {
		return "", f.err
	}
	return f.response, nil
}

// ============================================================
// Fixture
// ============================================================

type agenteFixture struct {
	uc            *AgenteUsecase
	agenteRepo    *fakeAgenteRepository
	llm           *fakeLLM
	empresaID     uuid.UUID
	lojaID        uuid.UUID
	usuarioID     uuid.UUID
	auditoriaRepo *mockAuditoriaRepository
}

func newAgenteFixture(t *testing.T, llm *fakeLLM) *agenteFixture {
	t.Helper()
	ctx := context.Background()

	empresaID := uuid.New()
	lojaID := uuid.New()
	usuarioID := uuid.New()

	lojaRepo := newMockLojaRepository()
	require.NoError(t, lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaID,
		EmpresaID: empresaID,
		Nome:      "Loja Centro",
		Ativo:     true,
	}))

	dashboardRepo := &mockDashboardRepository{
		response: &domain.DashboardResponse{
			EmpresaID: empresaID,
			LojaID:    lojaID,
			Vendas: domain.DashboardVendas{
				TotalVendas:  1000,
				TotalPedidos: 50,
				TicketMedio:  20,
			},
			CMV: domain.DashboardCMV{
				TotalCMV:             400,
				PercentualCMVReceita: 40,
				ProdutosSemFicha:     2,
			},
			Financeiro: domain.DashboardFinanceiro{
				ContasPagarVencidas:     1,
				ValorContasPagarVencido: 150,
			},
			DRE: domain.DashboardDRE{ResultadoLiquido: -50},
		},
	}

	fichaRepo := &fakeFichaTecnicaRepository{
		semFicha: []*domain.CMVProdutoDetalhe{
			{ProdutoNome: "Prato X", ReceitaTotal: 200},
		},
	}
	estoqueRepo := &fakeEstoqueRepositoryAgente{
		perdas: &domain.ResumoPerdas{CustoTotal: 30, Itens: []domain.ResumoPerdasItem{}},
	}
	auditoriaRepo := newMockAuditoriaRepository()
	agenteRepo := newFakeAgenteRepository()

	uc := NewAgenteUsecase(agenteRepo, dashboardRepo, fichaRepo, estoqueRepo, lojaRepo, auditoriaRepo, llm)

	return &agenteFixture{
		uc:            uc,
		agenteRepo:    agenteRepo,
		llm:           llm,
		empresaID:     empresaID,
		lojaID:        lojaID,
		usuarioID:     usuarioID,
		auditoriaRepo: auditoriaRepo,
	}
}

func respostaValidaJSON() string {
	return `{"resumo":"Vendas estáveis","diagnostico":"CMV elevado","evidencias":["CMV 40% no período"],"acoes_recomendadas":["Revisar fichas técnicas"],"limitacoes":["Sem dados de fornecedores"]}`
}

// ============================================================
// Testes
// ============================================================

func TestAgenteUsecase_CriarSessao_Sucesso(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})

	sessao, err := f.uc.CriarSessao(context.Background(), domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
		Titulo:    "Análise de outubro",
	})
	require.NoError(t, err)
	require.NotNil(t, sessao)
	assert.Equal(t, domain.AgenteSessaoStatusAtiva, sessao.Status)
	assert.Equal(t, "Análise de outubro", sessao.Titulo)
	assert.Len(t, f.auditoriaRepo.auditorias, 1)
}

func TestAgenteUsecase_CriarSessao_LojaNaoPertence(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})

	_, err := f.uc.CriarSessao(context.Background(), domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    uuid.New(),
		UsuarioID: f.usuarioID,
	})
	assert.ErrorIs(t, err, ErrAgenteLojaNaoPertence)
}

func TestAgenteUsecase_Perguntar_Sucesso(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})
	ctx := context.Background()

	sessao, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	resultado, err := f.uc.Perguntar(ctx, domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		UsuarioID: f.usuarioID,
		SessaoID:  sessao.ID,
		Pergunta:  "Como está o CMV?",
	})
	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, "Vendas estáveis", resultado.Resposta.Resumo)
	assert.Equal(t, "CMV elevado", resultado.Resposta.Diagnostico)
	require.Len(t, resultado.Resposta.Evidencias, 1)
	require.Len(t, resultado.Resposta.AcoesRecomendadas, 1)
	require.Len(t, resultado.Resposta.Limitacoes, 1)

	// Duas mensagens persistidas: pergunta do usuário e resposta do agente.
	mensagens := f.agenteRepo.mensagens[sessao.ID]
	require.Len(t, mensagens, 2)
	assert.Equal(t, domain.AgentePapelUsuario, mensagens[0].Papel)
	assert.Equal(t, domain.AgentePapelAgente, mensagens[1].Papel)

	// O contexto usado é armazenado para auditoria/rastreabilidade.
	require.NotEmpty(t, mensagens[0].ContextoUsado)
	var contexto domain.AgenteContexto
	require.NoError(t, json.Unmarshal(mensagens[0].ContextoUsado, &contexto))
	assert.Equal(t, f.lojaID, contexto.LojaID)
	assert.NotEmpty(t, contexto.Alertas)

	// O prompt enviado ao LLM contém o contexto e a pergunta.
	require.Len(t, f.llm.lastMsgs, 2)
	assert.Equal(t, ai.RoleSystem, f.llm.lastMsgs[0].Role)
	assert.Contains(t, f.llm.lastMsgs[1].Content, "Como está o CMV?")
	assert.Contains(t, f.llm.lastMsgs[1].Content, "CONTEXTO ESTRUTURADO")

	// Auditoria registrada (criação de sessão + pergunta).
	assert.GreaterOrEqual(t, len(f.auditoriaRepo.auditorias), 2)
}

func TestAgenteUsecase_Perguntar_PerguntaVazia(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})

	_, err := f.uc.Perguntar(context.Background(), domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		SessaoID:  uuid.New(),
		Pergunta:  "   ",
	})
	assert.ErrorIs(t, err, ErrAgentePerguntaObrigatoria)
}

func TestAgenteUsecase_Perguntar_SessaoNaoEncontrada(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})

	_, err := f.uc.Perguntar(context.Background(), domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		SessaoID:  uuid.New(),
		Pergunta:  "Pergunta válida",
	})
	assert.ErrorIs(t, err, ErrAgenteSessaoNaoEncontrada)
}

func TestAgenteUsecase_Perguntar_RespostaInvalida(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: "isto não é json"})
	ctx := context.Background()

	sessao, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	_, err = f.uc.Perguntar(ctx, domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		UsuarioID: f.usuarioID,
		SessaoID:  sessao.ID,
		Pergunta:  "Pergunta",
	})
	assert.ErrorIs(t, err, ErrAgenteRespostaInvalida)
}

func TestAgenteUsecase_Perguntar_ErroProvedor(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{err: errors.New("timeout")})
	ctx := context.Background()

	sessao, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	_, err = f.uc.Perguntar(ctx, domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		UsuarioID: f.usuarioID,
		SessaoID:  sessao.ID,
		Pergunta:  "Pergunta",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provedor de IA")
}

func TestAgenteUsecase_ObterSessao_Sucesso(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})
	ctx := context.Background()

	sessao, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	_, err = f.uc.Perguntar(ctx, domain.PerguntarAgenteInput{
		EmpresaID: f.empresaID,
		UsuarioID: f.usuarioID,
		SessaoID:  sessao.ID,
		Pergunta:  "Pergunta",
	})
	require.NoError(t, err)

	detalhe, err := f.uc.ObterSessao(ctx, f.empresaID, sessao.ID)
	require.NoError(t, err)
	require.NotNil(t, detalhe)
	assert.Equal(t, sessao.ID, detalhe.Sessao.ID)
	assert.Len(t, detalhe.Mensagens, 2)
}

func TestAgenteUsecase_ObterSessao_OutraEmpresa(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})
	ctx := context.Background()

	sessao, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	_, err = f.uc.ObterSessao(ctx, uuid.New(), sessao.ID)
	assert.ErrorIs(t, err, ErrAgenteSessaoNaoEncontrada)
}

func TestAgenteUsecase_ListarSessoes(t *testing.T) {
	f := newAgenteFixture(t, &fakeLLM{response: respostaValidaJSON()})
	ctx := context.Background()

	_, err := f.uc.CriarSessao(ctx, domain.CriarAgenteSessaoInput{
		EmpresaID: f.empresaID,
		LojaID:    f.lojaID,
		UsuarioID: f.usuarioID,
	})
	require.NoError(t, err)

	sessoes, err := f.uc.ListarSessoes(ctx, domain.AgenteSessaoFilter{EmpresaID: f.empresaID})
	require.NoError(t, err)
	assert.Len(t, sessoes, 1)
}

func TestParseAgenteResposta_RemoveCercasDeCodigo(t *testing.T) {
	raw := "```json\n" + respostaValidaJSON() + "\n```"
	resposta, err := parseAgenteResposta(raw)
	require.NoError(t, err)
	assert.Equal(t, "Vendas estáveis", resposta.Resumo)
}

func TestParseAgenteResposta_CamposAusentes(t *testing.T) {
	_, err := parseAgenteResposta(`{"evidencias":[]}`)
	assert.ErrorIs(t, err, ErrAgenteRespostaInvalida)
}
