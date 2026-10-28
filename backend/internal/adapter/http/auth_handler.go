package http

import (
	"errors"
	"net/http"

	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

type LoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	input := domain.LoginInput{
		Email: req.Email,
		Senha: req.Senha,
	}

	resp, err := h.authUsecase.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenciais inválidas", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioInativo) {
			RespondError(w, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	RespondJSON(w, http.StatusOK, usuario)
}

// MinhasPermissoes retorna as permissões efetivas do usuário autenticado.
// Útil para o frontend renderizar menus/ações conforme o RBAC (governança).
func (h *AuthHandler) MinhasPermissoes(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	permissoes, err := h.authUsecase.GetPermissoesUsuario(r.Context(), usuario.ID, usuario.EmpresaID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao carregar permissões", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"usuario_id": usuario.ID,
		"empresa_id": usuario.EmpresaID,
		"permissoes": permissoes,
	})
}
