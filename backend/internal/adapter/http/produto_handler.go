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

type ProdutoHandler struct {
	produtoUsecase *usecase.ProdutoUsecase
}

func NewProdutoHandler(produtoUsecase *usecase.ProdutoUsecase) *ProdutoHandler {
	return &ProdutoHandler{produtoUsecase: produtoUsecase}
}

type CreateProdutoRequest struct {
	Nome       string  `json:"nome" validate:"required"`
	SKUExterno *string `json:"sku_externo,omitempty"`
	Categoria  *string `json:"categoria,omitempty"`
}

type UpdateProdutoRequest struct {
	Nome       *string `json:"nome,omitempty"`
	SKUExterno *string `json:"sku_externo,omitempty"`
	Categoria  *string `json:"categoria,omitempty"`
	Ativo      *bool   `json:"ativo,omitempty"`
}

type ProdutoResponse struct {
	ID         uuid.UUID  `json:"id"`
	EmpresaID  uuid.UUID  `json:"empresa_id"`
	Nome       string     `json:"nome"`
	SKUExterno *string    `json:"sku_externo,omitempty"`
	Categoria  *string    `json:"categoria,omitempty"`
	Ativo      bool       `json:"ativo"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (h *ProdutoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProdutoRequest
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

	input := usecase.CreateProdutoInput{
		EmpresaID:  empresaID,
		Nome:       req.Nome,
		SKUExterno: req.SKUExterno,
		Categoria:  req.Categoria,
	}

	produto, err := h.produtoUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrProdutoJaExiste) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Produto com este SKU já existe", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toProdutoResponse(produto)
	RespondJSON(w, http.StatusCreated, resp)
}

func (h *ProdutoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	produto, err := h.produtoUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrProdutoCMVNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toProdutoResponse(produto)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *ProdutoHandler) List(w http.ResponseWriter, r *http.Request) {
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

	produtos, err := h.produtoUsecase.ListByEmpresa(r.Context(), empresaID, ativo, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]ProdutoResponse, len(produtos))
	for i, produto := range produtos {
		resp[i] = toProdutoResponse(produto)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *ProdutoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateProdutoRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := usecase.UpdateProdutoInput{
		Nome:       req.Nome,
		SKUExterno: req.SKUExterno,
		Categoria:  req.Categoria,
		Ativo:      req.Ativo,
	}

	produto, err := h.produtoUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrProdutoCMVNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := toProdutoResponse(produto)
	RespondJSON(w, http.StatusOK, resp)
}

func (h *ProdutoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err = h.produtoUsecase.Delete(r.Context(), id, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrProdutoCMVNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func toProdutoResponse(p *domain.Produto) ProdutoResponse {
	return ProdutoResponse{
		ID:         p.ID,
		EmpresaID:  p.EmpresaID,
		Nome:       p.Nome,
		SKUExterno: p.SKUExterno,
		Categoria:  p.Categoria,
		Ativo:      p.Ativo,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		DeletedAt:  p.DeletedAt,
	}
}
