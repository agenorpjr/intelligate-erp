package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type InsumoHandler struct {
	insumoUsecase *usecase.InsumoUsecase
}

func NewInsumoHandler(insumoUsecase *usecase.InsumoUsecase) *InsumoHandler {
	return &InsumoHandler{insumoUsecase: insumoUsecase}
}

type CreateInsumoRequest struct {
	Nome               string  `json:"nome" validate:"required"`
	UnidadeMedida      string  `json:"unidade_medida" validate:"required"`
	CustoUnitarioAtual float64 `json:"custo_unitario_atual"`
	Fornecedor         *string `json:"fornecedor,omitempty"`
}

type UpdateInsumoRequest struct {
	Nome               *string  `json:"nome,omitempty"`
	UnidadeMedida      *string  `json:"unidade_medida,omitempty"`
	CustoUnitarioAtual *float64 `json:"custo_unitario_atual,omitempty"`
	Fornecedor         *string  `json:"fornecedor,omitempty"`
	Ativo              *bool    `json:"ativo,omitempty"`
}

type CreateInsumoCustoRequest struct {
	DataVigencia  string  `json:"data_vigencia" validate:"required"`
	CustoUnitario float64 `json:"custo_unitario" validate:"required"`
	Observacao    *string `json:"observacao,omitempty"`
}

type InsumoResponse struct {
	ID                 uuid.UUID  `json:"id"`
	EmpresaID          uuid.UUID  `json:"empresa_id"`
	Nome               string     `json:"nome"`
	UnidadeMedida      string     `json:"unidade_medida"`
	CustoUnitarioAtual float64    `json:"custo_unitario_atual"`
	Fornecedor         *string    `json:"fornecedor,omitempty"`
	Ativo              bool       `json:"ativo"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type InsumoCustoResponse struct {
	ID            uuid.UUID `json:"id"`
	InsumoID      uuid.UUID `json:"insumo_id"`
	DataVigencia  time.Time `json:"data_vigencia"`
	CustoUnitario float64   `json:"custo_unitario"`
	Observacao    *string   `json:"observacao,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *InsumoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateInsumoRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	input := usecase.CreateInsumoInput{
		EmpresaID:          empresaID,
		Nome:               req.Nome,
		UnidadeMedida:      req.UnidadeMedida,
		CustoUnitarioAtual: req.CustoUnitarioAtual,
		Fornecedor:         req.Fornecedor,
	}

	insumo, err := h.insumoUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoJaExiste) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Insumo já existe", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoResponse(insumo)
	RespondJSON(w, http.StatusCreated, resp)
}

func (h *InsumoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	insumo, err := h.insumoUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoResponse(insumo)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *InsumoHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0
	var ativo *bool

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
	if ativoStr := r.URL.Query().Get("ativo"); ativoStr != "" {
		if a, err := strconv.ParseBool(ativoStr); err == nil {
			ativo = &a
		}
	}

	insumos, err := h.insumoUsecase.ListByEmpresa(r.Context(), empresaID, ativo, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]InsumoResponse, len(insumos))
	for i, insumo := range insumos {
		resp[i] = toInsumoResponse(insumo)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *InsumoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateInsumoRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := usecase.UpdateInsumoInput{
		Nome:               req.Nome,
		UnidadeMedida:      req.UnidadeMedida,
		CustoUnitarioAtual: req.CustoUnitarioAtual,
		Fornecedor:         req.Fornecedor,
		Ativo:              req.Ativo,
	}

	insumo, err := h.insumoUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoResponse(insumo)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *InsumoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	err = h.insumoUsecase.Delete(r.Context(), id, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *InsumoHandler) CreateCusto(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	insumoID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req CreateInsumoCustoRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	dataVigencia, err := time.Parse(time.RFC3339, req.DataVigencia)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_vigencia inválida (use RFC3339)", "")
		return
	}

	input := usecase.CreateInsumoCustoInput{
		InsumoID:      insumoID,
		DataVigencia:  dataVigencia,
		CustoUnitario: req.CustoUnitario,
		Observacao:    req.Observacao,
	}

	custo, err := h.insumoUsecase.CreateCusto(r.Context(), input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoCustoResponse(custo)
	RespondJSON(w, http.StatusCreated, resp)
}

func (h *InsumoHandler) GetCustoVigente(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	insumoID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	dataVigenciaStr := r.URL.Query().Get("data_vigencia")
	if dataVigenciaStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_vigencia é obrigatória", "")
		return
	}

	dataVigencia, err := time.Parse(time.RFC3339, dataVigenciaStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_vigencia inválida (use RFC3339)", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	custo, err := h.insumoUsecase.GetCustoVigente(r.Context(), insumoID, dataVigencia)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoCustoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Custo não encontrado para a data", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoCustoResponse(custo)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *InsumoHandler) GetCustoAtual(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	insumoID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	custo, err := h.insumoUsecase.GetCustoAtual(r.Context(), insumoID)
	if err != nil {
		if errors.Is(err, usecase.ErrInsumoCustoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Custo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toInsumoCustoResponse(custo)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *InsumoHandler) ListCustos(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	insumoID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

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

	custos, err := h.insumoUsecase.ListCustosByInsumo(r.Context(), insumoID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]InsumoCustoResponse, len(custos))
	for i, custo := range custos {
		resp[i] = toInsumoCustoResponse(custo)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func toInsumoResponse(i *domain.Insumo) InsumoResponse {
	return InsumoResponse{
		ID:                 i.ID,
		EmpresaID:          i.EmpresaID,
		Nome:               i.Nome,
		UnidadeMedida:      i.UnidadeMedida,
		CustoUnitarioAtual: i.CustoUnitarioAtual,
		Fornecedor:         i.Fornecedor,
		Ativo:              i.Ativo,
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
		DeletedAt:          i.DeletedAt,
	}
}

func toInsumoCustoResponse(c *domain.InsumoCusto) InsumoCustoResponse {
	return InsumoCustoResponse{
		ID:            c.ID,
		InsumoID:      c.InsumoID,
		DataVigencia:  c.DataVigencia,
		CustoUnitario: c.CustoUnitario,
		Observacao:    c.Observacao,
		CreatedAt:     c.CreatedAt,
	}
}
