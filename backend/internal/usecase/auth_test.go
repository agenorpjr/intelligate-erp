package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
)

type mockUsuarioRepository struct {
	usuarios map[string]*domain.Usuario
}

func newMockUsuarioRepository() *mockUsuarioRepository {
	return &mockUsuarioRepository{
		usuarios: make(map[string]*domain.Usuario),
	}
}

func (m *mockUsuarioRepository) Create(ctx context.Context, u *domain.Usuario) error {
	m.usuarios[u.Email] = u
	return nil
}

func (m *mockUsuarioRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	for _, u := range m.usuarios {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUsuarioRepository) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	if u, ok := m.usuarios[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUsuarioRepository) GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*domain.Usuario, error) {
	if u, ok := m.usuarios[email]; ok && u.EmpresaID == empresaID {
		return u, nil
	}
	return nil, nil
}

func (m *mockUsuarioRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Usuario, error) {
	var result []*domain.Usuario
	for _, u := range m.usuarios {
		if u.EmpresaID == empresaID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUsuarioRepository) Update(ctx context.Context, u *domain.Usuario) error {
	m.usuarios[u.Email] = u
	return nil
}

func (m *mockUsuarioRepository) UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error {
	for _, u := range m.usuarios {
		if u.ID == id {
			u.SenhaHash = senhaHash
			return nil
		}
	}
	return nil
}

func (m *mockUsuarioRepository) UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error {
	for _, u := range m.usuarios {
		if u.ID == id {
			now := time.Now()
			u.UltimoLogin = &now
			return nil
		}
	}
	return nil
}

func (m *mockUsuarioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for email, u := range m.usuarios {
		if u.ID == id {
			delete(m.usuarios, email)
			return nil
		}
	}
	return nil
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := NewAuthUsecase(repo, "test-secret", 24*time.Hour)

	// Create user with hashed password
	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	// Test login
	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "test@example.com", resp.Usuario.Email)
	assert.Equal(t, "Test User", resp.Usuario.Nome)
	assert.Empty(t, resp.Usuario.SenhaHash) // password should not be returned
	assert.Equal(t, int64(24*time.Hour.Seconds()), resp.ExpiresIn)
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := NewAuthUsecase(repo, "test-secret", 24*time.Hour)

	// Test with non-existent user
	_, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "nonexistent@example.com",
		Senha: "senha123",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// Test with wrong password
	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	_, err = uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "wrongpassword",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthUsecase_Login_InactiveUser(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := NewAuthUsecase(repo, "test-secret", 24*time.Hour)

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     false, // inactive user
	}
	repo.Create(context.Background(), usuario)

	_, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	assert.ErrorIs(t, err, ErrUsuarioInativo)
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := NewAuthUsecase(repo, "test-secret", 24*time.Hour)

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	// Login to get token
	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	require.NoError(t, err)

	// Validate token
	_, claims, err := uc.ValidateToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, usuario.ID.String(), claims["sub"])
	assert.Equal(t, usuario.EmpresaID.String(), claims["empresa_id"])
	assert.Equal(t, usuario.Email, claims["email"])
}

func TestAuthUsecase_GetUsuarioFromToken(t *testing.T) {
	repo := newMockUsuarioRepository()
	uc := NewAuthUsecase(repo, "test-secret", 24*time.Hour)

	senhaHash, _ := uc.HashSenha("senha123")
	usuario := &domain.Usuario{
		ID:        uuid.New(),
		EmpresaID: uuid.New(),
		Email:     "test@example.com",
		Nome:      "Test User",
		SenhaHash: senhaHash,
		Ativo:     true,
	}
	repo.Create(context.Background(), usuario)

	resp, err := uc.Login(context.Background(), domain.LoginInput{
		Email: "test@example.com",
		Senha: "senha123",
	})
	require.NoError(t, err)

	// Get user from token
	user, err := uc.GetUsuarioFromToken(context.Background(), resp.Token)
	require.NoError(t, err)
	assert.Equal(t, usuario.ID, user.ID)
	assert.Equal(t, usuario.Email, user.Email)
	assert.Empty(t, user.SenhaHash)
}
