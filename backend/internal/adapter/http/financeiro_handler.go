package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Conta Gerencial Handler
// ============================================================

type ContaGerencialHandler struct {
	usecase *usecase.ContaGerencialUsecase
}

func NewContaGerencialHandler(usecase *usecase.ContaGerencialUsecase) *ContaGerencialHandler {
	return &ContaGerencialHandler{usecase: usecase}
}

func (h *ContaGerencialHandler) Create(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	var input domain.CreateContaGerencialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}
	input.EmpresaID = empresaID

	account, err := h.usecase.Create(r.Context(), empresaID, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao criar conta gerencial", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, account)
}

func (h *ContaGerencialHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	account, err := h.usecase.GetByID(r.Context(), empresaID, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Conta gerencial não encontrada", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, account)
}

func (h *ContaGerencialHandler) List(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	limit, offset := parsePagination(r)
	filter := domain.ContaGerencialFilter{
		Limit:  limit,
		Offset: offset,
	}

	if tipo := r.URL.Query().Get("tipo"); tipo != "" {
		t := domain.ContaGerencialTipo(tipo)
		filter.Tipo = &t
	}
	if ativo := r.URL.Query().Get("ativo"); ativo != "" {
		a := ativo == "true"
		filter.Ativo = &a
	}

	accounts, err := h.usecase.List(r.Context(), empresaID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, accounts)
}

func (h *ContaGerencialHandler) ListWithHierarchy(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	accounts, err := h.usecase.ListWithHierarchy(r.Context(), empresaID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, accounts)
}

func (h *ContaGerencialHandler) Update(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var input domain.UpdateContaGerencialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}

	account, err := h.usecase.Update(r.Context(), empresaID, id, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao atualizar conta gerencial", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, account)
}

func (h *ContaGerencialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	if err := h.usecase.Delete(r.Context(), empresaID, id); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao excluir conta gerencial", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// Categoria Financeira Handler
// ============================================================

type CategoriaFinanceiraHandler struct {
	usecase *usecase.CategoriaFinanceiraUsecase
}

func NewCategoriaFinanceiraHandler(usecase *usecase.CategoriaFinanceiraUsecase) *CategoriaFinanceiraHandler {
	return &CategoriaFinanceiraHandler{usecase: usecase}
}

func (h *CategoriaFinanceiraHandler) Create(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	var input domain.CreateCategoriaFinanceiraInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}
	input.EmpresaID = empresaID

	category, err := h.usecase.Create(r.Context(), empresaID, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao criar categoria financeira", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, category)
}

func (h *CategoriaFinanceiraHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	category, err := h.usecase.GetByID(r.Context(), empresaID, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Categoria financeira não encontrada", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, category)
}

func (h *CategoriaFinanceiraHandler) List(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	limit, offset := parsePagination(r)
	filter := domain.CategoriaFinanceiraFilter{
		Limit:  limit,
		Offset: offset,
	}

	if contaGerStr := r.URL.Query().Get("conta_gerencial_id"); contaGerStr != "" {
		if id, err := uuid.Parse(contaGerStr); err == nil {
			filter.ContaGerencialID = &id
		}
	}
	if ativo := r.URL.Query().Get("ativo"); ativo != "" {
		a := ativo == "true"
		filter.Ativo = &a
	}

	categories, err := h.usecase.List(r.Context(), empresaID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, categories)
}

func (h *CategoriaFinanceiraHandler) Update(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var input domain.UpdateCategoriaFinanceiraInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}

	category, err := h.usecase.Update(r.Context(), empresaID, id, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao atualizar categoria financeira", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, category)
}

func (h *CategoriaFinanceiraHandler) Delete(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	if err := h.usecase.Delete(r.Context(), empresaID, id); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao excluir categoria financeira", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// Movimentação Caixa Handler
// ============================================================

type MovimentacaoCaixaHandler struct {
	usecase *usecase.MovimentacaoCaixaUsecase
}

func NewMovimentacaoCaixaHandler(usecase *usecase.MovimentacaoCaixaUsecase) *MovimentacaoCaixaHandler {
	return &MovimentacaoCaixaHandler{usecase: usecase}
}

func (h *MovimentacaoCaixaHandler) Create(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	var input domain.CreateMovimentacaoCaixaInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}
	input.EmpresaID = empresaID

	movement, err := h.usecase.Create(r.Context(), empresaID, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao criar movimentação de caixa", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, movement)
}

func (h *MovimentacaoCaixaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	movement, err := h.usecase.GetByID(r.Context(), empresaID, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Movimentação não encontrada", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, movement)
}

func (h *MovimentacaoCaixaHandler) List(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	limit, offset := parsePagination(r)
	filter := domain.MovimentacaoCaixaFilter{
		Limit:  limit,
		Offset: offset,
	}

	if lojaStr := r.URL.Query().Get("loja_id"); lojaStr != "" {
		if id, err := uuid.Parse(lojaStr); err == nil {
			filter.LojaID = &id
		}
	}
	if tipo := r.URL.Query().Get("tipo"); tipo != "" {
		t := domain.MovimentacaoCaixaTipo(tipo)
		filter.Tipo = &t
	}
	if catFinStr := r.URL.Query().Get("categoria_financeira_id"); catFinStr != "" {
		if id, err := uuid.Parse(catFinStr); err == nil {
			filter.CategoriaFinanceiraID = &id
		}
	}
	if contaGerStr := r.URL.Query().Get("conta_gerencial_id"); contaGerStr != "" {
		if id, err := uuid.Parse(contaGerStr); err == nil {
			filter.ContaGerencialID = &id
		}
	}
	if origem := r.URL.Query().Get("origem"); origem != "" {
		filter.Origem = &origem
	}
	if dataInicio := r.URL.Query().Get("data_inicio"); dataInicio != "" {
		if t, err := time.Parse(time.RFC3339, dataInicio); err == nil {
			filter.DataInicio = &t
		}
	}
	if dataFim := r.URL.Query().Get("data_fim"); dataFim != "" {
		if t, err := time.Parse(time.RFC3339, dataFim); err == nil {
			filter.DataFim = &t
		}
	}

	movements, err := h.usecase.List(r.Context(), empresaID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, movements)
}

func (h *MovimentacaoCaixaHandler) Update(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var input domain.UpdateMovimentacaoCaixaInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição inválido", err.Error())
		return
	}

	movement, err := h.usecase.Update(r.Context(), empresaID, id, input)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao atualizar movimentação de caixa", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, movement)
}

func (h *MovimentacaoCaixaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	if err := h.usecase.Delete(r.Context(), empresaID, id); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao excluir movimentação de caixa", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// DRE Gerencial Handler
// ============================================================

type DREGerencialHandler struct {
	usecase *usecase.DREGerencialUsecase
}

func NewDREGerencialHandler(usecase *usecase.DREGerencialUsecase) *DREGerencialHandler {
	return &DREGerencialHandler{usecase: usecase}
}

func (h *DREGerencialHandler) GetDRE(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	dataInicioStr := r.URL.Query().Get("data_inicio")
	dataFimStr := r.URL.Query().Get("data_fim")
	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}

	dataInicio, err := time.Parse(time.RFC3339, dataInicioStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Formato de data_inválido (usar RFC3339)", "")
		return
	}
	dataFim, err := time.Parse(time.RFC3339, dataFimStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Formato de data inválido (usar RFC3339)", "")
		return
	}

	var lojaID *uuid.UUID
	if lojaStr := r.URL.Query().Get("loja_id"); lojaStr != "" {
		if id, err := uuid.Parse(lojaStr); err == nil {
			lojaID = &id
		}
	}

	dre, err := h.usecase.GetDRE(r.Context(), usecase.DRERequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao gerar DRE", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, dre)
}

// ============================================================
// Fluxo de Caixa Handler
// ============================================================

type FluxoCaixaHandler struct {
	usecase *usecase.FluxoCaixaUsecase
}

func NewFluxoCaixaHandler(usecase *usecase.FluxoCaixaUsecase) *FluxoCaixaHandler {
	return &FluxoCaixaHandler{usecase: usecase}
}

func (h *FluxoCaixaHandler) GetFluxoCaixa(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	dataInicioStr := r.URL.Query().Get("data_inicio")
	dataFimStr := r.URL.Query().Get("data_fim")
	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}

	dataInicio, err := time.Parse(time.RFC3339, dataInicioStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Formato de data inválido (usar RFC3339)", "")
		return
	}
	dataFim, err := time.Parse(time.RFC3339, dataFimStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Formato de data inválido (usar RFC3339)", "")
		return
	}

	var lojaID *uuid.UUID
	if lojaStr := r.URL.Query().Get("loja_id"); lojaStr != "" {
		if id, err := uuid.Parse(lojaStr); err == nil {
			lojaID = &id
		}
	}

	fluxo, err := h.usecase.GetFluxoCaixa(r.Context(), usecase.FluxoCaixaRequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})
	if err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Erro ao gerar fluxo de caixa", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, fluxo)
}

// ============================================================
// Helper functions
// ============================================================

func getEmpresaID(r *http.Request) uuid.UUID {
	return GetEmpresaIDFromContext(r.Context())
}

func parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
