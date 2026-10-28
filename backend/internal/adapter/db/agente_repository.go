package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgenteRepository implementa domain.AgenteRepository sobre PostgreSQL.
type AgenteRepository struct {
	querier *Queries
}

// NewAgenteRepository cria uma nova instância de AgenteRepository.
func NewAgenteRepository(pool *pgxpool.Pool) *AgenteRepository {
	return &AgenteRepository{querier: New(pool)}
}

// CreateSessao insere uma nova sessão de conversa do agente.
func (r *AgenteRepository) CreateSessao(ctx context.Context, s *domain.AgenteSessao) error {
	row, err := r.querier.CreateAgenteSessao(ctx, CreateAgenteSessaoParams{
		EmpresaID: s.EmpresaID,
		LojaID:    s.LojaID,
		UsuarioID: s.UsuarioID,
		Titulo:    s.Titulo,
		Status:    s.Status,
	})
	if err != nil {
		return err
	}
	*s = *agenteSessaoToDomain(row)
	return nil
}

// GetSessao retorna uma sessão pelo ID. Retorna (nil, nil) quando não existe.
func (r *AgenteRepository) GetSessao(ctx context.Context, id uuid.UUID) (*domain.AgenteSessao, error) {
	row, err := r.querier.GetAgenteSessao(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return agenteSessaoToDomain(row), nil
}

// ListSessoes lista as sessões da empresa com filtros opcionais.
func (r *AgenteRepository) ListSessoes(ctx context.Context, filter domain.AgenteSessaoFilter) ([]*domain.AgenteSessao, error) {
	var status pgtype.Text
	if filter.Status != nil {
		status = pgtype.Text{String: *filter.Status, Valid: true}
	}

	rows, err := r.querier.ListAgenteSessoes(ctx, ListAgenteSessoesParams{
		EmpresaID:   filter.EmpresaID,
		LojaID:      pgUUIDFromPtr(filter.LojaID),
		UsuarioID:   pgUUIDFromPtr(filter.UsuarioID),
		Status:      status,
		LimitCount:  int32(filter.Limit),
		OffsetCount: int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	sessoes := make([]*domain.AgenteSessao, len(rows))
	for i, row := range rows {
		sessoes[i] = agenteSessaoToDomain(row)
	}
	return sessoes, nil
}

// UpdateSessaoStatus atualiza o status de uma sessão.
func (r *AgenteRepository) UpdateSessaoStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.querier.UpdateAgenteSessaoStatus(ctx, UpdateAgenteSessaoStatusParams{
		ID:     id,
		Status: status,
	})
}

// CreateMensagem insere uma mensagem na sessão.
func (r *AgenteRepository) CreateMensagem(ctx context.Context, m *domain.AgenteMensagem) error {
	var contexto []byte
	if len(m.ContextoUsado) > 0 {
		contexto = []byte(m.ContextoUsado)
	}

	row, err := r.querier.CreateAgenteMensagem(ctx, CreateAgenteMensagemParams{
		SessaoID:        m.SessaoID,
		Papel:           m.Papel,
		Conteudo:        m.Conteudo,
		ContextoUsado:   contexto,
		TokensEstimados: int32(m.TokensEstimados),
	})
	if err != nil {
		return err
	}
	*m = *agenteMensagemToDomain(row)
	return nil
}

// ListMensagensBySessao lista as mensagens de uma sessão em ordem cronológica.
func (r *AgenteRepository) ListMensagensBySessao(ctx context.Context, sessaoID uuid.UUID) ([]*domain.AgenteMensagem, error) {
	rows, err := r.querier.ListAgenteMensagensBySessao(ctx, sessaoID)
	if err != nil {
		return nil, err
	}
	mensagens := make([]*domain.AgenteMensagem, len(rows))
	for i, row := range rows {
		mensagens[i] = agenteMensagemToDomain(row)
	}
	return mensagens, nil
}

// ============================================================
// Conversões
// ============================================================

func agenteSessaoToDomain(row AgenteSesso) *domain.AgenteSessao {
	return &domain.AgenteSessao{
		ID:        row.ID,
		EmpresaID: row.EmpresaID,
		LojaID:    row.LojaID,
		UsuarioID: row.UsuarioID,
		Titulo:    row.Titulo,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func agenteMensagemToDomain(row AgenteMensagen) *domain.AgenteMensagem {
	var contexto json.RawMessage
	if len(row.ContextoUsado) > 0 {
		contexto = json.RawMessage(row.ContextoUsado)
	}
	return &domain.AgenteMensagem{
		ID:              row.ID,
		SessaoID:        row.SessaoID,
		Papel:           row.Papel,
		Conteudo:        row.Conteudo,
		ContextoUsado:   contexto,
		TokensEstimados: int(row.TokensEstimados),
		CreatedAt:       row.CreatedAt,
	}
}
