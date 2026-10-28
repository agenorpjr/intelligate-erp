-- +goose Up
-- Etapa 12: Estoque e Movimentações de Insumos
-- Estoque operacional/gerencial (não fiscal) por insumo e loja, com histórico
-- completo de movimentações (saldo anterior/posterior, custo médio móvel).

-- Saldos de estoque por insumo + loja
CREATE TABLE estoques (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    quantidade NUMERIC(14,4) NOT NULL DEFAULT 0,
    unidade_medida VARCHAR(20) NOT NULL,
    custo_medio_unitario NUMERIC(14,4) NOT NULL DEFAULT 0,
    -- Permissão explícita para saldo negativo (por padrão bloqueado)
    permite_negativo BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_estoques_insumo_loja UNIQUE (insumo_id, loja_id)
);

CREATE INDEX idx_estoques_empresa_id ON estoques(empresa_id);
CREATE INDEX idx_estoques_loja_id ON estoques(loja_id);
CREATE INDEX idx_estoques_insumo_id ON estoques(insumo_id);

-- Movimentações de estoque (append-only, sem exclusão física)
CREATE TABLE movimentacoes_estoque (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE RESTRICT,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    tipo VARCHAR(30) NOT NULL,
    quantidade NUMERIC(14,4) NOT NULL,
    unidade_medida VARCHAR(20) NOT NULL,
    -- Custo unitário da movimentação (0 quando não informado)
    custo_unitario NUMERIC(14,4) NOT NULL DEFAULT 0,
    custo_total NUMERIC(14,4) NOT NULL DEFAULT 0,
    saldo_anterior NUMERIC(14,4) NOT NULL,
    saldo_posterior NUMERIC(14,4) NOT NULL,
    motivo TEXT,
    observacao TEXT,
    -- Origem da movimentação: 'manual', 'recebimento_compra', 'transferencia', etc.
    origem VARCHAR(50) NOT NULL DEFAULT 'manual',
    referencia_externa VARCHAR(255),
    usuario_id UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_movimentacoes_estoque_tipo CHECK (tipo IN (
        'entrada_compra',
        'saida_consumo',
        'ajuste_positivo',
        'ajuste_negativo',
        'perda',
        'desperdicio',
        'transferencia_saida',
        'transferencia_entrada'
    ))
);

CREATE INDEX idx_movimentacoes_estoque_empresa_id ON movimentacoes_estoque(empresa_id);
CREATE INDEX idx_movimentacoes_estoque_loja_id ON movimentacoes_estoque(loja_id);
CREATE INDEX idx_movimentacoes_estoque_insumo_id ON movimentacoes_estoque(insumo_id);
CREATE INDEX idx_movimentacoes_estoque_tipo ON movimentacoes_estoque(tipo);
CREATE INDEX idx_movimentacoes_estoque_origem ON movimentacoes_estoque(origem);
CREATE INDEX idx_movimentacoes_estoque_created_at ON movimentacoes_estoque(created_at);
CREATE INDEX idx_movimentacoes_estoque_referencia_externa ON movimentacoes_estoque(referencia_externa);

-- Trigger para manter updated_at do saldo
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

CREATE TRIGGER update_estoques_updated_at BEFORE UPDATE ON estoques
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_estoques_updated_at ON estoques;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS movimentacoes_estoque;
DROP TABLE IF EXISTS estoques;
