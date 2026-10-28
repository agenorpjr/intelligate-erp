-- +goose Up
-- Etapa 13: Agente de IA para análise gerencial
-- Sessões de conversa e mensagens do agente. O contexto usado em cada
-- interação é persistido para auditoria e rastreabilidade. Nenhum segredo
-- (API key, token) é armazenado nestas tabelas.

-- Sessões de conversa do agente
CREATE TABLE agente_sessoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE RESTRICT,
    titulo VARCHAR(255) NOT NULL,
    -- Status da sessão: 'ativa' ou 'arquivada' (soft delete lógico)
    status VARCHAR(30) NOT NULL DEFAULT 'ativa',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_agente_sessoes_status CHECK (status IN ('ativa', 'arquivada'))
);

CREATE INDEX idx_agente_sessoes_empresa_id ON agente_sessoes(empresa_id);
CREATE INDEX idx_agente_sessoes_loja_id ON agente_sessoes(loja_id);
CREATE INDEX idx_agente_sessoes_usuario_id ON agente_sessoes(usuario_id);
CREATE INDEX idx_agente_sessoes_status ON agente_sessoes(status);
CREATE INDEX idx_agente_sessoes_created_at ON agente_sessoes(created_at DESC);

-- Mensagens da conversa (append-only, sem exclusão física)
CREATE TABLE agente_mensagens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sessao_id UUID NOT NULL REFERENCES agente_sessoes(id) ON DELETE CASCADE,
    -- Papel do autor: 'usuario' ou 'agente'
    papel VARCHAR(20) NOT NULL,
    conteudo TEXT NOT NULL,
    -- Contexto estruturado (dados reais) usado para gerar a resposta.
    -- Armazenado para auditoria/rastreabilidade; nunca contém segredos.
    contexto_usado JSONB,
    tokens_estimados INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_agente_mensagens_papel CHECK (papel IN ('usuario', 'agente'))
);

CREATE INDEX idx_agente_mensagens_sessao_id ON agente_mensagens(sessao_id);
CREATE INDEX idx_agente_mensagens_papel ON agente_mensagens(papel);
CREATE INDEX idx_agente_mensagens_created_at ON agente_mensagens(created_at);

-- Trigger para manter updated_at da sessão
CREATE TRIGGER update_agente_sessoes_updated_at BEFORE UPDATE ON agente_sessoes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_agente_sessoes_updated_at ON agente_sessoes;

DROP TABLE IF EXISTS agente_mensagens;
DROP TABLE IF EXISTS agente_sessoes;
