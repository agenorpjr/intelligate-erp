package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type contextKey string

const (
	UserContextKey            contextKey = "user"
	EmpresaContextKey         contextKey = "empresa_id"
	PermissoesContextKey      contextKey = "permissoes"
	LojasPermitidasContextKey contextKey = "lojas_permitidas"
	JTIContextKey             contextKey = "jti"
)

type AuthMiddleware struct {
	authUsecase   *usecase.AuthUsecase
	senhaSessaoUC *usecase.SenhaSessaoUsecase
}

func NewAuthMiddleware(authUsecase *usecase.AuthUsecase, senhaSessaoUC *usecase.SenhaSessaoUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: authUsecase, senhaSessaoUC: senhaSessaoUC}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token de autorização não fornecido", "")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Formato de token inválido", "")
			return
		}

		tokenString := parts[1]

		// Verificar se o token foi revogado (logout / invalidez de sessão).
		jti, err := m.authUsecase.ExtrairJTI(tokenString)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token inválido", "")
			return
		}

		revogado, err := m.senhaSessaoUC.ExisteSessaoRevogadaPorJTI(r.Context(), jti)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar sessão", err.Error())
			return
		}
		if revogado {
			RespondError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Token revogado", "")
			return
		}

		usuario, err := m.authUsecase.GetUsuarioFromToken(r.Context(), tokenString)
		if err != nil {
			if errors.Is(err, usecase.ErrTokenExpirado) {
				RespondError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token expirado", "")
				return
			}
			RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token inválido", "")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, usuario)
		ctx = context.WithValue(ctx, EmpresaContextKey, usuario.EmpresaID)
		ctx = context.WithValue(ctx, JTIContextKey, jti)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUsuarioFromContext(ctx context.Context) *domain.Usuario {
	if usuario, ok := ctx.Value(UserContextKey).(*domain.Usuario); ok {
		return usuario
	}
	return nil
}

func GetEmpresaIDFromContext(ctx context.Context) uuid.UUID {
	if empresaID, ok := ctx.Value(EmpresaContextKey).(uuid.UUID); ok {
		return empresaID
	}
	return uuid.Nil
}

func GetJTIFromContext(ctx context.Context) string {
	if v := ctx.Value(JTIContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
