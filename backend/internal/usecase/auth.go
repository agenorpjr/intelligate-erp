package usecase

import (
	"context"
	"errors"
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
	usuarioRepo domain.UsuarioRepository
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthUsecase(usuarioRepo domain.UsuarioRepository, jwtSecret string, jwtExpiry time.Duration) *AuthUsecase {
	return &AuthUsecase{
		usuarioRepo: usuarioRepo,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
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
