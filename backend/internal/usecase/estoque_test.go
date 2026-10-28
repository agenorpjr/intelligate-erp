package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intelligate/erp/backend/internal/domain"
)

// ============================================================
// Fakes
// ============================================================

type fakeEstoqueRepository struct {
	saldos        map[string]*domain.SaldoEstoque
	movimentacoes []*domain.MovimentacaoEstoque
}

func novoFakeEstoqueRepository() *fakeEstoqueRepository {
	return &fakeEstoqueRepository{saldos: make(map[string]*domain.SaldoEstoque)}
}

func chaveSaldo(insumoID, lojaID uuid.UUID) string {
	return insumoID.String() + ":" + lojaID.String()
}

func (r *fakeEstoqueRepository) semearSaldo(s *domain.SaldoEstoque) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	cp := *s
	r.saldos[chaveSaldo(s.InsumoID, s.LojaID)] = &cp
}

func (r *fakeEstoqueRepository) GetSaldo(_ context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	s, ok := r.saldos[chaveSaldo(insumoID, lojaID)]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeEstoqueRepository) ListSaldos(_ context.Context, filter domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	out := make([]*domain.SaldoEstoque, 0)
	for _, s := range r.saldos {
		if s.EmpresaID != filter.EmpresaID {
			continue
		}
		if filter.LojaID != nil && s.LojaID != *filter.LojaID {
			continue
		}
		if filter.InsumoID != nil && s.InsumoID != *filter.InsumoID {
			continue
		}
		cp := *s
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeEstoqueRepository) ListMovimentacoes(_ context.Context, filter domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	out := make([]*domain.MovimentacaoEstoque, 0)
	for _, m := range r.movimentacoes {
		if m.EmpresaID != filter.EmpresaID {
			continue
		}
		if filter.LojaID != nil && m.LojaID != *filter.LojaID {
			continue
		}
		if filter.InsumoID != nil && m.InsumoID != *filter.InsumoID {
			continue
		}
		if filter.Tipo != nil && m.Tipo != *filter.Tipo {
			continue
		}
		if filter.Origem != nil && m.Origem != *filter.Origem {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *fakeEstoqueRepository) GetResumoPerdas(_ context.Context, filter domain.ResumoPerdasFilter) (*domain.ResumoPerdas, error) {
	resumo := &domain.ResumoPerdas{
		EmpresaID:  filter.EmpresaID,
		LojaID:     filter.LojaID,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Itens:      make([]domain.ResumoPerdasItem, 0),
	}
	for _, m := range r.movimentacoes {
		if m.EmpresaID != filter.EmpresaID || m.LojaID != filter.LojaID {
			continue
		}
		if m.Tipo != domain.MovimentacaoPerda && m.Tipo != domain.MovimentacaoDesperdicio {
			continue
		}
		resumo.Itens = append(resumo.Itens, domain.ResumoPerdasItem{
			Tipo:          m.Tipo,
			InsumoID:      m.InsumoID,
			UnidadeMedida: m.UnidadeMedida,
			Quantidade:    m.Quantidade,
			CustoTotal:    m.CustoTotal,
		})
		resumo.QuantidadeTotal += m.Quantidade
		resumo.CustoTotal += m.CustoTotal
		if m.Tipo == domain.MovimentacaoPerda {
			resumo.CustoTotalPerdas += m.CustoTotal
		} else {
			resumo.CustoTotalDesperdicio += m.CustoTotal
		}
	}
	return resumo, nil
}

func (r *fakeEstoqueRepository) WithTx(_ context.Context, fn func(tx domain.EstoqueTxRepository) error) error {
	saldosSnapshot := make(map[string]*domain.SaldoEstoque, len(r.saldos))
	for k, v := range r.saldos {
		cp := *v
		saldosSnapshot[k] = &cp
	}
	movsSnapshot := make([]*domain.MovimentacaoEstoque, len(r.movimentacoes))
	copy(movsSnapshot, r.movimentacoes)

	if err := fn(&fakeEstoqueTx{repo: r}); err != nil {
		r.saldos = saldosSnapshot
		r.movimentacoes = movsSnapshot
		return err
	}
	return nil
}

type fakeEstoqueTx struct {
	repo *fakeEstoqueRepository
}

func (t *fakeEstoqueTx) GetSaldoForUpdate(ctx context.Context, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	return t.repo.GetSaldo(ctx, insumoID, lojaID)
}

func (t *fakeEstoqueTx) UpsertSaldo(_ context.Context, saldo *domain.SaldoEstoque) (*domain.SaldoEstoque, error) {
	cp := *saldo
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	t.repo.saldos[chaveSaldo(saldo.InsumoID, saldo.LojaID)] = &cp
	res := cp
	return &res, nil
}

func (t *fakeEstoqueTx) CreateMovimentacao(_ context.Context, mov *domain.MovimentacaoEstoque) (*domain.MovimentacaoEstoque, error) {
	cp := *mov
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	cp.CreatedAt = time.Now()
	t.repo.movimentacoes = append(t.repo.movimentacoes, &cp)
	res := cp
	return &res, nil
}

type mockInsumoRepository struct {
	insumos map[uuid.UUID]*domain.Insumo
}

func newMockInsumoRepository() *mockInsumoRepository {
	return &mockInsumoRepository{insumos: make(map[uuid.UUID]*domain.Insumo)}
}

func (m *mockInsumoRepository) Create(_ context.Context, i *domain.Insumo) error {
	m.insumos[i.ID] = i
	return nil
}

func (m *mockInsumoRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.Insumo, error) {
	if i, ok := m.insumos[id]; ok {
		return i, nil
	}
	return nil, errors.New("insumo não encontrado")
}

func (m *mockInsumoRepository) GetByNome(_ context.Context, _ uuid.UUID, _ string) (*domain.Insumo, error) {
	return nil, errors.New("não implementado")
}

func (m *mockInsumoRepository) ListByEmpresa(_ context.Context, _ uuid.UUID, _ *bool, _, _ int) ([]*domain.Insumo, error) {
	return nil, nil
}

func (m *mockInsumoRepository) Update(_ context.Context, _ *domain.Insumo) error { return nil }
func (m *mockInsumoRepository) Delete(_ context.Context, _ uuid.UUID) error      { return nil }

func (m *mockInsumoRepository) CreateCusto(_ context.Context, _ *domain.InsumoCusto) error {
	return nil
}

func (m *mockInsumoRepository) GetCustoVigente(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *mockInsumoRepository) GetCustoAtual(_ context.Context, _ uuid.UUID) (*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *mockInsumoRepository) ListCustosByInsumo(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.InsumoCusto, error) {
	return nil, nil
}

func (m *mockInsumoRepository) GetForUpdate(_ context.Context, id uuid.UUID) (*domain.Insumo, error) {
	return m.GetByID(context.Background(), id)
}

func (m *mockInsumoRepository) UpdateCustoUnitario(_ context.Context, _ uuid.UUID, _ float64) error {
	return nil
}

// ============================================================
// Helpers
// ============================================================

type estoqueFixture struct {
	uc          *EstoqueUsecase
	estoqueRepo *fakeEstoqueRepository
	lojaRepo    *mockLojaRepository
	insumoRepo  *mockInsumoRepository
	auditoria   *mockAuditoriaRepository
	empresaID   uuid.UUID
	lojaID      uuid.UUID
	insumoID    uuid.UUID
	usuarioID   uuid.UUID
}

func novoEstoqueFixture(t *testing.T) *estoqueFixture {
	t.Helper()
	ctx := context.Background()

	f := &estoqueFixture{
		estoqueRepo: novoFakeEstoqueRepository(),
		lojaRepo:    newMockLojaRepository(),
		insumoRepo:  newMockInsumoRepository(),
		auditoria:   newMockAuditoriaRepository(),
		empresaID:   uuid.New(),
		lojaID:      uuid.New(),
		insumoID:    uuid.New(),
		usuarioID:   uuid.New(),
	}

	require.NoError(t, f.lojaRepo.Create(ctx, &domain.Loja{
		ID:        f.lojaID,
		EmpresaID: f.empresaID,
		Nome:      "Loja Centro",
		Ativo:     true,
	}))
	require.NoError(t, f.insumoRepo.Create(ctx, &domain.Insumo{
		ID:            f.insumoID,
		EmpresaID:     f.empresaID,
		Nome:          "Farinha",
		UnidadeMedida: "kg",
		Ativo:         true,
	}))

	f.uc = NewEstoqueUsecase(f.estoqueRepo, f.lojaRepo, f.insumoRepo, f.auditoria)
	return f
}

func (f *estoqueFixture) semearSaldo(lojaID uuid.UUID, quantidade, custoMedio float64) {
	f.estoqueRepo.semearSaldo(&domain.SaldoEstoque{
		EmpresaID:          f.empresaID,
		LojaID:             lojaID,
		InsumoID:           f.insumoID,
		Quantidade:         quantidade,
		UnidadeMedida:      "kg",
		CustoMedioUnitario: custoMedio,
	})
}

// ============================================================
// Testes
// ============================================================

func TestEstoqueUsecase_RegistrarSaida_Sucesso(t *testing.T) {
	f := novoEstoqueFixture(t)
	f.semearSaldo(f.lojaID, 10, 5)

	mov, err := f.uc.RegistrarSaida(context.Background(), domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoSaidaConsumo),
		Quantidade: 3,
	}, f.usuarioID)

	require.NoError(t, err)
	require.NotNil(t, mov)
	assert.Equal(t, domain.MovimentacaoSaidaConsumo, mov.Tipo)
	assert.InDelta(t, 10, mov.SaldoAnterior, 1e-9)
	assert.InDelta(t, 7, mov.SaldoPosterior, 1e-9)
	assert.InDelta(t, 5, mov.CustoUnitario, 1e-9)
	assert.InDelta(t, 15, mov.CustoTotal, 1e-9)
	assert.Equal(t, domain.OrigemManual, mov.Origem)

	saldo, err := f.estoqueRepo.GetSaldo(context.Background(), f.insumoID, f.lojaID)
	require.NoError(t, err)
	require.NotNil(t, saldo)
	assert.InDelta(t, 7, saldo.Quantidade, 1e-9)

	require.Len(t, f.auditoria.auditorias, 1)
	assert.Equal(t, "SAIDA_ESTOQUE", f.auditoria.auditorias[0].Acao)
}

func TestEstoqueUsecase_RegistrarSaida_SaldoInsuficiente(t *testing.T) {
	f := novoEstoqueFixture(t)
	f.semearSaldo(f.lojaID, 2, 5)

	_, err := f.uc.RegistrarSaida(context.Background(), domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoPerda),
		Quantidade: 5,
	}, f.usuarioID)

	assert.ErrorIs(t, err, ErrEstoqueSaldoInsuficiente)

	// Saldo não deve sofrer alteração (rollback da transação).
	saldo, _ := f.estoqueRepo.GetSaldo(context.Background(), f.insumoID, f.lojaID)
	require.NotNil(t, saldo)
	assert.InDelta(t, 2, saldo.Quantidade, 1e-9)
	assert.Empty(t, f.estoqueRepo.movimentacoes)
}

func TestEstoqueUsecase_RegistrarSaida_TipoInvalido(t *testing.T) {
	f := novoEstoqueFixture(t)

	_, err := f.uc.RegistrarSaida(context.Background(), domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoEntradaCompra),
		Quantidade: 1,
	}, f.usuarioID)

	assert.ErrorIs(t, err, ErrEstoqueTipoInvalido)
}

func TestEstoqueUsecase_RegistrarAjuste_MotivoObrigatorio(t *testing.T) {
	f := novoEstoqueFixture(t)

	_, err := f.uc.RegistrarAjuste(context.Background(), domain.RegistrarAjusteEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoAjustePositivo),
		Quantidade: 5,
		Motivo:     "   ",
	}, f.usuarioID)

	assert.ErrorIs(t, err, ErrEstoqueMotivoObrigatorio)
	assert.Empty(t, f.estoqueRepo.movimentacoes)
}

func TestEstoqueUsecase_RegistrarAjuste_Positivo(t *testing.T) {
	f := novoEstoqueFixture(t)
	f.semearSaldo(f.lojaID, 4, 2)

	mov, err := f.uc.RegistrarAjuste(context.Background(), domain.RegistrarAjusteEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoAjustePositivo),
		Quantidade: 6,
		Motivo:     "Inventário",
	}, f.usuarioID)

	require.NoError(t, err)
	assert.InDelta(t, 10, mov.SaldoPosterior, 1e-9)
	assert.Equal(t, "AJUSTE_ESTOQUE", f.auditoria.auditorias[0].Acao)
}

func TestEstoqueUsecase_Transferir_Sucesso(t *testing.T) {
	f := novoEstoqueFixture(t)
	ctx := context.Background()

	lojaDestinoID := uuid.New()
	require.NoError(t, f.lojaRepo.Create(ctx, &domain.Loja{
		ID:        lojaDestinoID,
		EmpresaID: f.empresaID,
		Nome:      "Loja Zona Sul",
		Ativo:     true,
	}))

	f.semearSaldo(f.lojaID, 10, 5)     // origem
	f.semearSaldo(lojaDestinoID, 2, 8) // destino (custo médio próprio)

	movs, err := f.uc.Transferir(ctx, domain.TransferirEstoqueInput{
		EmpresaID:     f.empresaID,
		LojaOrigemID:  f.lojaID,
		LojaDestinoID: lojaDestinoID,
		InsumoID:      f.insumoID,
		Quantidade:    4,
	}, f.usuarioID)

	require.NoError(t, err)
	require.Len(t, movs, 2)
	assert.Equal(t, domain.MovimentacaoTransferenciaSaida, movs[0].Tipo)
	assert.Equal(t, domain.MovimentacaoTransferenciaEntrada, movs[1].Tipo)
	assert.Equal(t, f.lojaID, movs[0].LojaID)
	assert.Equal(t, lojaDestinoID, movs[1].LojaID)
	assert.InDelta(t, 5, movs[1].CustoUnitario, 1e-9) // custo herdado da origem

	origem, _ := f.estoqueRepo.GetSaldo(ctx, f.insumoID, f.lojaID)
	assert.InDelta(t, 6, origem.Quantidade, 1e-9)

	destino, _ := f.estoqueRepo.GetSaldo(ctx, f.insumoID, lojaDestinoID)
	assert.InDelta(t, 6, destino.Quantidade, 1e-9)
	// Custo médio do destino permanece inalterado.
	assert.InDelta(t, 8, destino.CustoMedioUnitario, 1e-9)
}

func TestEstoqueUsecase_Transferir_MesmaLoja(t *testing.T) {
	f := novoEstoqueFixture(t)

	_, err := f.uc.Transferir(context.Background(), domain.TransferirEstoqueInput{
		EmpresaID:     f.empresaID,
		LojaOrigemID:  f.lojaID,
		LojaDestinoID: f.lojaID,
		InsumoID:      f.insumoID,
		Quantidade:    1,
	}, f.usuarioID)

	assert.ErrorIs(t, err, ErrEstoqueMesmaLoja)
}

func TestEstoqueUsecase_Transferir_LojaOutraEmpresa(t *testing.T) {
	f := novoEstoqueFixture(t)
	ctx := context.Background()

	outraLojaID := uuid.New()
	require.NoError(t, f.lojaRepo.Create(ctx, &domain.Loja{
		ID:        outraLojaID,
		EmpresaID: uuid.New(), // outra empresa
		Ativo:     true,
	}))

	f.semearSaldo(f.lojaID, 10, 5)

	_, err := f.uc.Transferir(ctx, domain.TransferirEstoqueInput{
		EmpresaID:     f.empresaID,
		LojaOrigemID:  f.lojaID,
		LojaDestinoID: outraLojaID,
		InsumoID:      f.insumoID,
		Quantidade:    1,
	}, f.usuarioID)

	assert.ErrorIs(t, err, ErrEstoqueLojaNaoPertence)
}

func TestEstoqueUsecase_RegistrarEntradaCompra_MediaPonderada(t *testing.T) {
	f := novoEstoqueFixture(t)
	f.semearSaldo(f.lojaID, 10, 5)

	mov, err := f.uc.RegistrarEntradaCompra(context.Background(), domain.RegistrarEntradaCompraInput{
		EmpresaID:     f.empresaID,
		LojaID:        f.lojaID,
		InsumoID:      f.insumoID,
		Quantidade:    10,
		CustoUnitario: 7,
		UnidadeMedida: "kg",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.MovimentacaoEntradaCompra, mov.Tipo)
	assert.Equal(t, domain.OrigemRecebimentoCompra, mov.Origem)
	assert.InDelta(t, 20, mov.SaldoPosterior, 1e-9)

	saldo, _ := f.estoqueRepo.GetSaldo(context.Background(), f.insumoID, f.lojaID)
	// (10*5 + 10*7) / 20 = 6
	assert.InDelta(t, 6, saldo.CustoMedioUnitario, 1e-9)
}

func TestEstoqueUsecase_GetSaldo_LojaNaoPertence(t *testing.T) {
	f := novoEstoqueFixture(t)
	ctx := context.Background()

	outraLojaID := uuid.New()
	require.NoError(t, f.lojaRepo.Create(ctx, &domain.Loja{
		ID:        outraLojaID,
		EmpresaID: uuid.New(),
		Ativo:     true,
	}))

	_, err := f.uc.GetSaldo(ctx, f.empresaID, f.insumoID, outraLojaID)
	assert.ErrorIs(t, err, ErrEstoqueLojaNaoPertence)
}

func TestEstoqueUsecase_GetSaldo_NaoEncontrado(t *testing.T) {
	f := novoEstoqueFixture(t)

	_, err := f.uc.GetSaldo(context.Background(), f.empresaID, f.insumoID, f.lojaID)
	assert.ErrorIs(t, err, ErrEstoqueSaldoNaoEncontrado)
}

func TestEstoqueUsecase_ListMovimentacoes_FiltroPorTipo(t *testing.T) {
	f := novoEstoqueFixture(t)
	ctx := context.Background()
	f.semearSaldo(f.lojaID, 20, 5)

	_, err := f.uc.RegistrarSaida(ctx, domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoSaidaConsumo),
		Quantidade: 2,
	}, f.usuarioID)
	require.NoError(t, err)

	_, err = f.uc.RegistrarSaida(ctx, domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoPerda),
		Quantidade: 1,
		Motivo:     stringPtr("Quebra"),
	}, f.usuarioID)
	require.NoError(t, err)

	tipo := domain.MovimentacaoPerda
	movs, err := f.uc.ListMovimentacoes(ctx, domain.MovimentacaoEstoqueFilter{
		EmpresaID: f.empresaID,
		Tipo:      &tipo,
	})
	require.NoError(t, err)
	require.Len(t, movs, 1)
	assert.Equal(t, domain.MovimentacaoPerda, movs[0].Tipo)
}

func TestEstoqueUsecase_GetResumoPerdas(t *testing.T) {
	f := novoEstoqueFixture(t)
	ctx := context.Background()
	f.semearSaldo(f.lojaID, 20, 4)

	_, err := f.uc.RegistrarSaida(ctx, domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoPerda),
		Quantidade: 2,
	}, f.usuarioID)
	require.NoError(t, err)

	_, err = f.uc.RegistrarSaida(ctx, domain.RegistrarSaidaEstoqueInput{
		EmpresaID:  f.empresaID,
		LojaID:     f.lojaID,
		InsumoID:   f.insumoID,
		Tipo:       string(domain.MovimentacaoDesperdicio),
		Quantidade: 1,
	}, f.usuarioID)
	require.NoError(t, err)

	resumo, err := f.uc.GetResumoPerdas(ctx, f.empresaID, f.lojaID,
		time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	assert.InDelta(t, 8, resumo.CustoTotalPerdas, 1e-9)      // 2 * 4
	assert.InDelta(t, 4, resumo.CustoTotalDesperdicio, 1e-9) // 1 * 4
	assert.InDelta(t, 12, resumo.CustoTotal, 1e-9)
}
