package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsuarioPapelRepository struct {
	querier *Queries
}

func NewUsuarioPapelRepository(pool *pgxpool.Pool) *UsuarioPapelRepository {
	return &UsuarioPapelRepository{querier: New(pool)}
}

func (r *UsuarioPapelRepository) Create(ctx context.Context, up *domain.UsuarioPapel) error {
	var lojaID pgtype.UUID
	if up.LojaID != nil {
		lojaID = pgtype.UUID{Bytes: *up.LojaID, Valid: true}
	}
	params := CreateUsuarioPapelParams{
		UsuarioID: up.UsuarioID,
		PapelID:   up.PapelID,
		LojaID:    lojaID,
	}
	result, err := r.querier.CreateUsuarioPapel(ctx, params)
	if err != nil {
		return err
	}
	up.ID = result.ID
	up.CreatedAt = result.CreatedAt
	return nil
}

func (r *UsuarioPapelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UsuarioPapel, error) {
	result, err := r.querier.GetUsuarioPapel(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainUsuarioPapel(result), nil
}

func (r *UsuarioPapelRepository) ListByUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.UsuarioPapelComDetalhes, error) {
	results, err := r.querier.ListUsuarioPapeisByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	associacoes := make([]*domain.UsuarioPapelComDetalhes, len(results))
	for i, r := range results {
		var lojaID *uuid.UUID
		if r.LojaID.Valid {
			var u uuid.UUID
			copy(u[:], r.LojaID.Bytes[:])
			lojaID = &u
		}
		var lojaNome *string
		if r.LojaNome.Valid {
			lojaNome = &r.LojaNome.String
		}
		associacoes[i] = &domain.UsuarioPapelComDetalhes{
			ID:        r.ID,
			UsuarioID: r.UsuarioID,
			PapelID:   r.PapelID,
			LojaID:    lojaID,
			PapelNome: r.PapelNome,
			LojaNome:  lojaNome,
			CreatedAt: r.CreatedAt,
		}
	}
	return associacoes, nil
}

func (r *UsuarioPapelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteUsuarioPapel(ctx, id)
}

func toDomainUsuarioPapel(up UsuariosPapei) *domain.UsuarioPapel {
	var lojaID *uuid.UUID
	if up.LojaID.Valid {
		var u uuid.UUID
		copy(u[:], up.LojaID.Bytes[:])
		lojaID = &u
	}
	return &domain.UsuarioPapel{
		ID:        up.ID,
		UsuarioID: up.UsuarioID,
		PapelID:   up.PapelID,
		LojaID:    lojaID,
		CreatedAt: up.CreatedAt,
	}
}
