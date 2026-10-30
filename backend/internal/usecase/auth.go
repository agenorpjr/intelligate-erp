package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUsuarioInativo     = errors.New("usuário inativo")
	ErrTokenInvalido      = errors.New("token inválido")
	ErrTokenExpirado      = errors.New("token expirado")
)

type AuthUsecase struct {
	usuarioRepo        domain.UsuarioRepository
	usuarioPapelRepo   domain.UsuarioPapelRepository
	papelRepo          domain.PapelRepository
	papelPermissaoRepo domain.PapelPermissaoRepository
	jwtSecret          string
	jwtExpiry          time.Duration
}

func NewAuthUsecase(
	usuarioRepo domain.UsuarioRepository,
	usuarioPapelRepo domain.UsuarioPapelRepository,
	papelRepo domain.PapelRepository,
	papelPermissaoRepo domain.PapelPermissaoRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		usuarioRepo:        usuarioRepo,
		usuarioPapelRepo:   usuarioPapelRepo,
		papelRepo:          papelRepo,
		papelPermissaoRepo: papelPermissaoRepo,
		jwtSecret:          jwtSecret,
		jwtExpiry:          jwtExpiry,
	}
}

func (uc *AuthUsecase) Login(ctx context.Context, input domain.LoginInput) (*domain.LoginResponse, error) {
	usuario, err := uc.usuarioRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if usuario == nil {
		return nil, ErrInvalidCredentials
	}

	if !usuario.Ativo {
		return nil, ErrUsuarioInativo
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(input.Senha)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := uc.generateToken(usuario)
	if err != nil {
		return nil, err
	}

	if err := uc.usuarioRepo.UpdateUltimoLogin(ctx, usuario.ID); err != nil {
		return nil, err
	}

	usuario.SenhaHash = ""
	return &domain.LoginResponse{
		Token:     token,
		Usuario:   *usuario,
		ExpiresIn: int64(uc.jwtExpiry.Seconds()),
	}, nil
}

func (uc *AuthUsecase) generateToken(usuario *domain.Usuario) (string, error) {
	claims := jwt.MapClaims{
		"sub":        usuario.ID.String(),
		"empresa_id": usuario.EmpresaID.String(),
		"email":      usuario.Email,
		"jti":        uuid.New().String(),
		"exp":        time.Now().Add(uc.jwtExpiry).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUsecase) ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalido
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, ErrTokenExpirado
		}
		return nil, nil, ErrTokenInvalido
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, ErrTokenInvalido
	}

	return token, claims, nil
}

func (uc *AuthUsecase) GetUsuarioFromToken(ctx context.Context, tokenString string) (*domain.Usuario, error) {
	_, claims, err := uc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrTokenInvalido
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrTokenInvalido
	}

	usuario, err := uc.usuarioRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !usuario.Ativo {
		return nil, ErrUsuarioInativo
	}

	usuario.SenhaHash = ""
	return usuario, nil
}

func (uc *AuthUsecase) HashSenha(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// GetPermissoesUsuario retorna todas as permissões do usuário (agregadas de todos os seus papéis).
//
// Isolamento por empresa: apenas papéis pertencentes à empresa informada são
// considerados. Papéis de outra empresa (ou removidos) são ignorados, garantindo
// que um usuário nunca receba permissões fora do seu tenant.
func (uc *AuthUsecase) GetPermissoesUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]string, error) {
	// Buscar papéis do usuário
	papeisUsuario, err := uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	// Coletar todas as permissões dos papéis, filtrando por empresa
	permissoesSet := make(map[string]struct{})
	for _, up := range papeisUsuario {
		papel, err := uc.papelRepo.GetByID(ctx, up.PapelID)
		if err != nil || papel == nil {
			// Papel removido (soft delete) ou inacessível: ignora (fail-closed).
			continue
		}
		if papel.EmpresaID != empresaID {
			// Isolamento multi-tenant: nunca agregar permissões de outra empresa.
			continue
		}

		permissoes, err := uc.papelPermissaoRepo.ListPermissoesByPapel(ctx, up.PapelID)
		if err != nil {
			return nil, err
		}
		for _, p := range permissoes {
			permissoesSet[p.Chave] = struct{}{}
		}
	}

	// Converter set para slice ordenado (saída determinística)
	permissoes := make([]string, 0, len(permissoesSet))
	for p := range permissoesSet {
		permissoes = append(permissoes, p)
	}
	sort.Strings(permissoes)

	return permissoes, nil
}

// GetLojasUsuario retorna os IDs das lojas que o usuário tem acesso.
//
// Retorna slice vazio se o usuário tem acesso a todas as lojas da empresa
// (papel sem loja_id). Isolamento por empresa: papéis de outra empresa são
// ignorados.
func (uc *AuthUsecase) GetLojasUsuario(ctx context.Context, usuarioID, empresaID uuid.UUID) ([]uuid.UUID, error) {
	papeisUsuario, err := uc.usuarioPapelRepo.ListByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	lojasSet := make(map[uuid.UUID]struct{})
	temAcessoGlobal := false

	for _, up := range papeisUsuario {
		papel, err := uc.papelRepo.GetByID(ctx, up.PapelID)
		if err != nil || papel == nil {
			// Papel removido (soft delete) ou inacessível: ignora (fail-closed).
			continue
		}
		if papel.EmpresaID != empresaID {
			// Isolamento multi-tenant: nunca conceder acesso a lojas de outra empresa.
			continue
		}

		if up.LojaID == nil {
			// Papel sem loja específica = acesso a todas as lojas da empresa
			temAcessoGlobal = true
		} else {
			lojasSet[*up.LojaID] = struct{}{}
		}
	}

	if temAcessoGlobal {
		// Retornar slice vazio para indicar acesso a todas as lojas
		return []uuid.UUID{}, nil
	}

	lojas := make([]uuid.UUID, 0, len(lojasSet))
	for id := range lojasSet {
		lojas = append(lojas, id)
	}

	return lojas, nil
}

// ExtrairJTI extrai o identificador único (JTI) de um token JWT.
func (uc *AuthUsecase) ExtrairJTI(tokenString string) (string, error) {
	_, claims, err := uc.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return "", ErrTokenInvalido
	}
	return jti, nil
}

// ExtrairExpiracao extrai o tempo de expiração de um token JWT.
func (uc *AuthUsecase) ExtrairExpiracao(tokenString string) (time.Time, error) {
	_, claims, err := uc.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, ErrTokenInvalido
	}
	return time.Unix(int64(expFloat), 0), nil
}
