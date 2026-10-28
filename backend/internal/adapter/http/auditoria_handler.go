package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// AuditoriaHandler expõe a leitura dos logs de auditoria (governança de acesso).
type AuditoriaHandler struct {
	auditoriaUsecase *usecase.AuditoriaUsecase
}

func NewAuditoriaHandler(auditoriaUsecase *usecase.AuditoriaUsecase) *AuditoriaHandler {
	return &AuditoriaHandler{auditoriaUsecase: auditoriaUsecase}
}

// Listar retorna os logs de auditoria da empresa do usuário autenticado.
//
// Query params opcionais: entidade, acao, usuario_id, limit, offset.
func (h *AuditoriaHandler) Listar(w http.ResponseWriter, r *http.Request) {
	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	filter := domain.AuditoriaFilter{
		Limit:  parseQueryInt(r.URL.Query().Get("limit"), 20),
		Offset: parseQueryInt(r.URL.Query().Get("offset"), 0),
	}

	if entidade := r.URL.Query().Get("entidade"); entidade != "" {
		filter.Entidade = &entidade
	}
	if acao := r.URL.Query().Get("acao"); acao != "" {
		filter.Acao = &acao
	}
	if usuarioIDStr := r.URL.Query().Get("usuario_id"); usuarioIDStr != "" {
		usuarioID, err := uuid.Parse(usuarioIDStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_USUARIO_ID", "usuario_id inválido", "")
			return
		}
		filter.UsuarioID = &usuarioID
	}

	auditorias, err := h.auditoriaUsecase.Listar(r.Context(), empresaID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, auditorias)
}

// parseQueryInt converte um valor de query string em inteiro, retornando o
// valor padrão quando ausente ou inválido.
func parseQueryInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
