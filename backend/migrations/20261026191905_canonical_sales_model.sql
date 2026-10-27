-- +goose Up
-- Modelo Canônico de Vendas e Ingestão de Eventos de PDV

-- PDVs de Origem
CREATE TABLE pdvs_origem (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(100) NOT NULL,
    codigo VARCHAR(50) NOT NULL,
    tipo VARCHAR(50) NOT NULL DEFAULT 'gula',
    configuracao JSONB,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, codigo)
);

CREATE INDEX idx_pdvs_origem_empresa_id ON pdvs_origem(empresa_id);
CREATE INDEX idx_pdvs_origem_ativo ON pdvs_origem(ativo) WHERE ativo = true;

-- Categorias Canônicas
CREATE TABLE categorias_canonicas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    cor VARCHAR(7),
    icone VARCHAR(50),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, nome)
);

CREATE INDEX idx_categorias_canonicas_empresa_id ON categorias_canonicas(empresa_id);
CREATE INDEX idx_categorias_canonicas_ativo ON categorias_canonicas(ativo) WHERE ativo = true;

-- Produtos Canônicos
CREATE TABLE produtos_canonicos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    categoria_id UUID REFERENCES categorias_canonicas(id) ON DELETE SET NULL,
    sku VARCHAR(100) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    unidade VARCHAR(20) NOT NULL DEFAULT 'UN',
    preco_base NUMERIC(12,4) NOT NULL DEFAULT 0,
    ncm VARCHAR(8),
    cest VARCHAR(7),
    tributacao VARCHAR(50),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, sku)
);

CREATE INDEX idx_produtos_canonicos_empresa_id ON produtos_canonicos(empresa_id);
CREATE INDEX idx_produtos_canonicos_categoria_id ON produtos_canonicos(categoria_id);
CREATE INDEX idx_produtos_canonicos_ativo ON produtos_canonicos(ativo) WHERE ativo = true;
CREATE INDEX idx_produtos_canonicos_sku ON produtos_canonicos(sku);

-- Pedidos Canônicos
CREATE TABLE pedidos_canonicos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    pdv_origem_id UUID NOT NULL REFERENCES pdvs_origem(id) ON DELETE RESTRICT,
    numero_externo VARCHAR(100) NOT NULL,
    data_hora TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pendente',
    valor_total NUMERIC(12,4) NOT NULL DEFAULT 0,
    valor_desconto NUMERIC(12,4) NOT NULL DEFAULT 0,
    valor_acrescimo NUMERIC(12,4) NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(255) NOT NULL,
    cliente_nome VARCHAR(255),
    cliente_documento VARCHAR(20),
    cliente_email VARCHAR(255),
    observacoes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, idempotency_key)
);

CREATE INDEX idx_pedidos_canonicos_empresa_id ON pedidos_canonicos(empresa_id);
CREATE INDEX idx_pedidos_canonicos_loja_id ON pedidos_canonicos(loja_id);
CREATE INDEX idx_pedidos_canonicos_pdv_origem_id ON pedidos_canonicos(pdv_origem_id);
CREATE INDEX idx_pedidos_canonicos_data_hora ON pedidos_canonicos(data_hora DESC);
CREATE INDEX idx_pedidos_canonicos_status ON pedidos_canonicos(status);
CREATE INDEX idx_pedidos_canonicos_idempotency_key ON pedidos_canonicos(idempotency_key);

-- Itens de Pedido Canônicos
CREATE TABLE pedido_itens_canonicos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pedido_id UUID NOT NULL REFERENCES pedidos_canonicos(id) ON DELETE CASCADE,
    produto_canonico_id UUID REFERENCES produtos_canonicos(id) ON DELETE SET NULL,
    sku_externo VARCHAR(100),
    nome VARCHAR(255) NOT NULL,
    quantidade NUMERIC(12,4) NOT NULL,
    preco_unitario NUMERIC(12,4) NOT NULL,
    valor_total NUMERIC(12,4) NOT NULL,
    desconto NUMERIC(12,4) NOT NULL DEFAULT 0,
    modificadores JSONB,
    observacoes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_pedido_itens_canonicos_pedido_id ON pedido_itens_canonicos(pedido_id);
CREATE INDEX idx_pedido_itens_canonicos_produto_canonico_id ON pedido_itens_canonicos(produto_canonico_id);

-- Pagamentos de Pedido Canônicos
CREATE TABLE pedido_pagamentos_canonicos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pedido_id UUID NOT NULL REFERENCES pedidos_canonicos(id) ON DELETE CASCADE,
    forma_pagamento VARCHAR(50) NOT NULL,
    valor NUMERIC(12,4) NOT NULL,
    taxa NUMERIC(12,4) NOT NULL DEFAULT 0,
    troco NUMERIC(12,4) NOT NULL DEFAULT 0,
    referencia_externa VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'aprovado',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_pedido_pagamentos_canonicos_pedido_id ON pedido_pagamentos_canonicos(pedido_id);
CREATE INDEX idx_pedido_pagamentos_canonicos_forma_pagamento ON pedido_pagamentos_canonicos(forma_pagamento);
CREATE INDEX idx_pedido_pagamentos_canonicos_referencia_externa ON pedido_pagamentos_canonicos(referencia_externa);

-- Eventos de Ingestão
CREATE TABLE eventos_ingestao (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    pdv_origem_id UUID NOT NULL REFERENCES pdvs_origem(id) ON DELETE RESTRICT,
    tipo_evento VARCHAR(100) NOT NULL,
    payload_original JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pendente',
    erro TEXT,
    tentativas INTEGER NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(255) NOT NULL,
    processado_em TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, idempotency_key)
);

CREATE INDEX idx_eventos_ingestao_empresa_id ON eventos_ingestao(empresa_id);
CREATE INDEX idx_eventos_ingestao_loja_id ON eventos_ingestao(loja_id);
CREATE INDEX idx_eventos_ingestao_pdv_origem_id ON eventos_ingestao(pdv_origem_id);
CREATE INDEX idx_eventos_ingestao_tipo_evento ON eventos_ingestao(tipo_evento);
CREATE INDEX idx_eventos_ingestao_status ON eventos_ingestao(status);
CREATE INDEX idx_eventos_ingestao_idempotency_key ON eventos_ingestao(idempotency_key);
CREATE INDEX idx_eventos_ingestao_created_at ON eventos_ingestao(created_at DESC);

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
CREATE TRIGGER update_pdvs_origem_updated_at BEFORE UPDATE ON pdvs_origem
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categorias_canonicas_updated_at BEFORE UPDATE ON categorias_canonicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_produtos_canonicos_updated_at BEFORE UPDATE ON produtos_canonicos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pedidos_canonicos_updated_at BEFORE UPDATE ON pedidos_canonicos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pedido_itens_canonicos_updated_at BEFORE UPDATE ON pedido_itens_canonicos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pedido_pagamentos_canonicos_updated_at BEFORE UPDATE ON pedido_pagamentos_canonicos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_eventos_ingestao_updated_at BEFORE UPDATE ON eventos_ingestao
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_eventos_ingestao_updated_at ON eventos_ingestao;
DROP TRIGGER IF EXISTS update_pedido_pagamentos_canonicos_updated_at ON pedido_pagamentos_canonicos;
DROP TRIGGER IF EXISTS update_pedido_itens_canonicos_updated_at ON pedido_itens_canonicos;
DROP TRIGGER IF EXISTS update_pedidos_canonicos_updated_at ON pedidos_canonicos;
DROP TRIGGER IF EXISTS update_produtos_canonicos_updated_at ON produtos_canonicos;
DROP TRIGGER IF EXISTS update_categorias_canonicas_updated_at ON categorias_canonicas;
DROP TRIGGER IF EXISTS update_pdvs_origem_updated_at ON pdvs_origem;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS eventos_ingestao;
DROP TABLE IF EXISTS pedido_pagamentos_canonicos;
DROP TABLE IF EXISTS pedido_itens_canonicos;
DROP TABLE IF EXISTS pedidos_canonicos;
DROP TABLE IF EXISTS produtos_canonicos;
DROP TABLE IF EXISTS categorias_canonicas;
DROP TABLE IF EXISTS pdvs_origem;
