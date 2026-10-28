package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/usecase"
)

// ============================================================
// Dashboard Operacional Handler
// ============================================================

type DashboardHandler struct {
	usecase *usecase.DashboardUsecase
}

func NewDashboardHandler(usecase *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{usecase: usecase}
}

// GetDashboard trata GET /api/v1/dashboard?loja_id=...&data_inicio=...&data_fim=...
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	empresaID := getEmpresaID(r)
	if empresaID == uuid.Nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "empresa_id é obrigatório", "")
		return
	}

	q := r.URL.Query()

	lojaStr := q.Get("loja_id")
	if lojaStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "loja_id é obrigatório", "")
		return
	}
	lojaID, err := uuid.Parse(lojaStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "loja_id inválido", "")
		return
	}

	dataInicioStr := q.Get("data_inicio")
	dataFimStr := q.Get("data_fim")
	if dataInicioStr == "" || dataFimStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "data_inicio e data_fim são obrigatórios", "")
		return
	}

	dataInicio, err := parseDashboardDate(dataInicioStr, false)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_inicio inválida (usar RFC3339 ou YYYY-MM-DD)", "")
		return
	}
	dataFim, err := parseDashboardDate(dataFimStr, true)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_DATE", "data_fim inválida (usar RFC3339 ou YYYY-MM-DD)", "")
		return
	}

	dashboard, err := h.usecase.GetDashboard(r.Context(), usecase.DashboardRequest{
		EmpresaID:  empresaID,
		LojaID:     lojaID,
		DataInicio: dataInicio,
		DataFim:    dataFim,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrDashboardLojaObrigatoria),
			errors.Is(err, usecase.ErrDashboardPeriodoInvalido):
			RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Parâmetros inválidos", err.Error())
		case errors.Is(err, usecase.ErrDashboardLojaNaoPertence):
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Loja não encontrada", err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao gerar dashboard", err.Error())
		}
		return
	}

	RespondJSON(w, http.StatusOK, dashboard)
}

// parseDashboardDate aceita datas em RFC3339 ou no formato YYYY-MM-DD. Quando o
// valor é uma data sem hora e endOfDay=true, a data é estendida até o fim do dia.
func parseDashboardDate(value string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato de data inválido")
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return t, nil
}
