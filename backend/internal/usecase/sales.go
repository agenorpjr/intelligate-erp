package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/adapter/db"
	"github.com/intelligate/erp/backend/internal/domain"
)

var (
	ErrEventoJaProcessado      = errors.New("evento já processado")
	ErrPdvOrigemNaoEncontrado  = errors.New("PDV de origem não encontrado")
	ErrProdutoNaoEncontrado    = errors.New("produto canônico não encontrado")
	ErrPedidoNaoEncontrado     = errors.New("pedido não encontrado")
	ErrIdempotencyKeyDuplicada = errors.New("chave de idempotência já existe")
)

type SalesUsecase struct {
	pdvOrigemRepo    domain.PdvOrigemRepository
	produtoRepo      domain.ProdutoCanonicoRepository
	pedidoRepo       domain.PedidoCanonicoRepository
	itemRepo         domain.PedidoItemCanonicoRepository
	pagamentoRepo    domain.PedidoPagamentoCanonicoRepository
	eventoRepo       domain.EventoIngestaoRepository
	lojaRepo         domain.LojaRepository
	vendasResumoRepo domain.VendasResumoRepository
	salesTxRepo      db.SalesTxRepositoryInterface
}

func NewSalesUsecase(
	pdvOrigemRepo domain.PdvOrigemRepository,
	produtoRepo domain.ProdutoCanonicoRepository,
	pedidoRepo domain.PedidoCanonicoRepository,
	itemRepo domain.PedidoItemCanonicoRepository,
	pagamentoRepo domain.PedidoPagamentoCanonicoRepository,
	eventoRepo domain.EventoIngestaoRepository,
	lojaRepo domain.LojaRepository,
	vendasResumoRepo domain.VendasResumoRepository,
	salesTxRepo db.SalesTxRepositoryInterface,
) *SalesUsecase {
	return &SalesUsecase{
		pdvOrigemRepo:    pdvOrigemRepo,
		produtoRepo:      produtoRepo,
		pedidoRepo:       pedidoRepo,
		itemRepo:         itemRepo,
		pagamentoRepo:    pagamentoRepo,
		eventoRepo:       eventoRepo,
		lojaRepo:         lojaRepo,
		vendasResumoRepo: vendasResumoRepo,
		salesTxRepo:      salesTxRepo,
	}
}

type IngestSalesEventInput struct {
	EmpresaID      uuid.UUID
	LojaID         uuid.UUID
	PdvOrigemID    uuid.UUID
	TipoEvento     string
	Payload        map[string]interface{}
	IdempotencyKey string
}

type IngestSalesEventOutput struct {
	EventoID     uuid.UUID
	PedidoID     uuid.UUID
	JaProcessado bool
}

func (uc *SalesUsecase) IngestSalesEvent(ctx context.Context, input IngestSalesEventInput) (*IngestSalesEventOutput, error) {
	// Check if event already processed (outside transaction for early return)
	existingEvento, err := uc.eventoRepo.GetByIdempotencyKey(ctx, input.EmpresaID, input.IdempotencyKey)
	if err == nil && existingEvento != nil {
		if existingEvento.Status == "processado" {
			// Also need to get the associated pedido to return its ID
			existingPedido, _ := uc.pedidoRepo.GetByIdempotencyKey(ctx, input.EmpresaID, input.IdempotencyKey)
			var pedidoID uuid.UUID
			if existingPedido != nil {
				pedidoID = existingPedido.ID
			}
			return &IngestSalesEventOutput{
				EventoID:     existingEvento.ID,
				PedidoID:     pedidoID,
				JaProcessado: true,
			}, nil
		}
	}

	// Validate PDV origem
	pdvOrigem, err := uc.pdvOrigemRepo.GetByID(ctx, input.PdvOrigemID)
	if err != nil || pdvOrigem == nil {
		return nil, ErrPdvOrigemNaoEncontrado
	}

	// Validate loja
	loja, err := uc.lojaRepo.GetByID(ctx, input.LojaID)
	if err != nil || loja == nil {
		return nil, errors.New("loja não encontrada")
	}

	// Prepare evento_ingestao
	evento := &domain.EventoIngestao{
		EmpresaID:       input.EmpresaID,
		LojaID:          input.LojaID,
		PdvOrigemID:     input.PdvOrigemID,
		TipoEvento:      input.TipoEvento,
		PayloadOriginal: toJSON(input.Payload),
		Status:          "processando",
		IdempotencyKey:  input.IdempotencyKey,
		Tentativas:      1,
	}

	if existingEvento != nil {
		evento.ID = existingEvento.ID
		evento.Tentativas = existingEvento.Tentativas + 1
	} else {
		evento.ID = uuid.New()
	}

	// Process the sales event payload
	payload := input.Payload

	numeroExterno := getString(payload, "numero_externo")
	dataHoraStr := getString(payload, "data_hora")
	dataHora, _ := time.Parse(time.RFC3339, dataHoraStr)
	status := getString(payload, "status")
	if status == "" {
		status = "pendente"
	}
	valorTotal := getFloat64(payload, "valor_total")
	idempotencyKey := input.IdempotencyKey
	clienteNome := getStringPtr(payload, "cliente_nome")
	observacoes := getStringPtr(payload, "observacoes")

	// Check if pedido already exists by idempotency key
	existingPedido, err := uc.pedidoRepo.GetByIdempotencyKey(ctx, input.EmpresaID, idempotencyKey)
	var pedido *domain.PedidoCanonico
	var isNewPedido bool

	if err == nil && existingPedido != nil {
		pedido = existingPedido
		isNewPedido = false
	} else {
		pedido = &domain.PedidoCanonico{
			ID:               uuid.New(),
			EmpresaID:        input.EmpresaID,
			LojaID:           input.LojaID,
			PdvOrigemID:      input.PdvOrigemID,
			NumeroExterno:    numeroExterno,
			DataHora:         dataHora,
			Status:           status,
			ValorTotal:       valorTotal,
			ValorDesconto:    getFloat64(payload, "valor_desconto"),
			ValorAcrescimo:   getFloat64(payload, "valor_acrescimo"),
			IdempotencyKey:   idempotencyKey,
			ClienteNome:      clienteNome,
			ClienteDocumento: getStringPtr(payload, "cliente_documento"),
			ClienteEmail:     getStringPtr(payload, "cliente_email"),
			Observacoes:      observacoes,
		}
		isNewPedido = true
	}

	// Process items
	itensPayload := getArray(payload, "itens")
	var itens []*domain.PedidoItemCanonico

	for _, itemPayload := range itensPayload {
		itemMap := itemPayload.(map[string]interface{})

		produtoCanonicoID := getUUIDPtr(itemMap, "produto_canonico_id")
		skuExterno := getStringPtr(itemMap, "sku_externo")
		nome := getString(itemMap, "nome")
		quantidade := getFloat64(itemMap, "quantidade")
		precoUnitario := getFloat64(itemMap, "preco_unitario")
		valorTotalItem := getFloat64(itemMap, "valor_total")
		desconto := getFloat64(itemMap, "desconto")
		modificadores := getStringPtr(itemMap, "modificadores")
		observacoesItem := getStringPtr(itemMap, "observacoes")

		item := &domain.PedidoItemCanonico{
			ID:                uuid.New(),
			PedidoID:          uuid.Nil, // será preenchido depois
			ProdutoCanonicoID: produtoCanonicoID,
			SKUExterno:        skuExterno,
			Nome:              nome,
			Quantidade:        quantidade,
			PrecoUnitario:     precoUnitario,
			ValorTotal:        valorTotalItem,
			Desconto:          desconto,
			Modificadores:     modificadores,
			Observacoes:       observacoesItem,
		}
		itens = append(itens, item)
	}

	// Process payments
	pagamentosPayload := getArray(payload, "pagamentos")
	var pagamentos []*domain.PedidoPagamentoCanonico
	valorPagamentos := 0.0

	for _, pagPayload := range pagamentosPayload {
		pagMap := pagPayload.(map[string]interface{})

		formaPagamento := getString(pagMap, "forma_pagamento")
		valor := getFloat64(pagMap, "valor")
		taxa := getFloat64(pagMap, "taxa")
		troco := getFloat64(pagMap, "troco")
		referenciaExterna := getStringPtr(pagMap, "referencia_externa")
		statusPag := getString(pagMap, "status")
		if statusPag == "" {
			statusPag = "aprovado"
		}

		pagamento := &domain.PedidoPagamentoCanonico{
			ID:                uuid.New(),
			PedidoID:          uuid.Nil, // será preenchido depois
			FormaPagamento:    formaPagamento,
			Valor:             valor,
			Taxa:              taxa,
			Troco:             troco,
			ReferenciaExterna: referenciaExterna,
			Status:            statusPag,
		}
		pagamentos = append(pagamentos, pagamento)
		valorPagamentos += valor
	}

	// Validate totals
	if valorPagamentos < valorTotal {
		return nil, errors.New("valor dos pagamentos menor que valor total do pedido")
	}

	// Execute all writes within a transaction
	err = uc.salesTxRepo.WithTx(ctx, func(txQueries *db.SalesTxQueries) error {
		// Save pedido
		if isNewPedido {
			if err := txQueries.PedidoRepo.Create(ctx, pedido); err != nil {
				return err
			}
		} else {
			pedido.ValorTotal = valorTotal
			pedido.ValorDesconto = getFloat64(payload, "valor_desconto")
			pedido.ValorAcrescimo = getFloat64(payload, "valor_acrescimo")
			pedido.Status = status
			pedido.UpdatedAt = time.Now()
			if err := txQueries.PedidoRepo.Update(ctx, pedido); err != nil {
				return err
			}
		}

		// Save items
		if isNewPedido {
			for _, item := range itens {
				item.PedidoID = pedido.ID
				if err := txQueries.ItemRepo.Create(ctx, item); err != nil {
					return err
				}
			}
		} else {
			// Delete existing items and recreate
			if err := txQueries.ItemRepo.DeleteByPedido(ctx, pedido.ID); err != nil {
				return err
			}
			for _, item := range itens {
				item.PedidoID = pedido.ID
				if err := txQueries.ItemRepo.Create(ctx, item); err != nil {
					return err
				}
			}
		}

		// Save payments
		if isNewPedido {
			for _, pag := range pagamentos {
				pag.PedidoID = pedido.ID
				if err := txQueries.PagamentoRepo.Create(ctx, pag); err != nil {
					return err
				}
			}
		} else {
			if err := txQueries.PagamentoRepo.DeleteByPedido(ctx, pedido.ID); err != nil {
				return err
			}
			for _, pag := range pagamentos {
				pag.PedidoID = pedido.ID
				if err := txQueries.PagamentoRepo.Create(ctx, pag); err != nil {
					return err
				}
			}
		}

		// Update evento
		evento.Status = "processado"
		now := time.Now()
		evento.ProcessadoEm = &now
		evento.Erro = nil
		if err := txQueries.EventoRepo.Create(ctx, &domain.EventoIngestao{
			ID:              evento.ID,
			EmpresaID:       input.EmpresaID,
			LojaID:          input.LojaID,
			PdvOrigemID:     input.PdvOrigemID,
			TipoEvento:      input.TipoEvento,
			PayloadOriginal: evento.PayloadOriginal,
			Status:          "processado",
			Erro:            nil,
			Tentativas:      evento.Tentativas,
			IdempotencyKey:  input.IdempotencyKey,
			ProcessadoEm:    &now,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &IngestSalesEventOutput{
		EventoID:     evento.ID,
		PedidoID:     pedido.ID,
		JaProcessado: !isNewPedido,
	}, nil
}

func (uc *SalesUsecase) GetVendasResumo(ctx context.Context, empresaID uuid.UUID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time) (*domain.VendasResumo, error) {
	return uc.vendasResumoRepo.GetResumo(ctx, empresaID, lojaID, dataInicio, dataFim)
}

func (uc *SalesUsecase) ListPedidos(ctx context.Context, filter domain.PedidoCanonicoFilter) ([]*domain.PedidoCanonico, error) {
	return uc.pedidoRepo.List(ctx, filter)
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringPtr(m map[string]interface{}, key string) *string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return &s
		}
	}
	return nil
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func getUUIDPtr(m map[string]interface{}, key string) *uuid.UUID {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return &id
			}
		}
	}
	return nil
}

func getArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}
	return nil
}
