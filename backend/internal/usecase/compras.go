package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrFornecedorNaoEncontrado        = errors.New("fornecedor não encontrado")
	ErrFornecedorJaExiste             = errors.New("fornecedor com este CNPJ/CPF já existe")
	ErrOrdemCompraNaoEncontrada       = errors.New("ordem de compra não encontrada")
	ErrOrdemCompraNaoEstaEmRascunho   = errors.New("ordem de compra não está em rascunho")
	ErrOrdemCompraNaoPodeSerCancelada = errors.New("ordem de compra não pode ser cancelada")
	ErrOrdemCompraItemNaoEncontrado   = errors.New("item da ordem de compra não encontrado")
	ErrRecebimentoCompraNaoEncontrado = errors.New("recebimento de compra não encontrado")
	ErrQuantidadeExcedePedido         = errors.New("quantidade recebida excede a quantidade pedida")
	ErrContaPagarNaoEncontrada        = errors.New("conta a pagar não encontrada")
)

// ============================================================
// Fornecedor Usecase
// ============================================================

type FornecedorUsecase struct {
	fornecedorRepo domain.FornecedorRepository
	auditoriaRepo  domain.AuditoriaRepository
}

func NewFornecedorUsecase(fornecedorRepo domain.FornecedorRepository, auditoriaRepo domain.AuditoriaRepository) *FornecedorUsecase {
	return &FornecedorUsecase{
		fornecedorRepo: fornecedorRepo,
		auditoriaRepo:  auditoriaRepo,
	}
}

func (uc *FornecedorUsecase) Create(ctx context.Context, input domain.CreateFornecedorInput, usuarioID uuid.UUID) (*domain.Fornecedor, error) {
	// Verifica se já existe fornecedor com este CNPJ/CPF na empresa
	existing, err := uc.fornecedorRepo.GetByCNPJCPF(ctx, input.EmpresaID, input.CNPJCPF)
	if err == nil && existing != nil {
		return nil, ErrFornecedorJaExiste
	}

	fornecedor := &domain.Fornecedor{
		ID:          uuid.New(),
		EmpresaID:   input.EmpresaID,
		Nome:        input.Nome,
		CNPJCPF:     input.CNPJCPF,
		Email:       input.Email,
		Telefone:    input.Telefone,
		Banco:       input.Banco,
		Agencia:     input.Agencia,
		Conta:       input.Conta,
		TipoConta:   input.TipoConta,
		PixChave:    input.PixChave,
		Observacoes: input.Observacoes,
		Ativo:       true,
	}

	if err := uc.fornecedorRepo.Create(ctx, fornecedor); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "fornecedor",
		EntidadeID: fornecedor.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(fornecedor)),
	})

	return fornecedor, nil
}

func (uc *FornecedorUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error) {
	fornecedor, err := uc.fornecedorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFornecedorNaoEncontrado
	}
	return fornecedor, nil
}

func (uc *FornecedorUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Fornecedor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.fornecedorRepo.ListByEmpresa(ctx, empresaID, ativo, limit, offset)
}

func (uc *FornecedorUsecase) ListByFilter(ctx context.Context, empresaID uuid.UUID, ativo *bool, nome, cnpjCpf *string, limit, offset int) ([]*domain.Fornecedor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.fornecedorRepo.ListByFilter(ctx, empresaID, ativo, nome, cnpjCpf, limit, offset)
}

func (uc *FornecedorUsecase) Update(ctx context.Context, id uuid.UUID, input domain.UpdateFornecedorInput, usuarioID uuid.UUID) (*domain.Fornecedor, error) {
	fornecedor, err := uc.fornecedorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFornecedorNaoEncontrado
	}

	dadosAntigos := toJSON(fornecedor)

	if input.Nome != nil {
		fornecedor.Nome = *input.Nome
	}
	if input.CNPJCPF != nil {
		// Verifica se o novo CNPJ/CPF já existe para outra empresa
		existing, err := uc.fornecedorRepo.GetByCNPJCPF(ctx, fornecedor.EmpresaID, *input.CNPJCPF)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrFornecedorJaExiste
		}
		fornecedor.CNPJCPF = *input.CNPJCPF
	}
	if input.Email != nil {
		fornecedor.Email = input.Email
	}
	if input.Telefone != nil {
		fornecedor.Telefone = input.Telefone
	}
	if input.Banco != nil {
		fornecedor.Banco = input.Banco
	}
	if input.Agencia != nil {
		fornecedor.Agencia = input.Agencia
	}
	if input.Conta != nil {
		fornecedor.Conta = input.Conta
	}
	if input.TipoConta != nil {
		fornecedor.TipoConta = input.TipoConta
	}
	if input.PixChave != nil {
		fornecedor.PixChave = input.PixChave
	}
	if input.Observacoes != nil {
		fornecedor.Observacoes = input.Observacoes
	}
	if input.Ativo != nil {
		fornecedor.Ativo = *input.Ativo
	}

	if err := uc.fornecedorRepo.Update(ctx, fornecedor); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &fornecedor.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "fornecedor",
		EntidadeID: fornecedor.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(fornecedor)),
	})

	return fornecedor, nil
}

func (uc *FornecedorUsecase) Delete(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	fornecedor, err := uc.fornecedorRepo.GetByID(ctx, id)
	if err != nil {
		return ErrFornecedorNaoEncontrado
	}

	dadosAntigos := toJSON(fornecedor)

	if err := uc.fornecedorRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &fornecedor.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "fornecedor",
		EntidadeID: id,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}

// ============================================================
// Ordem de Compra Usecase
// ============================================================

type OrdemCompraUsecase struct {
	ordemCompraRepo domain.OrdemCompraRepository
	fornecedorRepo  domain.FornecedorRepository
	insumoRepo      domain.InsumoRepository
	auditoriaRepo   domain.AuditoriaRepository
}

func NewOrdemCompraUsecase(
	ordemCompraRepo domain.OrdemCompraRepository,
	fornecedorRepo domain.FornecedorRepository,
	insumoRepo domain.InsumoRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *OrdemCompraUsecase {
	return &OrdemCompraUsecase{
		ordemCompraRepo: ordemCompraRepo,
		fornecedorRepo:  fornecedorRepo,
		insumoRepo:      insumoRepo,
		auditoriaRepo:   auditoriaRepo,
	}
}

func (uc *OrdemCompraUsecase) Create(ctx context.Context, input domain.CreateOrdemCompraInput, usuarioID uuid.UUID) (*domain.OrdemCompra, error) {
	// Verifica se o fornecedor existe e pertence à empresa
	fornecedor, err := uc.fornecedorRepo.GetByID(ctx, input.FornecedorID)
	if err != nil {
		return nil, ErrFornecedorNaoEncontrado
	}
	if fornecedor.EmpresaID != input.EmpresaID {
		return nil, errors.New("fornecedor não pertence a esta empresa")
	}

	// Calcula valor total
	var valorTotal float64
	for _, itemInput := range input.Itens {
		valorTotal += itemInput.Quantidade * itemInput.PrecoUnitario
	}

	ordemCompra := &domain.OrdemCompra{
		ID:                  uuid.New(),
		EmpresaID:           input.EmpresaID,
		LojaID:              input.LojaID,
		FornecedorID:        input.FornecedorID,
		DataCompra:          input.DataCompra,
		DataPrevistaEntrega: input.DataPrevistaEntrega,
		Status:              domain.OrdemCompraStatusRascunho,
		Observacoes:         input.Observacoes,
		ValorTotal:          valorTotal,
	}

	if err := uc.ordemCompraRepo.Create(ctx, ordemCompra); err != nil {
		return nil, err
	}

	// Cria itens
	for _, itemInput := range input.Itens {
		item := &domain.OrdemCompraItem{
			ID:                 uuid.New(),
			OrdemCompraID:      ordemCompra.ID,
			InsumoID:           itemInput.InsumoID,
			Quantidade:         itemInput.Quantidade,
			UnidadeMedida:      itemInput.UnidadeMedida,
			PrecoUnitario:      itemInput.PrecoUnitario,
			QuantidadeRecebida: 0,
		}
		if err := uc.ordemCompraRepo.CreateItem(ctx, item); err != nil {
			return nil, err
		}
	}

	// Carrega itens para retorno
	ordemCompraItens, err := uc.ordemCompraRepo.ListItensByOrdemCompra(ctx, ordemCompra.ID)
	if err != nil {
		return nil, err
	}
	ordemCompra.Itens = ordemCompraItens

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra",
		EntidadeID: ordemCompra.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(ordemCompra)),
	})

	return ordemCompra, nil
}

func (uc *OrdemCompraUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrdemCompra, error) {
	ordemCompra, err := uc.ordemCompraRepo.GetWithItens(ctx, id)
	if err != nil {
		return nil, ErrOrdemCompraNaoEncontrada
	}
	return ordemCompra, nil
}

func (uc *OrdemCompraUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.OrdemCompra, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.ordemCompraRepo.ListByEmpresa(ctx, empresaID, limit, offset)
}

func (uc *OrdemCompraUsecase) ListByLoja(ctx context.Context, lojaID uuid.UUID, limit, offset int) ([]*domain.OrdemCompra, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.ordemCompraRepo.ListByLoja(ctx, lojaID, limit, offset)
}

func (uc *OrdemCompraUsecase) ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *domain.OrdemCompraStatus, dataInicio, dataFim *time.Time, limit, offset int) ([]*domain.OrdemCompra, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.ordemCompraRepo.ListByFilter(ctx, empresaID, lojaID, fornecedorID, status, dataInicio, dataFim, limit, offset)
}

func (uc *OrdemCompraUsecase) Update(ctx context.Context, id uuid.UUID, input domain.UpdateOrdemCompraInput, usuarioID uuid.UUID) (*domain.OrdemCompra, error) {
	ordemCompra, err := uc.ordemCompraRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status != domain.OrdemCompraStatusRascunho {
		return nil, ErrOrdemCompraNaoEstaEmRascunho
	}

	dadosAntigos := toJSON(ordemCompra)

	if input.LojaID != nil {
		ordemCompra.LojaID = *input.LojaID
	}
	if input.FornecedorID != nil {
		// Verifica se o fornecedor pertence à empresa
		fornecedor, err := uc.fornecedorRepo.GetByID(ctx, *input.FornecedorID)
		if err != nil {
			return nil, ErrFornecedorNaoEncontrado
		}
		if fornecedor.EmpresaID != ordemCompra.EmpresaID {
			return nil, errors.New("fornecedor não pertence a esta empresa")
		}
		ordemCompra.FornecedorID = *input.FornecedorID
	}
	if input.DataCompra != nil {
		ordemCompra.DataCompra = *input.DataCompra
	}
	if input.DataPrevistaEntrega != nil {
		ordemCompra.DataPrevistaEntrega = input.DataPrevistaEntrega
	}
	if input.Status != nil {
		ordemCompra.Status = domain.OrdemCompraStatus(*input.Status)
	}
	if input.Observacoes != nil {
		ordemCompra.Observacoes = input.Observacoes
	}

	if err := uc.ordemCompraRepo.Update(ctx, ordemCompra); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra",
		EntidadeID: ordemCompra.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(ordemCompra)),
	})

	return ordemCompra, nil
}

func (uc *OrdemCompraUsecase) Cancel(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	ordemCompra, err := uc.ordemCompraRepo.GetByID(ctx, id)
	if err != nil {
		return ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status == domain.OrdemCompraStatusCancelada {
		return ErrOrdemCompraNaoPodeSerCancelada
	}
	if ordemCompra.Status == domain.OrdemCompraStatusRecebida {
		return errors.New("não é possível cancelar uma ordem já recebida")
	}

	if err := uc.ordemCompraRepo.UpdateStatus(ctx, id, domain.OrdemCompraStatusCancelada); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra",
		EntidadeID: id,
		Acao:       "CANCEL",
		Depois:     stringPtr(toJSON(ordemCompra)),
	})

	return nil
}

func (uc *OrdemCompraUsecase) AddItem(ctx context.Context, ordemCompraID uuid.UUID, input domain.CreateOrdemCompraItemInput, usuarioID uuid.UUID) (*domain.OrdemCompraItem, error) {
	ordemCompra, err := uc.ordemCompraRepo.GetByID(ctx, ordemCompraID)
	if err != nil {
		return nil, ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status != domain.OrdemCompraStatusRascunho {
		return nil, ErrOrdemCompraNaoEstaEmRascunho
	}

	item := &domain.OrdemCompraItem{
		ID:                 uuid.New(),
		OrdemCompraID:      ordemCompraID,
		InsumoID:           input.InsumoID,
		Quantidade:         input.Quantidade,
		UnidadeMedida:      input.UnidadeMedida,
		PrecoUnitario:      input.PrecoUnitario,
		QuantidadeRecebida: 0,
	}

	if err := uc.ordemCompraRepo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	// Recalcula valor total
	itens, err := uc.ordemCompraRepo.ListItensByOrdemCompra(ctx, ordemCompraID)
	if err != nil {
		return nil, err
	}
	var valorTotal float64
	for _, it := range itens {
		valorTotal += it.Quantidade * it.PrecoUnitario
	}
	ordemCompra.ValorTotal = valorTotal
	if err := uc.ordemCompraRepo.Update(ctx, ordemCompra); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra_item",
		EntidadeID: item.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(item)),
	})

	return item, nil
}

func (uc *OrdemCompraUsecase) UpdateItem(ctx context.Context, itemID uuid.UUID, input domain.CreateOrdemCompraItemInput, usuarioID uuid.UUID) (*domain.OrdemCompraItem, error) {
	item, err := uc.ordemCompraRepo.GetItem(ctx, itemID)
	if err != nil {
		return nil, ErrOrdemCompraItemNaoEncontrado
	}

	ordemCompra, err := uc.ordemCompraRepo.GetByID(ctx, item.OrdemCompraID)
	if err != nil {
		return nil, ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status != domain.OrdemCompraStatusRascunho {
		return nil, ErrOrdemCompraNaoEstaEmRascunho
	}

	dadosAntigos := toJSON(item)

	item.InsumoID = input.InsumoID
	item.Quantidade = input.Quantidade
	item.UnidadeMedida = input.UnidadeMedida
	item.PrecoUnitario = input.PrecoUnitario

	if err := uc.ordemCompraRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	// Recalcula valor total
	itens, err := uc.ordemCompraRepo.ListItensByOrdemCompra(ctx, ordemCompra.ID)
	if err != nil {
		return nil, err
	}
	var valorTotal float64
	for _, it := range itens {
		valorTotal += it.Quantidade * it.PrecoUnitario
	}
	ordemCompra.ValorTotal = valorTotal
	if err := uc.ordemCompraRepo.Update(ctx, ordemCompra); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra_item",
		EntidadeID: itemID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(item)),
	})

	return item, nil
}

func (uc *OrdemCompraUsecase) RemoveItem(ctx context.Context, itemID uuid.UUID, usuarioID uuid.UUID) error {
	item, err := uc.ordemCompraRepo.GetItem(ctx, itemID)
	if err != nil {
		return ErrOrdemCompraItemNaoEncontrado
	}

	ordemCompra, err := uc.ordemCompraRepo.GetByID(ctx, item.OrdemCompraID)
	if err != nil {
		return ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status != domain.OrdemCompraStatusRascunho {
		return ErrOrdemCompraNaoEstaEmRascunho
	}

	dadosAntigos := toJSON(item)

	if err := uc.ordemCompraRepo.DeleteItem(ctx, itemID); err != nil {
		return err
	}

	// Recalcula valor total
	itens, err := uc.ordemCompraRepo.ListItensByOrdemCompra(ctx, ordemCompra.ID)
	if err != nil {
		return err
	}
	var valorTotal float64
	for _, it := range itens {
		valorTotal += it.Quantidade * it.PrecoUnitario
	}
	ordemCompra.ValorTotal = valorTotal
	if err := uc.ordemCompraRepo.Update(ctx, ordemCompra); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "ordem_compra_item",
		EntidadeID: itemID,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}

// ============================================================
// Recebimento de Compra Usecase
// ============================================================

type RecebimentoCompraUsecase struct {
	recebimentoRepo domain.RecebimentoCompraRepository
	ordemCompraRepo domain.OrdemCompraRepository
	fornecedorRepo  domain.FornecedorRepository
	insumoRepo      domain.InsumoRepository
	contaPagarRepo  domain.ContaPagarRepository
	auditoriaRepo   domain.AuditoriaRepository
}

func NewRecebimentoCompraUsecase(
	recebimentoRepo domain.RecebimentoCompraRepository,
	ordemCompraRepo domain.OrdemCompraRepository,
	fornecedorRepo domain.FornecedorRepository,
	insumoRepo domain.InsumoRepository,
	contaPagarRepo domain.ContaPagarRepository,
	auditoriaRepo domain.AuditoriaRepository,
) *RecebimentoCompraUsecase {
	return &RecebimentoCompraUsecase{
		recebimentoRepo: recebimentoRepo,
		ordemCompraRepo: ordemCompraRepo,
		fornecedorRepo:  fornecedorRepo,
		insumoRepo:      insumoRepo,
		contaPagarRepo:  contaPagarRepo,
		auditoriaRepo:   auditoriaRepo,
	}
}

func (uc *RecebimentoCompraUsecase) Create(ctx context.Context, input domain.CreateRecebimentoCompraInput, usuarioID uuid.UUID) (*domain.RecebimentoCompra, error) {
	ordemCompra, err := uc.ordemCompraRepo.GetWithItens(ctx, input.OrdemCompraID)
	if err != nil {
		return nil, ErrOrdemCompraNaoEncontrada
	}

	if ordemCompra.Status == domain.OrdemCompraStatusCancelada {
		return nil, errors.New("não é possível receber uma ordem cancelada")
	}

	// Valida quantidades recebidas
	for _, itemInput := range input.Itens {
		// Encontra o item da ordem de compra
		var ordemItem *domain.OrdemCompraItem
		for _, oi := range ordemCompra.Itens {
			if oi.ID == itemInput.OrdemCompraItemID {
				ordemItem = oi
				break
			}
		}
		if ordemItem == nil {
			return nil, errors.New("item da ordem de compra não encontrado")
		}

		// Verifica se não excede a quantidade pedida (a menos que permitir_excedente)
		if !input.PermitirExcedente {
			novaQuantidadeRecebida := ordemItem.QuantidadeRecebida + itemInput.QuantidadeRecebida
			if novaQuantidadeRecebida > ordemItem.Quantidade {
				return nil, ErrQuantidadeExcedePedido
			}
		}
	}

	// Cria o recebimento
	dataRecebimento := time.Now()
	if input.DataRecebimento != nil {
		dataRecebimento = *input.DataRecebimento
	}

	recebimento := &domain.RecebimentoCompra{
		ID:                uuid.New(),
		OrdemCompraID:     input.OrdemCompraID,
		LojaID:            input.LojaID,
		DataRecebimento:   dataRecebimento,
		Observacoes:       input.Observacoes,
		PermitirExcedente: input.PermitirExcedente,
	}

	if err := uc.recebimentoRepo.Create(ctx, recebimento); err != nil {
		return nil, err
	}

	// Cria itens do recebimento, atualiza quantidade recebida e custo do insumo
	for _, itemInput := range input.Itens {
		recebimentoItem := &domain.RecebimentoCompraItem{
			ID:                  uuid.New(),
			RecebimentoCompraID: recebimento.ID,
			OrdemCompraItemID:   itemInput.OrdemCompraItemID,
			InsumoID:            itemInput.InsumoID,
			QuantidadeRecebida:  itemInput.QuantidadeRecebida,
			PrecoUnitario:       itemInput.PrecoUnitario,
		}
		if err := uc.recebimentoRepo.CreateItem(ctx, recebimentoItem); err != nil {
			return nil, err
		}

		// Atualiza quantidade recebida no item da ordem de compra
		var ordemItem *domain.OrdemCompraItem
		for _, oi := range ordemCompra.Itens {
			if oi.ID == itemInput.OrdemCompraItemID {
				ordemItem = oi
				break
			}
		}
		if ordemItem != nil {
			novaQuantidadeRecebida := ordemItem.QuantidadeRecebida + itemInput.QuantidadeRecebida
			if err := uc.ordemCompraRepo.UpdateItemQuantidadeRecebida(ctx, ordemItem.ID, novaQuantidadeRecebida); err != nil {
				return nil, err
			}
		}

		// Atualiza custo do insumo (média ponderada ou último preço)
		insumo, err := uc.insumoRepo.GetForUpdate(ctx, itemInput.InsumoID)
		if err != nil {
			return nil, err
		}

		// Calcula novo custo unitário (média ponderada simples com o preço de compra)
		// Usa a quantidade já recebida ANTES deste recebimento
		qtdJaRecebida := 0.0
		if ordemItem != nil {
			qtdJaRecebida = ordemItem.QuantidadeRecebida
		}

		novoCusto := itemInput.PrecoUnitario
		if insumo.CustoUnitarioAtual > 0 && qtdJaRecebida > 0 {
			// Média ponderada: (custo_atual * qtd_anterior + novo_preco * nova_qtd) / qtd_total
			novoCusto = (insumo.CustoUnitarioAtual*qtdJaRecebida + itemInput.PrecoUnitario*itemInput.QuantidadeRecebida) / (qtdJaRecebida + itemInput.QuantidadeRecebida)
		}

		if err := uc.insumoRepo.UpdateCustoUnitario(ctx, itemInput.InsumoID, novoCusto); err != nil {
			return nil, err
		}

		// Cria entrada no histórico de custos do insumo
		custo := &domain.InsumoCusto{
			ID:            uuid.New(),
			InsumoID:      itemInput.InsumoID,
			DataVigencia:  dataRecebimento,
			CustoUnitario: novoCusto,
			Observacao:    stringPtr(fmt.Sprintf("Recebimento da ordem de compra %s", ordemCompra.ID)),
		}
		if err := uc.insumoRepo.CreateCusto(ctx, custo); err != nil {
			return nil, err
		}
	}

	// Atualiza status da ordem de compra
	totalRecebido := true
	for _, oi := range ordemCompra.Itens {
		if oi.QuantidadeRecebida < oi.Quantidade {
			totalRecebido = false
			break
		}
	}

	novoStatus := domain.OrdemCompraStatusRecebida
	if !totalRecebido {
		novoStatus = domain.OrdemCompraStatusRecebidaParcial
	}

	if ordemCompra.Status != novoStatus {
		if err := uc.ordemCompraRepo.UpdateStatus(ctx, ordemCompra.ID, novoStatus); err != nil {
			return nil, err
		}
	}

	// Carrega itens do recebimento para retorno
	recebimentoItens, err := uc.recebimentoRepo.ListItensByRecebimento(ctx, recebimento.ID)
	if err != nil {
		return nil, err
	}
	recebimento.Itens = recebimentoItens

	// Cria conta a pagar automaticamente
	valorRecebimento := 0.0
	for _, ri := range recebimentoItens {
		valorRecebimento += ri.QuantidadeRecebida * ri.PrecoUnitario
	}

	contaPagar := &domain.ContaPagar{
		ID:                  uuid.New(),
		EmpresaID:           ordemCompra.EmpresaID,
		LojaID:              input.LojaID,
		FornecedorID:        ordemCompra.FornecedorID,
		OrdemCompraID:       &ordemCompra.ID,
		RecebimentoCompraID: &recebimento.ID,
		Descricao:           fmt.Sprintf("Compra de %s - Ordem %s", fornecedorNome(ordemCompra.FornecedorID, uc.fornecedorRepo, ctx), ordemCompra.ID),
		Valor:               valorRecebimento,
		DataEmissao:         dataRecebimento,
		DataVencimento:      dataRecebimento.AddDate(0, 0, 30), // 30 dias por padrão
		Status:              domain.ContaPagarStatusPendente,
	}
	if err := uc.contaPagarRepo.Create(ctx, contaPagar); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &ordemCompra.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "recebimento_compra",
		EntidadeID: recebimento.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(recebimento)),
	})

	return recebimento, nil
}

func fornecedorNome(id uuid.UUID, repo domain.FornecedorRepository, ctx context.Context) string {
	f, err := repo.GetByID(ctx, id)
	if err != nil {
		return id.String()
	}
	return f.Nome
}

func (uc *RecebimentoCompraUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.RecebimentoCompra, error) {
	recebimento, err := uc.recebimentoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRecebimentoCompraNaoEncontrado
	}
	return recebimento, nil
}

func (uc *RecebimentoCompraUsecase) ListByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) ([]*domain.RecebimentoCompra, error) {
	return uc.recebimentoRepo.ListByOrdemCompra(ctx, ordemCompraID)
}

func (uc *RecebimentoCompraUsecase) ListByFilter(ctx context.Context, ordemCompraID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time, limit, offset int) ([]*domain.RecebimentoCompra, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.recebimentoRepo.ListByFilter(ctx, ordemCompraID, lojaID, dataInicio, dataFim, limit, offset)
}

// ============================================================
// Conta a Pagar Usecase
// ============================================================

type ContaPagarUsecase struct {
	contaPagarRepo domain.ContaPagarRepository
	auditoriaRepo  domain.AuditoriaRepository
}

func NewContaPagarUsecase(contaPagarRepo domain.ContaPagarRepository, auditoriaRepo domain.AuditoriaRepository) *ContaPagarUsecase {
	return &ContaPagarUsecase{
		contaPagarRepo: contaPagarRepo,
		auditoriaRepo:  auditoriaRepo,
	}
}

func (uc *ContaPagarUsecase) Create(ctx context.Context, input domain.CreateContaPagarInput, usuarioID uuid.UUID) (*domain.ContaPagar, error) {
	status := domain.ContaPagarStatusPendente
	if input.Status != nil {
		status = domain.ContaPagarStatus(*input.Status)
	}

	contaPagar := &domain.ContaPagar{
		ID:                  uuid.New(),
		EmpresaID:           input.EmpresaID,
		LojaID:              input.LojaID,
		FornecedorID:        input.FornecedorID,
		OrdemCompraID:       input.OrdemCompraID,
		RecebimentoCompraID: input.RecebimentoCompraID,
		Descricao:           input.Descricao,
		Valor:               input.Valor,
		DataEmissao:         input.DataEmissao,
		DataVencimento:      input.DataVencimento,
		Status:              status,
		FormaPagamento:      input.FormaPagamento,
		Observacoes:         input.Observacoes,
	}

	if err := uc.contaPagarRepo.Create(ctx, contaPagar); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &input.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "conta_pagar",
		EntidadeID: contaPagar.ID,
		Acao:       "CREATE",
		Depois:     stringPtr(toJSON(contaPagar)),
	})

	return contaPagar, nil
}

func (uc *ContaPagarUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContaPagar, error) {
	contaPagar, err := uc.contaPagarRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrContaPagarNaoEncontrada
	}
	return contaPagar, nil
}

func (uc *ContaPagarUsecase) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.ContaPagar, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.contaPagarRepo.ListByEmpresa(ctx, empresaID, limit, offset)
}

func (uc *ContaPagarUsecase) ListByFornecedor(ctx context.Context, fornecedorID uuid.UUID, limit, offset int) ([]*domain.ContaPagar, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.contaPagarRepo.ListByFornecedor(ctx, fornecedorID, limit, offset)
}

func (uc *ContaPagarUsecase) ListByStatus(ctx context.Context, empresaID uuid.UUID, status domain.ContaPagarStatus, limit, offset int) ([]*domain.ContaPagar, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.contaPagarRepo.ListByStatus(ctx, empresaID, status, limit, offset)
}

func (uc *ContaPagarUsecase) ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *domain.ContaPagarStatus, dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim *time.Time, categoriaFinanceiraID, contaGerencialID *uuid.UUID, limit, offset int) ([]*domain.ContaPagar, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.contaPagarRepo.ListByFilter(ctx, empresaID, lojaID, fornecedorID, status, dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim, categoriaFinanceiraID, contaGerencialID, limit, offset)
}

func (uc *ContaPagarUsecase) Update(ctx context.Context, id uuid.UUID, input domain.UpdateContaPagarInput, usuarioID uuid.UUID) (*domain.ContaPagar, error) {
	contaPagar, err := uc.contaPagarRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrContaPagarNaoEncontrada
	}

	dadosAntigos := toJSON(contaPagar)

	if input.Descricao != nil {
		contaPagar.Descricao = *input.Descricao
	}
	if input.Valor != nil {
		contaPagar.Valor = *input.Valor
	}
	if input.DataEmissao != nil {
		contaPagar.DataEmissao = *input.DataEmissao
	}
	if input.DataVencimento != nil {
		contaPagar.DataVencimento = *input.DataVencimento
	}
	if input.DataPagamento != nil {
		contaPagar.DataPagamento = input.DataPagamento
	}
	if input.Status != nil {
		contaPagar.Status = domain.ContaPagarStatus(*input.Status)
	}
	if input.FormaPagamento != nil {
		contaPagar.FormaPagamento = input.FormaPagamento
	}
	if input.Observacoes != nil {
		contaPagar.Observacoes = input.Observacoes
	}

	if err := uc.contaPagarRepo.Update(ctx, contaPagar); err != nil {
		return nil, err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &contaPagar.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "conta_pagar",
		EntidadeID: contaPagar.ID,
		Acao:       "UPDATE",
		Antes:      stringPtr(dadosAntigos),
		Depois:     stringPtr(toJSON(contaPagar)),
	})

	return contaPagar, nil
}

func (uc *ContaPagarUsecase) Pagar(ctx context.Context, id uuid.UUID, input domain.PagarContaInput, usuarioID uuid.UUID) (*domain.ContaPagar, error) {
	contaPagar, err := uc.contaPagarRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrContaPagarNaoEncontrada
	}

	if contaPagar.Status == domain.ContaPagarStatusPago {
		return nil, errors.New("conta já está paga")
	}

	now := time.Now()
	if err := uc.contaPagarRepo.UpdateStatus(ctx, id, domain.ContaPagarStatusPago, &now, &input.FormaPagamento); err != nil {
		return nil, err
	}

	contaPagar.Status = domain.ContaPagarStatusPago
	contaPagar.DataPagamento = &now
	contaPagar.FormaPagamento = &input.FormaPagamento

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &contaPagar.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "conta_pagar",
		EntidadeID: contaPagar.ID,
		Acao:       "PAGAR",
		Depois:     stringPtr(toJSON(contaPagar)),
	})

	return contaPagar, nil
}

func (uc *ContaPagarUsecase) Delete(ctx context.Context, id uuid.UUID, usuarioID uuid.UUID) error {
	contaPagar, err := uc.contaPagarRepo.GetByID(ctx, id)
	if err != nil {
		return ErrContaPagarNaoEncontrada
	}

	dadosAntigos := toJSON(contaPagar)

	if err := uc.contaPagarRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit
	uc.auditoriaRepo.Create(ctx, &domain.Auditoria{
		EmpresaID:  &contaPagar.EmpresaID,
		UsuarioID:  &usuarioID,
		Entidade:   "conta_pagar",
		EntidadeID: id,
		Acao:       "DELETE",
		Antes:      stringPtr(dadosAntigos),
	})

	return nil
}

// Helper para converter float64 para pgtype.Numeric
func floatToNumeric(v float64) pgtype.Numeric {
	intVal := big.NewInt(int64(v * 10000))
	return pgtype.Numeric{Int: intVal, Exp: -4, Valid: true}
}

// Helper para converter pgtype.Numeric para float64
func numericToFloat(n pgtype.Numeric) float64 {
	if n.Valid && n.Int != nil {
		return float64(n.Int.Int64()) / 10000
	}
	return 0
}
