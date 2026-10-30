package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SenhaRedefinicaoRepository struct {
	querier *Queries
}

func NewSenhaRedefinicaoRepository(pool *pgxpool.Pool) *SenhaRedefinicaoRepository {
	return &SenhaRedefinicaoRepository{querier: New(pool)}
}

func (r *SenhaRedefinicaoRepository) Create(ctx context.Context, s *domain.SenhaRedefinicao) error {
	params := CreateSenhaRedefinicaoParams{
		UsuarioID: s.UsuarioID,
		TokenHash: s.TokenHash,
		ExpiraEm:  s.ExpiraEm,
	}
	result, err := r.querier.CreateSenhaRedefinicao(ctx, params)
	if err != nil {
		return err
	}
	s.ID = result.ID
	s.CriadoEm = result.CriadoEm
	return nil
}

func (r *SenhaRedefinicaoRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.SenhaRedefinicao, error) {
	result, err := r.querier.GetSenhaRedefinicaoByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return toDomainSenhaRedefinicao(result), nil
}

func (r *SenhaRedefinicaoRepository) InvalidarPorUsuario(ctx context.Context, usuarioID uuid.UUID) error {
	return r.querier.InvalidateSenhaRedefinicoesByUsuario(ctx, usuarioID)
}

func (r *SenhaRedefinicaoRepository) MarcarUsado(ctx context.Context, id uuid.UUID) error {
	return r.querier.MarcarSenhaRedefinicaoUsada(ctx, id)
}

func (r *SenhaRedefinicaoRepository) ListarValidasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SenhaRedefinicao, error) {
	results, err := r.querier.ListSenhaRedefinicoesValidasPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	itens := make([]*domain.SenhaRedefinicao, len(results))
	for i, r := range results {
		itens[i] = toDomainSenhaRedefinicao(r)
	}
	return itens, nil
}

func toDomainSenhaRedefinicao(s SenhaRedefinico) *domain.SenhaRedefinicao {
	var usadoEm *time.Time
	if s.UsadoEm.Valid {
		usadoEm = &s.UsadoEm.Time
	}
	return &domain.SenhaRedefinicao{
		ID:        s.ID,
		UsuarioID: s.UsuarioID,
		TokenHash: s.TokenHash,
		Usado:     s.Usado,
		ExpiraEm:  s.ExpiraEm,
		CriadoEm:  s.CriadoEm,
		UsadoEm:   usadoEm,
	}
}
