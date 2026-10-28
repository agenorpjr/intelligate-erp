package db

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FornecedorRepository struct {
	querier *Queries
}

func NewFornecedorRepository(pool *pgxpool.Pool) *FornecedorRepository {
	return &FornecedorRepository{querier: New(pool)}
}

func (r *FornecedorRepository) Create(ctx context.Context, f *domain.Fornecedor) error {
	var email, telefone, banco, agencia, conta, tipoConta, pixChave, observacoes pgtype.Text
	if f.Email != nil {
		email = pgtype.Text{String: *f.Email, Valid: true}
	}
	if f.Telefone != nil {
		telefone = pgtype.Text{String: *f.Telefone, Valid: true}
	}
	if f.Banco != nil {
		banco = pgtype.Text{String: *f.Banco, Valid: true}
	}
	if f.Agencia != nil {
		agencia = pgtype.Text{String: *f.Agencia, Valid: true}
	}
	if f.Conta != nil {
		conta = pgtype.Text{String: *f.Conta, Valid: true}
	}
	if f.TipoConta != nil {
		tipoConta = pgtype.Text{String: *f.TipoConta, Valid: true}
	}
	if f.PixChave != nil {
		pixChave = pgtype.Text{String: *f.PixChave, Valid: true}
	}
	if f.Observacoes != nil {
		observacoes = pgtype.Text{String: *f.Observacoes, Valid: true}
	}

	params := CreateFornecedorParams{
		EmpresaID:   f.EmpresaID,
		Nome:        f.Nome,
		CnpjCpf:     f.CNPJCPF,
		Email:       email,
		Telefone:    telefone,
		Banco:       banco,
		Agencia:     agencia,
		Conta:       conta,
		TipoConta:   tipoConta,
		PixChave:    pixChave,
		Observacoes: observacoes,
	}
	result, err := r.querier.CreateFornecedor(ctx, params)
	if err != nil {
		return err
	}
	f.ID = result.ID
	f.CreatedAt = result.CreatedAt
	f.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *FornecedorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fornecedor, error) {
	result, err := r.querier.GetFornecedor(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainFornecedor(result), nil
}

func (r *FornecedorRepository) GetByCNPJCPF(ctx context.Context, empresaID uuid.UUID, cnpjCpf string) (*domain.Fornecedor, error) {
	result, err := r.querier.GetFornecedorByCNPJCPF(ctx, GetFornecedorByCNPJCPFParams{
		EmpresaID: empresaID,
		CnpjCpf:   cnpjCpf,
	})
	if err != nil {
		return nil, err
	}
	return toDomainFornecedor(result), nil
}

func (r *FornecedorRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, ativo *bool, limit, offset int) ([]*domain.Fornecedor, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	results, err := r.querier.ListFornecedoresByEmpresa(ctx, ListFornecedoresByEmpresaParams{
		EmpresaID: empresaID,
		Column2:   ativoParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	fornecedores := make([]*domain.Fornecedor, len(results))
	for i, r := range results {
		fornecedores[i] = toDomainFornecedor(r)
	}
	return fornecedores, nil
}

func (r *FornecedorRepository) ListByFilter(ctx context.Context, empresaID uuid.UUID, ativo *bool, nome, cnpjCpf *string, limit, offset int) ([]*domain.Fornecedor, error) {
	var ativoParam bool
	if ativo != nil {
		ativoParam = *ativo
	}
	var nomeParam, cnpjCpfParam string
	if nome != nil {
		nomeParam = *nome
	}
	if cnpjCpf != nil {
		cnpjCpfParam = *cnpjCpf
	}
	results, err := r.querier.ListFornecedoresByFilter(ctx, ListFornecedoresByFilterParams{
		EmpresaID: empresaID,
		Column2:   ativoParam,
		Column3:   nomeParam,
		Column4:   cnpjCpfParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	fornecedores := make([]*domain.Fornecedor, len(results))
	for i, r := range results {
		fornecedores[i] = toDomainFornecedor(r)
	}
	return fornecedores, nil
}

func (r *FornecedorRepository) Update(ctx context.Context, f *domain.Fornecedor) error {
	var email, telefone, banco, agencia, conta, tipoConta, pixChave, observacoes pgtype.Text
	if f.Email != nil {
		email = pgtype.Text{String: *f.Email, Valid: true}
	}
	if f.Telefone != nil {
		telefone = pgtype.Text{String: *f.Telefone, Valid: true}
	}
	if f.Banco != nil {
		banco = pgtype.Text{String: *f.Banco, Valid: true}
	}
	if f.Agencia != nil {
		agencia = pgtype.Text{String: *f.Agencia, Valid: true}
	}
	if f.Conta != nil {
		conta = pgtype.Text{String: *f.Conta, Valid: true}
	}
	if f.TipoConta != nil {
		tipoConta = pgtype.Text{String: *f.TipoConta, Valid: true}
	}
	if f.PixChave != nil {
		pixChave = pgtype.Text{String: *f.PixChave, Valid: true}
	}
	if f.Observacoes != nil {
		observacoes = pgtype.Text{String: *f.Observacoes, Valid: true}
	}

	params := UpdateFornecedorParams{
		ID:          f.ID,
		Nome:        f.Nome,
		CnpjCpf:     f.CNPJCPF,
		Email:       email,
		Telefone:    telefone,
		Banco:       banco,
		Agencia:     agencia,
		Conta:       conta,
		TipoConta:   tipoConta,
		PixChave:    pixChave,
		Observacoes: observacoes,
		Ativo:       f.Ativo,
	}
	_, err := r.querier.UpdateFornecedor(ctx, params)
	return err
}

func (r *FornecedorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteFornecedor(ctx, id)
}

func toDomainFornecedor(f Fornecedore) *domain.Fornecedor {
	var email, telefone, banco, agencia, conta, tipoConta, pixChave, observacoes *string
	if f.Email.Valid {
		email = &f.Email.String
	}
	if f.Telefone.Valid {
		telefone = &f.Telefone.String
	}
	if f.Banco.Valid {
		banco = &f.Banco.String
	}
	if f.Agencia.Valid {
		agencia = &f.Agencia.String
	}
	if f.Conta.Valid {
		conta = &f.Conta.String
	}
	if f.TipoConta.Valid {
		tipoConta = &f.TipoConta.String
	}
	if f.PixChave.Valid {
		pixChave = &f.PixChave.String
	}
	if f.Observacoes.Valid {
		observacoes = &f.Observacoes.String
	}
	var deletedAt *time.Time
	if f.DeletedAt.Valid {
		deletedAt = &f.DeletedAt.Time
	}
	return &domain.Fornecedor{
		ID:          f.ID,
		EmpresaID:   f.EmpresaID,
		Nome:        f.Nome,
		CNPJCPF:     f.CnpjCpf,
		Email:       email,
		Telefone:    telefone,
		Banco:       banco,
		Agencia:     agencia,
		Conta:       conta,
		TipoConta:   tipoConta,
		PixChave:    pixChave,
		Observacoes: observacoes,
		Ativo:       f.Ativo,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

// ============================================================
// Ordem de Compra Repository
// ============================================================

type OrdemCompraRepository struct {
	querier *Queries
}

func NewOrdemCompraRepository(pool *pgxpool.Pool) *OrdemCompraRepository {
	return &OrdemCompraRepository{querier: New(pool)}
}

func (r *OrdemCompraRepository) Create(ctx context.Context, oc *domain.OrdemCompra) error {
	var dataPrevistaEntrega pgtype.Date
	if oc.DataPrevistaEntrega != nil {
		dataPrevistaEntrega = pgtype.Date{Time: *oc.DataPrevistaEntrega, Valid: true}
	}
	var observacoes pgtype.Text
	if oc.Observacoes != nil {
		observacoes = pgtype.Text{String: *oc.Observacoes, Valid: true}
	}
	custoUnitarioAtual := big.NewInt(int64(oc.ValorTotal * 10000))
	params := CreateOrdemCompraParams{
		EmpresaID:           oc.EmpresaID,
		LojaID:              oc.LojaID,
		FornecedorID:        oc.FornecedorID,
		DataCompra:          pgtype.Date{Time: oc.DataCompra, Valid: true},
		DataPrevistaEntrega: dataPrevistaEntrega,
		Status:              string(oc.Status),
		Observacoes:         observacoes,
		ValorTotal:          pgtype.Numeric{Int: custoUnitarioAtual, Exp: -4, Valid: true},
	}
	result, err := r.querier.CreateOrdemCompra(ctx, params)
	if err != nil {
		return err
	}
	oc.ID = result.ID
	oc.CreatedAt = result.CreatedAt
	oc.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *OrdemCompraRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrdemCompra, error) {
	result, err := r.querier.GetOrdemCompra(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainOrdemCompra(result), nil
}

func (r *OrdemCompraRepository) GetWithItens(ctx context.Context, id uuid.UUID) (*domain.OrdemCompra, error) {
	results, err := r.querier.GetOrdemCompraWithItens(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainOrdemCompraWithItens(results), nil
}

func (r *OrdemCompraRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.OrdemCompra, error) {
	results, err := r.querier.ListOrdensCompraByEmpresa(ctx, ListOrdensCompraByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	ordens := make([]*domain.OrdemCompra, len(results))
	for i, r := range results {
		ordens[i] = toDomainOrdemCompra(r)
	}
	return ordens, nil
}

func (r *OrdemCompraRepository) ListByLoja(ctx context.Context, lojaID uuid.UUID, limit, offset int) ([]*domain.OrdemCompra, error) {
	results, err := r.querier.ListOrdensCompraByLoja(ctx, ListOrdensCompraByLojaParams{
		LojaID: lojaID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	ordens := make([]*domain.OrdemCompra, len(results))
	for i, r := range results {
		ordens[i] = toDomainOrdemCompra(r)
	}
	return ordens, nil
}

func (r *OrdemCompraRepository) ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *domain.OrdemCompraStatus, dataInicio, dataFim *time.Time, limit, offset int) ([]*domain.OrdemCompra, error) {
	var lojaIDParam, fornecedorIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	if fornecedorID != nil {
		fornecedorIDParam = *fornecedorID
	}
	var statusParam string
	if status != nil {
		statusParam = string(*status)
	}
	var dataInicioParam, dataFimParam pgtype.Date
	if dataInicio != nil {
		dataInicioParam = pgtype.Date{Time: *dataInicio, Valid: true}
	}
	if dataFim != nil {
		dataFimParam = pgtype.Date{Time: *dataFim, Valid: true}
	}
	results, err := r.querier.ListOrdensCompraByFilter(ctx, ListOrdensCompraByFilterParams{
		EmpresaID: empresaID,
		Column2:   lojaIDParam,
		Column3:   fornecedorIDParam,
		Column4:   statusParam,
		Column5:   dataInicioParam,
		Column6:   dataFimParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	ordens := make([]*domain.OrdemCompra, len(results))
	for i, r := range results {
		ordens[i] = toDomainOrdemCompra(r)
	}
	return ordens, nil
}

func (r *OrdemCompraRepository) Update(ctx context.Context, oc *domain.OrdemCompra) error {
	var dataPrevistaEntrega pgtype.Date
	if oc.DataPrevistaEntrega != nil {
		dataPrevistaEntrega = pgtype.Date{Time: *oc.DataPrevistaEntrega, Valid: true}
	}
	var observacoes pgtype.Text
	if oc.Observacoes != nil {
		observacoes = pgtype.Text{String: *oc.Observacoes, Valid: true}
	}
	custoUnitarioAtual := big.NewInt(int64(oc.ValorTotal * 10000))
	params := UpdateOrdemCompraParams{
		ID:                  oc.ID,
		LojaID:              oc.LojaID,
		FornecedorID:        oc.FornecedorID,
		DataCompra:          pgtype.Date{Time: oc.DataCompra, Valid: true},
		DataPrevistaEntrega: dataPrevistaEntrega,
		Status:              string(oc.Status),
		Observacoes:         observacoes,
		ValorTotal:          pgtype.Numeric{Int: custoUnitarioAtual, Exp: -4, Valid: true},
	}
	_, err := r.querier.UpdateOrdemCompra(ctx, params)
	return err
}

func (r *OrdemCompraRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrdemCompraStatus) error {
	_, err := r.querier.UpdateOrdemCompraStatus(ctx, UpdateOrdemCompraStatusParams{
		ID:     id,
		Status: string(status),
	})
	return err
}

func (r *OrdemCompraRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteOrdemCompra(ctx, id)
}

func (r *OrdemCompraRepository) CreateItem(ctx context.Context, item *domain.OrdemCompraItem) error {
	quantidade := big.NewInt(int64(item.Quantidade * 10000))
	precoUnitario := big.NewInt(int64(item.PrecoUnitario * 10000))
	params := CreateOrdemCompraItemParams{
		OrdemCompraID: item.OrdemCompraID,
		InsumoID:      item.InsumoID,
		Quantidade:    pgtype.Numeric{Int: quantidade, Exp: -4, Valid: true},
		UnidadeMedida: item.UnidadeMedida,
		PrecoUnitario: pgtype.Numeric{Int: precoUnitario, Exp: -4, Valid: true},
	}
	result, err := r.querier.CreateOrdemCompraItem(ctx, params)
	if err != nil {
		return err
	}
	item.ID = result.ID
	item.CreatedAt = result.CreatedAt
	item.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *OrdemCompraRepository) GetItem(ctx context.Context, id uuid.UUID) (*domain.OrdemCompraItem, error) {
	result, err := r.querier.GetOrdemCompraItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainOrdemCompraItem(result), nil
}

func (r *OrdemCompraRepository) ListItensByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) ([]*domain.OrdemCompraItem, error) {
	results, err := r.querier.ListOrdemCompraItensByOrdemCompra(ctx, ordemCompraID)
	if err != nil {
		return nil, err
	}
	itens := make([]*domain.OrdemCompraItem, len(results))
	for i, r := range results {
		itens[i] = toDomainOrdemCompraItem(r)
	}
	return itens, nil
}

func (r *OrdemCompraRepository) UpdateItem(ctx context.Context, item *domain.OrdemCompraItem) error {
	quantidade := big.NewInt(int64(item.Quantidade * 10000))
	precoUnitario := big.NewInt(int64(item.PrecoUnitario * 10000))
	params := UpdateOrdemCompraItemParams{
		ID:            item.ID,
		InsumoID:      item.InsumoID,
		Quantidade:    pgtype.Numeric{Int: quantidade, Exp: -4, Valid: true},
		UnidadeMedida: item.UnidadeMedida,
		PrecoUnitario: pgtype.Numeric{Int: precoUnitario, Exp: -4, Valid: true},
	}
	_, err := r.querier.UpdateOrdemCompraItem(ctx, params)
	return err
}

func (r *OrdemCompraRepository) UpdateItemQuantidadeRecebida(ctx context.Context, itemID uuid.UUID, quantidadeRecebida float64) error {
	qtd := big.NewInt(int64(quantidadeRecebida * 10000))
	_, err := r.querier.UpdateOrdemCompraItemQuantidadeRecebida(ctx, UpdateOrdemCompraItemQuantidadeRecebidaParams{
		ID:                 itemID,
		QuantidadeRecebida: pgtype.Numeric{Int: qtd, Exp: -4, Valid: true},
	})
	return err
}

func (r *OrdemCompraRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteOrdemCompraItem(ctx, id)
}

func (r *OrdemCompraRepository) DeleteItensByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) error {
	return r.querier.DeleteOrdemCompraItensByOrdemCompra(ctx, ordemCompraID)
}

func toDomainOrdemCompra(o OrdensCompra) *domain.OrdemCompra {
	var dataPrevistaEntrega *time.Time
	if o.DataPrevistaEntrega.Valid {
		dataPrevistaEntrega = &o.DataPrevistaEntrega.Time
	}
	var observacoes *string
	if o.Observacoes.Valid {
		observacoes = &o.Observacoes.String
	}
	var deletedAt *time.Time
	if o.DeletedAt.Valid {
		deletedAt = &o.DeletedAt.Time
	}
	valorTotal := 0.0
	if o.ValorTotal.Valid && o.ValorTotal.Int != nil {
		valorTotal = float64(o.ValorTotal.Int.Int64()) / 10000
	}
	return &domain.OrdemCompra{
		ID:                  o.ID,
		EmpresaID:           o.EmpresaID,
		LojaID:              o.LojaID,
		FornecedorID:        o.FornecedorID,
		DataCompra:          o.DataCompra.Time,
		DataPrevistaEntrega: dataPrevistaEntrega,
		Status:              domain.OrdemCompraStatus(o.Status),
		Observacoes:         observacoes,
		ValorTotal:          valorTotal,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
		DeletedAt:           deletedAt,
	}
}

func toDomainOrdemCompraWithItens(rows []GetOrdemCompraWithItensRow) *domain.OrdemCompra {
	if len(rows) == 0 {
		return nil
	}

	r := rows[0]
	var dataPrevistaEntrega *time.Time
	if r.DataPrevistaEntrega.Valid {
		dataPrevistaEntrega = &r.DataPrevistaEntrega.Time
	}
	var observacoes *string
	if r.Observacoes.Valid {
		observacoes = &r.Observacoes.String
	}
	var deletedAt *time.Time
	if r.DeletedAt.Valid {
		deletedAt = &r.DeletedAt.Time
	}
	valorTotal := 0.0
	if r.ValorTotal.Valid && r.ValorTotal.Int != nil {
		valorTotal = float64(r.ValorTotal.Int.Int64()) / 10000
	}

	oc := &domain.OrdemCompra{
		ID:                  r.ID,
		EmpresaID:           r.EmpresaID,
		LojaID:              r.LojaID,
		FornecedorID:        r.FornecedorID,
		DataCompra:          r.DataCompra.Time,
		DataPrevistaEntrega: dataPrevistaEntrega,
		Status:              domain.OrdemCompraStatus(r.Status),
		Observacoes:         observacoes,
		ValorTotal:          valorTotal,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		DeletedAt:           deletedAt,
	}

	itens := make([]*domain.OrdemCompraItem, 0, len(rows))
	for _, row := range rows {
		if row.ItemID.Valid {
			item := toDomainOrdemCompraItemRow(row)
			itens = append(itens, item)
		}
	}
	oc.Itens = itens

	return oc
}

func toDomainOrdemCompraItemRow(r GetOrdemCompraWithItensRow) *domain.OrdemCompraItem {
	quantidade := 0.0
	if r.Quantidade.Valid && r.Quantidade.Int != nil {
		quantidade = float64(r.Quantidade.Int.Int64()) / 10000
	}
	precoUnitario := 0.0
	if r.PrecoUnitario.Valid && r.PrecoUnitario.Int != nil {
		precoUnitario = float64(r.PrecoUnitario.Int.Int64()) / 10000
	}
	qtdRecebida := 0.0
	if r.QuantidadeRecebida.Valid && r.QuantidadeRecebida.Int != nil {
		qtdRecebida = float64(r.QuantidadeRecebida.Int.Int64()) / 10000
	}
	var itemID uuid.UUID
	if r.ItemID.Valid {
		itemID = r.ItemID.Bytes
	}
	var ordemCompraID uuid.UUID
	if r.OrdemCompraID.Valid {
		ordemCompraID = r.OrdemCompraID.Bytes
	}
	var insumoID uuid.UUID
	if r.InsumoID.Valid {
		insumoID = r.InsumoID.Bytes
	}
	var unidadeMedida string
	if r.UnidadeMedida.Valid {
		unidadeMedida = r.UnidadeMedida.String
	}
	var createdAt time.Time
	if r.ItemCreatedAt.Valid {
		createdAt = r.ItemCreatedAt.Time
	}
	var updatedAt time.Time
	if r.ItemUpdatedAt.Valid {
		updatedAt = r.ItemUpdatedAt.Time
	}
	return &domain.OrdemCompraItem{
		ID:                 itemID,
		OrdemCompraID:      ordemCompraID,
		InsumoID:           insumoID,
		Quantidade:         quantidade,
		UnidadeMedida:      unidadeMedida,
		PrecoUnitario:      precoUnitario,
		QuantidadeRecebida: qtdRecebida,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}

func toDomainOrdemCompraItem(i OrdemCompraIten) *domain.OrdemCompraItem {
	quantidade := 0.0
	if i.Quantidade.Valid && i.Quantidade.Int != nil {
		quantidade = float64(i.Quantidade.Int.Int64()) / 10000
	}
	precoUnitario := 0.0
	if i.PrecoUnitario.Valid && i.PrecoUnitario.Int != nil {
		precoUnitario = float64(i.PrecoUnitario.Int.Int64()) / 10000
	}
	qtdRecebida := 0.0
	if i.QuantidadeRecebida.Valid && i.QuantidadeRecebida.Int != nil {
		qtdRecebida = float64(i.QuantidadeRecebida.Int.Int64()) / 10000
	}
	return &domain.OrdemCompraItem{
		ID:                 i.ID,
		OrdemCompraID:      i.OrdemCompraID,
		InsumoID:           i.InsumoID,
		Quantidade:         quantidade,
		UnidadeMedida:      i.UnidadeMedida,
		PrecoUnitario:      precoUnitario,
		QuantidadeRecebida: qtdRecebida,
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
	}
}

// ============================================================
// Recebimento de Compra Repository
// ============================================================

type RecebimentoCompraRepository struct {
	querier *Queries
}

func NewRecebimentoCompraRepository(pool *pgxpool.Pool) *RecebimentoCompraRepository {
	return &RecebimentoCompraRepository{querier: New(pool)}
}

func (r *RecebimentoCompraRepository) Create(ctx context.Context, rc *domain.RecebimentoCompra) error {
	var observacoes pgtype.Text
	if rc.Observacoes != nil {
		observacoes = pgtype.Text{String: *rc.Observacoes, Valid: true}
	}
	params := CreateRecebimentoCompraParams{
		OrdemCompraID:     rc.OrdemCompraID,
		LojaID:            rc.LojaID,
		DataRecebimento:   rc.DataRecebimento,
		Observacoes:       observacoes,
		PermitirExcedente: rc.PermitirExcedente,
	}
	result, err := r.querier.CreateRecebimentoCompra(ctx, params)
	if err != nil {
		return err
	}
	rc.ID = result.ID
	rc.CreatedAt = result.CreatedAt
	rc.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *RecebimentoCompraRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RecebimentoCompra, error) {
	result, err := r.querier.GetRecebimentoCompra(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainRecebimentoCompra(result), nil
}

func (r *RecebimentoCompraRepository) ListByOrdemCompra(ctx context.Context, ordemCompraID uuid.UUID) ([]*domain.RecebimentoCompra, error) {
	results, err := r.querier.ListRecebimentosCompraByOrdemCompra(ctx, ordemCompraID)
	if err != nil {
		return nil, err
	}
	recebimentos := make([]*domain.RecebimentoCompra, len(results))
	for i, r := range results {
		recebimentos[i] = toDomainRecebimentoCompra(r)
	}
	return recebimentos, nil
}

func (r *RecebimentoCompraRepository) ListByFilter(ctx context.Context, ordemCompraID, lojaID *uuid.UUID, dataInicio, dataFim *time.Time, limit, offset int) ([]*domain.RecebimentoCompra, error) {
	var ordemCompraIDParam, lojaIDParam uuid.UUID
	if ordemCompraID != nil {
		ordemCompraIDParam = *ordemCompraID
	}
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	var dataInicioParam, dataFimParam pgtype.Timestamptz
	if dataInicio != nil {
		dataInicioParam = pgtype.Timestamptz{Time: *dataInicio, Valid: true}
	}
	if dataFim != nil {
		dataFimParam = pgtype.Timestamptz{Time: *dataFim, Valid: true}
	}
	results, err := r.querier.ListRecebimentosCompraByFilter(ctx, ListRecebimentosCompraByFilterParams{
		Column1: ordemCompraIDParam,
		Column2: lojaIDParam,
		Column3: dataInicioParam.Time,
		Column4: dataFimParam.Time,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	recebimentos := make([]*domain.RecebimentoCompra, len(results))
	for i, r := range results {
		recebimentos[i] = toDomainRecebimentoCompra(r)
	}
	return recebimentos, nil
}

func (r *RecebimentoCompraRepository) CreateItem(ctx context.Context, item *domain.RecebimentoCompraItem) error {
	quantidadeRecebida := big.NewInt(int64(item.QuantidadeRecebida * 10000))
	precoUnitario := big.NewInt(int64(item.PrecoUnitario * 10000))
	params := CreateRecebimentoCompraItemParams{
		RecebimentoCompraID: item.RecebimentoCompraID,
		OrdemCompraItemID:   item.OrdemCompraItemID,
		InsumoID:            item.InsumoID,
		QuantidadeRecebida:  pgtype.Numeric{Int: quantidadeRecebida, Exp: -4, Valid: true},
		PrecoUnitario:       pgtype.Numeric{Int: precoUnitario, Exp: -4, Valid: true},
	}
	result, err := r.querier.CreateRecebimentoCompraItem(ctx, params)
	if err != nil {
		return err
	}
	item.ID = result.ID
	item.CreatedAt = result.CreatedAt
	return nil
}

func (r *RecebimentoCompraRepository) ListItensByRecebimento(ctx context.Context, recebimentoID uuid.UUID) ([]*domain.RecebimentoCompraItem, error) {
	results, err := r.querier.ListRecebimentoCompraItensByRecebimento(ctx, recebimentoID)
	if err != nil {
		return nil, err
	}
	itens := make([]*domain.RecebimentoCompraItem, len(results))
	for i, r := range results {
		itens[i] = toDomainRecebimentoCompraItem(r)
	}
	return itens, nil
}

func toDomainRecebimentoCompra(r RecebimentosCompra) *domain.RecebimentoCompra {
	return &domain.RecebimentoCompra{
		ID:                r.ID,
		OrdemCompraID:     r.OrdemCompraID,
		LojaID:            r.LojaID,
		DataRecebimento:   r.DataRecebimento,
		Observacoes:       ptrStr(r.Observacoes),
		PermitirExcedente: r.PermitirExcedente,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func toDomainRecebimentoCompraItem(i RecebimentoCompraIten) *domain.RecebimentoCompraItem {
	quantidadeRecebida := 0.0
	if i.QuantidadeRecebida.Valid && i.QuantidadeRecebida.Int != nil {
		quantidadeRecebida = float64(i.QuantidadeRecebida.Int.Int64()) / 10000
	}
	precoUnitario := 0.0
	if i.PrecoUnitario.Valid && i.PrecoUnitario.Int != nil {
		precoUnitario = float64(i.PrecoUnitario.Int.Int64()) / 10000
	}
	return &domain.RecebimentoCompraItem{
		ID:                  i.ID,
		RecebimentoCompraID: i.RecebimentoCompraID,
		OrdemCompraItemID:   i.OrdemCompraItemID,
		InsumoID:            i.InsumoID,
		QuantidadeRecebida:  quantidadeRecebida,
		PrecoUnitario:       precoUnitario,
		CreatedAt:           i.CreatedAt,
	}
}

func ptrStr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

// ============================================================
// Conta a Pagar Repository
// ============================================================

type ContaPagarRepository struct {
	querier *Queries
}

func NewContaPagarRepository(pool *pgxpool.Pool) *ContaPagarRepository {
	return &ContaPagarRepository{querier: New(pool)}
}

func (r *ContaPagarRepository) Create(ctx context.Context, cp *domain.ContaPagar) error {
	var ordemCompraID, recebimentoCompraID pgtype.UUID
	if cp.OrdemCompraID != nil {
		ordemCompraID = pgtype.UUID{Bytes: *cp.OrdemCompraID, Valid: true}
	}
	if cp.RecebimentoCompraID != nil {
		recebimentoCompraID = pgtype.UUID{Bytes: *cp.RecebimentoCompraID, Valid: true}
	}
	var formaPagamento, observacoes pgtype.Text
	if cp.FormaPagamento != nil {
		formaPagamento = pgtype.Text{String: *cp.FormaPagamento, Valid: true}
	}
	if cp.Observacoes != nil {
		observacoes = pgtype.Text{String: *cp.Observacoes, Valid: true}
	}
	valor := big.NewInt(int64(cp.Valor * 10000))
	params := CreateContaPagarParams{
		EmpresaID:           cp.EmpresaID,
		LojaID:              cp.LojaID,
		FornecedorID:        cp.FornecedorID,
		OrdemCompraID:       ordemCompraID,
		RecebimentoCompraID: recebimentoCompraID,
		Descricao:           cp.Descricao,
		Valor:               pgtype.Numeric{Int: valor, Exp: -4, Valid: true},
		DataEmissao:         pgtype.Date{Time: cp.DataEmissao, Valid: true},
		DataVencimento:      pgtype.Date{Time: cp.DataVencimento, Valid: true},
		Status:              string(cp.Status),
		FormaPagamento:      formaPagamento,
		Observacoes:         observacoes,
	}
	result, err := r.querier.CreateContaPagar(ctx, params)
	if err != nil {
		return err
	}
	cp.ID = result.ID
	cp.CreatedAt = result.CreatedAt
	cp.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ContaPagarRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContaPagar, error) {
	result, err := r.querier.GetContaPagar(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainContaPagar(result), nil
}

func (r *ContaPagarRepository) ListByEmpresa(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]*domain.ContaPagar, error) {
	results, err := r.querier.ListContasPagarByEmpresa(ctx, ListContasPagarByEmpresaParams{
		EmpresaID: empresaID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	contas := make([]*domain.ContaPagar, len(results))
	for i, r := range results {
		contas[i] = toDomainContaPagar(r)
	}
	return contas, nil
}

func (r *ContaPagarRepository) ListByFornecedor(ctx context.Context, fornecedorID uuid.UUID, limit, offset int) ([]*domain.ContaPagar, error) {
	results, err := r.querier.ListContasPagarByFornecedor(ctx, ListContasPagarByFornecedorParams{
		FornecedorID: fornecedorID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, err
	}
	contas := make([]*domain.ContaPagar, len(results))
	for i, r := range results {
		contas[i] = toDomainContaPagar(r)
	}
	return contas, nil
}

func (r *ContaPagarRepository) ListByStatus(ctx context.Context, empresaID uuid.UUID, status domain.ContaPagarStatus, limit, offset int) ([]*domain.ContaPagar, error) {
	results, err := r.querier.ListContasPagarByStatus(ctx, ListContasPagarByStatusParams{
		EmpresaID: empresaID,
		Status:    string(status),
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	contas := make([]*domain.ContaPagar, len(results))
	for i, r := range results {
		contas[i] = toDomainContaPagar(r)
	}
	return contas, nil
}

func (r *ContaPagarRepository) ListByFilter(ctx context.Context, empresaID uuid.UUID, lojaID, fornecedorID *uuid.UUID, status *domain.ContaPagarStatus, dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim *time.Time, limit, offset int) ([]*domain.ContaPagar, error) {
	var lojaIDParam, fornecedorIDParam uuid.UUID
	if lojaID != nil {
		lojaIDParam = *lojaID
	}
	if fornecedorID != nil {
		fornecedorIDParam = *fornecedorID
	}
	var statusParam string
	if status != nil {
		statusParam = string(*status)
	}
	var dataVencimentoInicioParam, dataVencimentoFimParam, dataPagamentoInicioParam, dataPagamentoFimParam pgtype.Date
	if dataVencimentoInicio != nil {
		dataVencimentoInicioParam = pgtype.Date{Time: *dataVencimentoInicio, Valid: true}
	}
	if dataVencimentoFim != nil {
		dataVencimentoFimParam = pgtype.Date{Time: *dataVencimentoFim, Valid: true}
	}
	if dataPagamentoInicio != nil {
		dataPagamentoInicioParam = pgtype.Date{Time: *dataPagamentoInicio, Valid: true}
	}
	if dataPagamentoFim != nil {
		dataPagamentoFimParam = pgtype.Date{Time: *dataPagamentoFim, Valid: true}
	}
	results, err := r.querier.ListContasPagarByFilter(ctx, ListContasPagarByFilterParams{
		EmpresaID: empresaID,
		Column2:   lojaIDParam,
		Column3:   fornecedorIDParam,
		Column4:   statusParam,
		Column5:   dataVencimentoInicioParam,
		Column6:   dataVencimentoFimParam,
		Column7:   dataPagamentoInicioParam,
		Column8:   dataPagamentoFimParam,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	contas := make([]*domain.ContaPagar, len(results))
	for i, r := range results {
		contas[i] = toDomainContaPagar(r)
	}
	return contas, nil
}

func (r *ContaPagarRepository) Update(ctx context.Context, cp *domain.ContaPagar) error {
	var formaPagamento, observacoes pgtype.Text
	if cp.FormaPagamento != nil {
		formaPagamento = pgtype.Text{String: *cp.FormaPagamento, Valid: true}
	}
	if cp.Observacoes != nil {
		observacoes = pgtype.Text{String: *cp.Observacoes, Valid: true}
	}
	var dataPagamento pgtype.Date
	if cp.DataPagamento != nil {
		dataPagamento = pgtype.Date{Time: *cp.DataPagamento, Valid: true}
	}
	valor := big.NewInt(int64(cp.Valor * 10000))
	params := UpdateContaPagarParams{
		ID:             cp.ID,
		Descricao:      cp.Descricao,
		Valor:          pgtype.Numeric{Int: valor, Exp: -4, Valid: true},
		DataEmissao:    pgtype.Date{Time: cp.DataEmissao, Valid: true},
		DataVencimento: pgtype.Date{Time: cp.DataVencimento, Valid: true},
		DataPagamento:  dataPagamento,
		Status:         string(cp.Status),
		FormaPagamento: formaPagamento,
		Observacoes:    observacoes,
	}
	_, err := r.querier.UpdateContaPagar(ctx, params)
	return err
}

func (r *ContaPagarRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ContaPagarStatus, dataPagamento *time.Time, formaPagamento *string) error {
	var dp pgtype.Date
	if dataPagamento != nil {
		dp = pgtype.Date{Time: *dataPagamento, Valid: true}
	}
	var fp pgtype.Text
	if formaPagamento != nil {
		fp = pgtype.Text{String: *formaPagamento, Valid: true}
	}
	_, err := r.querier.UpdateContaPagarStatus(ctx, UpdateContaPagarStatusParams{
		ID:             id,
		Status:         string(status),
		DataPagamento:  dp,
		FormaPagamento: fp,
	})
	return err
}

func (r *ContaPagarRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier.DeleteContaPagar(ctx, id)
}

func toDomainContaPagar(c ContasPagar) *domain.ContaPagar {
	var ordemCompraID, recebimentoCompraID *uuid.UUID
	if c.OrdemCompraID.Valid {
		id := uuid.UUID(c.OrdemCompraID.Bytes)
		ordemCompraID = &id
	}
	if c.RecebimentoCompraID.Valid {
		id := uuid.UUID(c.RecebimentoCompraID.Bytes)
		recebimentoCompraID = &id
	}
	var formaPagamento, observacoes *string
	if c.FormaPagamento.Valid {
		formaPagamento = &c.FormaPagamento.String
	}
	if c.Observacoes.Valid {
		observacoes = &c.Observacoes.String
	}
	var dataPagamento *time.Time
	if c.DataPagamento.Valid {
		dataPagamento = &c.DataPagamento.Time
	}
	var deletedAt *time.Time
	if c.DeletedAt.Valid {
		deletedAt = &c.DeletedAt.Time
	}
	valor := 0.0
	if c.Valor.Valid && c.Valor.Int != nil {
		valor = float64(c.Valor.Int.Int64()) / 10000
	}
	return &domain.ContaPagar{
		ID:                  c.ID,
		EmpresaID:           c.EmpresaID,
		LojaID:              c.LojaID,
		FornecedorID:        c.FornecedorID,
		OrdemCompraID:       ordemCompraID,
		RecebimentoCompraID: recebimentoCompraID,
		Descricao:           c.Descricao,
		Valor:               valor,
		DataEmissao:         c.DataEmissao.Time,
		DataVencimento:      c.DataVencimento.Time,
		DataPagamento:       dataPagamento,
		Status:              domain.ContaPagarStatus(c.Status),
		FormaPagamento:      formaPagamento,
		Observacoes:         observacoes,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		DeletedAt:           deletedAt,
	}
}
