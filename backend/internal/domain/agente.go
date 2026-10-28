package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Agente de IA para análise gerencial (Etapa 13)
// ============================================================

// Status possíveis de uma sessão de conversa do agente.
const (
	AgenteSessaoStatusAtiva     = "ativa"
	AgenteSessaoStatusArquivada = "arquivada"
)

// Papéis possíveis de uma mensagem.
const (
	AgentePapelUsuario = "usuario"
	AgentePapelAgente  = "agente"
)

// AgenteSessao representa uma sessão de conversa do agente gerencial.
type AgenteSessao struct {
	ID        uuid.UUID `json:"id"`
	EmpresaID uuid.UUID `json:"empresa_id"`
	LojaID    uuid.UUID `json:"loja_id"`
	UsuarioID uuid.UUID `json:"usuario_id"`
	Titulo    string    `json:"titulo"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgenteMensagem representa uma mensagem trocada em uma sessão.
//
// ContextoUsado armazena o contexto estruturado (dados reais) utilizado para
// gerar a resposta, para fins de auditoria e rastreabilidade. Nunca contém
// segredos (API keys, tokens).
type AgenteMensagem struct {
	ID              uuid.UUID       `json:"id"`
	SessaoID        uuid.UUID       `json:"sessao_id"`
	Papel           string          `json:"papel"`
	Conteudo        string          `json:"conteudo"`
	ContextoUsado   json.RawMessage `json:"contexto_usado,omitempty"`
	TokensEstimados int             `json:"tokens_estimados"`
	CreatedAt       time.Time       `json:"created_at"`
}

// AgenteSessaoFilter filtra a listagem de sessões.
type AgenteSessaoFilter struct {
	EmpresaID uuid.UUID
	LojaID    *uuid.UUID
	UsuarioID *uuid.UUID
	Status    *string
	Limit     int
	Offset    int
}

// AgenteSessaoDetalhe agrega a sessão e suas mensagens.
type AgenteSessaoDetalhe struct {
	Sessao    *AgenteSessao     `json:"sessao"`
	Mensagens []*AgenteMensagem `json:"mensagens"`
}

// ============================================================
// Contexto estruturado fornecido ao agente
// ============================================================

// AgenteContexto reúne os dados reais da loja/período usados como base para a
// análise. Reutiliza os tipos do dashboard e do estoque para evitar duplicação.
type AgenteContexto struct {
	EmpresaID  uuid.UUID `json:"empresa_id"`
	LojaID     uuid.UUID `json:"loja_id"`
	DataInicio time.Time `json:"data_inicio"`
	DataFim    time.Time `json:"data_fim"`

	Vendas     DashboardVendas     `json:"vendas"`
	CMV        DashboardCMV        `json:"cmv"`
	Financeiro DashboardFinanceiro `json:"financeiro"`
	DRE        DashboardDRE        `json:"dre"`

	// ProdutosSemFichaTecnica lista produtos vendidos sem ficha técnica.
	ProdutosSemFichaTecnica []*CMVProdutoDetalhe `json:"produtos_sem_ficha_tecnica"`

	// Perdas consolida perdas e desperdícios de estoque no período.
	Perdas ResumoPerdas `json:"perdas"`

	// Alertas resume pontos de atenção derivados dos indicadores.
	Alertas []string `json:"alertas"`
}

// ============================================================
// Resposta estruturada do agente
// ============================================================

// AgenteResposta é o formato estruturado devolvido pelo agente.
type AgenteResposta struct {
	Resumo            string   `json:"resumo"`
	Diagnostico       string   `json:"diagnostico"`
	Evidencias        []string `json:"evidencias"`
	AcoesRecomendadas []string `json:"acoes_recomendadas"`
	Limitacoes        []string `json:"limitacoes"`
}

// ============================================================
// Entradas dos casos de uso
// ============================================================

// CriarAgenteSessaoInput cria uma nova sessão de conversa.
type CriarAgenteSessaoInput struct {
	EmpresaID uuid.UUID `json:"-"`
	LojaID    uuid.UUID `json:"loja_id" validate:"required"`
	UsuarioID uuid.UUID `json:"-"`
	Titulo    string    `json:"titulo"`
}

// PerguntarAgenteInput registra uma pergunta do gestor em uma sessão.
type PerguntarAgenteInput struct {
	EmpresaID  uuid.UUID `json:"-"`
	UsuarioID  uuid.UUID `json:"-"`
	SessaoID   uuid.UUID `json:"-"`
	Pergunta   string    `json:"pergunta" validate:"required"`
	DataInicio time.Time `json:"data_inicio"`
	DataFim    time.Time `json:"data_fim"`
}
