package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type LojaHandler struct {
	lojaUsecase *usecase.LojaUsecase
}

func NewLojaHandler(lojaUsecase *usecase.LojaUsecase) *LojaHandler {
	return &LojaHandler{lojaUsecase: lojaUsecase}
}

type CreateLojaRequest struct {
	EmpresaID   uuid.UUID `json:"empresa_id" validate:"required"`
	Codigo      string    `json:"codigo" validate:"required"`
	Nome        string    `json:"nome" validate:"required"`
	Endereco    *string   `json:"endereco,omitempty"`
	FusoHorario string    `json:"fuso_horario" validate:"required"`
}

type UpdateLojaRequest struct {
	Codigo      *string `json:"codigo,omitempty"`
	Nome        *string `json:"nome,omitempty"`
	Endereco    *string `json:"endereco,omitempty"`
	FusoHorario *string `json:"fuso_horario,omitempty"`
	Ativo       *bool   `json:"ativo,omitempty"`
}

func (h *LojaHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var req CreateLojaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	input := domain.CreateLojaInput{
		EmpresaID:   req.EmpresaID,
		Codigo:      req.Codigo,
		Nome:        req.Nome,
		Endereco:    req.Endereco,
		FusoHorario: req.FusoHorario,
	}

	loja, err := h.lojaUsecase.Criar(r.Context(), input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrCodigoDuplicado) {
			RespondError(w, http.StatusConflict, "DUPLICATE_CODE", "Código de loja já existe para esta empresa", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, loja)
}

func (h *LojaHandler) Listar(w http.ResponseWriter, r *http.Request) {
	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0

	lojas, err := h.lojaUsecase.Listar(r.Context(), empresaID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, lojas)
}

func (h *LojaHandler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	loja, err := h.lojaUsecase.BuscarPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, loja)
}

func (h *LojaHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var req UpdateLojaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	input := domain.UpdateLojaInput{
		Codigo:      req.Codigo,
		Nome:        req.Nome,
		Endereco:    req.Endereco,
		FusoHorario: req.FusoHorario,
		Ativo:       req.Ativo,
	}

	loja, err := h.lojaUsecase.Atualizar(r.Context(), id, input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrCodigoDuplicado) {
			RespondError(w, http.StatusConflict, "DUPLICATE_CODE", "Código de loja já existe para esta empresa", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, loja)
}

func (h *LojaHandler) Deletar(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	err = h.lojaUsecase.Deletar(r.Context(), id, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}
