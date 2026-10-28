package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
)

// Erros esperados do estoque.
var (
	ErrEstoqueSaldoInsuficiente   = errors.New("saldo insuficiente para a movimentação")
	ErrEstoqueTipoInvalido        = errors.New("tipo de movimentação inválido")
	ErrEstoqueMotivoObrigatorio   = errors.New("motivo é obrigatório")
	ErrEstoqueQuantidadeInvalida  = errors.New("quantidade deve ser maior que zero")
	ErrEstoqueLojaNaoPertence     = errors.New("loja não pertence à empresa ou está inativa")
	ErrEstoqueInsumoNaoEncontrado = errors.New("insumo não encontrado")
	ErrEstoqueMesmaLoja           = errors.New("loja de origem e destino devem ser diferentes")
	ErrEstoqueSaldoNaoEncontrado  = errors.New("saldo de estoque não encontrado")
	ErrEstoquePeriodoInvalido     = errors.New("período inválido")
	ErrEstoqueInsumoObrigatorio   = errors.New("insumo_id é obrigatório")
	ErrEstoqueLojaObrigatoria     = errors.New("loja_id é obrigatório")
)

// EstoqueUsecase orquestra saldos e movimentações de estoque.
type EstoqueUsecase struct {
	estoqueRepo   domain.EstoqueRepository
	lojaRepo      domain.LojaRepository
	insumoRepo    domain.InsumoRepository
	auditoriaRepo domain.AuditoriaRepository
}

// NewEstoqueUsecase cria uma nova instância de EstoqueUsecase.
func NewEstoqueUsecase(
	estoqueRepo domain.EstoqueRepository,
	lojaRepo domain.LojaRepository,
	insumoRepo domain.InsumoRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *EstoqueUsecase {
	return &EstoqueUsecase{
		estoqueRepo:   estoqueRepo,
		lojaRepo:      lojaRepo,
		insumoRepo:    insumoRepo,
		auditoriaRepo: auditoriaRepo,
	}
}

// ============================================================
// Saídas manuais
// ============================================================

// RegistrarSaida registra saída por consumo, perda, desperdício ou ajuste negativo.
func (uc *EstoqueUsecase) RegistrarSaida(ctx context.Context, input domain.RegistrarSaidaEstoqueInput, usuarioID uuid.UUID) (*domain.MovimentacaoEstoque, error) {
	tipo := domain.TipoMovimentacaoEstoque(input.Tipo)
	switch tipo {
	case domain.MovimentacaoSaidaConsumo, domain.MovimentacaoPerda, domain.MovimentacaoDesperdicio, domain.MovimentacaoAjusteNegativo:
	default:
		return nil, fmt.Errorf("%w: %s", ErrEstoqueTipoInvalido, input.Tipo)
	}

	if err := validarQuantidade(input.Quantidade); err != nil {
		return nil, err
	}

	// Ajuste negativo exige justificativa explícita.
	if tipo == domain.MovimentacaoAjusteNegativo && !motivoValido(input.Motivo) {
		return nil, ErrEstoqueMotivoObrigatorio
	}

	insumo, err := uc.preparar(ctx, input.EmpresaID, input.LojaID, input.InsumoID)
	if err != nil {
		return nil, err
	}

	var mov *domain.MovimentacaoEstoque
	err = uc.estoqueRepo.WithTx(ctx, func(tx domain.EstoqueTxRepository) error {
		m, err := aplicarMovimentacao(ctx, tx, movimentacaoParams{
			EmpresaID:           input.EmpresaID,
			LojaID:              input.LojaID,
			InsumoID:            input.InsumoID,
			UnidadeMedida:       insumo.UnidadeMedida,
			Tipo:                tipo,
			Quantidade:          input.Quantidade,
			UsarCustoMedioAtual: true,
			Motivo:              input.Motivo,
			Observacao:          input.Observacao,
			Origem:              domain.OrigemManual,
			ReferenciaExterna:   input.ReferenciaExterna,
			UsuarioID:           &usuarioID,
		})
		if err != nil {
			return err
		}
		mov = m
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.registrarAuditoria(ctx, input.EmpresaID, usuarioID, "SAIDA_ESTOQUE", mov)
	return mov, nil
}

// ============================================================
// Ajustes
// ============================================================

// RegistrarAjuste registra um ajuste positivo ou negativo (motivo obrigatório).
func (uc *EstoqueUsecase) RegistrarAjuste(ctx context.Context, input domain.RegistrarAjusteEstoqueInput, usuarioID uuid.UUID) (*domain.MovimentacaoEstoque, error) {
	tipo := domain.TipoMovimentacaoEstoque(input.Tipo)
	switch tipo {
	case domain.MovimentacaoAjustePositivo, domain.MovimentacaoAjusteNegativo:
	default:
		return nil, fmt.Errorf("%w: %s", ErrEstoqueTipoInvalido, input.Tipo)
	}

	if !motivoValido(&input.Motivo) {
		return nil, ErrEstoqueMotivoObrigatorio
	}

	if err := validarQuantidade(input.Quantidade); err != nil {
		return nil, err
	}

	insumo, err := uc.preparar(ctx, input.EmpresaID, input.LojaID, input.InsumoID)
	if err != nil {
		return nil, err
	}

	var mov *domain.MovimentacaoEstoque
	err = uc.estoqueRepo.WithTx(ctx, func(tx domain.EstoqueTxRepository) error {
		m, err := aplicarMovimentacao(ctx, tx, movimentacaoParams{
			EmpresaID:           input.EmpresaID,
			LojaID:              input.LojaID,
			InsumoID:            input.InsumoID,
			UnidadeMedida:       insumo.UnidadeMedida,
			Tipo:                tipo,
			Quantidade:          input.Quantidade,
			UsarCustoMedioAtual: true,
			Motivo:              &input.Motivo,
			Observacao:          input.Observacao,
			Origem:              domain.OrigemManual,
			UsuarioID:           &usuarioID,
		})
		if err != nil {
			return err
		}
		mov = m
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.registrarAuditoria(ctx, input.EmpresaID, usuarioID, "AJUSTE_ESTOQUE", mov)
	return mov, nil
}

// ============================================================
// Transferência entre lojas
// ============================================================

// Transferir transfere insumo entre lojas da mesma empresa, gerando duas
// movimentações (saída na origem e entrada no destino) na mesma transação.
func (uc *EstoqueUsecase) Transferir(ctx context.Context, input domain.TransferirEstoqueInput, usuarioID uuid.UUID) ([]*domain.MovimentacaoEstoque, error) {
	if input.LojaOrigemID == uuid.Nil || input.LojaDestinoID == uuid.Nil {
		return nil, ErrEstoqueLojaObrigatoria
	}
	if input.LojaOrigemID == input.LojaDestinoID {
		return nil, ErrEstoqueMesmaLoja
	}
	if err := validarQuantidade(input.Quantidade); err != nil {
		return nil, err
	}

	if err := uc.validarLoja(ctx, input.EmpresaID, input.LojaOrigemID); err != nil {
		return nil, err
	}
	if err := uc.validarLoja(ctx, input.EmpresaID, input.LojaDestinoID); err != nil {
		return nil, err
	}
	insumo, err := uc.validarInsumo(ctx, input.EmpresaID, input.InsumoID)
	if err != nil {
		return nil, err
	}

	movimentacoes := make([]*domain.MovimentacaoEstoque, 0, 2)
	err = uc.estoqueRepo.WithTx(ctx, func(tx domain.EstoqueTxRepository) error {
		saida, err := aplicarMovimentacao(ctx, tx, movimentacaoParams{
			EmpresaID:           input.EmpresaID,
			LojaID:              input.LojaOrigemID,
			InsumoID:            input.InsumoID,
			UnidadeMedida:       insumo.UnidadeMedida,
			Tipo:                domain.MovimentacaoTransferenciaSaida,
			Quantidade:          input.Quantidade,
			UsarCustoMedioAtual: true,
			Motivo:              input.Motivo,
			Observacao:          input.Observacao,
			Origem:              domain.OrigemTransferencia,
			UsuarioID:           &usuarioID,
		})
		if err != nil {
			return err
		}

		// Mantém o custo médio da loja de destino: registra a entrada ao custo da origem.
		entrada, err := aplicarMovimentacao(ctx, tx, movimentacaoParams{
			EmpresaID:     input.EmpresaID,
			LojaID:        input.LojaDestinoID,
			InsumoID:      input.InsumoID,
			UnidadeMedida: insumo.UnidadeMedida,
			Tipo:          domain.MovimentacaoTransferenciaEntrada,
			Quantidade:    input.Quantidade,
			CustoUnitario: saida.CustoUnitario,
			Motivo:        input.Motivo,
			Observacao:    input.Observacao,
			Origem:        domain.OrigemTransferencia,
			UsuarioID:     &usuarioID,
		})
		if err != nil {
			return err
		}

		movimentacoes = append(movimentacoes, saida, entrada)
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "movimentacao_estoque",
		EntidadeID: movimentacoes[0].ID,
		Acao:       "TRANSFERENCIA_ESTOQUE",
		Depois:     stringPtr(toJSON(movimentacoes)),
	})

	return movimentacoes, nil
}

// ============================================================
// Entrada por recebimento de compra
// ============================================================

// RegistrarEntradaCompra cria a entrada de estoque correspondente a um
// recebimento de compra, atualizando saldo e custo médio ponderado.
func (uc *EstoqueUsecase) RegistrarEntradaCompra(ctx context.Context, input domain.RegistrarEntradaCompraInput) (*domain.MovimentacaoEstoque, error) {
	if input.Quantidade <= 0 {
		return nil, ErrEstoqueQuantidadeInvalida
	}

	insumo, err := uc.validarInsumo(ctx, input.EmpresaID, input.InsumoID)
	if err != nil {
		return nil, err
	}

	unidade := input.UnidadeMedida
	if unidade == "" {
		unidade = insumo.UnidadeMedida
	}

	var mov *domain.MovimentacaoEstoque
	err = uc.estoqueRepo.WithTx(ctx, func(tx domain.EstoqueTxRepository) error {
		m, err := aplicarMovimentacao(ctx, tx, movimentacaoParams{
			EmpresaID:            input.EmpresaID,
			LojaID:               input.LojaID,
			InsumoID:             input.InsumoID,
			UnidadeMedida:        unidade,
			Tipo:                 domain.MovimentacaoEntradaCompra,
			Quantidade:           input.Quantidade,
			CustoUnitario:        input.CustoUnitario,
			RecalcularCustoMedio: true,
			Observacao:           input.Observacao,
			Origem:               domain.OrigemRecebimentoCompra,
			ReferenciaExterna:    input.ReferenciaExterna,
			UsuarioID:            input.UsuarioID,
		})
		if err != nil {
			return err
		}
		mov = m
		return nil
	})
	if err != nil {
		return nil, err
	}

	return mov, nil
}

// ============================================================
// Consultas
// ============================================================

// ListSaldos lista os saldos da empresa, opcionalmente filtrados por loja/insumo.
func (uc *EstoqueUsecase) ListSaldos(ctx context.Context, filter domain.SaldoEstoqueFilter) ([]*domain.SaldoEstoque, error) {
	if filter.EmpresaID == uuid.Nil {
		return nil, fmt.Errorf("empresa_id é obrigatório")
	}
	if filter.LojaID != nil {
		if err := uc.validarLoja(ctx, filter.EmpresaID, *filter.LojaID); err != nil {
			return nil, err
		}
	}
	filter.Limit, filter.Offset = normalizePagination(filter.Limit, filter.Offset)
	return uc.estoqueRepo.ListSaldos(ctx, filter)
}

// GetSaldo retorna o saldo de um insumo em uma loja.
func (uc *EstoqueUsecase) GetSaldo(ctx context.Context, empresaID, insumoID, lojaID uuid.UUID) (*domain.SaldoEstoque, error) {
	if insumoID == uuid.Nil {
		return nil, ErrEstoqueInsumoObrigatorio
	}
	if lojaID == uuid.Nil {
		return nil, ErrEstoqueLojaObrigatoria
	}
	if err := uc.validarLoja(ctx, empresaID, lojaID); err != nil {
		return nil, err
	}
	if _, err := uc.validarInsumo(ctx, empresaID, insumoID); err != nil {
		return nil, err
	}

	saldo, err := uc.estoqueRepo.GetSaldo(ctx, insumoID, lojaID)
	if err != nil {
		return nil, err
	}
	if saldo == nil {
		return nil, ErrEstoqueSaldoNaoEncontrado
	}
	return saldo, nil
}

// ListMovimentacoes lista as movimentações com filtros.
func (uc *EstoqueUsecase) ListMovimentacoes(ctx context.Context, filter domain.MovimentacaoEstoqueFilter) ([]*domain.MovimentacaoEstoque, error) {
	if filter.EmpresaID == uuid.Nil {
		return nil, fmt.Errorf("empresa_id é obrigatório")
	}
	if filter.LojaID != nil {
		if err := uc.validarLoja(ctx, filter.EmpresaID, *filter.LojaID); err != nil {
			return nil, err
		}
	}
	if filter.Tipo != nil && !domain.IsTipoMovimentacaoValido(*filter.Tipo) {
		return nil, fmt.Errorf("%w: %s", ErrEstoqueTipoInvalido, *filter.Tipo)
	}
	filter.Limit, filter.Offset = normalizePagination(filter.Limit, filter.Offset)
	return uc.estoqueRepo.ListMovimentacoes(ctx, filter)
}

// GetResumoPerdas consolida perdas e desperdícios da loja no período.
func (uc *EstoqueUsecase) GetResumoPerdas(ctx context.Context, empresaID, lojaID uuid.UUID, dataInicio, dataFim time.Time) (*domain.ResumoPerdas, error) {
	if lojaID == uuid.Nil {
		return nil, ErrEstoqueLojaObrigatoria
	}
	if dataInicio.IsZero() || dataFim.IsZero() {
		return nil, fmt.Errorf("%w: data_inicio e data_fim são obrigatórios", ErrEstoquePeriodoInvalido)
	}
	if dataFim.Before(dataInicio) {
		return nil, fmt.Errorf("%w: data_inicio deve ser anterior a data_fim", ErrEstoquePeriodoInvalido)
	}
	if err := uc.validarLoja(ctx, empresaID, lojaID); err != nil {
		return nil, err
	}

	return uc.estoqueRepo.GetResumoPerdas(ctx, domain.ResumoPerdasFilter{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})
}

// ============================================================
// Helpers internos
// ============================================================

type movimentacaoParams struct {
	EmpresaID            uuid.UUID
	LojaID               uuid.UUID
	InsumoID             uuid.UUID
	UnidadeMedida        string
	Tipo                 domain.TipoMovimentacaoEstoque
	Quantidade           float64
	CustoUnitario        float64
	UsarCustoMedioAtual  bool
	RecalcularCustoMedio bool
	Motivo               *string
	Observacao           *string
	Origem               string
	ReferenciaExterna    *string
	UsuarioID            *uuid.UUID
}

// aplicarMovimentacao lê o saldo (com bloqueio), valida o saldo resultante,
// atualiza o saldo e cria a movimentação — tudo dentro da transação recebida.
func aplicarMovimentacao(ctx context.Context, tx domain.EstoqueTxRepository, p movimentacaoParams) (*domain.MovimentacaoEstoque, error) {
	saldo, err := tx.GetSaldoForUpdate(ctx, p.InsumoID, p.LojaID)
	if err != nil {
		return nil, err
	}

	saldoAnterior := 0.0
	custoMedioAtual := 0.0
	permiteNegativo := false
	if saldo != nil {
		saldoAnterior = saldo.Quantidade
		custoMedioAtual = saldo.CustoMedioUnitario
		permiteNegativo = saldo.PermiteNegativo
	} else {
		saldo = &domain.SaldoEstoque{
			EmpresaID:       p.EmpresaID,
			LojaID:          p.LojaID,
			InsumoID:        p.InsumoID,
			UnidadeMedida:   p.UnidadeMedida,
			PermiteNegativo: false,
		}
	}
	if saldo.UnidadeMedida == "" {
		saldo.UnidadeMedida = p.UnidadeMedida
	}

	sinal := sinalMovimentacao(p.Tipo)
	saldoPosterior := saldoAnterior + sinal*p.Quantidade
	if saldoPosterior < 0 && !permiteNegativo {
		return nil, ErrEstoqueSaldoInsuficiente
	}

	custoUnitario := p.CustoUnitario
	if p.UsarCustoMedioAtual {
		custoUnitario = custoMedioAtual
	}

	novoCustoMedio := custoMedioAtual
	if p.RecalcularCustoMedio {
		total := saldoAnterior + p.Quantidade
		if total > 0 {
			novoCustoMedio = (saldoAnterior*custoMedioAtual + p.Quantidade*custoUnitario) / total
		}
	}

	saldo.Quantidade = saldoPosterior
	saldo.CustoMedioUnitario = novoCustoMedio

	if _, err := tx.UpsertSaldo(ctx, saldo); err != nil {
		return nil, err
	}

	mov := &domain.MovimentacaoEstoque{
		EmpresaID:         p.EmpresaID,
		LojaID:            p.LojaID,
		InsumoID:          p.InsumoID,
		Tipo:              p.Tipo,
		Quantidade:        p.Quantidade,
		UnidadeMedida:     saldo.UnidadeMedida,
		CustoUnitario:     custoUnitario,
		CustoTotal:        custoUnitario * p.Quantidade,
		SaldoAnterior:     saldoAnterior,
		SaldoPosterior:    saldoPosterior,
		Motivo:            p.Motivo,
		Observacao:        p.Observacao,
		Origem:            p.Origem,
		ReferenciaExterna: p.ReferenciaExterna,
		UsuarioID:         p.UsuarioID,
	}

	return tx.CreateMovimentacao(ctx, mov)
}

// sinalMovimentacao indica se o tipo aumenta (+1) ou reduz (-1) o saldo.
func sinalMovimentacao(tipo domain.TipoMovimentacaoEstoque) float64 {
	switch tipo {
	case domain.MovimentacaoEntradaCompra, domain.MovimentacaoAjustePositivo, domain.MovimentacaoTransferenciaEntrada:
		return 1
	default:
		return -1
	}
}

func (uc *EstoqueUsecase) preparar(ctx context.Context, empresaID, lojaID, insumoID uuid.UUID) (*domain.Insumo, error) {
	if lojaID == uuid.Nil {
		return nil, ErrEstoqueLojaObrigatoria
	}
	if insumoID == uuid.Nil {
		return nil, ErrEstoqueInsumoObrigatorio
	}
	if err := uc.validarLoja(ctx, empresaID, lojaID); err != nil {
		return nil, err
	}
	return uc.validarInsumo(ctx, empresaID, insumoID)
}

func (uc *EstoqueUsecase) validarLoja(ctx context.Context, empresaID, lojaID uuid.UUID) error {
	loja, err := uc.lojaRepo.GetByID(ctx, lojaID)
	if err != nil || loja == nil || loja.EmpresaID != empresaID || !loja.Ativo {
		return ErrEstoqueLojaNaoPertence
	}
	return nil
}

func (uc *EstoqueUsecase) validarInsumo(ctx context.Context, empresaID, insumoID uuid.UUID) (*domain.Insumo, error) {
	insumo, err := uc.insumoRepo.GetByID(ctx, insumoID)
	if err != nil || insumo == nil || insumo.EmpresaID != empresaID {
		return nil, ErrEstoqueInsumoNaoEncontrado
	}
	return insumo, nil
}

func (uc *EstoqueUsecase) registrarAuditoria(ctx context.Context, empresaID, usuarioID uuid.UUID, acao string, mov *domain.MovimentacaoEstoque) {
	if mov == nil {
		return
	}
	_ = uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &empresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "movimentacao_estoque",
		EntidadeID: mov.ID,
		Acao:       acao,
		Depois:     stringPtr(toJSON(mov)),
	})
}

func validarQuantidade(q float64) error {
	if q <= 0 {
		return ErrEstoqueQuantidadeInvalida
	}
	return nil
}

func motivoValido(motivo *string) bool {
	return motivo != nil && strings.TrimSpace(*motivo) != ""
}

func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
