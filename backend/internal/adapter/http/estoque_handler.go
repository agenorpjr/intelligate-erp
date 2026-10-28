package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Estoque Handler (Etapa 12)
// ============================================================

type EstoqueHandler struct {
	usecase *usecase.EstoqueUsecase
}

func NewEstoqueHandler(estoqueUsecase *usecase.EstoqueUsecase) *EstoqueHandler {
	return &EstoqueHandler{usecase: estoqueUsecase}
}

type registrarSaidaRequest struct {
	LojaID            uuid.UUID `json:"loja_id"`
	InsumoID          uuid.UUID `json:"insumo_id"`
	Tipo              string    `json:"tipo"`
	Quantidade        float64   `json:"quantidade"`
	Motivo            *string   `json:"motivo,omitempty"`
	Observacao        *string   `json:"observacao,omitempty"`
	ReferenciaExterna *string   `json:"referencia_externa,omitempty"`
}

type registrarAjusteRequest struct {
	LojaID     uuid.UUID `json:"loja_id"`
	InsumoID   uuid.UUID `json:"insumo_id"`
	Tipo       string    `json:"tipo"`
	Quantidade float64   `json:"quantidade"`
	Motivo     string    `json:"motivo"`
	Observacao *string   `json:"observacao,omitempty"`
}

type transferirEstoqueRequest struct {
	LojaOrigemID  uuid.UUID `json:"loja_origem_id"`
	LojaDestinoID uuid.UUID `json:"loja_destino_id"`
	InsumoID      uuid.UUID `json:"insumo_id"`
	Quantidade    float64   `json:"quantidade"`
	Motivo        *string   `json:"motivo,omitempty"`
	Observacao    *string   `json:"observacao,omitempty"`
}

// ListSaldos trata GET /api/v1/estoque/saldos?loja_id=...&insumo_id=...
func (h *EstoqueHandler) ListSaldos(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	filter := domain.SaldoEstoqueFilter{
		EmpresaID: empresaID,
		Limit:     parseLimit(r, 20),
		Offset:    parseOffset(r),
	}

	if lojaStr := r.URL.Query().Get("loja_id"); lojaStr != "" {
		lojaID, err := uuid.Parse(lojaStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_ID", "loja_id inválido", "")
			return
		}
		filter.LojaID = &lojaID
	}
	if insumoStr := r.URL.Query().Get("insumo_id"); insumoStr != "" {
		insumoID, err := uuid.Parse(insumoStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_ID", "insumo_id inválido", "")
			return
		}
		filter.InsumoID = &insumoID
	}

	saldos, err := h.usecase.ListSaldos(r.Context(), filter)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, saldos)
}

// GetSaldo trata GET /api/v1/estoque/saldos/{insumo_id}?loja_id=...
func (h *EstoqueHandler) GetSaldo(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	insumoID, err := uuid.Parse(chi.URLParam(r, "insumo_id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "insumo_id inválido", "")
		return
	}

	lojaStr := r.URL.Query().Get("loja_id")
	if lojaStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "loja_id é obrigatório", "")
		return
	}
	lojaID, err := uuid.Parse(lojaStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "loja_id inválido", "")
		return
	}

	saldo, err := h.usecase.GetSaldo(r.Context(), empresaID, insumoID, lojaID)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, saldo)
}

// RegistrarSaida trata POST /api/v1/estoque/saidas
func (h *EstoqueHandler) RegistrarSaida(w http.ResponseWriter, r *http.Request) {
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

	var req registrarSaidaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	mov, err := h.usecase.RegistrarSaida(r.Context(), domain.RegistrarSaidaEstoqueInput{
		EmpresaID:         empresaID,
		LojaID:            req.LojaID,
		InsumoID:          req.InsumoID,
		Tipo:              req.Tipo,
		Quantidade:        req.Quantidade,
		Motivo:            req.Motivo,
		Observacao:        req.Observacao,
		ReferenciaExterna: req.ReferenciaExterna,
	}, usuario.ID)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, mov)
}

// RegistrarAjuste trata POST /api/v1/estoque/ajustes
func (h *EstoqueHandler) RegistrarAjuste(w http.ResponseWriter, r *http.Request) {
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

	var req registrarAjusteRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	mov, err := h.usecase.RegistrarAjuste(r.Context(), domain.RegistrarAjusteEstoqueInput{
		EmpresaID:  empresaID,
		LojaID:     req.LojaID,
		InsumoID:   req.InsumoID,
		Tipo:       req.Tipo,
		Quantidade: req.Quantidade,
		Motivo:     req.Motivo,
		Observacao: req.Observacao,
	}, usuario.ID)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, mov)
}

// Transferir trata POST /api/v1/estoque/transferencias
func (h *EstoqueHandler) Transferir(w http.ResponseWriter, r *http.Request) {
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

	var req transferirEstoqueRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	movimentacoes, err := h.usecase.Transferir(r.Context(), domain.TransferirEstoqueInput{
		EmpresaID:     empresaID,
		LojaOrigemID:  req.LojaOrigemID,
		LojaDestinoID: req.LojaDestinoID,
		InsumoID:      req.InsumoID,
		Quantidade:    req.Quantidade,
		Motivo:        req.Motivo,
		Observacao:    req.Observacao,
	}, usuario.ID)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, movimentacoes)
}

// ListMovimentacoes trata GET /api/v1/estoque/movimentacoes
func (h *EstoqueHandler) ListMovimentacoes(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	q := r.URL.Query()
	filter := domain.MovimentacaoEstoqueFilter{
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
	if insumoStr := q.Get("insumo_id"); insumoStr != "" {
		insumoID, err := uuid.Parse(insumoStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_ID", "insumo_id inválido", "")
			return
		}
		filter.InsumoID = &insumoID
	}
	if tipoStr := q.Get("tipo"); tipoStr != "" {
		tipo := domain.TipoMovimentacaoEstoque(tipoStr)
		filter.Tipo = &tipo
	}
	if origemStr := q.Get("origem"); origemStr != "" {
		filter.Origem = &origemStr
	}
	if dataInicioStr := q.Get("data_inicio"); dataInicioStr != "" {
		dataInicio, err := parseDashboardDate(dataInicioStr, false)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_inicio inválida", "")
			return
		}
		filter.DataInicio = &dataInicio
	}
	if dataFimStr := q.Get("data_fim"); dataFimStr != "" {
		dataFim, err := parseDashboardDate(dataFimStr, true)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_fim inválida", "")
			return
		}
		filter.DataFim = &dataFim
	}

	movimentacoes, err := h.usecase.ListMovimentacoes(r.Context(), filter)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, movimentacoes)
}

// GetResumoPerdas trata GET /api/v1/estoque/perdas?loja_id=...&data_inicio=...&data_fim=...
func (h *EstoqueHandler) GetResumoPerdas(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	q := r.URL.Query()
	lojaStr := q.Get("loja_id")
	if lojaStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "loja_id é obrigatório", "")
		return
	}
	lojaID, err := uuid.Parse(lojaStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "loja_id inválido", "")
		return
	}

	dataInicioStr := q.Get("data_inicio")
	dataFimStr := q.Get("data_fim")
	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}
	dataInicio, err := parseDashboardDate(dataInicioStr, false)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_inicio inválida", "")
		return
	}
	dataFim, err := parseDashboardDate(dataFimStr, true)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_fim inválida", "")
		return
	}

	resumo, err := h.usecase.GetResumoPerdas(r.Context(), empresaID, lojaID, dataInicio, dataFim)
	if err != nil {
		responderErroEstoque(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, resumo)
}

// responderErroEstoque mapeia os erros do estoque para respostas HTTP.
func responderErroEstoque(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrEstoqueSaldoInsuficiente):
		RespondError(w, http.StatusConflict, "SALDO_INSUFICIENTE", "Saldo insuficiente", err.Error())
	case errors.Is(err, usecase.ErrEstoqueTipoInvalido),
		errors.Is(err, usecase.ErrEstoqueMotivoObrigatorio),
		errors.Is(err, usecase.ErrEstoqueQuantidadeInvalida),
		errors.Is(err, usecase.ErrEstoqueMesmaLoja),
		errors.Is(err, usecase.ErrEstoquePeriodoInvalido),
		errors.Is(err, usecase.ErrEstoqueInsumoObrigatorio),
		errors.Is(err, usecase.ErrEstoqueLojaObrigatoria):
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Parâmetros inválidos", err.Error())
	case errors.Is(err, usecase.ErrEstoqueLojaNaoPertence):
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", err.Error())
	case errors.Is(err, usecase.ErrEstoqueInsumoNaoEncontrado):
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", err.Error())
	case errors.Is(err, usecase.ErrEstoqueSaldoNaoEncontrado):
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Saldo não encontrado", err.Error())
	default:
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
	}
}

func parseLimit(r *http.Request, def int) int {
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			return l
		}
	}
	return def
}

func parseOffset(r *http.Request) int {
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			return o
		}
	}
	return 0
}
