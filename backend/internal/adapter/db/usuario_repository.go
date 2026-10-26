package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsuarioRepository struct {
	querier *Queries
}

func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{querier: New(pool)}
}

func (r *UsuarioRepository) Create(ctx context.Context, u *domain.Usuario) error {
	params := CreateUsuarioParams{
		EmpresaID: u.EmpresaID,
		Email:     u.Email,
		Nome:      u.Nome,
		SenhaHash: u.SenhaHash,
	}
	result, err := r.querier.CreateUsuario(ctx, params)
	if err != nil {
		return err
	}
	u.ID = result.ID
	u.CreatedAt = result.CreatedAt
	u.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *UsuarioRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Usuario, error) {
	result, err := r.querier.GetUsuario(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainUsuario(result), nil
}

func (r *UsuarioRepository) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	result, err := r.querier.GetUsuarioByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toDomainUsuario(result), nil
}

func (r *UsuarioRepository) GetByEmailAndEmpresa(ctx context.Context, email string, empresaID uuid.UUID) (*domain.Usuario, error) {
	result, err := r.querier.GetUsuarioByEmailAndEmpresa(ctx, GetUsuarioByEmailAndEmpresaParams{
		Email:     email,
		EmpresaID: empresaID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainUsuario(result), nil
}

func (r *UsuarioRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Usuario, error) {
	results, err := r.querier.ListUsuariosByEmpresa(ctx, ListUsuariosByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	usuarios := make([]*domain.Usuario, len(results))
	for i, r := range results {
		usuarios[i] = toDomainUsuario(r)
	}
	return usuarios, nil
}

func (r *UsuarioRepository) Update(ctx context.Context, u *domain.Usuario) error {
	params := UpdateUsuarioParams{
		ID:    u.ID,
		Nome:  u.Nome,
		Email: u.Email,
		Ativo: u.Ativo,
	}
	result, err := r.querier.UpdateUsuario(ctx, params)
	if err != nil {
		return err
	}
	u.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *UsuarioRepository) UpdateSenha(ctx context.Context, id uuid.UUID, senhaHash string) error {
	return r.querier.UpdateUsuarioSenha(ctx, UpdateUsuarioSenhaParams{
		ID:        id,
		SenhaHash: senhaHash,
	})
}

func (r *UsuarioRepository) UpdateUltimoLogin(ctx context.Context, id uuid.UUID) error {
	return r.querier.UpdateUsuarioUltimoLogin(ctx, id)
}

func (r *UsuarioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteUsuario(ctx, id)
}

func toDomainUsuario(u Usuario) *domain.Usuario {
	var ultimoLogin *time.Time
	if u.UltimoLogin.Valid {
		ultimoLogin = &u.UltimoLogin.Time
	}
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}
	return &domain.Usuario{
		ID:          u.ID,
		EmpresaID:   u.EmpresaID,
		Email:       u.Email,
		Nome:        u.Nome,
		SenhaHash:   u.SenhaHash,
		Ativo:       u.Ativo,
		UltimoLogin: ultimoLogin,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
