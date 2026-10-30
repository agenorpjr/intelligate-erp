-- +goose Up
-- Tabela de associação usuário-loja (acesso direto a lojas, independente de papéis)
CREATE TABLE usuario_lojas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (usuario_id, loja_id)
);

CREATE INDEX idx_usuario_lojas_usuario_id ON usuario_lojas(usuario_id);
CREATE INDEX idx_usuario_lojas_loja_id ON usuario_lojas(loja_id);

-- Trigger para updated_at (se necessário no futuro)
-- CREATE TRIGGER update_usuario_lojas_updated_at BEFORE UPDATE ON usuario_lojas
--     FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TABLE IF EXISTS usuario_lojas;