package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Fornecedor Handler
// ============================================================

type FornecedorHandler struct {
	fornecedorUsecase *usecase.FornecedorUsecase
}

func NewFornecedorHandler(fornecedorUsecase *usecase.FornecedorUsecase) *FornecedorHandler {
	return &FornecedorHandler{fornecedorUsecase: fornecedorUsecase}
}

type CreateFornecedorRequest struct {
	Nome        string  `json:"nome" validate:"required"`
	CNPJCPF     string  `json:"cnpj_cpf" validate:"required"`
	Email       *string `json:"email,omitempty"`
	Telefone    *string `json:"telefone,omitempty"`
	Banco       *string `json:"banco,omitempty"`
	Agencia     *string `json:"agencia,omitempty"`
	Conta       *string `json:"conta,omitempty"`
	TipoConta   *string `json:"tipo_conta,omitempty"`
	PixChave    *string `json:"pix_chave,omitempty"`
	Observacoes *string `json:"observacoes,omitempty"`
}

type UpdateFornecedorRequest struct {
	Nome        *string `json:"nome,omitempty"`
	CNPJCPF     *string `json:"cnpj_cpf,omitempty"`
	Email       *string `json:"email,omitempty"`
	Telefone    *string `json:"telefone,omitempty"`
	Banco       *string `json:"banco,omitempty"`
	Agencia     *string `json:"agencia,omitempty"`
	Conta       *string `json:"conta,omitempty"`
	TipoConta   *string `json:"tipo_conta,omitempty"`
	PixChave    *string `json:"pix_chave,omitempty"`
	Observacoes *string `json:"observacoes,omitempty"`
	Ativo       *bool   `json:"ativo,omitempty"`
}

type FornecedorResponse struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	Nome        string     `json:"nome"`
	CNPJCPF     string     `json:"cnpj_cpf"`
	Email       *string    `json:"email,omitempty"`
	Telefone    *string    `json:"telefone,omitempty"`
	Banco       *string    `json:"banco,omitempty"`
	Agencia     *string    `json:"agencia,omitempty"`
	Conta       *string    `json:"conta,omitempty"`
	TipoConta   *string    `json:"tipo_conta,omitempty"`
	PixChave    *string    `json:"pix_chave,omitempty"`
	Observacoes *string    `json:"observacoes,omitempty"`
	Ativo       bool       `json:"ativo"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func (h *FornecedorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateFornecedorRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	input := domain.CreateFornecedorInput{
		EmpresaID:   empresaID,
		Nome:        req.Nome,
		CNPJCPF:     req.CNPJCPF,
		Email:       req.Email,
		Telefone:    req.Telefone,
		Banco:       req.Banco,
		Agencia:     req.Agencia,
		Conta:       req.Conta,
		TipoConta:   req.TipoConta,
		PixChave:    req.PixChave,
		Observacoes: req.Observacoes,
	}

	fornecedor, err := h.fornecedorUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrFornecedorJaExiste) {
			RespondError(w, http.StatusConflict, "CONFLICT", "Fornecedor com este CNPJ/CPF já existe", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, toFornecedorResponse(fornecedor))
}

func (h *FornecedorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	fornecedor, err := h.fornecedorUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrFornecedorNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Fornecedor não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toFornecedorResponse(fornecedor))
}

func (h *FornecedorHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0
	var ativo *bool
	var nome, cnpjCpf *string

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if ativoStr := r.URL.Query().Get("ativo"); ativoStr != "" {
		if a, err := strconv.ParseBool(ativoStr); err == nil {
			ativo = &a
		}
	}
	if nomeStr := r.URL.Query().Get("nome"); nomeStr != "" {
		nome = &nomeStr
	}
	if cnpjCpfStr := r.URL.Query().Get("cnpj_cpf"); cnpjCpfStr != "" {
		cnpjCpf = &cnpjCpfStr
	}

	fornecedores, err := h.fornecedorUsecase.ListByFilter(r.Context(), empresaID, ativo, nome, cnpjCpf, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]FornecedorResponse, len(fornecedores))
	for i, f := range fornecedores {
		resp[i] = toFornecedorResponse(f)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *FornecedorHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateFornecedorRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	input := domain.UpdateFornecedorInput{
		Nome:        req.Nome,
		CNPJCPF:     req.CNPJCPF,
		Email:       req.Email,
		Telefone:    req.Telefone,
		Banco:       req.Banco,
		Agencia:     req.Agencia,
		Conta:       req.Conta,
		TipoConta:   req.TipoConta,
		PixChave:    req.PixChave,
		Observacoes: req.Observacoes,
		Ativo:       req.Ativo,
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	fornecedor, err := h.fornecedorUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrFornecedorNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Fornecedor não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrFornecedorJaExiste) {
			RespondError(w, http.StatusConflict, "CONFLICT", "CNPJ/CPF já existe", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toFornecedorResponse(fornecedor))
}

func (h *FornecedorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	if err := h.fornecedorUsecase.Delete(r.Context(), id, usuario.ID); err != nil {
		if errors.Is(err, usecase.ErrFornecedorNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Fornecedor não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func toFornecedorResponse(f *domain.Fornecedor) FornecedorResponse {
	return FornecedorResponse{
		ID:          f.ID,
		EmpresaID:   f.EmpresaID,
		Nome:        f.Nome,
		CNPJCPF:     f.CNPJCPF,
		Email:       f.Email,
		Telefone:    f.Telefone,
		Banco:       f.Banco,
		Agencia:     f.Agencia,
		Conta:       f.Conta,
		TipoConta:   f.TipoConta,
		PixChave:    f.PixChave,
		Observacoes: f.Observacoes,
		Ativo:       f.Ativo,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		DeletedAt:   f.DeletedAt,
	}
}

// ============================================================
// Ordem de Compra Handler
// ============================================================

type OrdemCompraHandler struct {
	ordemCompraUsecase *usecase.OrdemCompraUsecase
}

func NewOrdemCompraHandler(ordemCompraUsecase *usecase.OrdemCompraUsecase) *OrdemCompraHandler {
	return &OrdemCompraHandler{ordemCompraUsecase: ordemCompraUsecase}
}

type CreateOrdemCompraRequest struct {
	LojaID              uuid.UUID                      `json:"loja_id" validate:"required"`
	FornecedorID        uuid.UUID                      `json:"fornecedor_id" validate:"required"`
	DataCompra          time.Time                      `json:"data_compra" validate:"required"`
	DataPrevistaEntrega *time.Time                     `json:"data_prevista_entrega,omitempty"`
	Observacoes         *string                        `json:"observacoes,omitempty"`
	Itens               []CreateOrdemCompraItemRequest `json:"itens" validate:"required,min=1"`
}

type CreateOrdemCompraItemRequest struct {
	InsumoID      uuid.UUID `json:"insumo_id" validate:"required"`
	Quantidade    float64   `json:"quantidade" validate:"required,gt=0"`
	UnidadeMedida string    `json:"unidade_medida" validate:"required"`
	PrecoUnitario float64   `json:"preco_unitario" validate:"required,gte=0"`
}

type UpdateOrdemCompraRequest struct {
	LojaID              *uuid.UUID `json:"loja_id,omitempty"`
	FornecedorID        *uuid.UUID `json:"fornecedor_id,omitempty"`
	DataCompra          *time.Time `json:"data_compra,omitempty"`
	DataPrevistaEntrega *time.Time `json:"data_prevista_entrega,omitempty"`
	Status              *string    `json:"status,omitempty"`
	Observacoes         *string    `json:"observacoes,omitempty"`
}

type OrdemCompraResponse struct {
	ID                  uuid.UUID                 `json:"id"`
	EmpresaID           uuid.UUID                 `json:"empresa_id"`
	LojaID              uuid.UUID                 `json:"loja_id"`
	FornecedorID        uuid.UUID                 `json:"fornecedor_id"`
	DataCompra          time.Time                 `json:"data_compra"`
	DataPrevistaEntrega *time.Time                `json:"data_prevista_entrega,omitempty"`
	Status              string                    `json:"status"`
	Observacoes         *string                   `json:"observacoes,omitempty"`
	ValorTotal          float64                   `json:"valor_total"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
	DeletedAt           *time.Time                `json:"deleted_at,omitempty"`
	Itens               []OrdemCompraItemResponse `json:"itens,omitempty"`
	Fornecedor          *FornecedorResponse       `json:"fornecedor,omitempty"`
	Loja                *LojaResponse             `json:"loja,omitempty"`
}

type OrdemCompraItemResponse struct {
	ID                 uuid.UUID `json:"id"`
	OrdemCompraID      uuid.UUID `json:"ordem_compra_id"`
	InsumoID           uuid.UUID `json:"insumo_id"`
	Quantidade         float64   `json:"quantidade"`
	UnidadeMedida      string    `json:"unidade_medida"`
	PrecoUnitario      float64   `json:"preco_unitario"`
	QuantidadeRecebida float64   `json:"quantidade_recebida"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (h *OrdemCompraHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrdemCompraRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	itens := make([]domain.CreateOrdemCompraItemInput, len(req.Itens))
	for i, item := range req.Itens {
		itens[i] = domain.CreateOrdemCompraItemInput{
			InsumoID:      item.InsumoID,
			Quantidade:    item.Quantidade,
			UnidadeMedida: item.UnidadeMedida,
			PrecoUnitario: item.PrecoUnitario,
		}
	}

	input := domain.CreateOrdemCompraInput{
		EmpresaID:           empresaID,
		LojaID:              req.LojaID,
		FornecedorID:        req.FornecedorID,
		DataCompra:          req.DataCompra,
		DataPrevistaEntrega: req.DataPrevistaEntrega,
		Observacoes:         req.Observacoes,
		Itens:               itens,
	}

	ordemCompra, err := h.ordemCompraUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, toOrdemCompraResponse(ordemCompra))
}

func (h *OrdemCompraHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	ordemCompra, err := h.ordemCompraUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ordem de compra não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toOrdemCompraResponse(ordemCompra))
}

func (h *OrdemCompraHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0
	var lojaID, fornecedorID *uuid.UUID
	var status *string
	var dataInicio, dataFim *time.Time

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}
	if fornecedorIDStr := r.URL.Query().Get("fornecedor_id"); fornecedorIDStr != "" {
		if id, err := uuid.Parse(fornecedorIDStr); err == nil {
			fornecedorID = &id
		}
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status = &statusStr
	}
	if dataInicioStr := r.URL.Query().Get("data_inicio"); dataInicioStr != "" {
		if t, err := time.Parse(time.RFC3339, dataInicioStr); err == nil {
			dataInicio = &t
		}
	}
	if dataFimStr := r.URL.Query().Get("data_fim"); dataFimStr != "" {
		if t, err := time.Parse(time.RFC3339, dataFimStr); err == nil {
			dataFim = &t
		}
	}

	ordens, err := h.ordemCompraUsecase.ListByFilter(r.Context(), empresaID, lojaID, fornecedorID, (*domain.OrdemCompraStatus)(status), dataInicio, dataFim, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]OrdemCompraResponse, len(ordens))
	for i, oc := range ordens {
		resp[i] = toOrdemCompraResponse(oc)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *OrdemCompraHandler) ListByLoja(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	lojaIDStr := chi.URLParam(r, "loja_id")
	lojaID, err := uuid.Parse(lojaIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID da loja inválido", "")
		return
	}

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ordens, err := h.ordemCompraUsecase.ListByLoja(r.Context(), lojaID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]OrdemCompraResponse, len(ordens))
	for i, oc := range ordens {
		resp[i] = toOrdemCompraResponse(oc)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *OrdemCompraHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateOrdemCompraRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := domain.UpdateOrdemCompraInput{
		LojaID:              req.LojaID,
		FornecedorID:        req.FornecedorID,
		DataCompra:          req.DataCompra,
		DataPrevistaEntrega: req.DataPrevistaEntrega,
		Status:              req.Status,
		Observacoes:         req.Observacoes,
	}

	ordemCompra, err := h.ordemCompraUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ordem de compra não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrOrdemCompraNaoEstaEmRascunho) {
			RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Ordem de compra não está em rascunho", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toOrdemCompraResponse(ordemCompra))
}

func (h *OrdemCompraHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	if err := h.ordemCompraUsecase.Cancel(r.Context(), id, usuario.ID); err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ordem de compra não encontrada", "")
			return
		}
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), "")
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *OrdemCompraHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	ordemCompraID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req CreateOrdemCompraItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := domain.CreateOrdemCompraItemInput{
		InsumoID:      req.InsumoID,
		Quantidade:    req.Quantidade,
		UnidadeMedida: req.UnidadeMedida,
		PrecoUnitario: req.PrecoUnitario,
	}

	item, err := h.ordemCompraUsecase.AddItem(r.Context(), ordemCompraID, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ordem de compra não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrOrdemCompraNaoEstaEmRascunho) {
			RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Ordem de compra não está em rascunho", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, toOrdemCompraItemResponse(item))
}

func (h *OrdemCompraHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID do item inválido", "")
		return
	}

	var req CreateOrdemCompraItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := domain.CreateOrdemCompraItemInput{
		InsumoID:      req.InsumoID,
		Quantidade:    req.Quantidade,
		UnidadeMedida: req.UnidadeMedida,
		PrecoUnitario: req.PrecoUnitario,
	}

	item, err := h.ordemCompraUsecase.UpdateItem(r.Context(), itemID, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraItemNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Item da ordem de compra não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrOrdemCompraNaoEstaEmRascunho) {
			RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Ordem de compra não está em rascunho", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toOrdemCompraItemResponse(item))
}

func (h *OrdemCompraHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID do item inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	if err := h.ordemCompraUsecase.RemoveItem(r.Context(), itemID, usuario.ID); err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraItemNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Item da ordem de compra não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrOrdemCompraNaoEstaEmRascunho) {
			RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Ordem de compra não está em rascunho", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func toOrdemCompraResponse(oc *domain.OrdemCompra) OrdemCompraResponse {
	resp := OrdemCompraResponse{
		ID:                  oc.ID,
		EmpresaID:           oc.EmpresaID,
		LojaID:              oc.LojaID,
		FornecedorID:        oc.FornecedorID,
		DataCompra:          oc.DataCompra,
		DataPrevistaEntrega: oc.DataPrevistaEntrega,
		Status:              string(oc.Status),
		Observacoes:         oc.Observacoes,
		ValorTotal:          oc.ValorTotal,
		CreatedAt:           oc.CreatedAt,
		UpdatedAt:           oc.UpdatedAt,
		DeletedAt:           oc.DeletedAt,
	}

	if oc.Itens != nil {
		resp.Itens = make([]OrdemCompraItemResponse, len(oc.Itens))
		for i, item := range oc.Itens {
			resp.Itens[i] = toOrdemCompraItemResponse(item)
		}
	}

	if oc.Fornecedor != nil {
		fr := toFornecedorResponse(oc.Fornecedor)
		resp.Fornecedor = &fr
	}

	if oc.Loja != nil {
		lr := toLojaResponse(oc.Loja)
		resp.Loja = &lr
	}

	return resp
}

func toOrdemCompraItemResponse(item *domain.OrdemCompraItem) OrdemCompraItemResponse {
	return OrdemCompraItemResponse{
		ID:                 item.ID,
		OrdemCompraID:      item.OrdemCompraID,
		InsumoID:           item.InsumoID,
		Quantidade:         item.Quantidade,
		UnidadeMedida:      item.UnidadeMedida,
		PrecoUnitario:      item.PrecoUnitario,
		QuantidadeRecebida: item.QuantidadeRecebida,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

// ============================================================
// Recebimento de Compra Handler
// ============================================================

type RecebimentoCompraHandler struct {
	recebimentoCompraUsecase *usecase.RecebimentoCompraUsecase
}

func NewRecebimentoCompraHandler(recebimentoCompraUsecase *usecase.RecebimentoCompraUsecase) *RecebimentoCompraHandler {
	return &RecebimentoCompraHandler{recebimentoCompraUsecase: recebimentoCompraUsecase}
}

type CreateRecebimentoCompraRequest struct {
	OrdemCompraID     uuid.UUID                            `json:"ordem_compra_id" validate:"required"`
	LojaID            uuid.UUID                            `json:"loja_id" validate:"required"`
	DataRecebimento   *time.Time                           `json:"data_recebimento,omitempty"`
	Observacoes       *string                              `json:"observacoes,omitempty"`
	PermitirExcedente bool                                 `json:"permitir_excedente"`
	Itens             []CreateRecebimentoCompraItemRequest `json:"itens" validate:"required,min=1"`
}

type CreateRecebimentoCompraItemRequest struct {
	OrdemCompraItemID  uuid.UUID `json:"ordem_compra_item_id" validate:"required"`
	InsumoID           uuid.UUID `json:"insumo_id" validate:"required"`
	QuantidadeRecebida float64   `json:"quantidade_recebida" validate:"required,gt=0"`
	PrecoUnitario      float64   `json:"preco_unitario" validate:"required,gte=0"`
}

type RecebimentoCompraResponse struct {
	ID                uuid.UUID                       `json:"id"`
	OrdemCompraID     uuid.UUID                       `json:"ordem_compra_id"`
	LojaID            uuid.UUID                       `json:"loja_id"`
	DataRecebimento   time.Time                       `json:"data_recebimento"`
	Observacoes       *string                         `json:"observacoes,omitempty"`
	PermitirExcedente bool                            `json:"permitir_excedente"`
	CreatedAt         time.Time                       `json:"created_at"`
	UpdatedAt         time.Time                       `json:"updated_at"`
	Itens             []RecebimentoCompraItemResponse `json:"itens,omitempty"`
}

type RecebimentoCompraItemResponse struct {
	ID                  uuid.UUID `json:"id"`
	RecebimentoCompraID uuid.UUID `json:"recebimento_compra_id"`
	OrdemCompraItemID   uuid.UUID `json:"ordem_compra_item_id"`
	InsumoID            uuid.UUID `json:"insumo_id"`
	QuantidadeRecebida  float64   `json:"quantidade_recebida"`
	PrecoUnitario       float64   `json:"preco_unitario"`
	CreatedAt           time.Time `json:"created_at"`
}

func (h *RecebimentoCompraHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRecebimentoCompraRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	itens := make([]domain.CreateRecebimentoCompraItemInput, len(req.Itens))
	for i, item := range req.Itens {
		itens[i] = domain.CreateRecebimentoCompraItemInput{
			OrdemCompraItemID:  item.OrdemCompraItemID,
			InsumoID:           item.InsumoID,
			QuantidadeRecebida: item.QuantidadeRecebida,
			PrecoUnitario:      item.PrecoUnitario,
		}
	}

	input := domain.CreateRecebimentoCompraInput{
		OrdemCompraID:     req.OrdemCompraID,
		LojaID:            req.LojaID,
		DataRecebimento:   req.DataRecebimento,
		Observacoes:       req.Observacoes,
		PermitirExcedente: req.PermitirExcedente,
		Itens:             itens,
	}

	recebimento, err := h.recebimentoCompraUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdemCompraNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ordem de compra não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrQuantidadeExcedePedido) {
			RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Quantidade recebida excede a quantidade pedida", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, toRecebimentoCompraResponse(recebimento))
}

func (h *RecebimentoCompraHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	recebimento, err := h.recebimentoCompraUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrRecebimentoCompraNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Recebimento de compra não encontrado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toRecebimentoCompraResponse(recebimento))
}

func (h *RecebimentoCompraHandler) ListByOrdemCompra(w http.ResponseWriter, r *http.Request) {
	ordemCompraIDStr := chi.URLParam(r, "ordem_compra_id")
	ordemCompraID, err := uuid.Parse(ordemCompraIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID da ordem de compra inválido", "")
		return
	}

	recebimentos, err := h.recebimentoCompraUsecase.ListByOrdemCompra(r.Context(), ordemCompraID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]RecebimentoCompraResponse, len(recebimentos))
	for i, rc := range recebimentos {
		resp[i] = toRecebimentoCompraResponse(rc)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *RecebimentoCompraHandler) ListByFilter(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	var ordemCompraID, lojaID *uuid.UUID
	var dataInicio, dataFim *time.Time

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if ordemCompraIDStr := r.URL.Query().Get("ordem_compra_id"); ordemCompraIDStr != "" {
		if id, err := uuid.Parse(ordemCompraIDStr); err == nil {
			ordemCompraID = &id
		}
	}
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}
	if dataInicioStr := r.URL.Query().Get("data_inicio"); dataInicioStr != "" {
		if t, err := time.Parse(time.RFC3339, dataInicioStr); err == nil {
			dataInicio = &t
		}
	}
	if dataFimStr := r.URL.Query().Get("data_fim"); dataFimStr != "" {
		if t, err := time.Parse(time.RFC3339, dataFimStr); err == nil {
			dataFim = &t
		}
	}

	recebimentos, err := h.recebimentoCompraUsecase.ListByFilter(r.Context(), ordemCompraID, lojaID, dataInicio, dataFim, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]RecebimentoCompraResponse, len(recebimentos))
	for i, rc := range recebimentos {
		resp[i] = toRecebimentoCompraResponse(rc)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func toRecebimentoCompraResponse(rc *domain.RecebimentoCompra) RecebimentoCompraResponse {
	resp := RecebimentoCompraResponse{
		ID:                rc.ID,
		OrdemCompraID:     rc.OrdemCompraID,
		LojaID:            rc.LojaID,
		DataRecebimento:   rc.DataRecebimento,
		Observacoes:       rc.Observacoes,
		PermitirExcedente: rc.PermitirExcedente,
		CreatedAt:         rc.CreatedAt,
		UpdatedAt:         rc.UpdatedAt,
	}

	if rc.Itens != nil {
		resp.Itens = make([]RecebimentoCompraItemResponse, len(rc.Itens))
		for i, item := range rc.Itens {
			resp.Itens[i] = toRecebimentoCompraItemResponse(item)
		}
	}

	return resp
}

func toRecebimentoCompraItemResponse(item *domain.RecebimentoCompraItem) RecebimentoCompraItemResponse {
	return RecebimentoCompraItemResponse{
		ID:                  item.ID,
		RecebimentoCompraID: item.RecebimentoCompraID,
		OrdemCompraItemID:   item.OrdemCompraItemID,
		InsumoID:            item.InsumoID,
		QuantidadeRecebida:  item.QuantidadeRecebida,
		PrecoUnitario:       item.PrecoUnitario,
		CreatedAt:           item.CreatedAt,
	}
}

// ============================================================
// Conta a Pagar Handler
// ============================================================

type ContaPagarHandler struct {
	contaPagarUsecase *usecase.ContaPagarUsecase
}

func NewContaPagarHandler(contaPagarUsecase *usecase.ContaPagarUsecase) *ContaPagarHandler {
	return &ContaPagarHandler{contaPagarUsecase: contaPagarUsecase}
}

type CreateContaPagarRequest struct {
	LojaID              uuid.UUID  `json:"loja_id" validate:"required"`
	FornecedorID        uuid.UUID  `json:"fornecedor_id" validate:"required"`
	OrdemCompraID       *uuid.UUID `json:"ordem_compra_id,omitempty"`
	RecebimentoCompraID *uuid.UUID `json:"recebimento_compra_id,omitempty"`
	Descricao           string     `json:"descricao" validate:"required"`
	Valor               float64    `json:"valor" validate:"required,gt=0"`
	DataEmissao         time.Time  `json:"data_emissao" validate:"required"`
	DataVencimento      time.Time  `json:"data_vencimento" validate:"required"`
	Status              *string    `json:"status,omitempty"`
	FormaPagamento      *string    `json:"forma_pagamento,omitempty"`
	Observacoes         *string    `json:"observacoes,omitempty"`
}

type UpdateContaPagarRequest struct {
	Descricao      *string    `json:"descricao,omitempty"`
	Valor          *float64   `json:"valor,omitempty"`
	DataEmissao    *time.Time `json:"data_emissao,omitempty"`
	DataVencimento *time.Time `json:"data_vencimento,omitempty"`
	DataPagamento  *time.Time `json:"data_pagamento,omitempty"`
	Status         *string    `json:"status,omitempty"`
	FormaPagamento *string    `json:"forma_pagamento,omitempty"`
	Observacoes    *string    `json:"observacoes,omitempty"`
}

type PagarContaRequest struct {
	DataPagamento  time.Time `json:"data_pagamento" validate:"required"`
	FormaPagamento string    `json:"forma_pagamento" validate:"required"`
}

type ContaPagarResponse struct {
	ID                  uuid.UUID           `json:"id"`
	EmpresaID           uuid.UUID           `json:"empresa_id"`
	LojaID              uuid.UUID           `json:"loja_id"`
	FornecedorID        uuid.UUID           `json:"fornecedor_id"`
	OrdemCompraID       *uuid.UUID          `json:"ordem_compra_id,omitempty"`
	RecebimentoCompraID *uuid.UUID          `json:"recebimento_compra_id,omitempty"`
	Descricao           string              `json:"descricao"`
	Valor               float64             `json:"valor"`
	DataEmissao         time.Time           `json:"data_emissao"`
	DataVencimento      time.Time           `json:"data_vencimento"`
	DataPagamento       *time.Time          `json:"data_pagamento,omitempty"`
	Status              string              `json:"status"`
	FormaPagamento      *string             `json:"forma_pagamento,omitempty"`
	Observacoes         *string             `json:"observacoes,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	DeletedAt           *time.Time          `json:"deleted_at,omitempty"`
	Fornecedor          *FornecedorResponse `json:"fornecedor,omitempty"`
	Loja                *LojaResponse       `json:"loja,omitempty"`
}

func (h *ContaPagarHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateContaPagarRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	input := domain.CreateContaPagarInput{
		EmpresaID:           empresaID,
		LojaID:              req.LojaID,
		FornecedorID:        req.FornecedorID,
		OrdemCompraID:       req.OrdemCompraID,
		RecebimentoCompraID: req.RecebimentoCompraID,
		Descricao:           req.Descricao,
		Valor:               req.Valor,
		DataEmissao:         req.DataEmissao,
		DataVencimento:      req.DataVencimento,
		Status:              req.Status,
		FormaPagamento:      req.FormaPagamento,
		Observacoes:         req.Observacoes,
	}

	contaPagar, err := h.contaPagarUsecase.Create(r.Context(), input, usuario.ID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, toContaPagarResponse(contaPagar))
}

func (h *ContaPagarHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	contaPagar, err := h.contaPagarUsecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrContaPagarNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Conta a pagar não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toContaPagarResponse(contaPagar))
}

func (h *ContaPagarHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	limit := 20
	offset := 0
	var lojaID, fornecedorID *uuid.UUID
	var status *string
	var dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim *time.Time

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}
	if fornecedorIDStr := r.URL.Query().Get("fornecedor_id"); fornecedorIDStr != "" {
		if id, err := uuid.Parse(fornecedorIDStr); err == nil {
			fornecedorID = &id
		}
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status = &statusStr
	}
	if dataVencimentoInicioStr := r.URL.Query().Get("data_vencimento_inicio"); dataVencimentoInicioStr != "" {
		if t, err := time.Parse(time.RFC3339, dataVencimentoInicioStr); err == nil {
			dataVencimentoInicio = &t
		}
	}
	if dataVencimentoFimStr := r.URL.Query().Get("data_vencimento_fim"); dataVencimentoFimStr != "" {
		if t, err := time.Parse(time.RFC3339, dataVencimentoFimStr); err == nil {
			dataVencimentoFim = &t
		}
	}
	if dataPagamentoInicioStr := r.URL.Query().Get("data_pagamento_inicio"); dataPagamentoInicioStr != "" {
		if t, err := time.Parse(time.RFC3339, dataPagamentoInicioStr); err == nil {
			dataPagamentoInicio = &t
		}
	}
	if dataPagamentoFimStr := r.URL.Query().Get("data_pagamento_fim"); dataPagamentoFimStr != "" {
		if t, err := time.Parse(time.RFC3339, dataPagamentoFimStr); err == nil {
			dataPagamentoFim = &t
		}
	}

	// Parse categoria_financeira_id and conta_gerencial_id filters
	var categoriaFinanceiraID, contaGerencialID *uuid.UUID
	if catFinStr := r.URL.Query().Get("categoria_financeira_id"); catFinStr != "" {
		if id, err := uuid.Parse(catFinStr); err == nil {
			categoriaFinanceiraID = &id
		}
	}
	if contaGerStr := r.URL.Query().Get("conta_gerencial_id"); contaGerStr != "" {
		if id, err := uuid.Parse(contaGerStr); err == nil {
			contaGerencialID = &id
		}
	}

	contas, err := h.contaPagarUsecase.ListByFilter(r.Context(), empresaID, lojaID, fornecedorID, (*domain.ContaPagarStatus)(status), dataVencimentoInicio, dataVencimentoFim, dataPagamentoInicio, dataPagamentoFim, categoriaFinanceiraID, contaGerencialID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]ContaPagarResponse, len(contas))
	for i, cp := range contas {
		resp[i] = toContaPagarResponse(cp)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *ContaPagarHandler) ListByFornecedor(w http.ResponseWriter, r *http.Request) {
	fornecedorIDStr := chi.URLParam(r, "fornecedor_id")
	fornecedorID, err := uuid.Parse(fornecedorIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID do fornecedor inválido", "")
		return
	}

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	contas, err := h.contaPagarUsecase.ListByFornecedor(r.Context(), fornecedorID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]ContaPagarResponse, len(contas))
	for i, cp := range contas {
		resp[i] = toContaPagarResponse(cp)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *ContaPagarHandler) ListByStatus(w http.ResponseWriter, r *http.Request) {
	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	empresaID := GetEmpresaIDFromContext(r.Context())
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Empresa não identificada", "")
		return
	}

	statusStr := r.URL.Query().Get("status")
	if statusStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "status é obrigatório", "")
		return
	}

	status := domain.ContaPagarStatus(statusStr)

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	contas, err := h.contaPagarUsecase.ListByStatus(r.Context(), empresaID, status, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]ContaPagarResponse, len(contas))
	for i, cp := range contas {
		resp[i] = toContaPagarResponse(cp)
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *ContaPagarHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req UpdateContaPagarRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := domain.UpdateContaPagarInput{
		Descricao:      req.Descricao,
		Valor:          req.Valor,
		DataEmissao:    req.DataEmissao,
		DataVencimento: req.DataVencimento,
		DataPagamento:  req.DataPagamento,
		Status:         req.Status,
		FormaPagamento: req.FormaPagamento,
		Observacoes:    req.Observacoes,
	}

	contaPagar, err := h.contaPagarUsecase.Update(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrContaPagarNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Conta a pagar não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, toContaPagarResponse(contaPagar))
}

func (h *ContaPagarHandler) Pagar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	var req PagarContaRequest
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida", err.Error())
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	input := domain.PagarContaInput{
		DataPagamento:  req.DataPagamento,
		FormaPagamento: req.FormaPagamento,
	}

	contaPagar, err := h.contaPagarUsecase.Pagar(r.Context(), id, input, usuario.ID)
	if err != nil {
		if errors.Is(err, usecase.ErrContaPagarNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Conta a pagar não encontrada", "")
			return
		}
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), "")
		return
	}

	RespondJSON(w, http.StatusOK, toContaPagarResponse(contaPagar))
}

func (h *ContaPagarHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "ID inválido", "")
		return
	}

	usuario := GetUsuarioFromContext(r.Context())
	if usuario == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", "")
		return
	}

	if err := h.contaPagarUsecase.Delete(r.Context(), id, usuario.ID); err != nil {
		if errors.Is(err, usecase.ErrContaPagarNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Conta a pagar não encontrada", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func toContaPagarResponse(cp *domain.ContaPagar) ContaPagarResponse {
	resp := ContaPagarResponse{
		ID:                  cp.ID,
		EmpresaID:           cp.EmpresaID,
		LojaID:              cp.LojaID,
		FornecedorID:        cp.FornecedorID,
		OrdemCompraID:       cp.OrdemCompraID,
		RecebimentoCompraID: cp.RecebimentoCompraID,
		Descricao:           cp.Descricao,
		Valor:               cp.Valor,
		DataEmissao:         cp.DataEmissao,
		DataVencimento:      cp.DataVencimento,
		DataPagamento:       cp.DataPagamento,
		Status:              string(cp.Status),
		FormaPagamento:      cp.FormaPagamento,
		Observacoes:         cp.Observacoes,
		CreatedAt:           cp.CreatedAt,
		UpdatedAt:           cp.UpdatedAt,
		DeletedAt:           cp.DeletedAt,
	}

	if cp.Fornecedor != nil {
		fr := toFornecedorResponse(cp.Fornecedor)
		resp.Fornecedor = &fr
	}

	if cp.Loja != nil {
		lr := toLojaResponse(cp.Loja)
		resp.Loja = &lr
	}

	return resp
}

// ============================================================
// Helper: LojaResponse
// ============================================================

type LojaResponse struct {
	ID        uuid.UUID  `json:"id"`
	EmpresaID uuid.UUID  `json:"empresa_id"`
	Nome      string     `json:"nome"`
	Ativo     bool       `json:"ativo"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func toLojaResponse(l *domain.Loja) LojaResponse {
	return LojaResponse{
		ID:        l.ID,
		EmpresaID: l.EmpresaID,
		Nome:      l.Nome,
		Ativo:     l.Ativo,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
	}
}
