-- +goose Up
-- Etapa 9: Compras, Contas a Pagar e Atualização de Custo dos Insumos

-- Fornecedores
CREATE TABLE fornecedores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    cnpj_cpf VARCHAR(18) NOT NULL,
    email VARCHAR(255),
    telefone VARCHAR(20),
    banco VARCHAR(100),
    agencia VARCHAR(20),
    conta VARCHAR(20),
    tipo_conta VARCHAR(20), -- 'corrente', 'poupanca'
    pix_chave VARCHAR(255),
    observacoes TEXT,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, cnpj_cpf)
);

CREATE INDEX idx_fornecedores_empresa_id ON fornecedores(empresa_id);
CREATE INDEX idx_fornecedores_ativo ON fornecedores(ativo) WHERE ativo = true;
CREATE INDEX idx_fornecedores_cnpj_cpf ON fornecedores(cnpj_cpf);

-- Ordens de Compra
CREATE TABLE ordens_compra (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    fornecedor_id UUID NOT NULL REFERENCES fornecedores(id) ON DELETE RESTRICT,
    data_compra DATE NOT NULL,
    data_prevista_entrega DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'rascunho', -- 'rascunho', 'enviada', 'recebida_parcial', 'recebida', 'cancelada'
    observacoes TEXT,
    valor_total NUMERIC(14,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_ordens_compra_empresa_id ON ordens_compra(empresa_id);
CREATE INDEX idx_ordens_compra_loja_id ON ordens_compra(loja_id);
CREATE INDEX idx_ordens_compra_fornecedor_id ON ordens_compra(fornecedor_id);
CREATE INDEX idx_ordens_compra_status ON ordens_compra(status);
CREATE INDEX idx_ordens_compra_data_compra ON ordens_compra(data_compra);

-- Itens da Ordem de Compra
CREATE TABLE ordem_compra_itens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_compra_id UUID NOT NULL REFERENCES ordens_compra(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    quantidade NUMERIC(12,4) NOT NULL,
    unidade_medida VARCHAR(20) NOT NULL,
    preco_unitario NUMERIC(12,4) NOT NULL,
    quantidade_recebida NUMERIC(12,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ordem_compra_id, insumo_id)
);

CREATE INDEX idx_ordem_compra_itens_ordem_compra_id ON ordem_compra_itens(ordem_compra_id);
CREATE INDEX idx_ordem_compra_itens_insumo_id ON ordem_compra_itens(insumo_id);

-- Recebimentos de Compra
CREATE TABLE recebimentos_compra (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_compra_id UUID NOT NULL REFERENCES ordens_compra(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    data_recebimento TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    observacoes TEXT,
    permitir_excedente BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recebimentos_compra_ordem_compra_id ON recebimentos_compra(ordem_compra_id);
CREATE INDEX idx_recebimentos_compra_loja_id ON recebimentos_compra(loja_id);
CREATE INDEX idx_recebimentos_compra_data_recebimento ON recebimentos_compra(data_recebimento);

-- Itens do Recebimento
CREATE TABLE recebimento_compra_itens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recebimento_compra_id UUID NOT NULL REFERENCES recebimentos_compra(id) ON DELETE CASCADE,
    ordem_compra_item_id UUID NOT NULL REFERENCES ordem_compra_itens(id) ON DELETE RESTRICT,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    quantidade_recebida NUMERIC(12,4) NOT NULL,
    preco_unitario NUMERIC(12,4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recebimento_compra_itens_recebimento_id ON recebimento_compra_itens(recebimento_compra_id);
CREATE INDEX idx_recebimento_compra_itens_ordem_compra_item_id ON recebimento_compra_itens(ordem_compra_item_id);
CREATE INDEX idx_recebimento_compra_itens_insumo_id ON recebimento_compra_itens(insumo_id);

-- Contas a Pagar
CREATE TABLE contas_pagar (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    fornecedor_id UUID NOT NULL REFERENCES fornecedores(id) ON DELETE RESTRICT,
    ordem_compra_id UUID REFERENCES ordens_compra(id) ON DELETE SET NULL,
    recebimento_compra_id UUID REFERENCES recebimentos_compra(id) ON DELETE SET NULL,
    descricao VARCHAR(255) NOT NULL,
    valor NUMERIC(14,4) NOT NULL,
    data_emissao DATE NOT NULL,
    data_vencimento DATE NOT NULL,
    data_pagamento DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'pendente', -- 'pendente', 'pago', 'cancelado', 'vencido'
    forma_pagamento VARCHAR(50), -- 'dinheiro', 'pix', 'ted', 'boleto', 'cartao'
    observacoes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_contas_pagar_empresa_id ON contas_pagar(empresa_id);
CREATE INDEX idx_contas_pagar_loja_id ON contas_pagar(loja_id);
CREATE INDEX idx_contas_pagar_fornecedor_id ON contas_pagar(fornecedor_id);
CREATE INDEX idx_contas_pagar_ordem_compra_id ON contas_pagar(ordem_compra_id);
CREATE INDEX idx_contas_pagar_status ON contas_pagar(status);
CREATE INDEX idx_contas_pagar_data_vencimento ON contas_pagar(data_vencimento);
CREATE INDEX idx_contas_pagar_data_pagamento ON contas_pagar(data_pagamento);

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
CREATE TRIGGER update_fornecedores_updated_at BEFORE UPDATE ON fornecedores
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ordens_compra_updated_at BEFORE UPDATE ON ordens_compra
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ordem_compra_itens_updated_at BEFORE UPDATE ON ordem_compra_itens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recebimentos_compra_updated_at BEFORE UPDATE ON recebimentos_compra
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contas_pagar_updated_at BEFORE UPDATE ON contas_pagar
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_contas_pagar_updated_at ON contas_pagar;
DROP TRIGGER IF EXISTS update_recebimentos_compra_updated_at ON recebimentos_compra;
DROP TRIGGER IF EXISTS update_ordem_compra_itens_updated_at ON ordem_compra_itens;
DROP TRIGGER IF EXISTS update_ordens_compra_updated_at ON ordens_compra;
DROP TRIGGER IF EXISTS update_fornecedores_updated_at ON fornecedores;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS contas_pagar;
DROP TABLE IF EXISTS recebimento_compra_itens;
DROP TABLE IF EXISTS recebimentos_compra;
DROP TABLE IF EXISTS ordem_compra_itens;
DROP TABLE IF EXISTS ordens_compra;
DROP TABLE IF EXISTS fornecedores;