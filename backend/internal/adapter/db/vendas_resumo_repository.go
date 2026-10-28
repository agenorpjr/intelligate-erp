package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VendasResumoRepository struct {
	querier *Queries
}

func NewVendasResumoRepository(pool *pgxpool.Pool) *VendasResumoRepository {
	return &VendasResumoRepository{querier: New(pool)}
}

func (r *VendasResumoRepository) GetResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time) (*domain.VendasResumo, error) {
	var lojaIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	var dataInicioParam, dataFimParam time.Time
	if dataInicio != nil {
		dataInicioParam = *dataInicio
	}
	if dataFim != nil {
		dataFimParam = *dataFim
	}

	result, err := r.querier.GetVendasResumo(ctx, GetVendasResumoParams{
		EmpresaID: empresaID,
		Column2:   lojaIDParam,
		Column3:   dataInicioParam,
		Column4:   dataFimParam,
	})
	if err != nil {
		return nil, err
	}

	return &domain.VendasResumo{
		TotalVendas:    float64FromInterface(result.TotalVendas),
		TotalPedidos:   result.TotalPedidos,
		TicketMedio:    float64FromInterface(result.TicketMedio),
		TotalDescontos: float64FromInterface(result.TotalDescontos),
	}, nil
}
