package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// AuditoriaUsecase expõe a leitura dos logs de auditoria para fins de
// governança de acesso (Etapa 14).
type AuditoriaUsecase struct {
	auditoriaRepo domain.AuditoriaRepository
}

func NewAuditoriaUsecase(auditoriaRepo domain.AuditoriaRepository) *AuditoriaUsecase {
	return &AuditoriaUsecase{auditoriaRepo: auditoriaRepo}
}

// Listar retorna os logs de auditoria da empresa informada.
//
// Isolamento por empresa: o filtro de empresa é sempre forçado a partir do
// contexto autenticado, impedindo que um usuário leia auditoria de outro tenant.
func (uc *AuditoriaUsecase) Listar(ctx context.Context, empresaID uuid.UUID, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	// Isolamento multi-tenant: sobrescreve qualquer empresa vinda do cliente.
	filter.EmpresaID = &empresaID

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return uc.auditoriaRepo.List(ctx, filter)
}
