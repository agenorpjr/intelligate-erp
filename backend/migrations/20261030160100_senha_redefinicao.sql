-- +goose Up
-- Tabela de tokens de redefinicao de senha (esqueci-senha / redefinir-senha)
CREATE TABLE senha_redefinicoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    usado BOOLEAN NOT NULL DEFAULT false,
    expira_em TIMESTAMPTZ NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    usado_em TIMESTAMPTZ
);

CREATE INDEX idx_senha_redefinicoes_usuario_id ON senha_redefinicoes(usuario_id);
CREATE INDEX idx_senha_redefinicoes_token_hash ON senha_redefinicoes(token_hash);
CREATE INDEX idx_senha_redefinicoes_expira_em ON senha_redefinicoes(expira_em);

-- +goose Down
DROP TABLE IF EXISTS senha_redefinicoes;
