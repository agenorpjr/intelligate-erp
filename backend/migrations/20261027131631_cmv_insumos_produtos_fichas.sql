-- +goose Up
-- Etapa 7: Insumos, Produtos, Fichas Técnicas e CMV

-- Insumos
CREATE TABLE insumos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    unidade_medida VARCHAR(20) NOT NULL,
    custo_unitario_atual NUMERIC(12,4) NOT NULL DEFAULT 0,
    fornecedor VARCHAR(255),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, nome)
);

CREATE INDEX idx_insumos_empresa_id ON insumos(empresa_id);
CREATE INDEX idx_insumos_ativo ON insumos(ativo) WHERE ativo = true;

-- Histórico de Custos de Insumos
CREATE TABLE insumo_custos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    data_vigencia TIMESTAMPTZ NOT NULL,
    custo_unitario NUMERIC(12,4) NOT NULL,
    observacao TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (insumo_id, data_vigencia)
);

CREATE INDEX idx_insumo_custos_insumo_id ON insumo_custos(insumo_id);
CREATE INDEX idx_insumo_custos_data_vigencia ON insumo_custos(data_vigencia DESC);

-- Produtos (itens vendidos no PDV)
CREATE TABLE produtos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    sku_externo VARCHAR(100),
    categoria VARCHAR(100),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, sku_externo)
);

CREATE INDEX idx_produtos_empresa_id ON produtos(empresa_id);
CREATE INDEX idx_produtos_ativo ON produtos(ativo) WHERE ativo = true;
CREATE INDEX idx_produtos_sku_externo ON produtos(sku_externo);

-- Fichas Técnicas
CREATE TABLE fichas_tecnicas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    produto_id UUID NOT NULL REFERENCES produtos(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    rendimento NUMERIC(12,4) NOT NULL DEFAULT 1,
    unidade_rendimento VARCHAR(20) NOT NULL DEFAULT 'UN',
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (produto_id, loja_id)
);

CREATE INDEX idx_fichas_tecnicas_produto_id ON fichas_tecnicas(produto_id);
CREATE INDEX idx_fichas_tecnicas_loja_id ON fichas_tecnicas(loja_id);
CREATE INDEX idx_fichas_tecnicas_ativo ON fichas_tecnicas(ativo) WHERE ativo = true;

-- Componentes da Ficha Técnica
CREATE TABLE ficha_tecnica_componentes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ficha_tecnica_id UUID NOT NULL REFERENCES fichas_tecnicas(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    quantidade NUMERIC(12,4) NOT NULL,
    unidade_medida VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ficha_tecnica_id, insumo_id)
);

CREATE INDEX idx_ficha_tecnica_componentes_ficha_id ON ficha_tecnica_componentes(ficha_tecnica_id);
CREATE INDEX idx_ficha_tecnica_componentes_insumo_id ON ficha_tecnica_componentes(insumo_id);

-- Trigger function for updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- Apply updated_at triggers
CREATE TRIGGER update_insumos_updated_at BEFORE UPDATE ON insumos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_produtos_updated_at BEFORE UPDATE ON produtos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fichas_tecnicas_updated_at BEFORE UPDATE ON fichas_tecnicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ficha_tecnica_componentes_updated_at BEFORE UPDATE ON ficha_tecnica_componentes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_ficha_tecnica_componentes_updated_at ON ficha_tecnica_componentes;
DROP TRIGGER IF EXISTS update_fichas_tecnicas_updated_at ON fichas_tecnicas;
DROP TRIGGER IF EXISTS update_produtos_updated_at ON produtos;
DROP TRIGGER IF EXISTS update_insumos_updated_at ON insumos;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS ficha_tecnica_componentes;
DROP TABLE IF EXISTS fichas_tecnicas;
DROP TABLE IF EXISTS produtos;
DROP TABLE IF EXISTS insumo_custos;
DROP TABLE IF EXISTS insumos;