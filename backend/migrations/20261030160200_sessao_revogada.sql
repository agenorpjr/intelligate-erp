-- +goose Up
-- Tabela de tokens JWT revogados (logout / invalidez de sessao)
CREATE TABLE sessoes_revogadas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    jti VARCHAR(255) NOT NULL UNIQUE,
    expira_em TIMESTAMPTZ NOT NULL,
    revogado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    motivo VARCHAR(100) NOT NULL DEFAULT 'logout'
);

CREATE INDEX idx_sessoes_revogadas_usuario_id ON sessoes_revogadas(usuario_id);
CREATE INDEX idx_sessoes_revogadas_jti ON sessoes_revogadas(jti);
CREATE INDEX idx_sessoes_revogadas_expira_em ON sessoes_revogadas(expira_em);

-- +goose Down
DROP TABLE IF EXISTS sessoes_revogadas;
