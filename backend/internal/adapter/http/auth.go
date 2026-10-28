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
)

type AuthMiddleware struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthMiddleware(authUsecase *usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: authUsecase}
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
