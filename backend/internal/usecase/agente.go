package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/adapter/ai"
	"github.com/intelligate/erp/backend/internal/domain"
)

// Erros esperados do agente de IA.
var (
	ErrAgenteSessaoObrigatoria   = errors.New("sessao_id é obrigatório")
	ErrAgenteSessaoNaoEncontrada = errors.New("sessão não encontrada")
	ErrAgenteSessaoArquivada     = errors.New("sessão está arquivada")
	ErrAgentePerguntaObrigatoria = errors.New("pergunta é obrigatória")
	ErrAgenteLojaObrigatoria     = errors.New("loja_id é obrigatório")
	ErrAgenteLojaNaoPertence     = errors.New("loja não pertence à empresa ou está inativa")
	ErrAgentePeriodoInvalido     = errors.New("período inválido")
	ErrAgenteRespostaInvalida    = errors.New("resposta do agente inválida")
)

const (
	agenteContextoProdutoLimit = 20
	agentePeriodoPadraoDias    = 30
)

// agenteSystemPrompt define as regras de comportamento do agente. O modelo só
// pode responder com base no contexto fornecido e nunca deve inventar dados.
const agenteSystemPrompt = `Você é um analista gerencial do IntelliGate ERP, especializado em food service.

Você recebe um CONTEXTO ESTRUTURADO com dados reais de uma loja em um período e uma PERGUNTA do gestor.

Regras obrigatórias:
1. Responda APENAS com base nos dados presentes no contexto fornecido.
2. NUNCA invente valores, produtos, fornecedores, causas ou períodos.
3. Se os dados forem insuficientes para responder, diga explicitamente que não há informação suficiente e explique o que falta.
4. Não faça suposições sobre causas que não estejam evidenciadas nos dados.
5. Responda em português do Brasil, de forma objetiva e prática.

Formato da resposta: retorne SOMENTE um objeto JSON válido, sem texto adicional e sem cercas de código, com exatamente os campos:
{
  "resumo": "string curta com a conclusão principal",
  "diagnostico": "string com a análise dos indicadores",
  "evidencias": ["indicador com valor e período", "..."],
  "acoes_recomendadas": ["ação prática e priorizada", "..."],
  "limitacoes": ["dado ausente ou premissa usada", "..."]
}

Cada item de "evidencias" deve citar o indicador usado, seu valor e o período.
Cada item de "acoes_recomendadas" deve ser prático e priorizado.
Cada item de "limitacoes" deve indicar dados ausentes ou premissas adotadas.`

// AgenteUsecase orquestra o agente de análise gerencial: monta o contexto com
// dados reais, consulta o LLM e persiste a conversa para auditoria.
type AgenteUsecase struct {
	agenteRepo       domain.AgenteRepository
	dashboardRepo    domain.DashboardRepository
	fichaTecnicaRepo domain.FichaTecnicaRepository
	estoqueRepo      domain.EstoqueRepository
	lojaRepo         domain.LojaRepository
	auditoriaRepo    domain.AuditoriaRepository
	llm              ai.LLMClient
}

// NewAgenteUsecase cria uma nova instância de AgenteUsecase.
func NewAgenteUsecase(
	agenteRepo domain.AgenteRepository,
	dashboardRepo domain.DashboardRepository,
	fichaTecnicaRepo domain.FichaTecnicaRepository,
	estoqueRepo domain.EstoqueRepository,
	lojaRepo domain.LojaRepository,
	auditoriaRepo domain.AuditoriaRepository,
	llm ai.LLMClient,
) *AgenteUsecase {
	return &AgenteUsecase{
		agenteRepo:       agenteRepo,
		dashboardRepo:    dashboardRepo,
		fichaTecnicaRepo: fichaTecnicaRepo,
		estoqueRepo:      estoqueRepo,
		lojaRepo:         lojaRepo,
		auditoriaRepo:    auditoriaRepo,
		llm:              llm,
	}
}

// PerguntarResult agrega a resposta estruturada e a mensagem persistida.
type PerguntarResult struct {
	Resposta *domain.AgenteResposta `json:"resposta"`
	Mensagem *domain.AgenteMensagem `json:"mensagem"`
}

// ============================================================
// Sessões
// ============================================================

// CriarSessao cria uma nova sessão de conversa do agente.
func (uc *AgenteUsecase) CriarSessao(ctx context.Context, input domain.CriarAgenteSessaoInput) (*domain.AgenteSessao, error) {
	if input.EmpresaID == uuid.Nil {
		return nil, fmt.Errorf("empresa_id é obrigatório")
	}
	if input.LojaID == uuid.Nil {
		return nil, ErrAgenteLojaObrigatoria
	}
	if err := uc.validarLoja(ctx, input.EmpresaID, input.LojaID); err != nil {
		return nil, err
	}

	titulo := strings.TrimSpace(input.Titulo)
	if titulo == "" {
		titulo = "Análise gerencial"
	}

	sessao := &domain.AgenteSessao{
		EmpresaID: input.EmpresaID,
		LojaID:    input.LojaID,
		UsuarioID: input.UsuarioID,
		Titulo:    titulo,
		Status:    domain.AgenteSessaoStatusAtiva,
	}
	if err := uc.agenteRepo.CreateSessao(ctx, sessao); err != nil {
		return nil, fmt.Errorf("falha ao criar sessão do agente: %w", err)
	}

	uc.registrarAuditoria(ctx, input.EmpresaID, input.UsuarioID, "CRIAR_SESSAO_AGENTE", sessao.ID, sessao)
	return sessao, nil
}

// ListarSessoes lista as sessões da empresa com filtros opcionais.
func (uc *AgenteUsecase) ListarSessoes(ctx context.Context, filter domain.AgenteSessaoFilter) ([]*domain.AgenteSessao, error) {
	if filter.EmpresaID == uuid.Nil {
		return nil, fmt.Errorf("empresa_id é obrigatório")
	}
	if filter.LojaID != nil {
		if err := uc.validarLoja(ctx, filter.EmpresaID, *filter.LojaID); err != nil {
			return nil, err
		}
	}
	filter.Limit, filter.Offset = normalizePagination(filter.Limit, filter.Offset)
	return uc.agenteRepo.ListSessoes(ctx, filter)
}

// ObterSessao retorna a sessão com suas mensagens.
func (uc *AgenteUsecase) ObterSessao(ctx context.Context, empresaID, sessaoID uuid.UUID) (*domain.AgenteSessaoDetalhe, error) {
	if sessaoID == uuid.Nil {
		return nil, ErrAgenteSessaoObrigatoria
	}
	sessao, err := uc.agenteRepo.GetSessao(ctx, sessaoID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar sessão do agente: %w", err)
	}
	if sessao == nil || sessao.EmpresaID != empresaID {
		return nil, ErrAgenteSessaoNaoEncontrada
	}

	mensagens, err := uc.agenteRepo.ListMensagensBySessao(ctx, sessaoID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar mensagens da sessão: %w", err)
	}
	return &domain.AgenteSessaoDetalhe{Sessao: sessao, Mensagens: mensagens}, nil
}

// ============================================================
// Pergunta ao agente
// ============================================================

// Perguntar registra a pergunta do gestor, monta o contexto com dados reais,
// consulta o LLM e persiste a resposta estruturada.
func (uc *AgenteUsecase) Perguntar(ctx context.Context, input domain.PerguntarAgenteInput) (*PerguntarResult, error) {
	if input.SessaoID == uuid.Nil {
		return nil, ErrAgenteSessaoObrigatoria
	}
	pergunta := strings.TrimSpace(input.Pergunta)
	if pergunta == "" {
		return nil, ErrAgentePerguntaObrigatoria
	}

	sessao, err := uc.agenteRepo.GetSessao(ctx, input.SessaoID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar sessão do agente: %w", err)
	}
	if sessao == nil || sessao.EmpresaID != input.EmpresaID {
		return nil, ErrAgenteSessaoNaoEncontrada
	}
	if sessao.Status != domain.AgenteSessaoStatusAtiva {
		return nil, ErrAgenteSessaoArquivada
	}

	dataInicio, dataFim, err := resolverPeriodo(input.DataInicio, input.DataFim)
	if err != nil {
		return nil, err
	}

	contexto, err := uc.montarContexto(ctx, input.EmpresaID, sessao.LojaID, dataInicio, dataFim)
	if err != nil {
		return nil, err
	}
	contextoJSON, err := json.Marshal(contexto)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar contexto do agente: %w", err)
	}

	// Persiste a pergunta do usuário com o contexto usado (rastreabilidade).
	mensagemUsuario := &domain.AgenteMensagem{
		SessaoID:        sessao.ID,
		Papel:           domain.AgentePapelUsuario,
		Conteudo:        pergunta,
		ContextoUsado:   contextoJSON,
		TokensEstimados: estimarTokens(pergunta),
	}
	if err := uc.agenteRepo.CreateMensagem(ctx, mensagemUsuario); err != nil {
		return nil, fmt.Errorf("falha ao registrar pergunta do usuário: %w", err)
	}

	// Consulta o LLM com o contexto estruturado e a pergunta.
	mensagens := []ai.Message{
		{Role: ai.RoleSystem, Content: agenteSystemPrompt},
		{Role: ai.RoleUser, Content: montarPromptUsuario(contextoJSON, pergunta)},
	}
	raw, err := uc.llm.Generate(ctx, mensagens)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar o provedor de IA: %w", err)
	}

	resposta, err := parseAgenteResposta(raw)
	if err != nil {
		// Persiste a resposta bruta para rastreabilidade antes de falhar.
		_ = uc.agenteRepo.CreateMensagem(ctx, &domain.AgenteMensagem{
			SessaoID:        sessao.ID,
			Papel:           domain.AgentePapelAgente,
			Conteudo:        raw,
			ContextoUsado:   contextoJSON,
			TokensEstimados: estimarTokens(raw),
		})
		return nil, err
	}

	conteudoAgente, err := json.Marshal(resposta)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar resposta do agente: %w", err)
	}

	mensagemAgente := &domain.AgenteMensagem{
		SessaoID:        sessao.ID,
		Papel:           domain.AgentePapelAgente,
		Conteudo:        string(conteudoAgente),
		ContextoUsado:   contextoJSON,
		TokensEstimados: estimarTokens(string(conteudoAgente)),
	}
	if err := uc.agenteRepo.CreateMensagem(ctx, mensagemAgente); err != nil {
		return nil, fmt.Errorf("falha ao registrar resposta do agente: %w", err)
	}

	uc.registrarAuditoria(ctx, input.EmpresaID, input.UsuarioID, "PERGUNTAR_AGENTE", sessao.ID, resposta)

	return &PerguntarResult{Resposta: resposta, Mensagem: mensagemAgente}, nil
}

// ============================================================
// Montagem do contexto
// ============================================================

// montarContexto reúne os dados reais da loja/período usados como base da
// análise: vendas, CMV, produtos sem ficha, despesas, contas a pagar, fluxo de
// caixa, DRE, alertas e perdas de estoque.
func (uc *AgenteUsecase) montarContexto(ctx context.Context, empresaID, lojaID uuid.UUID, dataInicio, dataFim time.Time) (*domain.AgenteContexto, error) {
	dash, err := uc.dashboardRepo.GetDashboard(ctx, domain.DashboardRequest{
		EmpresaID:       empresaID,
		LojaID:          lojaID,
		DataInicio:      dataInicio,
		DataFim:         dataFim,
		CmvProdutoLimit: agenteContextoProdutoLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao montar indicadores do dashboard: %w", err)
	}

	produtosSemFicha, err := uc.fichaTecnicaRepo.GetProdutosSemFichaTecnica(
		ctx, empresaID, &lojaID, dataInicio, dataFim, agenteContextoProdutoLimit, 0,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar produtos sem ficha técnica: %w", err)
	}

	perdas, err := uc.estoqueRepo.GetResumoPerdas(ctx, domain.ResumoPerdasFilter{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar perdas de estoque: %w", err)
	}

	contexto := &domain.AgenteContexto{
		EmpresaID:               empresaID,
		LojaID:                  lojaID,
		DataInicio:              dataInicio,
		DataFim:                 dataFim,
		Vendas:                  dash.Vendas,
		CMV:                     dash.CMV,
		Financeiro:              dash.Financeiro,
		DRE:                     dash.DRE,
		ProdutosSemFichaTecnica: produtosSemFicha,
		Perdas:                  *perdas,
	}
	contexto.Alertas = gerarAlertas(contexto)
	return contexto, nil
}

// gerarAlertas deriva pontos de atenção objetivos a partir dos indicadores.
func gerarAlertas(c *domain.AgenteContexto) []string {
	alertas := make([]string, 0)

	if c.Financeiro.ContasPagarVencidas > 0 {
		alertas = append(alertas, fmt.Sprintf(
			"%d conta(s) a pagar vencida(s), totalizando %.2f.",
			c.Financeiro.ContasPagarVencidas, c.Financeiro.ValorContasPagarVencido,
		))
	}
	if c.CMV.PercentualCMVReceita > 35 {
		alertas = append(alertas, fmt.Sprintf(
			"CMV representa %.2f%% da receita líquida no período (acima de 35%%).",
			c.CMV.PercentualCMVReceita,
		))
	}
	if c.CMV.ProdutosSemFicha > 0 {
		alertas = append(alertas, fmt.Sprintf(
			"%d produto(s) vendido(s) sem ficha técnica, comprometendo o cálculo de CMV.",
			c.CMV.ProdutosSemFicha,
		))
	}
	if c.Perdas.CustoTotal > 0 {
		alertas = append(alertas, fmt.Sprintf(
			"Perdas e desperdícios de estoque somam %.2f no período.",
			c.Perdas.CustoTotal,
		))
	}
	if c.DRE.ResultadoLiquido < 0 {
		alertas = append(alertas, fmt.Sprintf(
			"Resultado líquido negativo (%.2f) no período.", c.DRE.ResultadoLiquido,
		))
	}
	return alertas
}

// ============================================================
// Helpers
// ============================================================

// resolverPeriodo aplica um período padrão (últimos 30 dias) quando não informado.
func resolverPeriodo(dataInicio, dataFim time.Time) (time.Time, time.Time, error) {
	if dataInicio.IsZero() && dataFim.IsZero() {
		dataFim = time.Now()
		dataInicio = dataFim.AddDate(0, 0, -agentePeriodoPadraoDias)
		return dataInicio, dataFim, nil
	}
	if dataInicio.IsZero() || dataFim.IsZero() {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: data_inicio e data_fim devem ser informadas juntas", ErrAgentePeriodoInvalido)
	}
	if !dataInicio.Before(dataFim) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: data_inicio deve ser anterior a data_fim", ErrAgentePeriodoInvalido)
	}
	return dataInicio, dataFim, nil
}

// montarPromptUsuario combina o contexto estruturado e a pergunta do gestor.
func montarPromptUsuario(contextoJSON []byte, pergunta string) string {
	var sb strings.Builder
	sb.WriteString("CONTEXTO ESTRUTURADO (dados reais da loja e período):\n")
	sb.Write(contextoJSON)
	sb.WriteString("\n\nPERGUNTA DO GESTOR:\n")
	sb.WriteString(pergunta)
	return sb.String()
}

// parseAgenteResposta extrai e valida a resposta estruturada do modelo.
func parseAgenteResposta(raw string) (*domain.AgenteResposta, error) {
	cleaned := stripCodeFences(raw)
	if cleaned == "" {
		return nil, fmt.Errorf("%w: resposta vazia", ErrAgenteRespostaInvalida)
	}

	var resposta domain.AgenteResposta
	if err := json.Unmarshal([]byte(cleaned), &resposta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAgenteRespostaInvalida, err)
	}
	if strings.TrimSpace(resposta.Resumo) == "" && strings.TrimSpace(resposta.Diagnostico) == "" {
		return nil, fmt.Errorf("%w: campos obrigatórios ausentes", ErrAgenteRespostaInvalida)
	}
	if resposta.Evidencias == nil {
		resposta.Evidencias = []string{}
	}
	if resposta.AcoesRecomendadas == nil {
		resposta.AcoesRecomendadas = []string{}
	}
	if resposta.Limitacoes == nil {
		resposta.Limitacoes = []string{}
	}
	return &resposta, nil
}

// stripCodeFences remove cercas de código markdown que o modelo possa incluir.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	// Remove a primeira linha (```json) e a cerca de fechamento.
	if idx := strings.Index(s, "\n"); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// estimarTokens faz uma estimativa simples de tokens (~4 caracteres por token).
func estimarTokens(s string) int {
	n := len([]rune(s))
	if n == 0 {
		return 0
	}
	return n/4 + 1
}

func (uc *AgenteUsecase) validarLoja(ctx context.Context, empresaID, lojaID uuid.UUID) error {
	loja, err := uc.lojaRepo.GetByID(ctx, lojaID)
	if err != nil || loja == nil || loja.EmpresaID != empresaID || !loja.Ativo {
		return ErrAgenteLojaNaoPertence
	}
	return nil
}

func (uc *AgenteUsecase) registrarAuditoria(ctx context.Context, empresaID, usuarioID uuid.UUID, acao string, entidadeID uuid.UUID, payload any) {
	_ = uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "agente_ia",
		EntidadeID: entidadeID,
		Acao:       acao,
		Depois:     stringPtr(toJSON(payload)),
	})
}
