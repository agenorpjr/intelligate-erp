package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// AuthUsecaseInterface define a interface necessária para o RBACMiddleware
// Isso permite usar mocks nos testes
type AuthUsecaseInterface interface {
	GetPermissoesUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]string, error)
	GetLojasUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]uuid.UUID, error)
	GetUsuarioFromToken(ctx context.Context, tokenString string) (*domain.Usuario, error)
}

// RBACMiddleware gerencia autorização baseada em permissões
type RBACMiddleware struct {
	authUsecase AuthUsecaseInterface
}

// NewRBACMiddleware cria uma nova instância do middleware RBAC
func NewRBACMiddleware(authUsecase AuthUsecaseInterface) *RBACMiddleware {
	return &RBACMiddleware{authUsecase: authUsecase}
}

// RequirePermission retorna um middleware que verifica se o usuário tem a permissão necessária
func (m *RBACMiddleware) RequirePermission(permissao string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Primeiro, garantir que o usuário está autenticado
			usuario := GetUsuarioFromContext(r.Context())
			if usuario == nil {
				RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
				return
			}

			// Buscar permissões do usuário (com cache no contexto para evitar múltiplas queries)
			permissoes := GetPermissoesFromContext(r.Context())
			if permissoes == nil {
				// Carregar permissões do usuário
				loaded, err := m.authUsecase.GetPermissoesUsuario(r.Context(), usuario.ID, GetEmpresaIDFromContext(r.Context()))
				if err != nil {
					RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar permissões", "")
					return
				}
				permissoes = loaded
				// Armazenar no contexto para reutilização
				ctx := context.WithValue(r.Context(), PermissoesContextKey, permissoes)
				r = r.WithContext(ctx)
			}

			// Verificar se o usuário tem a permissão
			if !hasPermission(permissoes, permissao) {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente: "+permissao, "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission retorna um middleware que verifica se o usuário tem PELO MENOS UMA das permissões
func (m *RBACMiddleware) RequireAnyPermission(permissoes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usuario := GetUsuarioFromContext(r.Context())
			if usuario == nil {
				RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
				return
			}

			userPerms := GetPermissoesFromContext(r.Context())
			if userPerms == nil {
				perms, err := m.authUsecase.GetPermissoesUsuario(r.Context(), usuario.ID, GetEmpresaIDFromContext(r.Context()))
				if err != nil {
					RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar permissões", "")
					return
				}
				ctx := context.WithValue(r.Context(), PermissoesContextKey, perms)
				r = r.WithContext(ctx)
				userPerms = perms
			}

			// Verificar se tem pelo menos uma das permissões
			hasAny := false
			for _, p := range permissoes {
				if hasPermission(userPerms, p) {
					hasAny = true
					break
				}
			}

			if !hasAny {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente: uma de "+joinPermissoes(permissoes), "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireLojaAccess retorna um middleware que verifica se o usuário tem acesso à loja
// especificada no path param informado.
func (m *RBACMiddleware) RequireLojaAccess(lojaIDParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usuario := GetUsuarioFromContext(r.Context())
			if usuario == nil {
				RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
				return
			}

			// Obter loja_id da URL
			lojaIDStr := chi.URLParam(r, lojaIDParam)
			if lojaIDStr == "" {
				// Se não tem loja_id na URL, permitir (pode ser endpoint de nível empresa)
				next.ServeHTTP(w, r)
				return
			}

			lojaID, err := uuid.Parse(lojaIDStr)
			if err != nil {
				RespondError(w, http.StatusBadRequest, "INVALID_LOJA_ID", "ID de loja inválido", "")
				return
			}

			lojasPermitidas, newReq, err := m.lojasPermitidas(r, usuario)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar acesso à loja", "")
				return
			}
			r = newReq

			if !temAcessoLoja(lojasPermitidas, lojaID) {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "Sem acesso a esta loja", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireLojaAccessBody valida o acesso à loja cujo ID está no corpo JSON da
// requisição, no campo informado. É útil para rotas POST/PUT onde o loja_id não
// está no path. O corpo é restaurado para o handler seguinte.
func (m *RBACMiddleware) RequireLojaAccessBody(field string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usuario := GetUsuarioFromContext(r.Context())
			if usuario == nil {
				RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Não foi possível ler o corpo da requisição", "")
				return
			}
			// Restaura o corpo para o próximo handler/middleware.
			r.Body = io.NopCloser(bytes.NewReader(body))

			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Corpo da requisição inválido", "")
				return
			}

			raw, ok := payload[field].(string)
			if !ok || raw == "" {
				// Campo ausente: o handler é responsável por validar a obrigatoriedade.
				next.ServeHTTP(w, r)
				return
			}

			lojaID, err := uuid.Parse(raw)
			if err != nil {
				RespondError(w, http.StatusBadRequest, "INVALID_LOJA_ID", "ID de loja inválido", "")
				return
			}

			lojasPermitidas, newReq, err := m.lojasPermitidas(r, usuario)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar acesso à loja", "")
				return
			}
			r = newReq

			if !temAcessoLoja(lojasPermitidas, lojaID) {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "Sem acesso a esta loja", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// lojasPermitidas retorna as lojas permitidas do usuário, reutilizando o cache
// do contexto quando disponível. Retorna também a request possivelmente
// atualizada com o cache.
func (m *RBACMiddleware) lojasPermitidas(r *http.Request, usuario *domain.Usuario) ([]uuid.UUID, *http.Request, error) {
	if lojas := GetLojasPermitidasFromContext(r.Context()); lojas != nil {
		return lojas, r, nil
	}

	lojas, err := m.authUsecase.GetLojasUsuario(r.Context(), usuario.ID, GetEmpresaIDFromContext(r.Context()))
	if err != nil {
		return nil, r, err
	}

	ctx := context.WithValue(r.Context(), LojasPermitidasContextKey, lojas)
	return lojas, r.WithContext(ctx), nil
}

// temAcessoLoja indica se o usuário pode acessar a loja informada.
// Lista vazia significa acesso a todas as lojas da empresa.
func temAcessoLoja(lojasPermitidas []uuid.UUID, lojaID uuid.UUID) bool {
	if len(lojasPermitidas) == 0 {
		return true
	}
	for _, id := range lojasPermitidas {
		if id == lojaID {
			return true
		}
	}
	return false
}

// GetPermissoesFromContext obtém as permissões do usuário do contexto
func GetPermissoesFromContext(ctx context.Context) []string {
	if perms, ok := ctx.Value(PermissoesContextKey).([]string); ok {
		return perms
	}
	return nil
}

// GetLojasPermitidasFromContext obtém as lojas permitidas do usuário do contexto
func GetLojasPermitidasFromContext(ctx context.Context) []uuid.UUID {
	if lojas, ok := ctx.Value(LojasPermitidasContextKey).([]uuid.UUID); ok {
		return lojas
	}
	return nil
}

// hasPermission verifica se a lista de permissões contém a permissão necessária
func hasPermission(permissoes []string, permissao string) bool {
	for _, p := range permissoes {
		if p == permissao {
			return true
		}
	}
	return false
}

// joinPermissoes junta as permissões em uma string para mensagem de erro
func joinPermissoes(permissoes []string) string {
	if len(permissoes) == 0 {
		return ""
	}
	result := permissoes[0]
	for i := 1; i < len(permissoes); i++ {
		result += ", " + permissoes[i]
	}
	return result
}
