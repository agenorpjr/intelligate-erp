package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditoriaRepository struct {
	querier *Queries
}

func NewAuditoriaRepository(pool *pgxpool.Pool) *AuditoriaRepository {
	return &AuditoriaRepository{querier: New(pool)}
}

func (r *AuditoriaRepository) Create(ctx context.Context, a *domain.Auditoria) error {
	var antes, depois []byte
	if a.Antes != nil {
		antes = []byte(*a.Antes)
	}
	if a.Depois != nil {
		depois = []byte(*a.Depois)
	}
	var empresaID, usuarioID pgtype.UUID
	if a.EmpresaID != nil {
		empresaID = pgtype.UUID{Bytes: *a.EmpresaID, Valid: true}
	}
	if a.UsuarioID != nil {
		usuarioID = pgtype.UUID{Bytes: *a.UsuarioID, Valid: true}
	}
	var ip, userAgent pgtype.Text
	if a.IP != nil {
		ip = pgtype.Text{String: *a.IP, Valid: true}
	}
	if a.UserAgent != nil {
		userAgent = pgtype.Text{String: *a.UserAgent, Valid: true}
	}

	params := CreateAuditoriaParams{
		EmpresaID:  empresaID,
		UsuarioID:  usuarioID,
		Entidade:   a.Entidade,
		EntidadeID: a.EntidadeID,
		Acao:       a.Acao,
		Antes:      antes,
		Depois:     depois,
		Ip:         ip,
		UserAgent:  userAgent,
	}
	_, err := r.querier.CreateAuditoria(ctx, params)
	return err
}

func (r *AuditoriaRepository) List(ctx context.Context, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	var empresaID, usuarioID uuid.UUID
	if filter.EmpresaID != nil {
		empresaID = *filter.EmpresaID
	}
	if filter.UsuarioID != nil {
		usuarioID = *filter.UsuarioID
	}
	var entidade, acao string
	if filter.Entidade != nil {
		entidade = *filter.Entidade
	}
	if filter.Acao != nil {
		acao = *filter.Acao
	}

	results, err := r.querier.ListAuditoria(ctx, ListAuditoriaParams{
		Column1: empresaID,
		Column2: usuarioID,
		Column3: entidade,
		Column4: acao,
		Limit:   int32(filter.Limit),
		Offset:  int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	auditorias := make([]*domain.Auditoria, len(results))
	for i, r := range results {
		var antes, depois *string
		if len(r.Antes) > 0 {
			s := string(r.Antes)
			antes = &s
		}
		if len(r.Depois) > 0 {
			s := string(r.Depois)
			depois = &s
		}
		var empresaID, usuarioID *uuid.UUID
		if r.EmpresaID.Valid {
			var u uuid.UUID
			copy(u[:], r.EmpresaID.Bytes[:])
			empresaID = &u
		}
		if r.UsuarioID.Valid {
			var u uuid.UUID
			copy(u[:], r.UsuarioID.Bytes[:])
			usuarioID = &u
		}
		var ip, userAgent *string
		if r.Ip.Valid {
			ip = &r.Ip.String
		}
		if r.UserAgent.Valid {
			userAgent = &r.UserAgent.String
		}

		auditorias[i] = &domain.Auditoria{
			ID:         r.ID,
			EmpresaID:  empresaID,
			UsuarioID:  usuarioID,
			Entidade:   r.Entidade,
			EntidadeID: r.EntidadeID,
			Acao:       r.Acao,
			Antes:      antes,
			Depois:     depois,
			IP:         ip,
			UserAgent:  userAgent,
			CreatedAt:  r.CreatedAt,
		}
	}
	return auditorias, nil
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
