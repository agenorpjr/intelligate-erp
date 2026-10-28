-- ============================================================
-- Dashboard Operacional
-- ============================================================
-- Indicadores consolidados por loja e período.
-- Valores monetários e percentuais são convertidos para float8
-- de forma explícita para que o sqlc gere tipos float64.

-- name: GetDashboardVendasResumo :one
SELECT
    COALESCE(SUM(valor_total), 0)::float8 AS total_vendas,
    COUNT(*)::bigint AS total_pedidos,
    (CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(valor_total), 0) / COUNT(*) ELSE 0 END)::float8 AS ticket_medio,
    COALESCE(SUM(valor_desconto), 0)::float8 AS total_descontos
FROM pedidos_canonicos
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data_hora >= sqlc.arg(data_inicio)
  AND data_hora <= sqlc.arg(data_fim)
  AND status <> 'cancelado';

-- name: GetDashboardVendasPorDia :many
SELECT
    DATE(data_hora AT TIME ZONE 'UTC')::date AS data,
    COALESCE(SUM(valor_total), 0)::float8 AS total_vendas,
    COUNT(*)::bigint AS total_pedidos,
    (CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(valor_total), 0) / COUNT(*) ELSE 0 END)::float8 AS ticket_medio
FROM pedidos_canonicos
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data_hora >= sqlc.arg(data_inicio)
  AND data_hora <= sqlc.arg(data_fim)
  AND status <> 'cancelado'
GROUP BY DATE(data_hora AT TIME ZONE 'UTC')::date
ORDER BY data;

-- name: GetDashboardVendasPorStatus :many
SELECT
    status,
    COALESCE(SUM(valor_total), 0)::float8 AS total_vendas,
    COUNT(*)::bigint AS total_pedidos
FROM pedidos_canonicos
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data_hora >= sqlc.arg(data_inicio)
  AND data_hora <= sqlc.arg(data_fim)
GROUP BY status
ORDER BY total_vendas DESC;

-- name: GetDashboardVendasResumoPeriodoAnterior :one
SELECT
    COALESCE(SUM(valor_total), 0)::float8 AS total_vendas,
    COUNT(*)::bigint AS total_pedidos,
    (CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(valor_total), 0) / COUNT(*) ELSE 0 END)::float8 AS ticket_medio,
    COALESCE(SUM(valor_desconto), 0)::float8 AS total_descontos
FROM pedidos_canonicos
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data_hora >= sqlc.arg(data_inicio)
  AND data_hora <= sqlc.arg(data_fim)
  AND status <> 'cancelado';

-- ============================================================
-- CMV
-- ============================================================

-- name: GetDashboardCMVResumo :one
WITH vendas_itens AS (
    SELECT
        pc.id AS pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.valor_total AS item_valor_total,
        pi.nome AS item_nome,
        pi.sku_externo AS item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = sqlc.arg(empresa_id)
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND pc.loja_id = sqlc.arg(loja_id)
      AND pc.data_hora >= sqlc.arg(data_inicio)
      AND pc.data_hora <= sqlc.arg(data_fim)
),
itens_custo AS (
    SELECT
        vi.*,
        p.id AS produto_id,
        p.nome AS produto_nome,
        p.sku_externo AS produto_sku,
        (SELECT COALESCE(SUM(ftc.quantidade * ic.custo_unitario / NULLIF(ft.rendimento, 0)), 0)
           FROM fichas_tecnicas ft
           JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
           JOIN LATERAL (
                SELECT insumo_custos.custo_unitario
                FROM insumo_custos
                WHERE insumo_custos.insumo_id = ftc.insumo_id
                  AND insumo_custos.data_vigencia <= vi.data_hora
                ORDER BY insumo_custos.data_vigencia DESC
                LIMIT 1
           ) ic ON true
           WHERE ft.produto_id = p.id
             AND ft.loja_id = vi.loja_id
             AND ft.deleted_at IS NULL
             AND ft.ativo = true
        )::float8 AS custo_unitario_teorico
    FROM vendas_itens vi
    LEFT JOIN LATERAL (
        SELECT pr.id, pr.nome, pr.sku_externo
        FROM produtos pr
        WHERE pr.empresa_id = sqlc.arg(empresa_id)
          AND pr.deleted_at IS NULL
          AND (
              (pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku)
              OR (pr.nome = vi.item_nome)
          )
        ORDER BY (CASE WHEN pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku THEN 0 ELSE 1 END)
        LIMIT 1
    ) p ON true
)
SELECT
    COALESCE(SUM(item_valor_total), 0)::float8 AS total_vendas,
    COALESCE(SUM(quantidade * custo_unitario_teorico), 0)::float8 AS total_cmv,
    (COALESCE(SUM(item_valor_total), 0) - COALESCE(SUM(quantidade * custo_unitario_teorico), 0))::float8 AS margem_bruta,
    (CASE WHEN COALESCE(SUM(item_valor_total), 0) > 0
          THEN ((COALESCE(SUM(item_valor_total), 0) - COALESCE(SUM(quantidade * custo_unitario_teorico), 0)) / COALESCE(SUM(item_valor_total), 0)) * 100
          ELSE 0 END)::float8 AS margem_bruta_percent,
    COUNT(*)::bigint AS itens_vendidos,
    COUNT(*) FILTER (WHERE custo_unitario_teorico > 0)::bigint AS itens_com_ficha,
    COUNT(DISTINCT produto_id) FILTER (WHERE custo_unitario_teorico = 0)::bigint AS produtos_sem_ficha_tecnica
FROM itens_custo;

-- name: GetDashboardCMVPorDia :many
WITH vendas_itens AS (
    SELECT
        pc.id AS pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.valor_total AS item_valor_total,
        pi.nome AS item_nome,
        pi.sku_externo AS item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = sqlc.arg(empresa_id)
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND pc.loja_id = sqlc.arg(loja_id)
      AND pc.data_hora >= sqlc.arg(data_inicio)
      AND pc.data_hora <= sqlc.arg(data_fim)
),
itens_custo AS (
    SELECT
        vi.*,
        p.id AS produto_id,
        (SELECT COALESCE(SUM(ftc.quantidade * ic.custo_unitario / NULLIF(ft.rendimento, 0)), 0)
           FROM fichas_tecnicas ft
           JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
           JOIN LATERAL (
                SELECT insumo_custos.custo_unitario
                FROM insumo_custos
                WHERE insumo_custos.insumo_id = ftc.insumo_id
                  AND insumo_custos.data_vigencia <= vi.data_hora
                ORDER BY insumo_custos.data_vigencia DESC
                LIMIT 1
           ) ic ON true
           WHERE ft.produto_id = p.id
             AND ft.loja_id = vi.loja_id
             AND ft.deleted_at IS NULL
             AND ft.ativo = true
        )::float8 AS custo_unitario_teorico
    FROM vendas_itens vi
    LEFT JOIN LATERAL (
        SELECT pr.id
        FROM produtos pr
        WHERE pr.empresa_id = sqlc.arg(empresa_id)
          AND pr.deleted_at IS NULL
          AND (
              (pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku)
              OR (pr.nome = vi.item_nome)
          )
        ORDER BY (CASE WHEN pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku THEN 0 ELSE 1 END)
        LIMIT 1
    ) p ON true
)
SELECT
    DATE(data_hora AT TIME ZONE 'UTC')::date AS data,
    COALESCE(SUM(item_valor_total), 0)::float8 AS total_vendas,
    COALESCE(SUM(quantidade * custo_unitario_teorico), 0)::float8 AS total_cmv
FROM itens_custo
GROUP BY DATE(data_hora AT TIME ZONE 'UTC')::date
ORDER BY data;

-- name: GetDashboardCMVPorProduto :many
WITH vendas_itens AS (
    SELECT
        pc.id AS pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.valor_total AS item_valor_total,
        pi.nome AS item_nome,
        pi.sku_externo AS item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = sqlc.arg(empresa_id)
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND pc.loja_id = sqlc.arg(loja_id)
      AND pc.data_hora >= sqlc.arg(data_inicio)
      AND pc.data_hora <= sqlc.arg(data_fim)
),
itens_custo AS (
    SELECT
        vi.*,
        p.id AS produto_id,
        p.nome AS produto_nome,
        p.sku_externo AS produto_sku,
        (SELECT COALESCE(SUM(ftc.quantidade * ic.custo_unitario / NULLIF(ft.rendimento, 0)), 0)
           FROM fichas_tecnicas ft
           JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
           JOIN LATERAL (
                SELECT insumo_custos.custo_unitario
                FROM insumo_custos
                WHERE insumo_custos.insumo_id = ftc.insumo_id
                  AND insumo_custos.data_vigencia <= vi.data_hora
                ORDER BY insumo_custos.data_vigencia DESC
                LIMIT 1
           ) ic ON true
           WHERE ft.produto_id = p.id
             AND ft.loja_id = vi.loja_id
             AND ft.deleted_at IS NULL
             AND ft.ativo = true
        )::float8 AS custo_unitario_teorico
    FROM vendas_itens vi
    LEFT JOIN LATERAL (
        SELECT pr.id, pr.nome, pr.sku_externo
        FROM produtos pr
        WHERE pr.empresa_id = sqlc.arg(empresa_id)
          AND pr.deleted_at IS NULL
          AND (
              (pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku)
              OR (pr.nome = vi.item_nome)
          )
        ORDER BY (CASE WHEN pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku THEN 0 ELSE 1 END)
        LIMIT 1
    ) p ON true
)
SELECT
    produto_id,
    produto_nome,
    produto_sku,
    SUM(quantidade)::float8 AS quantidade_vendida,
    (COALESCE(SUM(quantidade * custo_unitario_teorico), 0) / NULLIF(SUM(quantidade), 0))::float8 AS custo_unitario_teorico,
    COALESCE(SUM(quantidade * custo_unitario_teorico), 0)::float8 AS cmv_total,
    COALESCE(SUM(item_valor_total), 0)::float8 AS receita_total,
    (COALESCE(SUM(item_valor_total), 0) - COALESCE(SUM(quantidade * custo_unitario_teorico), 0))::float8 AS margem_bruta,
    (CASE WHEN COALESCE(SUM(item_valor_total), 0) > 0
          THEN ((COALESCE(SUM(item_valor_total), 0) - COALESCE(SUM(quantidade * custo_unitario_teorico), 0)) / COALESCE(SUM(item_valor_total), 0)) * 100
          ELSE 0 END)::float8 AS margem_bruta_percent,
    bool_and(custo_unitario_teorico > 0) AS tem_ficha_tecnica
FROM itens_custo
WHERE produto_id IS NOT NULL
GROUP BY produto_id, produto_nome, produto_sku
ORDER BY cmv_total DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: GetDashboardProdutosSemFichaTecnica :many
WITH vendas_itens AS (
    SELECT
        pc.id AS pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.valor_total AS item_valor_total,
        pi.nome AS item_nome,
        pi.sku_externo AS item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = sqlc.arg(empresa_id)
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND pc.loja_id = sqlc.arg(loja_id)
      AND pc.data_hora >= sqlc.arg(data_inicio)
      AND pc.data_hora <= sqlc.arg(data_fim)
),
itens_custo AS (
    SELECT
        vi.*,
        p.id AS produto_id,
        p.nome AS produto_nome,
        p.sku_externo AS produto_sku,
        (SELECT COALESCE(SUM(ftc.quantidade * ic.custo_unitario / NULLIF(ft.rendimento, 0)), 0)
           FROM fichas_tecnicas ft
           JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
           JOIN LATERAL (
                SELECT insumo_custos.custo_unitario
                FROM insumo_custos
                WHERE insumo_custos.insumo_id = ftc.insumo_id
                  AND insumo_custos.data_vigencia <= vi.data_hora
                ORDER BY insumo_custos.data_vigencia DESC
                LIMIT 1
           ) ic ON true
           WHERE ft.produto_id = p.id
             AND ft.loja_id = vi.loja_id
             AND ft.deleted_at IS NULL
             AND ft.ativo = true
        )::float8 AS custo_unitario_teorico
    FROM vendas_itens vi
    LEFT JOIN LATERAL (
        SELECT pr.id, pr.nome, pr.sku_externo
        FROM produtos pr
        WHERE pr.empresa_id = sqlc.arg(empresa_id)
          AND pr.deleted_at IS NULL
          AND (
              (pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku)
              OR (pr.nome = vi.item_nome)
          )
        ORDER BY (CASE WHEN pr.sku_externo IS NOT NULL AND vi.item_sku IS NOT NULL AND pr.sku_externo = vi.item_sku THEN 0 ELSE 1 END)
        LIMIT 1
    ) p ON true
)
SELECT
    produto_id,
    produto_nome,
    produto_sku,
    SUM(quantidade)::float8 AS quantidade_vendida,
    COALESCE(SUM(item_valor_total), 0)::float8 AS receita_total
FROM itens_custo
WHERE produto_id IS NOT NULL
  AND custo_unitario_teorico = 0
GROUP BY produto_id, produto_nome, produto_sku
ORDER BY receita_total DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- ============================================================
-- Financeiro
-- ============================================================

-- name: GetDashboardFinanceiroResumo :one
SELECT
    COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE 0 END), 0)::float8 AS total_entradas,
    COALESCE(SUM(CASE WHEN tipo = 'saida' THEN valor ELSE 0 END), 0)::float8 AS total_saidas,
    COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE -valor END), 0)::float8 AS saldo_final
FROM movimentacoes_caixa
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data >= sqlc.arg(data_inicio)
  AND data <= sqlc.arg(data_fim);

-- name: GetDashboardDespesasPorCategoria :many
SELECT
    cf.id AS categoria_financeira_id,
    cf.nome AS categoria_nome,
    COALESCE(SUM(mc.valor), 0)::float8 AS total
FROM movimentacoes_caixa mc
JOIN categorias_financeiras cf ON cf.id = mc.categoria_financeira_id AND cf.deleted_at IS NULL
WHERE mc.empresa_id = sqlc.arg(empresa_id)
  AND mc.deleted_at IS NULL
  AND mc.loja_id = sqlc.arg(loja_id)
  AND mc.data >= sqlc.arg(data_inicio)
  AND mc.data <= sqlc.arg(data_fim)
  AND mc.tipo = 'saida'
GROUP BY cf.id, cf.nome
ORDER BY total DESC;

-- name: GetDashboardContasPagarResumo :one
SELECT
    COUNT(*) FILTER (WHERE status = 'pendente')::bigint AS contas_pagar_em_aberto,
    COUNT(*) FILTER (WHERE status = 'vencido' OR (status = 'pendente' AND data_vencimento < CURRENT_DATE))::bigint AS contas_pagar_vencidas,
    COALESCE(SUM(valor) FILTER (WHERE status = 'pendente'), 0)::float8 AS valor_contas_pagar_aberto,
    COALESCE(SUM(valor) FILTER (WHERE status = 'vencido' OR (status = 'pendente' AND data_vencimento < CURRENT_DATE)), 0)::float8 AS valor_contas_pagar_vencido
FROM contas_pagar
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id);

-- name: GetDashboardFluxoCaixaPorDia :many
SELECT
    DATE(data AT TIME ZONE 'UTC')::date AS data,
    COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE 0 END), 0)::float8 AS entradas,
    COALESCE(SUM(CASE WHEN tipo = 'saida' THEN valor ELSE 0 END), 0)::float8 AS saidas,
    COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE -valor END), 0)::float8 AS saldo_diario
FROM movimentacoes_caixa
WHERE empresa_id = sqlc.arg(empresa_id)
  AND deleted_at IS NULL
  AND loja_id = sqlc.arg(loja_id)
  AND data >= sqlc.arg(data_inicio)
  AND data <= sqlc.arg(data_fim)
GROUP BY DATE(data AT TIME ZONE 'UTC')::date
ORDER BY data;

-- ============================================================
-- Totais por tipo de conta gerencial (base para o DRE)
-- ============================================================

-- name: GetDashboardTotaisPorTipoContaGerencial :many
SELECT
    cg.tipo AS tipo,
    COALESCE(SUM(CASE WHEN mc.tipo = 'saida' THEN mc.valor ELSE -mc.valor END), 0)::float8 AS total
FROM movimentacoes_caixa mc
JOIN contas_gerenciais cg ON cg.id = mc.conta_gerencial_id
WHERE mc.empresa_id = sqlc.arg(empresa_id)
  AND mc.deleted_at IS NULL
  AND cg.deleted_at IS NULL
  AND mc.loja_id = sqlc.arg(loja_id)
  AND mc.data >= sqlc.arg(data_inicio)
  AND mc.data <= sqlc.arg(data_fim)
GROUP BY cg.tipo;