package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
)

type mockLojaRepository struct {
	lojas map[uuid.UUID]*domain.Loja
}

func newMockLojaRepository() *mockLojaRepository {
	return &mockLojaRepository{
		lojas: make(map[uuid.UUID]*domain.Loja),
	}
}

func (m *mockLojaRepository) Create(ctx context.Context, l *domain.Loja) error {
	m.lojas[l.ID] = l
	return nil
}

func (m *mockLojaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Loja, error) {
	if l, ok := m.lojas[id]; ok {
		return l, nil
	}
	return nil, ErrLojaNaoEncontrada
}

func (m *mockLojaRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.Loja, error) {
	var result []*domain.Loja
	for _, l := range m.lojas {
		if l.EmpresaID == empresaID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLojaRepository) Update(ctx context.Context, l *domain.Loja) error {
	m.lojas[l.ID] = l
	return nil
}

func (m *mockLojaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.lojas, id)
	return nil
}

type mockEmpresaRepository struct {
	empresas map[uuid.UUID]*domain.Empresa
}

func newMockEmpresaRepository() *mockEmpresaRepository {
	return &mockEmpresaRepository{
		empresas: make(map[uuid.UUID]*domain.Empresa),
	}
}

func (m *mockEmpresaRepository) Create(ctx context.Context, e *domain.Empresa) error {
	m.empresas[e.ID] = e
	return nil
}

func (m *mockEmpresaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Empresa, error) {
	if e, ok := m.empresas[id]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockEmpresaRepository) GetByCNPJ(ctx context.Context, cnpj string) (*domain.Empresa, error) {
	for _, e := range m.empresas {
		if e.CNPJ == cnpj {
			return e, nil
		}
	}
	return nil, nil
}

func (m *mockEmpresaRepository) List(ctx context.Context, limit, offset int) ([]*domain.Empresa, error) {
	var result []*domain.Empresa
	for _, e := range m.empresas {
		result = append(result, e)
	}
	return result, nil
}

func (m *mockEmpresaRepository) Update(ctx context.Context, e *domain.Empresa) error {
	m.empresas[e.ID] = e
	return nil
}

func (m *mockEmpresaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.empresas, id)
	return nil
}

type mockAuditoriaRepository struct {
	auditorias []*domain.Auditoria
	lastFilter domain.AuditoriaFilter
}

func newMockAuditoriaRepository() *mockAuditoriaRepository {
	return &mockAuditoriaRepository{
		auditorias: make([]*domain.Auditoria, 0),
	}
}

func (m *mockAuditoriaRepository) Create(ctx context.Context, a *domain.Auditoria) error {
	m.auditorias = append(m.auditorias, a)
	return nil
}

func (m *mockAuditoriaRepository) List(ctx context.Context, filter domain.AuditoriaFilter) ([]*domain.Auditoria, error) {
	m.lastFilter = filter
	return m.auditorias, nil
}

func TestLojaUsecase_Criar_Success(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	empresaID := uuid.New()
	empresa := &domain.Empresa{
		ID:          empresaID,
		CNPJ:        "12.345.678/0001-90",
		RazaoSocial: "Empresa Teste",
		Ativo:       true,
	}
	empresaRepo.Create(context.Background(), empresa)

	input := domain.CreateLojaInput{
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		Endereco:    stringPtr("Rua Teste, 123"),
		FusoHorario: "America/Sao_Paulo",
	}

	loja, err := uc.Criar(context.Background(), input, uuid.New(), "127.0.0.1", "test-agent")

	require.NoError(t, err)
	assert.NotNil(t, loja)
	assert.Equal(t, "LOJA001", loja.Codigo)
	assert.Equal(t, "Loja Teste", loja.Nome)
	assert.Equal(t, empresaID, loja.EmpresaID)
	assert.True(t, loja.Ativo)
	assert.Len(t, auditoriaRepo.auditorias, 1)
	assert.Equal(t, "create", auditoriaRepo.auditorias[0].Acao)
	assert.Equal(t, "loja", auditoriaRepo.auditorias[0].Entidade)
}

func TestLojaUsecase_Criar_DuplicateCode(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	empresaID := uuid.New()
	empresa := &domain.Empresa{
		ID:          empresaID,
		CNPJ:        "12.345.678/0001-90",
		RazaoSocial: "Empresa Teste",
		Ativo:       true,
	}
	empresaRepo.Create(context.Background(), empresa)

	// Create first loja
	input1 := domain.CreateLojaInput{
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja 1",
		FusoHorario: "America/Sao_Paulo",
	}
	_, err := uc.Criar(context.Background(), input1, uuid.New(), "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Try to create second loja with same code
	input2 := domain.CreateLojaInput{
		EmpresaID:   empresaID,
		Codigo:      "LOJA001",
		Nome:        "Loja 2",
		FusoHorario: "America/Sao_Paulo",
	}
	_, err = uc.Criar(context.Background(), input2, uuid.New(), "127.0.0.1", "test-agent")
	assert.ErrorIs(t, err, ErrCodigoDuplicado)
}

func TestLojaUsecase_Listar(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	empresaID := uuid.New()
	empresa := &domain.Empresa{
		ID:          empresaID,
		CNPJ:        "12.345.678/0001-90",
		RazaoSocial: "Empresa Teste",
		Ativo:       true,
	}
	empresaRepo.Create(context.Background(), empresa)

	// Create multiple lojas
	for i := 1; i <= 3; i++ {
		loja := &domain.Loja{
			ID:          uuid.New(),
			EmpresaID:   empresaID,
			Codigo:      "LOJA" + string(rune('0'+i)),
			Nome:        "Loja " + string(rune('0'+i)),
			FusoHorario: "America/Sao_Paulo",
			Ativo:       true,
		}
		lojaRepo.Create(context.Background(), loja)
	}

	lojas, err := uc.Listar(context.Background(), empresaID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, lojas, 3)
}

func TestLojaUsecase_BuscarPorID(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	loja := &domain.Loja{
		ID:          uuid.New(),
		EmpresaID:   uuid.New(),
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	found, err := uc.BuscarPorID(context.Background(), loja.ID)
	require.NoError(t, err)
	assert.Equal(t, loja.ID, found.ID)
	assert.Equal(t, loja.Codigo, found.Codigo)

	// Test not found
	_, err = uc.BuscarPorID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, ErrLojaNaoEncontrada)
}

func TestLojaUsecase_Atualizar(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	loja := &domain.Loja{
		ID:          uuid.New(),
		EmpresaID:   uuid.New(),
		Codigo:      "LOJA001",
		Nome:        "Loja Original",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	input := domain.UpdateLojaInput{
		Nome:  stringPtr("Loja Atualizada"),
		Ativo: boolPtr(false),
	}

	updated, err := uc.Atualizar(context.Background(), loja.ID, input, uuid.New(), "127.0.0.1", "test-agent")
	require.NoError(t, err)
	assert.Equal(t, "Loja Atualizada", updated.Nome)
	assert.False(t, updated.Ativo)
	assert.Len(t, auditoriaRepo.auditorias, 1)
	assert.Equal(t, "update", auditoriaRepo.auditorias[0].Acao)
}

func TestLojaUsecase_Deletar(t *testing.T) {
	lojaRepo := newMockLojaRepository()
	empresaRepo := newMockEmpresaRepository()
	auditoriaRepo := newMockAuditoriaRepository()

	uc := NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)

	loja := &domain.Loja{
		ID:          uuid.New(),
		EmpresaID:   uuid.New(),
		Codigo:      "LOJA001",
		Nome:        "Loja Teste",
		FusoHorario: "America/Sao_Paulo",
		Ativo:       true,
	}
	lojaRepo.Create(context.Background(), loja)

	err := uc.Deletar(context.Background(), loja.ID, uuid.New(), "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Verify deleted
	_, err = lojaRepo.GetByID(context.Background(), loja.ID)
	assert.Error(t, err) // should return error (not found)

	assert.Len(t, auditoriaRepo.auditorias, 1)
	assert.Equal(t, "delete", auditoriaRepo.auditorias[0].Acao)
}
