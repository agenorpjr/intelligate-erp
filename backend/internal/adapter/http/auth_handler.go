package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type AuthHandler struct {
	authUsecase   *usecase.AuthUsecase
	senhaSessaoUC *usecase.SenhaSessaoUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase, senhaSessaoUC *usecase.SenhaSessaoUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, senhaSessaoUC: senhaSessaoUC}
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

// EsqueciSenha solicita um token de redefinição de senha.
// Responde sempre com sucesso genérico para evitar enumeração de usuários.
func (h *AuthHandler) EsqueciSenha(w http.ResponseWriter, r *http.Request) {
	var req domain.EsqueciSenhaInput
	if err := DecodeJSON(r, &req); err != nil {
		// Mesmo com requisição inválida, responder sucesso genérico para não expor informações.
		RespondJSON(w, http.StatusOK, map[string]string{"message": "Se o e-mail existir, um token de redefinição será enviado."})
		return
	}

	tokenDev, err := h.senhaSessaoUC.EsqueciSenha(r.Context(), req)
	if err != nil {
		// Erro interno: não expor detalhes, responder sucesso genérico.
		RespondJSON(w, http.StatusOK, map[string]string{"message": "Se o e-mail existir, um token de redefinição será enviado."})
		return
	}

	resp := map[string]string{"message": "Se o e-mail existir, um token de redefinição será enviado."}
	if tokenDev != "" {
		resp["token"] = tokenDev
	}
	RespondJSON(w, http.StatusOK, resp)
}

// RedefinirSenha redefine a senha do usuário usando um token de redefinição.
func (h *AuthHandler) RedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var req domain.RedefinirSenhaInput
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	if err := h.senhaSessaoUC.RedefinirSenha(r.Context(), req); err != nil {
		if errors.Is(err, usecase.ErrSenhaRedefinicaoNaoEncontrada) || errors.Is(err, usecase.ErrSenhaRedefinicaoJaUsada) {
			RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "Token de redefinição inválido ou expirado", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioInativo) {
			RespondError(w, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Senha redefinida com sucesso."})
}

// AlterarSenha altera a senha de um usuário autenticado.
func (h *AuthHandler) AlterarSenha(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	var req domain.AlterarSenhaInput
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	if err := h.senhaSessaoUC.AlterarSenha(r.Context(), usuario.ID, req); err != nil {
		if errors.Is(err, usecase.ErrSenhaAtualInvalida) {
			RespondError(w, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "Senha atual inválida", "")
			return
		}
		if errors.Is(err, usecase.ErrNovaSenhaIgualAtual) {
			RespondError(w, http.StatusBadRequest, "SAME_PASSWORD", "Nova senha deve ser diferente da senha atual", "")
			return
		}
		if errors.Is(err, usecase.ErrUsuarioInativo) {
			RespondError(w, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Senha alterada com sucesso."})
}

// Logout revoga o token JWT atual do usuário.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Token de autorização não fornecido", "")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Formato de token inválido", "")
		return
	}

	tokenString := parts[1]

	jti, err := h.authUsecase.ExtrairJTI(tokenString)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "Não foi possível extrair identificador do token", "")
		return
	}

	expiraEm, err := h.authUsecase.ExtrairExpiracao(tokenString)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "Não foi possível extrair expiração do token", "")
		return
	}

	if err := h.senhaSessaoUC.Logout(r.Context(), usuario.ID, jti, expiraEm); err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao fazer logout", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Logout realizado com sucesso."})
}
