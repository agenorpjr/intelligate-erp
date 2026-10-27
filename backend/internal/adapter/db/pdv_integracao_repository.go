package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PdvIntegracaoRepository struct {
	querier *Queries
}

func NewPdvIntegracaoRepository(pool *pgxpool.Pool) *PdvIntegracaoRepository {
	return &PdvIntegracaoRepository{querier: New(pool)}
}

func (r *PdvIntegracaoRepository) Create(ctx context.Context, p *domain.PdvIntegracao) error {
	params := CreatePdvIntegracaoParams{
		EmpresaID:             p.EmpresaID,
		LojaID:                p.LojaID,
		Provedor:              p.Provedor,
		Ambiente:              p.Ambiente,
		UrlBase:               p.URLBase,
		ApiKeyEncrypted:       p.APIKeyEncrypted,
		ClientID:              pgtype.Text{String: p.ClientID, Valid: p.ClientID != ""},
		ClientSecretEncrypted: pgtype.Text{String: p.ClientSecretEncrypted, Valid: p.ClientSecretEncrypted != ""},
	}
	result, err := r.querier.CreatePdvIntegracao(ctx, params)
	if err != nil {
		return err
	}
	p.ID = result.ID
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PdvIntegracaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PdvIntegracao, error) {
	result, err := r.querier.GetPdvIntegracao(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPdvIntegracao(result), nil
}

func (r *PdvIntegracaoRepository) GetByLojaProvedor(ctx context.Context, lojaID uuid.UUID, provedor string) (*domain.PdvIntegracao, error) {
	result, err := r.querier.GetPdvIntegracaoByLojaProvedor(ctx, GetPdvIntegracaoByLojaProvedorParams{
		LojaID:   lojaID,
		Provedor: provedor,
	})
	if err != nil {
		return nil, err
	}
	return toDomainPdvIntegracao(result), nil
}

func (r *PdvIntegracaoRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, filter domain.PdvIntegracaoFilter) ([]*domain.PdvIntegracao, error) {
	var provedor string
	if filter.Provedor != nil {
		provedor = *filter.Provedor
	}
	var ativo bool
	if filter.Ativo != nil {
		ativo = *filter.Ativo
	}

	results, err := r.querier.ListPdvIntegracoesByEmpresa(ctx, ListPdvIntegracoesByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   provedor,
		Column3:   ativo,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	integracoes := make([]*domain.PdvIntegracao, len(results))
	for i, r := range results {
		integracoes[i] = toDomainPdvIntegracao(r)
	}
	return integracoes, nil
}

func (r *PdvIntegracaoRepository) Update(ctx context.Context, p *domain.PdvIntegracao) error {
	params := UpdatePdvIntegracaoParams{
		ID:                    p.ID,
		Provedor:              p.Provedor,
		Ambiente:              p.Ambiente,
		UrlBase:               p.URLBase,
		ApiKeyEncrypted:       p.APIKeyEncrypted,
		ClientID:              pgtype.Text{String: p.ClientID, Valid: p.ClientID != ""},
		ClientSecretEncrypted: pgtype.Text{String: p.ClientSecretEncrypted, Valid: p.ClientSecretEncrypted != ""},
		Ativo:                 p.Ativo,
	}
	result, err := r.querier.UpdatePdvIntegracao(ctx, params)
	if err != nil {
		return err
	}
	p.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *PdvIntegracaoRepository) UpdateUltimaSincronizacao(ctx context.Context, id uuid.UUID, ultimaSincronizacao time.Time) error {
	params := UpdatePdvIntegracaoUltimaSincronizacaoParams{
		ID:                  id,
		UltimaSincronizacao: pgtype.Timestamptz{Time: ultimaSincronizacao, Valid: true},
	}
	_, err := r.querier.UpdatePdvIntegracaoUltimaSincronizacao(ctx, params)
	return err
}

func (r *PdvIntegracaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeletePdvIntegracao(ctx, id)
}

func toDomainPdvIntegracao(p PdvIntegraco) *domain.PdvIntegracao {
	var clientID string
	if p.ClientID.Valid {
		clientID = p.ClientID.String
	}
	var clientSecretEncrypted string
	if p.ClientSecretEncrypted.Valid {
		clientSecretEncrypted = p.ClientSecretEncrypted.String
	}
	var ultimaSincronizacao *time.Time
	if p.UltimaSincronizacao.Valid {
		ultimaSincronizacao = &p.UltimaSincronizacao.Time
	}
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	return &domain.PdvIntegracao{
		ID:                    p.ID,
		EmpresaID:             p.EmpresaID,
		LojaID:                p.LojaID,
		Provedor:              p.Provedor,
		Ambiente:              p.Ambiente,
		URLBase:               p.UrlBase,
		APIKeyEncrypted:       p.ApiKeyEncrypted,
		ClientID:              clientID,
		ClientSecretEncrypted: clientSecretEncrypted,
		Ativo:                 p.Ativo,
		UltimaSincronizacao:   ultimaSincronizacao,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
		DeletedAt:             deletedAt,
	}
}
