package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type UsuarioHandler struct {
	usuarioUsecase *usecase.UsuarioUsecase
}

func NewUsuarioHandler(usuarioUsecase *usecase.UsuarioUsecase) *UsuarioHandler {
	return &UsuarioHandler{usuarioUsecase: usuarioUsecase}
}

// ========== Request/Response Types ==========

type CreateUsuarioRequest struct {
	EmpresaID uuid.UUID `json:"empresa_id" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Nome      string    `json:"nome" validate:"required"`
	Senha     string    `json:"senha" validate:"required,min=8"`
}

type UpdateUsuarioRequest struct {
	Nome  *string `json:"nome,omitempty"`
	Email *string `json:"email,omitempty"`
	Ativo *bool   `json:"ativo,omitempty"`
}

type UsuarioLojaRequest struct {
	LojaID uuid.UUID `json:"loja_id" validate:"required"`
}

// ========== CRUD Usuário ==========

func (h *UsuarioHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var req CreateUsuarioRequest
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

	input := domain.CreateUsuarioInput{
		EmpresaID: req.EmpresaID,
		Email:     req.Email,
		Nome:      req.Nome,
		Senha:     req.Senha,
	}

	u, err := h.usuarioUsecase.Criar(r.Context(), input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailJaExiste) {
			RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "E-mail já cadastrado para esta empresa", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, u)
}

func (h *UsuarioHandler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	u, err := h.usuarioUsecase.BuscarPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, u)
}

func (h *UsuarioHandler) Listar(w http.ResponseWriter, r *http.Request) {
	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0

	// Parse query params for pagination
	if l := r.URL.Query().Get("limit"); l != "" {
		// Simple parse, in production use strconv
		// limit = parseInt(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		// offset = parseInt(o)
	}

	usuarios, err := h.usuarioUsecase.Listar(r.Context(), empresaID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, usuarios)
}

func (h *UsuarioHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido", "")
		return
	}

	var req UpdateUsuarioRequest
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

	input := domain.UpdateUsuarioInput{
		Nome:  req.Nome,
		Email: req.Email,
		Ativo: req.Ativo,
	}

	u, err := h.usuarioUsecase.Atualizar(r.Context(), id, input, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrEmailJaExiste) {
			RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "E-mail já cadastrado para esta empresa", "")
			return
		}
		if errors.Is(err, usecase.ErrNaoPodeDesativarSiMesmo) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível desativar a si mesmo", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, u)
}

func (h *UsuarioHandler) Desativar(w http.ResponseWriter, r *http.Request) {
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

	err = h.usuarioUsecase.Desativar(r.Context(), id, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrNaoPodeDesativarSiMesmo) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível desativar a si mesmo", "")
			return
		}
		if errors.Is(err, usecase.ErrNaoPodeRemoverPapelAdmin) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível remover o último administrador ativo da empresa", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Usuário desativado com sucesso"})
}

func (h *UsuarioHandler) Reativar(w http.ResponseWriter, r *http.Request) {
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

	err = h.usuarioUsecase.Reativar(r.Context(), id, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Usuário reativado com sucesso"})
}

// ========== Gestão de Papéis ==========

func (h *UsuarioHandler) AssociarPapel(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	type AssociarPapelRequest struct {
		PapelID uuid.UUID  `json:"papel_id" validate:"required"`
		LojaID  *uuid.UUID `json:"loja_id,omitempty"`
	}

	var req AssociarPapelRequest
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

	err = h.usuarioUsecase.AssociarPapel(r.Context(), usuarioID, req.PapelID, req.LojaID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Papel não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelDeOutraEmpresa) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Papel pertence a outra empresa", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaDeOutraEmpresa) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Loja pertence a outra empresa", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioJaTemPapel) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Usuário já possui este papel", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, map[string]string{"message": "Papel associado com sucesso"})
}

func (h *UsuarioHandler) RemoverPapel(w http.ResponseWriter, r *http.Request) {
	usuarioPapelIDStr := r.PathValue("papel_id")
	usuarioPapelID, err := uuid.Parse(usuarioPapelIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de associação usuário-papel inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	err = h.usuarioUsecase.RemoverPapel(r.Context(), usuarioPapelID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoTemPapel) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Associação usuário-papel não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrPapelSistema) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível remover papel do sistema", "")
			return
		}
		if errors.Is(err, usecase.ErrNaoPodeRemoverPapelAdmin) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Não é possível remover o último administrador ativo da empresa", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Papel removido com sucesso"})
}

func (h *UsuarioHandler) ListarPapeis(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	papeis, err := h.usuarioUsecase.ListarPapeis(r.Context(), usuarioID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, papeis)
}

func (h *UsuarioHandler) ListarPermissoesEfetivas(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	permissoes, err := h.usuarioUsecase.ListarPermissoesEfetivas(r.Context(), usuarioID, empresaID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"usuario_id": usuarioID,
		"empresa_id": empresaID,
		"permissoes": permissoes,
	})
}

// ========== Gestão de Lojas ==========

func (h *UsuarioHandler) AdicionarLoja(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	var req UsuarioLojaRequest
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

	err = h.usuarioUsecase.AdicionarLoja(r.Context(), usuarioID, req.LojaID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaDeOutraEmpresa) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Loja pertence a outra empresa", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioJaTemAcessoLoja) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Usuário já tem acesso a esta loja", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, map[string]string{"message": "Acesso à loja adicionado com sucesso"})
}

func (h *UsuarioHandler) RemoverLoja(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	lojaIDStr := r.PathValue("loja_id")
	lojaID, err := uuid.Parse(lojaIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de loja inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	err = h.usuarioUsecase.RemoverLoja(r.Context(), usuarioID, lojaID, usuario.ID, ip, userAgent)
	if err != nil {
		if errors.Is(err, usecase.ErrUsuarioNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioNaoTemAcessoLoja) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Usuário não tem acesso a esta loja", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Acesso à loja removido com sucesso"})
}

func (h *UsuarioHandler) ListarLojas(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.PathValue("id")
	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuário inválido", "")
		return
	}

	lojas, err := h.usuarioUsecase.ListarLojas(r.Context(), usuarioID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, lojas)
}
