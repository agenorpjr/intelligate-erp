package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Agente de IA Handler (Etapa 13)
// ============================================================

type AgenteHandler struct {
	usecase *usecase.AgenteUsecase
}

func NewAgenteHandler(agenteUsecase *usecase.AgenteUsecase) *AgenteHandler {
	return &AgenteHandler{usecase: agenteUsecase}
}

type criarSessaoAgenteRequest struct {
	LojaID uuid.UUID `json:"loja_id"`
	Titulo string    `json:"titulo"`
}

type perguntarAgenteRequest struct {
	Pergunta   string `json:"pergunta"`
	DataInicio string `json:"data_inicio,omitempty"`
	DataFim    string `json:"data_fim,omitempty"`
}

// CriarSessao trata POST /api/v1/agente/sessoes
func (h *AgenteHandler) CriarSessao(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	var req criarSessaoAgenteRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	sessao, err := h.usecase.CriarSessao(r.Context(), domain.CriarAgenteSessaoInput{
		EmpresaID: empresaID,
		LojaID:    req.LojaID,
		UsuarioID: usuario.ID,
		Titulo:    req.Titulo,
	})
	if err != nil {
		responderErroAgente(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, sessao)
}

// ListarSessoes trata GET /api/v1/agente/sessoes?loja_id=...&status=...
func (h *AgenteHandler) ListarSessoes(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	q := r.URL.Query()
	filter := domain.AgenteSessaoFilter{
		EmpresaID: empresaID,
		Limit:     parseLimit(r, 20),
		Offset:    parseOffset(r),
	}

	if lojaStr := q.Get("loja_id"); lojaStr != "" {
		lojaID, err := uuid.Parse(lojaStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_ID", "loja_id inválido", "")
			return
		}
		filter.LojaID = &lojaID
	}
	if statusStr := q.Get("status"); statusStr != "" {
		filter.Status = &statusStr
	}

	sessoes, err := h.usecase.ListarSessoes(r.Context(), filter)
	if err != nil {
		responderErroAgente(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, sessoes)
}

// ObterSessao trata GET /api/v1/agente/sessoes/{id}
func (h *AgenteHandler) ObterSessao(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	sessaoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "id inválido", "")
		return
	}

	detalhe, err := h.usecase.ObterSessao(r.Context(), empresaID, sessaoID)
	if err != nil {
		responderErroAgente(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, detalhe)
}

// Perguntar trata POST /api/v1/agente/sessoes/{id}/perguntas
func (h *AgenteHandler) Perguntar(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	sessaoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "id inválido", "")
		return
	}

	var req perguntarAgenteRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	input := domain.PerguntarAgenteInput{
		EmpresaID: empresaID,
		UsuarioID: usuario.ID,
		SessaoID:  sessaoID,
		Pergunta:  req.Pergunta,
	}

	if req.DataInicio != "" {
		dataInicio, err := parseDashboardDate(req.DataInicio, false)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_inicio inválida (usar RFC3339 ou YYYY-MM-DD)", "")
			return
		}
		input.DataInicio = dataInicio
	}
	if req.DataFim != "" {
		dataFim, err := parseDashboardDate(req.DataFim, true)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_fim inválida (usar RFC3339 ou YYYY-MM-DD)", "")
			return
		}
		input.DataFim = dataFim
	}

	resultado, err := h.usecase.Perguntar(r.Context(), input)
	if err != nil {
		responderErroAgente(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, resultado)
}

// responderErroAgente mapeia os erros do agente para respostas HTTP.
func responderErroAgente(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrAgenteSessaoObrigatoria),
		errors.Is(err, usecase.ErrAgentePerguntaObrigatoria),
		errors.Is(err, usecase.ErrAgenteLojaObrigatoria),
		errors.Is(err, usecase.ErrAgentePeriodoInvalido):
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Parâmetros inválidos", err.Error())
	case errors.Is(err, usecase.ErrAgenteSessaoNaoEncontrada):
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Sessão não encontrada", err.Error())
	case errors.Is(err, usecase.ErrAgenteLojaNaoPertence):
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", err.Error())
	case errors.Is(err, usecase.ErrAgenteSessaoArquivada):
		RespondError(w, http.StatusConflict, "SESSAO_ARQUIVADA", "Sessão arquivada", err.Error())
	case errors.Is(err, usecase.ErrAgenteRespostaInvalida):
		RespondError(w, http.StatusBadGateway, "AGENTE_RESPOSTA_INVALIDA", "Resposta inválida do provedor de IA", err.Error())
	default:
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
	}
}
