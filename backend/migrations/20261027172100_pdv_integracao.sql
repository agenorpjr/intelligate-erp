-- +goose Up
-- Integrações PDV Externas (Gula Menu, etc.)

CREATE TABLE pdv_integracoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    provedor VARCHAR(50) NOT NULL DEFAULT 'gula',
    ambiente VARCHAR(20) NOT NULL DEFAULT 'sandbox',
    url_base VARCHAR(500) NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    client_id VARCHAR(255),
    client_secret_encrypted TEXT,
    ativo BOOLEAN NOT NULL DEFAULT true,
    ultima_sincronizacao TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (loja_id, provedor)
);

CREATE INDEX idx_pdv_integracoes_empresa_id ON pdv_integracoes(empresa_id);
CREATE INDEX idx_pdv_integracoes_loja_id ON pdv_integracoes(loja_id);
CREATE INDEX idx_pdv_integracoes_provedor ON pdv_integracoes(provedor);
CREATE INDEX idx_pdv_integracoes_ativo ON pdv_integracoes(ativo) WHERE ativo = true;

-- Trigger for updated_at
CREATE TRIGGER update_pdv_integracoes_updated_at BEFORE UPDATE ON pdv_integracoes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_pdv_integracoes_updated_at ON pdv_integracoes;
DROP TABLE IF EXISTS pdv_integracoes;