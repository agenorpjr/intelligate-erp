package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventoIngestaoRepository struct {
	querier *Queries
}

func NewEventoIngestaoRepository(pool *pgxpool.Pool) *EventoIngestaoRepository {
	return &EventoIngestaoRepository{querier: New(pool)}
}

func (r *EventoIngestaoRepository) Create(ctx context.Context, e *domain.EventoIngestao) error {
	var empresaID, lojaID, pdvOrigemID uuid.UUID
	empresaID = e.EmpresaID
	lojaID = e.LojaID
	pdvOrigemID = e.PdvOrigemID

	var payloadOriginal []byte
	if e.PayloadOriginal != "" {
		payloadOriginal = []byte(e.PayloadOriginal)
	}

	var erro pgtype.Text
	if e.Erro != nil {
		erro = pgtype.Text{String: *e.Erro, Valid: true}
	}

	params := CreateEventoIngestaoParams{
		EmpresaID:       empresaID,
		LojaID:          lojaID,
		PdvOrigemID:     pdvOrigemID,
		TipoEvento:      e.TipoEvento,
		PayloadOriginal: payloadOriginal,
		Status:          e.Status,
		Erro:            erro,
		Tentativas:      int32(e.Tentativas),
		IdempotencyKey:  e.IdempotencyKey,
	}
	_, err := r.querier.CreateEventoIngestao(ctx, params)
	return err
}

func (r *EventoIngestaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.EventoIngestao, error) {
	result, err := r.querier.GetEventoIngestao(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainEventoIngestao(result), nil
}

func (r *EventoIngestaoRepository) GetByIdempotencyKey(ctx context.Context, empresaID uuid.UUID, idempotencyKey string) (*domain.EventoIngestao, error) {
	result, err := r.querier.GetEventoIngestaoByIdempotencyKey(ctx, GetEventoIngestaoByIdempotencyKeyParams{
		EmpresaID:      empresaID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, err
	}
	return toDomainEventoIngestao(result), nil
}

func (r *EventoIngestaoRepository) List(ctx context.Context, filter domain.EventoIngestaoFilter) ([]*domain.EventoIngestao, error) {
	var empresaID uuid.UUID
	if filter.EmpresaID != nil {
		empresaID = *filter.EmpresaID
	}
	var lojaID, pdvOrigemID uuid.UUID
	if filter.LojaID != nil {
		lojaID = *filter.LojaID
	}
	if filter.PdvOrigemID != nil {
		pdvOrigemID = *filter.PdvOrigemID
	}
	var tipoEvento, status string
	if filter.TipoEvento != nil {
		tipoEvento = *filter.TipoEvento
	}
	if filter.Status != nil {
		status = *filter.Status
	}

	results, err := r.querier.ListEventosIngestao(ctx, ListEventosIngestaoParams{
		EmpresaID: empresaID,
		Column2:   lojaID,
		Column3:   pdvOrigemID,
		Column4:   tipoEvento,
		Column5:   status,
		Limit:     int32(filter.Limit),
		Offset:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	eventos := make([]*domain.EventoIngestao, len(results))
	for i, r := range results {
		eventos[i] = toDomainEventoIngestao(r)
	}
	return eventos, nil
}

func (r *EventoIngestaoRepository) Update(ctx context.Context, e *domain.EventoIngestao) error {
	var erro pgtype.Text
	if e.Erro != nil {
		erro = pgtype.Text{String: *e.Erro, Valid: true}
	}
	var processadoEm pgtype.Timestamptz
	if e.ProcessadoEm != nil {
		processadoEm = pgtype.Timestamptz{Time: *e.ProcessadoEm, Valid: true}
	}

	params := UpdateEventoIngestaoParams{
		ID:           e.ID,
		Status:       e.Status,
		Erro:         erro,
		Tentativas:   int32(e.Tentativas),
		ProcessadoEm: processadoEm,
	}
	_, err := r.querier.UpdateEventoIngestao(ctx, params)
	return err
}

func (r *EventoIngestaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteEventoIngestao(ctx, id)
}

func toDomainEventoIngestao(e EventosIngestao) *domain.EventoIngestao {
	var erro *string
	if e.Erro.Valid {
		erro = &e.Erro.String
	}
	var processadoEm *time.Time
	if e.ProcessadoEm.Valid {
		processadoEm = &e.ProcessadoEm.Time
	}
	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}
	return &domain.EventoIngestao{
		ID:              e.ID,
		EmpresaID:       e.EmpresaID,
		LojaID:          e.LojaID,
		PdvOrigemID:     e.PdvOrigemID,
		TipoEvento:      e.TipoEvento,
		PayloadOriginal: string(e.PayloadOriginal),
		Status:          e.Status,
		Erro:            erro,
		Tentativas:      int(e.Tentativas),
		IdempotencyKey:  e.IdempotencyKey,
		ProcessadoEm:    processadoEm,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}
