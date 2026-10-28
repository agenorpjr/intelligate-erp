-- +goose Up
-- Etapa 10: Financeiro Operacional - Plano de Contas, Categorias Financeiras e Movimentações de Caixa

-- Plano de Contas Gerencial
CREATE TABLE contas_gerenciais (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    codigo VARCHAR(20) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    tipo VARCHAR(30) NOT NULL, -- 'receita', 'custo', 'despesa', 'investimento'
    grupo VARCHAR(100), -- agrupamento opcional (ex: 'pessoal', 'operacional', 'administrativo')
    conta_pai_id UUID REFERENCES contas_gerenciais(id) ON DELETE SET NULL,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, codigo)
);

CREATE INDEX idx_contas_gerenciais_empresa_id ON contas_gerenciais(empresa_id);
CREATE INDEX idx_contas_gerenciais_tipo ON contas_gerenciais(tipo);
CREATE INDEX idx_contas_gerenciais_conta_pai_id ON contas_gerenciais(conta_pai_id);
CREATE INDEX idx_contas_gerenciais_ativo ON contas_gerenciais(ativo) WHERE ativo = true;

-- Categorias Financeiras
CREATE TABLE categorias_financeiras (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    conta_gerencial_id UUID REFERENCES contas_gerenciais(id) ON DELETE SET NULL,
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, nome)
);

CREATE INDEX idx_categorias_financeiras_empresa_id ON categorias_financeiras(empresa_id);
CREATE INDEX idx_categorias_financeiras_conta_gerencial_id ON categorias_financeiras(conta_gerencial_id);
CREATE INDEX idx_categorias_financeiras_ativo ON categorias_financeiras(ativo) WHERE ativo = true;

-- Adicionar colunas de categoria_financeira_id e conta_gerencial_id em contas_pagar
ALTER TABLE contas_pagar
    ADD COLUMN categoria_financeira_id UUID REFERENCES categorias_financeiras(id) ON DELETE SET NULL,
    ADD COLUMN conta_gerencial_id UUID REFERENCES contas_gerenciais(id) ON DELETE SET NULL;

CREATE INDEX idx_contas_pagar_categoria_financeira_id ON contas_pagar(categoria_financeira_id);
CREATE INDEX idx_contas_pagar_conta_gerencial_id ON contas_pagar(conta_gerencial_id);

-- Movimentações de Caixa
CREATE TABLE movimentacoes_caixa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    data TIMESTAMPTZ NOT NULL,
    tipo VARCHAR(10) NOT NULL, -- 'entrada', 'saida'
    valor NUMERIC(14,4) NOT NULL,
    descricao VARCHAR(255) NOT NULL,
    categoria_financeira_id UUID REFERENCES categorias_financeiras(id) ON DELETE SET NULL,
    conta_gerencial_id UUID REFERENCES contas_gerenciais(id) ON DELETE SET NULL,
    origem VARCHAR(50), -- 'venda', 'conta_pagar', 'manual', 'outros'
    origem_id UUID, -- ID da entidade de origem (ex: pedido_canonico_id, conta_pagar_id)
    observacoes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_movimentacoes_caixa_empresa_id ON movimentacoes_caixa(empresa_id);
CREATE INDEX idx_movimentacoes_caixa_loja_id ON movimentacoes_caixa(loja_id);
CREATE INDEX idx_movimentacoes_caixa_data ON movimentacoes_caixa(data);
CREATE INDEX idx_movimentacoes_caixa_tipo ON movimentacoes_caixa(tipo);
CREATE INDEX idx_movimentacoes_caixa_categoria_financeira_id ON movimentacoes_caixa(categoria_financeira_id);
CREATE INDEX idx_movimentacoes_caixa_conta_gerencial_id ON movimentacoes_caixa(conta_gerencial_id);
CREATE INDEX idx_movimentacoes_caixa_origem ON movimentacoes_caixa(origem);
CREATE INDEX idx_movimentacoes_caixa_origem_id ON movimentacoes_caixa(origem_id);

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
CREATE TRIGGER update_contas_gerenciais_updated_at BEFORE UPDATE ON contas_gerenciais
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categorias_financeiras_updated_at BEFORE UPDATE ON categorias_financeiras
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_movimentacoes_caixa_updated_at BEFORE UPDATE ON movimentacoes_caixa
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_movimentacoes_caixa_updated_at ON movimentacoes_caixa;
DROP TRIGGER IF EXISTS update_categorias_financeiras_updated_at ON categorias_financeiras;
DROP TRIGGER IF EXISTS update_contas_gerenciais_updated_at ON contas_gerenciais;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS movimentacoes_caixa;

ALTER TABLE contas_pagar
    DROP COLUMN IF EXISTS categoria_financeira_id,
    DROP COLUMN IF EXISTS conta_gerencial_id;

DROP TABLE IF EXISTS categorias_financeiras;
DROP TABLE IF EXISTS contas_gerenciais;