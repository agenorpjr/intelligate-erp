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

type FichaTecnicaHandler struct {
	fichaUsecase *usecase.FichaTecnicaUsecase
}

func NewFichaTecnicaHandler(fichaUsecase *usecase.FichaTecnicaUsecase) *FichaTecnicaHandler {
	return &FichaTecnicaHandler{fichaUsecase: fichaUsecase}
}

type CreateFichaTecnicaRequest struct {
	ProdutoID         uuid.UUID                             `json:"produto_id,omitempty"` // Optional, comes from path parameter
	LojaID            uuid.UUID                             `json:"loja_id" validate:"required"`
	Rendimento        float64                               `json:"rendimento"`
	UnidadeRendimento string                                `json:"unidade_rendimento"`
	Componentes       []CreateFichaTecnicaComponenteRequest `json:"componentes" validate:"required,min=1"`
}

type CreateFichaTecnicaComponenteRequest struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string    `json:"unidade_medida" validate:"required"`
}

type UpdateFichaTecnicaRequest struct {
	Rendimento        *float64                              `json:"rendimento,omitempty"`
	UnidadeRendimento *string                               `json:"unidade_rendimento,omitempty"`
	Ativo             *bool                                 `json:"ativo,omitempty"`
	Componentes       []UpdateFichaTecnicaComponenteRequest `json:"componentes,omitempty"`
}

type UpdateFichaTecnicaComponenteRequest struct {
	ID            *uuid.UUID `json:"id,omitempty"`
	InsumoID      uuid.UUID  `json:"insumo_id" validate:"required"`
	Quantidade    float64    `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string     `json:"unidade_medida" validate:"required"`
}

type FichaTecnicaResponse struct {
	ID                uuid.UUID                        `json:"id"`
	ProdutoID         uuid.UUID                        `json:"produto_id"`
	LojaID            uuid.UUID                        `json:"loja_id"`
	Rendimento        float64                          `json:"rendimento"`
	UnidadeRendimento string                           `json:"unidade_rendimento"`
	Ativo             bool                             `json:"ativo"`
	CustoTeorico      float64                          `json:"custo_teorico,omitempty"`
	CreatedAt         time.Time                        `json:"created_at"`
	UpdatedAt         time.Time                        `json:"updated_at"`
	DeletedAt         *time.Time                       `json:"deleted_at,omitempty"`
	Componentes       []FichaTecnicaComponenteResponse `json:"componentes,omitempty"`
}

type FichaTecnicaComponenteResponse struct {
	ID             uuid.UUID `json:"id"`
	FichaTecnicaID uuid.UUID `json:"ficha_tecnica_id"`
	InsumoID       uuid.UUID `json:"insumo_id"`
	Quantidade     float64   `json:"quantidade"`
	UnidadeMedida  string    `json:"unidade_medida"`
	InsumoNome     string    `json:"insumo_nome,omitempty"`
	InsumoUnidade  string    `json:"insumo_unidade,omitempty"`
	CustoUnitario  float64   `json:"custo_unitario,omitempty"`
	CustoTotal     float64   `json:"custo_total,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (h *FichaTecnicaHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extract produto_id from URL path parameter {id}
	produtoIDStr := chi.URLParam(r, "id")
	produtoID, err := uuid.Parse(produtoIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID do produto inválido", "")
		return
	}

	var req CreateFichaTecnicaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	// Validate that body produto_id matches path parameter (if provided)
	if req.ProdutoID != uuid.Nil && req.ProdutoID != produtoID {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "produto_id no body não corresponde ao path parameter", "")
		return
	}

	input := usecase.CreateFichaTecnicaInput{
		ProdutoID:         produtoID,
		LojaID:            req.LojaID,
		Rendimento:        req.Rendimento,
		UnidadeRendimento: req.UnidadeRendimento,
		Componentes:       make([]usecase.CreateFichaTecnicaComponenteInput, len(req.Componentes)),
	}

	for i, comp := range req.Componentes {
		input.Componentes[i] = usecase.CreateFichaTecnicaComponenteInput{
			InsumoID:      comp.InsumoID,
			Quantidade:    comp.Quantidade,
			UnidadeMedida: comp.UnidadeMedida,
		}
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	ficha, err := h.fichaUsecase.Create(r.Context(), input, usuario.ID, empresaID)
	if err != nil {
		if errors.Is(err, usecase.ErrFichaTecnicaJaExiste) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Ficha técnica já existe para este produto e loja", "")
			return
		}
		if errors.Is(err, usecase.ErrProdutoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrInsumoNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Insumo não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toFichaTecnicaResponse(ficha)
	RespondJSON(w, http.StatusCreated, resp)
}

func (h *FichaTecnicaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	ficha, err := h.fichaUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrFichaTecnicaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ficha técnica não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toFichaTecnicaResponse(ficha)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *FichaTecnicaHandler) ListByProduto(w http.ResponseWriter, r *http.Request) {
	produtoIDStr := chi.URLParam(r, "id")
	produtoID, err := uuid.Parse(produtoIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID do produto inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
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

	fichas, err := h.fichaUsecase.ListByProduto(r.Context(), produtoID, ativo, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]FichaTecnicaResponse, len(fichas))
	for i, ficha := range fichas {
		resp[i] = toFichaTecnicaResponse(ficha)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *FichaTecnicaHandler) ListByLoja(w http.ResponseWriter, r *http.Request) {
	lojaIDStr := chi.URLParam(r, "id")
	lojaID, err := uuid.Parse(lojaIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID da loja inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
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

	fichas, err := h.fichaUsecase.ListByLoja(r.Context(), lojaID, ativo, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]FichaTecnicaResponse, len(fichas))
	for i, ficha := range fichas {
		resp[i] = toFichaTecnicaResponse(ficha)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *FichaTecnicaHandler) ListByEmpresa(w http.ResponseWriter, r *http.Request) {
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

	fichas, err := h.fichaUsecase.ListByEmpresa(r.Context(), empresaID, ativo, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]FichaTecnicaResponse, len(fichas))
	for i, ficha := range fichas {
		resp[i] = toFichaTecnicaResponse(ficha)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *FichaTecnicaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateFichaTecnicaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := usecase.UpdateFichaTecnicaInput{
		Rendimento:        req.Rendimento,
		UnidadeRendimento: req.UnidadeRendimento,
		Ativo:             req.Ativo,
		Componentes:       make([]usecase.UpdateFichaTecnicaComponenteInput, len(req.Componentes)),
	}

	for i, comp := range req.Componentes {
		input.Componentes[i] = usecase.UpdateFichaTecnicaComponenteInput{
			ID:            comp.ID,
			InsumoID:      comp.InsumoID,
			Quantidade:    comp.Quantidade,
			UnidadeMedida: comp.UnidadeMedida,
		}
	}

	ficha, err := h.fichaUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrFichaTecnicaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ficha técnica não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toFichaTecnicaResponse(ficha)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *FichaTecnicaHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err = h.fichaUsecase.Delete(r.Context(), id, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrFichaTecnicaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ficha técnica não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func toFichaTecnicaResponse(f *domain.FichaTecnica) FichaTecnicaResponse {
	componentes := make([]FichaTecnicaComponenteResponse, len(f.Componentes))
	for i, comp := range f.Componentes {
		componentes[i] = FichaTecnicaComponenteResponse{
			ID:             comp.ID,
			FichaTecnicaID: comp.FichaTecnicaID,
			InsumoID:       comp.InsumoID,
			Quantidade:     comp.Quantidade,
			UnidadeMedida:  comp.UnidadeMedida,
			InsumoNome:     comp.InsumoNome,
			InsumoUnidade:  comp.InsumoUnidade,
			CustoUnitario:  comp.CustoUnitario,
			CustoTotal:     comp.CustoTotal,
			CreatedAt:      comp.CreatedAt,
			UpdatedAt:      comp.UpdatedAt,
		}
	}

	return FichaTecnicaResponse{
		ID:                f.ID,
		ProdutoID:         f.ProdutoID,
		LojaID:            f.LojaID,
		Rendimento:        f.Rendimento,
		UnidadeRendimento: f.UnidadeRendimento,
		Ativo:             f.Ativo,
		CustoTeorico:      f.CustoTeorico,
		CreatedAt:         f.CreatedAt,
		UpdatedAt:         f.UpdatedAt,
		DeletedAt:         f.DeletedAt,
		Componentes:       componentes,
	}
}
