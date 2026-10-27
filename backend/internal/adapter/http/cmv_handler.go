package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/usecase"
)

type CMVHandler struct {
	fichaUsecase *usecase.FichaTecnicaUsecase
}

func NewCMVHandler(fichaUsecase *usecase.FichaTecnicaUsecase) *CMVHandler {
	return &CMVHandler{fichaUsecase: fichaUsecase}
}

type CMVResumoResponse struct {
	EmpresaID          uuid.UUID `json:"empresa_id"`
	LojaID             uuid.UUID `json:"loja_id"`
	DataInicio         time.Time `json:"data_inicio"`
	DataFim            time.Time `json:"data_fim"`
	TotalVendas        float64   `json:"total_vendas"`
	TotalCMV           float64   `json:"total_cmv"`
	MargemBruta        float64   `json:"margem_bruta"`
	MargemBrutaPercent float64   `json:"margem_bruta_percent"`
	ProdutosComFicha   int       `json:"produtos_com_ficha"`
	ProdutosSemFicha   int       `json:"produtos_sem_ficha_tecnica"`
	ItensVendidos      int       `json:"itens_vendidos"`
}

type CMVProdutoDetalheResponse struct {
	ProdutoID            uuid.UUID `json:"produto_id"`
	ProdutoNome          string    `json:"produto_nome"`
	ProdutoSKU           *string   `json:"produto_sku,omitempty"`
	QuantidadeVendida    float64   `json:"quantidade_vendida"`
	CustoUnitarioTeorico float64   `json:"custo_unitario_teorico"`
	CMVTotal             float64   `json:"cmv_total"`
	ReceitaTotal         float64   `json:"receita_total"`
	MargemBruta          float64   `json:"margem_bruta"`
	MargemBrutaPercent   float64   `json:"margem_bruta_percent"`
	TemFichaTecnica      bool      `json:"tem_ficha_tecnica"`
}

func (h *CMVHandler) GetCMVResumo(w http.ResponseWriter, r *http.Request) {
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
	var lojaID *uuid.UUID
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}

	dataInicioStr := r.URL.Query().Get("data_inicio")
	dataFimStr := r.URL.Query().Get("data_fim")

	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}

	dataInicio, err := time.Parse(time.RFC3339, dataInicioStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio inválida (use RFC3339)", "")
		return
	}

	dataFim, err := time.Parse(time.RFC3339, dataFimStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_fim inválida (use RFC3339)", "")
		return
	}

	filter := usecase.CMVFilter{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	}

	resumo, err := h.fichaUsecase.GetCMVResumo(r.Context(), filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := CMVResumoResponse{
		EmpresaID:          resumo.EmpresaID,
		LojaID:             resumo.LojaID,
		DataInicio:         resumo.DataInicio,
		DataFim:            resumo.DataFim,
		TotalVendas:        resumo.TotalVendas,
		TotalCMV:           resumo.TotalCMV,
		MargemBruta:        resumo.MargemBruta,
		MargemBrutaPercent: resumo.MargemBrutaPercent,
		ProdutosComFicha:   resumo.ProdutosComFicha,
		ProdutosSemFicha:   resumo.ProdutosSemFicha,
		ItensVendidos:      resumo.ItensVendidos,
	}

	RespondJSON(w, http.StatusOK, resp)
}

func (h *CMVHandler) GetCMVProdutosDetalhe(w http.ResponseWriter, r *http.Request) {
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
	var lojaID *uuid.UUID
	if lojaIDStr := r.URL.Query().Get("loja_id"); lojaIDStr != "" {
		if id, err := uuid.Parse(lojaIDStr); err == nil {
			lojaID = &id
		}
	}

	dataInicioStr := r.URL.Query().Get("data_inicio")
	dataFimStr := r.URL.Query().Get("data_fim")

	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}

	dataInicio, err := time.Parse(time.RFC3339, dataInicioStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio inválida (use RFC3339)", "")
		return
	}

	dataFim, err := time.Parse(time.RFC3339, dataFimStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_fim inválida (use RFC3339)", "")
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

	filter := usecase.CMVFilter{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Limit:      limit,
		Offset:     offset,
	}

	detalhes, err := h.fichaUsecase.GetCMVProdutosDetalhe(r.Context(), filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", err.Error())
		return
	}

	resp := make([]CMVProdutoDetalheResponse, len(detalhes))
	for i, d := range detalhes {
		resp[i] = CMVProdutoDetalheResponse{
			ProdutoID:            d.ProdutoID,
			ProdutoNome:          d.ProdutoNome,
			ProdutoSKU:           d.ProdutoSKU,
			QuantidadeVendida:    d.QuantidadeVendida,
			CustoUnitarioTeorico: d.CustoUnitarioTeorico,
			CMVTotal:             d.CMVTotal,
			ReceitaTotal:         d.ReceitaTotal,
			MargemBruta:          d.MargemBruta,
			MargemBrutaPercent:   d.MargemBrutaPercent,
			TemFichaTecnica:      d.TemFichaTecnica,
		}
	}

	RespondJSON(w, http.StatusOK, resp)
}
