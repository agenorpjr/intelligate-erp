package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessaoRevogadaRepository struct {
	querier *Queries
}

func NewSessaoRevogadaRepository(pool *pgxpool.Pool) *SessaoRevogadaRepository {
	return &SessaoRevogadaRepository{querier: New(pool)}
}

func (r *SessaoRevogadaRepository) Create(ctx context.Context, s *domain.SessaoRevogada) error {
	params := CreateSessaoRevogadaParams{
		UsuarioID: s.UsuarioID,
		Jti:       s.JTI,
		ExpiraEm:  s.ExpiraEm,
		Motivo:    s.Motivo,
	}
	result, err := r.querier.CreateSessaoRevogada(ctx, params)
	if err != nil {
		return err
	}
	s.ID = result.ID
	s.RevogadoEm = result.RevogadoEm
	return nil
}

func (r *SessaoRevogadaRepository) ExistePorJTI(ctx context.Context, jti string) (bool, error) {
	result, err := r.querier.GetSessaoRevogadaByJTI(ctx, jti)
	if err != nil {
		return false, err
	}
	return result.ID != uuid.Nil, nil
}

func (r *SessaoRevogadaRepository) RevogarPorUsuario(ctx context.Context, usuarioID uuid.UUID, excecaoJTI *string, motivo string) error {
	if excecaoJTI != nil && *excecaoJTI != "" {
		return r.querier.RevogarSessoesPorUsuarioExcetoJTI(ctx, RevogarSessoesPorUsuarioExcetoJTIParams{
			UsuarioID: usuarioID,
			Jti:       *excecaoJTI,
		})
	}
	return r.querier.RevogarSessoesPorUsuario(ctx, usuarioID)
}

func (r *SessaoRevogadaRepository) ListarAtivasPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]*domain.SessaoRevogada, error) {
	results, err := r.querier.ListSessoesRevogadasAtivasPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	itens := make([]*domain.SessaoRevogada, len(results))
	for i, s := range results {
		itens[i] = toDomainSessaoRevogada(s)
	}
	return itens, nil
}

func (r *SessaoRevogadaRepository) LimparExpiradas(ctx context.Context) error {
	return r.querier.LimparSessoesRevogadasExpiradas(ctx)
}

func toDomainSessaoRevogada(s SessoesRevogada) *domain.SessaoRevogada {
	return &domain.SessaoRevogada{
		ID:         s.ID,
		UsuarioID:  s.UsuarioID,
		JTI:        s.Jti,
		ExpiraEm:   s.ExpiraEm,
		RevogadoEm: s.RevogadoEm,
		Motivo:     s.Motivo,
	}
}

var _ = time.Second
var _ = time.Second
