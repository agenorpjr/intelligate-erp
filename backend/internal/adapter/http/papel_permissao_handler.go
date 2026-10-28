package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type PapelPermissaoHandler struct {
	papelPermissaoUsecase *usecase.PapelPermissaoUsecase
}

func NewPapelPermissaoHandler(papelPermissaoUsecase *usecase.PapelPermissaoUsecase) *PapelPermissaoHandler {
	return &PapelPermissaoHandler{papelPermissaoUsecase: papelPermissaoUsecase}
}

// ========== Papel Handlers ==========

type CreatePapelRequest struct {
	EmpresaID uuid.UUID `json:"empresa_id" validate:"required"`
	Nome      string    `json:"nome" validate:"required"`
	Descricao *string   `json:"descricao,omitempty"`
	Sistema   bool      `json:"sistema"`
}

type UpdatePapelRequest struct {
	Nome      *string `json:"nome,omitempty"`
	Descricao *string `json:"descricao,omitempty"`
	Sistema   *bool   `json:"sistema,omitempty"`
}

func (h *PapelPermissaoHandler) CriarPapel(w http.ResponseWriter, r *http.Request) {
	var req CreatePapelRequest
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

	input := domain.CreatePapelInput{
		EmpresaID: req.EmpresaID,
		Nome:      req.Nome,
		Descricao: req.Descricao,
		Sistema:   req.Sistema,
	}

	papel, err := h.papelPermissaoUsecase.CriarPapel(r.Context(), input, usuario.ID, ip, userAgent)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, papel)
}

func (h *PapelPermissaoHandler) ListarPapeis(w http.ResponseWriter, r *http.Request) {
	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0

	papeis, err := h.papelPermissaoUsecase.ListarPapeis(r.Context(), empresaID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, papeis)
}

func (h *PapelPermissaoHandler) BuscarPapel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	papel, err := h.papelPermissaoUsecase.BuscarPapel(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, papel)
}

func (h *PapelPermissaoHandler) AtualizarPapel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var req UpdatePapelRequest
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

	input := domain.UpdatePapelInput{
		Nome:      req.Nome,
		Descricao: req.Descricao,
		Sistema:   req.Sistema,
	}

	papel, err := h.papelPermissaoUsecase.AtualizarPapel(r.Context(), id, input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelSistema) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível modificar papel do sistema", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, papel)
}

func (h *PapelPermissaoHandler) DeletarPapel(w http.ResponseWriter, r *http.Request) {
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

	err = h.papelPermissaoUsecase.DeletarPapel(r.Context(), id, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelSistema) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível excluir papel do sistema", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Papel excluído com sucesso"})
}

// ========== Permissão Handlers ==========

func (h *PapelPermissaoHandler) ListarPermissoes(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	permissoes, err := h.papelPermissaoUsecase.ListarPermissoes(r.Context(), limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, permissoes)
}

func (h *PapelPermissaoHandler) BuscarPermissao(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	permissao, err := h.papelPermissaoUsecase.BuscarPermissao(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrPermissaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Permissão não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, permissao)
}

// ========== Papel-Permissão Association Handlers ==========

// PapelPermissaoRequest é o corpo de POST /papeis/{papel_id}/permissoes.
// O papel_id vem do path param; apenas a permissão é enviada no corpo.
type PapelPermissaoRequest struct {
	PermissaoID uuid.UUID `json:"permissao_id" validate:"required"`
}

func (h *PapelPermissaoHandler) AdicionarPermissaoAoPapel(w http.ResponseWriter, r *http.Request) {
	papelID, err := uuid.Parse(r.PathValue("papel_id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de papel inválido", "")
		return
	}

	var req PapelPermissaoRequest
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

	err = h.papelPermissaoUsecase.AdicionarPermissaoAoPapel(r.Context(), papelID, req.PermissaoID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPermissaoNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Permissão não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelSistema) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível modificar papel do sistema", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Permissão adicionada ao papel com sucesso"})
}

func (h *PapelPermissaoHandler) RemoverPermissaoDoPapel(w http.ResponseWriter, r *http.Request) {
	papelID, err := uuid.Parse(r.PathValue("papel_id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de papel inválido", "")
		return
	}

	permissaoID, err := uuid.Parse(r.PathValue("permissao_id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de permissão inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	err = h.papelPermissaoUsecase.RemoverPermissaoDoPapel(r.Context(), papelID, permissaoID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelSistema) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível modificar papel do sistema", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Permissão removida do papel com sucesso"})
}

func (h *PapelPermissaoHandler) ListarPermissoesDoPapel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	permissoes, err := h.papelPermissaoUsecase.ListarPermissoesDoPapel(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, permissoes)
}

// ========== Usuario-Papel Association Handlers ==========

// UsuarioPapelRequest é o corpo de POST /usuarios/{id}/papeis.
// O usuario_id vem do path param; papel_id e loja_id (opcional) no corpo.
type UsuarioPapelRequest struct {
	PapelID uuid.UUID  `json:"papel_id" validate:"required"`
	LojaID  *uuid.UUID `json:"loja_id,omitempty"`
}

func (h *PapelPermissaoHandler) AssociarUsuarioPapel(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	var req UsuarioPapelRequest
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

	input := domain.UsuarioPapel{
		UsuarioID: usuarioID,
		PapelID:   req.PapelID,
		LojaID:    req.LojaID,
	}

	err = h.papelPermissaoUsecase.AssociarUsuarioPapel(r.Context(), input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Usuário associado ao papel com sucesso"})
}

func (h *PapelPermissaoHandler) RemoverUsuarioPapel(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	papelID, err := uuid.Parse(r.PathValue("papel_id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de papel inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	// Localiza as associações usuário-papel correspondentes ao papel informado.
	associacoes, err := h.papelPermissaoUsecase.ListarPapeisDoUsuario(r.Context(), usuarioID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	removidas := 0
	for _, a := range associacoes {
		if a.PapelID != papelID {
			continue
		}
		if err := h.papelPermissaoUsecase.RemoverUsuarioPapel(r.Context(), a.ID, usuario.ID, ip, userAgent); err != nil {
			if errors.Is(err, usecase.ErrPapelSistema) {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível modificar papel do sistema", "")
				return
			}
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
			return
		}
		removidas++
	}

	if removidas == 0 {
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Associação usuário-papel não encontrada", "")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Associação usuário-papel removida com sucesso"})
}

func (h *PapelPermissaoHandler) ListarPapeisDoUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	papeis, err := h.papelPermissaoUsecase.ListarPapeisDoUsuario(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, papeis)
}
