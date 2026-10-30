-- +goose Up
-- Adiciona campo tecnico imutavel 'codigo' à tabela de papeis
-- para padronizar identificacao do papel administrador (ex: 'ADMIN').
ALTER TABLE papeis ADD COLUMN codigo VARCHAR(50);

-- Unicidade de codigo por empresa
CREATE UNIQUE INDEX idx_papeis_codigo_empresa_unique ON papeis (empresa_id, codigo) WHERE deleted_at IS NULL;

-- Index para busca por codigo
CREATE INDEX idx_papeis_codigo ON papeis (codigo);

-- +goose Down
DROP INDEX IF EXISTS idx_papeis_codigo;
DROP INDEX IF EXISTS idx_papeis_codigo_empresa_unique;
ALTER TABLE papeis DROP COLUMN IF EXISTS codigo;
