package db

import (
	"context"
	"math/big"
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

	totalVendas := 0.0
	if result.TotalVendas != nil {
		switch v := result.TotalVendas.(type) {
		case *big.Int:
			if v != nil {
				totalVendas = float64(v.Int64()) / 10000
			}
		case float64:
			totalVendas = v
		case int64:
			totalVendas = float64(v) / 10000
		}
	}

	ticketMedio := 0.0
	if result.TicketMedio != 0 {
		ticketMedio = float64(result.TicketMedio) / 10000
	}

	totalDescontos := 0.0
	if result.TotalDescontos != nil {
		switch v := result.TotalDescontos.(type) {
		case *big.Int:
			if v != nil {
				totalDescontos = float64(v.Int64()) / 10000
			}
		case float64:
			totalDescontos = v
		case int64:
			totalDescontos = float64(v) / 10000
		}
	}

	return &domain.VendasResumo{
		TotalVendas:    totalVendas,
		TotalPedidos:   result.TotalPedidos,
		TicketMedio:    ticketMedio,
		TotalDescontos: totalDescontos,
	}, nil
}
