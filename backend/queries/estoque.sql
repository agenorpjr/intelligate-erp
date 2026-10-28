-- ============================================================
-- Estoque e Movimentações de Insumos (Etapa 12)
-- ============================================================
-- Estoque operacional/gerencial por insumo e loja.
-- Colunas NUMERIC são convertidas para float8 na leitura para que o
-- sqlc gere tipos float64 (mesmo padrão do dashboard).

-- name: GetEstoqueSaldo :one
SELECT
    id, empresa_id, loja_id, insumo_id,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_medio_unitario::float8 AS custo_medio_unitario,
    permite_negativo, created_at, updated_at
FROM estoques
WHERE insumo_id = sqlc.arg(insumo_id)
  AND loja_id = sqlc.arg(loja_id);

-- name: GetEstoqueSaldoForUpdate :one
SELECT
    id, empresa_id, loja_id, insumo_id,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_medio_unitario::float8 AS custo_medio_unitario,
    permite_negativo, created_at, updated_at
FROM estoques
WHERE insumo_id = sqlc.arg(insumo_id)
  AND loja_id = sqlc.arg(loja_id)
FOR UPDATE;

-- name: ListEstoquesSaldos :many
SELECT
    id, empresa_id, loja_id, insumo_id,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_medio_unitario::float8 AS custo_medio_unitario,
    permite_negativo, created_at, updated_at
FROM estoques
WHERE empresa_id = sqlc.arg(empresa_id)
  AND (sqlc.narg(loja_id)::uuid IS NULL OR loja_id = sqlc.narg(loja_id))
  AND (sqlc.narg(insumo_id)::uuid IS NULL OR insumo_id = sqlc.narg(insumo_id))
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpsertEstoqueSaldo :one
INSERT INTO estoques (
    empresa_id, loja_id, insumo_id, quantidade, unidade_medida,
    custo_medio_unitario, permite_negativo
) VALUES (
    sqlc.arg(empresa_id), sqlc.arg(loja_id), sqlc.arg(insumo_id),
    sqlc.arg(quantidade), sqlc.arg(unidade_medida),
    sqlc.arg(custo_medio_unitario), sqlc.arg(permite_negativo)
)
ON CONFLICT (insumo_id, loja_id) DO UPDATE
SET quantidade = EXCLUDED.quantidade,
    unidade_medida = EXCLUDED.unidade_medida,
    custo_medio_unitario = EXCLUDED.custo_medio_unitario,
    updated_at = NOW()
RETURNING
    id, empresa_id, loja_id, insumo_id,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_medio_unitario::float8 AS custo_medio_unitario,
    permite_negativo, created_at, updated_at;

-- name: CreateMovimentacaoEstoque :one
INSERT INTO movimentacoes_estoque (
    empresa_id, loja_id, insumo_id, tipo, quantidade, unidade_medida,
    custo_unitario, custo_total, saldo_anterior, saldo_posterior,
    motivo, observacao, origem, referencia_externa, usuario_id
) VALUES (
    sqlc.arg(empresa_id), sqlc.arg(loja_id), sqlc.arg(insumo_id),
    sqlc.arg(tipo), sqlc.arg(quantidade), sqlc.arg(unidade_medida),
    sqlc.narg(custo_unitario), sqlc.arg(custo_total),
    sqlc.arg(saldo_anterior), sqlc.arg(saldo_posterior),
    sqlc.narg(motivo), sqlc.narg(observacao), sqlc.arg(origem),
    sqlc.narg(referencia_externa), sqlc.narg(usuario_id)
)
RETURNING
    id, empresa_id, loja_id, insumo_id, tipo,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_unitario::float8 AS custo_unitario,
    custo_total::float8 AS custo_total,
    saldo_anterior::float8 AS saldo_anterior,
    saldo_posterior::float8 AS saldo_posterior,
    motivo, observacao, origem, referencia_externa, usuario_id, created_at;

-- name: ListMovimentacoesEstoque :many
SELECT
    id, empresa_id, loja_id, insumo_id, tipo,
    quantidade::float8 AS quantidade,
    unidade_medida,
    custo_unitario::float8 AS custo_unitario,
    custo_total::float8 AS custo_total,
    saldo_anterior::float8 AS saldo_anterior,
    saldo_posterior::float8 AS saldo_posterior,
    motivo, observacao, origem, referencia_externa, usuario_id, created_at
FROM movimentacoes_estoque
WHERE empresa_id = sqlc.arg(empresa_id)
  AND (sqlc.narg(loja_id)::uuid IS NULL OR loja_id = sqlc.narg(loja_id))
  AND (sqlc.narg(insumo_id)::uuid IS NULL OR insumo_id = sqlc.narg(insumo_id))
  AND (sqlc.narg(tipo)::varchar IS NULL OR tipo = sqlc.narg(tipo))
  AND (sqlc.narg(origem)::varchar IS NULL OR origem = sqlc.narg(origem))
  AND (sqlc.narg(data_inicio)::timestamptz IS NULL OR created_at >= sqlc.narg(data_inicio))
  AND (sqlc.narg(data_fim)::timestamptz IS NULL OR created_at <= sqlc.narg(data_fim))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: GetResumoPerdasEstoque :many
SELECT
    me.tipo,
    me.insumo_id,
    i.nome AS insumo_nome,
    me.unidade_medida,
    COALESCE(SUM(me.quantidade), 0)::float8 AS quantidade_total,
    COALESCE(SUM(me.custo_total), 0)::float8 AS custo_total
FROM movimentacoes_estoque me
JOIN insumos i ON i.id = me.insumo_id
WHERE me.empresa_id = sqlc.arg(empresa_id)
  AND me.loja_id = sqlc.arg(loja_id)
  AND me.tipo IN ('perda', 'desperdicio')
  AND me.created_at >= sqlc.arg(data_inicio)
  AND me.created_at <= sqlc.arg(data_fim)
GROUP BY me.tipo, me.insumo_id, i.nome, me.unidade_medida
ORDER BY custo_total DESC;
