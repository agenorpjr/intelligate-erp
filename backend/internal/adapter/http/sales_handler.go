package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type SalesHandler struct {
	salesUsecase *usecase.SalesUsecase
}

func NewSalesHandler(salesUsecase *usecase.SalesUsecase) *SalesHandler {
	return &SalesHandler{salesUsecase: salesUsecase}
}

type IngestSalesEventRequest struct {
	TipoEvento     string                 `json:"tipo_evento" validate:"required"`
	Payload        map[string]interface{} `json:"payload" validate:"required"`
	IdempotencyKey string                 `json:"idempotency_key" validate:"required"`
}

type IngestSalesEventResponse struct {
	EventoID     uuid.UUID `json:"evento_id"`
	PedidoID     uuid.UUID `json:"pedido_id"`
	JaProcessado bool      `json:"ja_processado"`
}

func (h *SalesHandler) IngestSalesEvent(w http.ResponseWriter, r *http.Request) {
	var req IngestSalesEventRequest
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

	// Get loja_id from payload or use user's default loja
	lojaID := uuid.Nil
	if lojaIDStr, ok := req.Payload["loja_id"].(string); ok {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = id
		}
	}
	if lojaID == uuid.Nil {
		// Try to get from user's associated loja
		// For now, require it in payload
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "loja_id é obrigatório no payload", "")
		return
	}

	// Get pdv_origem_id from payload
	pdvOrigemID := uuid.Nil
	if pdvOrigemIDStr, ok := req.Payload["pdv_origem_id"].(string); ok {
		if id, err := uuid.Parse(pdvOrigemIDStr); err == nil {
			pdvOrigemID = id
		}
	}
	if pdvOrigemID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "pdv_origem_id é obrigatório no payload", "")
		return
	}

	input := usecase.IngestSalesEventInput{
		EmpresaID:      usuario.EmpresaID,
		LojaID:         lojaID,
		PdvOrigemID:    pdvOrigemID,
		TipoEvento:     req.TipoEvento,
		Payload:        req.Payload,
		IdempotencyKey: req.IdempotencyKey,
	}

	output, err := h.salesUsecase.IngestSalesEvent(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrPdvOrigemNaoEncontrado) {
			RespondError(w, http.StatusNotFound, "PDV_NOT_FOUND", "PDV de origem não encontrado", "")
			return
		}
		if errors.Is(err, usecase.ErrLojaNaoEncontrada) {
			RespondError(w, http.StatusNotFound, "LOJA_NOT_FOUND", "Loja não encontrada", "")
			return
		}
		if errors.Is(err, usecase.ErrEventoJaProcessado) {
			RespondError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Evento já processado", "")
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := IngestSalesEventResponse{
		EventoID:     output.EventoID,
		PedidoID:     output.PedidoID,
		JaProcessado: output.JaProcessado,
	}

	statusCode := http.StatusCreated
	if output.JaProcessado {
		statusCode = http.StatusOK
	}
	RespondJSON(w, statusCode, resp)
}

type VendasResumoResponse struct {
	TotalVendas    float64 `json:"total_vendas"`
	TotalPedidos   int64   `json:"total_pedidos"`
	TicketMedio    float64 `json:"ticket_medio"`
	TotalDescontos float64 `json:"total_descontos"`
}

func (h *SalesHandler) GetVendasResumo(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	_ = r.URL.Query().Get("loja_id")
	_ = r.URL.Query().Get("data_inicio")
	_ = r.URL.Query().Get("data_fim")

	// For now, return empty - would need proper repository implementation
	resp := VendasResumoResponse{
		TotalVendas:    0,
		TotalPedidos:   0,
		TicketMedio:    0,
		TotalDescontos: 0,
	}

	RespondJSON(w, http.StatusOK, resp)
}

type ListPedidosResponse struct {
	Pedidos []*domain.PedidoCanonico `json:"pedidos"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
}

func (h *SalesHandler) ListPedidos(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	limit := 20
	offset := 0

	var lojaID *uuid.UUID
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}

	var dataInicio, dataFim *time.Time
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

	var status *string
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status = &statusStr
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filter := domain.PedidoCanonicoFilter{
		EmpresaID:  &usuario.EmpresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	}

	pedidos, err := h.salesUsecase.ListPedidos(r.Context(), filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := ListPedidosResponse{
		Pedidos: pedidos,
		Total:   len(pedidos),
		Limit:   limit,
		Offset:  offset,
	}

	RespondJSON(w, http.StatusOK, resp)
}
