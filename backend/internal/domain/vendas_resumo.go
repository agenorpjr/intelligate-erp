package domain

type VendasResumo struct {
	TotalVendas    float64 `json:"total_vendas"`
	TotalPedidos   int64   `json:"total_pedidos"`
	TicketMedio    float64 `json:"ticket_medio"`
	TotalDescontos float64 `json:"total_descontos"`
}
