package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type PdvIntegracaoHandler struct {
	pdvIntegracaoUsecase *usecase.PdvIntegracaoUsecase
}

func NewPdvIntegracaoHandler(pdvIntegracaoUsecase *usecase.PdvIntegracaoUsecase) *PdvIntegracaoHandler {
	return &PdvIntegracaoHandler{pdvIntegracaoUsecase: pdvIntegracaoUsecase}
}

type CreatePdvIntegracaoRequest struct {
	EmpresaID    uuid.UUID `json:"empresa_id" validate:"required"`
	LojaID       uuid.UUID `json:"loja_id" validate:"required"`
	Provedor     string    `json:"provedor" validate:"required,oneof=gula"`
	Ambiente     string    `json:"ambiente" validate:"required,oneof=sandbox production"`
	URLBase      string    `json:"url_base" validate:"required,url"`
	APIKey       string    `json:"api_key" validate:"required"`
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
}

type UpdatePdvIntegracaoRequest struct {
	Provedor     *string `json:"provedor,omitempty" validate:"omitempty,oneof=gula"`
	Ambiente     *string `json:"ambiente,omitempty" validate:"omitempty,oneof=sandbox production"`
	URLBase      *string `json:"url_base,omitempty" validate:"omitempty,url"`
	APIKey       *string `json:"api_key,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`
	Ativo        *bool   `json:"ativo,omitempty"`
}

type SincronizarRequest struct {
	DataInicio string `json:"data_inicio" validate:"required"`
	DataFim    string `json:"data_fim" validate:"required"`
}

func (h *PdvIntegracaoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var req CreatePdvIntegracaoRequest
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

	input := domain.CreatePdvIntegracaoInput{
		EmpresaID:    req.EmpresaID,
		LojaID:       req.LojaID,
		Provedor:     req.Provedor,
		Ambiente:     req.Ambiente,
		URLBase:      req.URLBase,
		APIKey:       req.APIKey,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
	}

	integracao, err := h.pdvIntegracaoUsecase.Criar(r.Context(), input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvIntegracaoJaExiste) {
			RespondError(w, http.StatusConflict, "DUPLICATE_INTEGRATION", "Integração PDV já existe para esta loja e provedor", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "LOJA_NOT_FOUND", "Loja não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, integracao)
}

func (h *PdvIntegracaoHandler) Listar(w http.ResponseWriter, r *http.Request) {
	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0

	var provedor *string
	if p := r.URL.Query().Get("provedor"); p != "" {
		provedor = &p
	}

	var ativo *bool
	if a := r.URL.Query().Get("ativo"); a != "" {
		val := a == "true"
		ativo = &val
	}

	filter := domain.PdvIntegracaoFilter{
		Provedor: provedor,
		Ativo:    ativo,
		Limit:    limit,
		Offset:   offset,
	}

	integracoes, err := h.pdvIntegracaoUsecase.Listar(r.Context(), empresaID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, integracoes)
}

func (h *PdvIntegracaoHandler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	integracao, err := h.pdvIntegracaoUsecase.BuscarPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvIntegracaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Integração PDV não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, integracao)
}

func (h *PdvIntegracaoHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var req UpdatePdvIntegracaoRequest
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

	input := domain.UpdatePdvIntegracaoInput{
		Provedor:     req.Provedor,
		Ambiente:     req.Ambiente,
		URLBase:      req.URLBase,
		APIKey:       req.APIKey,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		Ativo:        req.Ativo,
	}

	integracao, err := h.pdvIntegracaoUsecase.Atualizar(r.Context(), id, input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvIntegracaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Integração PDV não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, integracao)
}

func (h *PdvIntegracaoHandler) Deletar(w http.ResponseWriter, r *http.Request) {
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

	err = h.pdvIntegracaoUsecase.Deletar(r.Context(), id, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvIntegracaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Integração PDV não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *PdvIntegracaoHandler) Sincronizar(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var req SincronizarRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	dataInicio, err := time.Parse(time.RFC3339, req.DataInicio)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_inicio inválida, use formato RFC3339", "")
		return
	}

	dataFim, err := time.Parse(time.RFC3339, req.DataFim)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_fim inválida, use formato RFC3339", "")
		return
	}

	if dataFim.Before(dataInicio) {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE_RANGE", "data_fim deve ser posterior a data_inicio", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	result, err := h.pdvIntegracaoUsecase.SincronizarVendas(r.Context(), id, dataInicio, dataFim, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvIntegracaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Integração PDV não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, result)
}
