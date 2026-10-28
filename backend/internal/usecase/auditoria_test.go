package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
)

func TestAuditoriaUsecase_Listar_ForcaEmpresaDoContexto(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()
	outraEmpresa := uuid.New()

	repo := newMockAuditoriaRepository()
	repo.auditorias = append(repo.auditorias, &domain.Auditoria{ID: 1, Entidade: "papel", Acao: "criar"})
	uc := NewAuditoriaUsecase(repo)

	// Cliente tenta forçar outra empresa; deve ser sobrescrita.
	filter := domain.AuditoriaFilter{EmpresaID: &outraEmpresa, Limit: 10, Offset: 0}

	items, err := uc.Listar(ctx, empresaID, filter)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	require.NotNil(t, repo.lastFilter.EmpresaID)
	assert.Equal(t, empresaID, *repo.lastFilter.EmpresaID)
}

func TestAuditoriaUsecase_Listar_AplicaDefaultsDePaginacao(t *testing.T) {
	ctx := context.Background()
	empresaID := uuid.New()

	repo := newMockAuditoriaRepository()
	uc := NewAuditoriaUsecase(repo)

	// Limit inválido (0) e offset negativo devem ser normalizados.
	_, err := uc.Listar(ctx, empresaID, domain.AuditoriaFilter{Limit: 0, Offset: -5})
	require.NoError(t, err)
	assert.Equal(t, 20, repo.lastFilter.Limit)
	assert.Equal(t, 0, repo.lastFilter.Offset)

	// Limit acima do máximo deve ser limitado a 20.
	_, err = uc.Listar(ctx, empresaID, domain.AuditoriaFilter{Limit: 500, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 20, repo.lastFilter.Limit)
}
