-- Insumos
-- name: CreateInsumo :one
INSERT INTO insumos (empresa_id, nome, unidade_medida, custo_unitario_atual, fornecedor)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInsumo :one
SELECT * FROM insumos WHERE id = $1 AND deleted_at IS NULL;

-- name: ListInsumosByEmpresa :many
SELECT * FROM insumos
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::boolean IS NULL OR ativo = $2)
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: UpdateInsumo :one
UPDATE insumos
SET nome = COALESCE($2, nome),
    unidade_medida = COALESCE($3, unidade_medida),
    custo_unitario_atual = COALESCE($4, custo_unitario_atual),
    fornecedor = COALESCE($5, fornecedor),
    ativo = COALESCE($6, ativo),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteInsumo :exec
UPDATE insumos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: GetInsumoByNome :one
SELECT * FROM insumos WHERE empresa_id = $1 AND nome = $2 AND deleted_at IS NULL;

-- Insumo Custos (Histórico)
-- name: CreateInsumoCusto :one
INSERT INTO insumo_custos (insumo_id, data_vigencia, custo_unitario, observacao)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetInsumoCusto :one
SELECT * FROM insumo_custos WHERE id = $1;

-- name: ListInsumoCustosByInsumo :many
SELECT * FROM insumo_custos
WHERE insumo_id = $1
ORDER BY data_vigencia DESC LIMIT $2 OFFSET $3;

-- name: GetInsumoCustoVigente :one
-- Get the cost valid at a specific date (most recent cost with data_vigencia <= target date)
SELECT * FROM insumo_custos
WHERE insumo_id = $1 AND data_vigencia <= $2
ORDER BY data_vigencia DESC LIMIT 1;

-- name: GetInsumoCustoAtual :one
-- Get the most recent cost for an insumo
SELECT * FROM insumo_custos
WHERE insumo_id = $1
ORDER BY data_vigencia DESC LIMIT 1;

-- Produtos
-- name: CreateProduto :one
INSERT INTO produtos (empresa_id, nome, sku_externo, categoria)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProduto :one
SELECT * FROM produtos WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProdutosByEmpresa :many
SELECT * FROM produtos
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::boolean IS NULL OR ativo = $2)
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: UpdateProduto :one
UPDATE produtos
SET nome = COALESCE($2, nome),
    sku_externo = COALESCE($3, sku_externo),
    categoria = COALESCE($4, categoria),
    ativo = COALESCE($5, ativo),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProduto :exec
UPDATE produtos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: GetProdutoBySKUExterno :one
SELECT * FROM produtos WHERE empresa_id = $1 AND sku_externo = $2 AND deleted_at IS NULL;

-- Fichas Técnicas
-- name: CreateFichaTecnica :one
INSERT INTO fichas_tecnicas (produto_id, loja_id, rendimento, unidade_rendimento)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFichaTecnica :one
SELECT * FROM fichas_tecnicas WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFichaTecnicaByProdutoLoja :one
SELECT * FROM fichas_tecnicas WHERE produto_id = $1 AND loja_id = $2 AND deleted_at IS NULL;

-- name: ListFichasTecnicasByProduto :many
SELECT * FROM fichas_tecnicas
WHERE produto_id = $1
  AND deleted_at IS NULL
  AND ($2::boolean IS NULL OR ativo = $2)
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: ListFichasTecnicasByLoja :many
SELECT * FROM fichas_tecnicas
WHERE loja_id = $1
  AND deleted_at IS NULL
  AND ($2::boolean IS NULL OR ativo = $2)
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: ListFichasTecnicasByEmpresa :many
SELECT ft.* FROM fichas_tecnicas ft
JOIN produtos p ON ft.produto_id = p.id
WHERE p.empresa_id = $1
  AND ft.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND ($2::boolean IS NULL OR ft.ativo = $2)
ORDER BY ft.created_at DESC LIMIT $3 OFFSET $4;

-- name: UpdateFichaTecnica :one
UPDATE fichas_tecnicas
SET rendimento = COALESCE($2, rendimento),
    unidade_rendimento = COALESCE($3, unidade_rendimento),
    ativo = COALESCE($4, ativo),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteFichaTecnica :exec
UPDATE fichas_tecnicas SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Ficha Técnica Componentes
-- name: CreateFichaTecnicaComponente :one
INSERT INTO ficha_tecnica_componentes (ficha_tecnica_id, insumo_id, quantidade, unidade_medida)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFichaTecnicaComponente :one
SELECT * FROM ficha_tecnica_componentes WHERE id = $1;

-- name: ListFichaTecnicaComponentes :many
SELECT ftc.*, i.nome as insumo_nome, i.unidade_medida as insumo_unidade
FROM ficha_tecnica_componentes ftc
JOIN insumos i ON ftc.insumo_id = i.id
WHERE ftc.ficha_tecnica_id = $1
ORDER BY ftc.created_at;

-- name: UpdateFichaTecnicaComponente :one
UPDATE ficha_tecnica_componentes
SET insumo_id = $2, quantidade = $3, unidade_medida = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteFichaTecnicaComponente :exec
DELETE FROM ficha_tecnica_componentes WHERE id = $1;

-- name: DeleteFichaTecnicaComponentesByFicha :exec
DELETE FROM ficha_tecnica_componentes WHERE ficha_tecnica_id = $1;

-- CMV Queries
-- name: GetCMVResumo :one
WITH vendas_filtradas AS (
    SELECT
        pc.id as pedido_id,
        pc.loja_id,
        pc.data_hora,
        pc.valor_total,
        pi.quantidade,
        pi.preco_unitario,
        pi.valor_total as item_valor_total,
        pi.produto_canonico_id,
        pi.nome as item_nome,
        pi.sku_externo as item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = $1
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND ($2::uuid IS NULL OR pc.loja_id = $2)
      AND pc.data_hora >= $3
      AND pc.data_hora <= $4
),
produtos_com_ficha AS (
    SELECT
        vf.*,
        p.id as produto_id,
        p.nome as produto_nome,
        p.sku_externo as produto_sku,
        ft.id as ficha_id,
        ft.rendimento,
        ft.unidade_rendimento,
        COALESCE(SUM(ftc.quantidade * ic.custo_unitario / ft.rendimento), 0) as custo_unitario_teorico
    FROM vendas_filtradas vf
    JOIN produtos p ON (
        p.empresa_id = $1
        AND p.deleted_at IS NULL
        AND (
            (p.sku_externo IS NOT NULL AND vf.item_sku IS NOT NULL AND p.sku_externo = vf.item_sku)
            OR (p.nome = vf.item_nome)
        )
    )
    JOIN fichas_tecnicas ft ON ft.produto_id = p.id AND ft.loja_id = vf.loja_id AND ft.deleted_at IS NULL AND ft.ativo = true
    JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
    JOIN LATERAL (
        SELECT custo_unitario FROM insumo_custos
        WHERE insumo_id = ftc.insumo_id AND data_vigencia <= vf.data_hora
        ORDER BY data_vigencia DESC LIMIT 1
    ) ic ON true
    GROUP BY vf.pedido_id, vf.loja_id, vf.data_hora, vf.valor_total, vf.quantidade, vf.preco_unitario, vf.item_valor_total,
             vf.produto_canonico_id, vf.item_nome, vf.item_sku, p.id, p.nome, p.sku_externo, ft.id, ft.rendimento, ft.unidade_rendimento
),
produtos_sem_ficha AS (
    SELECT
        vf.*,
        p.id as produto_id,
        p.nome as produto_nome,
        p.sku_externo as produto_sku
    FROM vendas_filtradas vf
    LEFT JOIN produtos p ON (
        p.empresa_id = $1
        AND p.deleted_at IS NULL
        AND (
            (p.sku_externo IS NOT NULL AND vf.item_sku IS NOT NULL AND p.sku_externo = vf.item_sku)
            OR (p.nome = vf.item_nome)
        )
    )
    LEFT JOIN fichas_tecnicas ft ON ft.produto_id = p.id AND ft.loja_id = vf.loja_id AND ft.deleted_at IS NULL AND ft.ativo = true
    WHERE ft.id IS NULL
)
SELECT
    $1::uuid as empresa_id,
    $2::uuid as loja_id,
    $3::timestamptz as data_inicio,
    $4::timestamptz as data_fim,
    COALESCE(SUM(vf.valor_total), 0) as total_vendas,
    COALESCE(SUM(pcf.quantidade * pcf.custo_unitario_teorico), 0) as total_cmv,
    COALESCE(SUM(vf.valor_total), 0) - COALESCE(SUM(pcf.quantidade * pcf.custo_unitario_teorico), 0) as margem_bruta,
    CASE WHEN COALESCE(SUM(vf.valor_total), 0) > 0
         THEN ((COALESCE(SUM(vf.valor_total), 0) - COALESCE(SUM(pcf.quantidade * pcf.custo_unitario_teorico), 0)) / COALESCE(SUM(vf.valor_total), 0)) * 100
         ELSE 0 END as margem_bruta_percent,
    COUNT(DISTINCT pcf.produto_id) as produtos_com_ficha,
    COUNT(DISTINCT psf.produto_id) as produtos_sem_ficha_tecnica,
    COUNT(*) as itens_vendidos
FROM vendas_filtradas vf
LEFT JOIN produtos_com_ficha pcf ON pcf.pedido_id = vf.pedido_id AND pcf.item_nome = vf.item_nome
LEFT JOIN produtos_sem_ficha psf ON psf.pedido_id = vf.pedido_id AND psf.item_nome = vf.item_nome
LIMIT 1;

-- name: GetCMVProdutosDetalhe :many
WITH vendas_filtradas AS (
    SELECT
        pc.id as pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.preco_unitario,
        pi.valor_total as item_valor_total,
        pi.produto_canonico_id,
        pi.nome as item_nome,
        pi.sku_externo as item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = $1
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND ($2::uuid IS NULL OR pc.loja_id = $2)
      AND pc.data_hora >= $3
      AND pc.data_hora <= $4
),
produtos_com_ficha AS (
    SELECT
        vf.item_nome,
        vf.item_sku,
        p.id as produto_id,
        p.nome as produto_nome,
        p.sku_externo as produto_sku,
        SUM(vf.quantidade) as quantidade_vendida,
        SUM(vf.item_valor_total) as receita_total,
        COALESCE(SUM(ftc.quantidade * ic.custo_unitario / ft.rendimento * vf.quantidade), 0) as cmv_total,
        COALESCE(SUM(ftc.quantidade * ic.custo_unitario / ft.rendimento * vf.quantidade) / NULLIF(SUM(vf.quantidade), 0), 0) as custo_unitario_teorico
    FROM vendas_filtradas vf
    JOIN produtos p ON (
        p.empresa_id = $1
        AND p.deleted_at IS NULL
        AND (
            (p.sku_externo IS NOT NULL AND vf.item_sku IS NOT NULL AND p.sku_externo = vf.item_sku)
            OR (p.nome = vf.item_nome)
        )
    )
    JOIN fichas_tecnicas ft ON ft.produto_id = p.id AND ft.loja_id = vf.loja_id AND ft.deleted_at IS NULL AND ft.ativo = true
    JOIN ficha_tecnica_componentes ftc ON ftc.ficha_tecnica_id = ft.id
    JOIN LATERAL (
        SELECT custo_unitario FROM insumo_custos
        WHERE insumo_id = ftc.insumo_id AND data_vigencia <= vf.data_hora
        ORDER BY data_vigencia DESC LIMIT 1
    ) ic ON true
    GROUP BY vf.item_nome, vf.item_sku, p.id, p.nome, p.sku_externo
)
SELECT
    produto_id,
    produto_nome,
    produto_sku,
    quantidade_vendida,
    custo_unitario_teorico,
    cmv_total,
    receita_total,
    receita_total - cmv_total as margem_bruta,
    CASE WHEN receita_total > 0 THEN ((receita_total - cmv_total) / receita_total) * 100 ELSE 0 END as margem_bruta_percent,
    true as tem_ficha_tecnica
FROM produtos_com_ficha
ORDER BY cmv_total DESC LIMIT $5 OFFSET $6;

-- name: GetProdutosSemFichaTecnica :many
WITH vendas_filtradas AS (
    SELECT
        pc.id as pedido_id,
        pc.loja_id,
        pc.data_hora,
        pi.quantidade,
        pi.preco_unitario,
        pi.valor_total as item_valor_total,
        pi.nome as item_nome,
        pi.sku_externo as item_sku
    FROM pedidos_canonicos pc
    JOIN pedido_itens_canonicos pi ON pi.pedido_id = pc.id
    WHERE pc.empresa_id = $1
      AND pc.deleted_at IS NULL
      AND pi.deleted_at IS NULL
      AND pc.status = 'concluido'
      AND ($2::uuid IS NULL OR pc.loja_id = $2)
      AND pc.data_hora >= $3
      AND pc.data_hora <= $4
),
produtos_sem_ficha AS (
    SELECT
        vf.item_nome,
        vf.item_sku,
        p.id as produto_id,
        p.nome as produto_nome,
        p.sku_externo as produto_sku,
        SUM(vf.quantidade) as quantidade_vendida,
        SUM(vf.item_valor_total) as receita_total
    FROM vendas_filtradas vf
    LEFT JOIN produtos p ON (
        p.empresa_id = $1
        AND p.deleted_at IS NULL
        AND (
            (p.sku_externo IS NOT NULL AND vf.item_sku IS NOT NULL AND p.sku_externo = vf.item_sku)
            OR (p.nome = vf.item_nome)
        )
    )
    LEFT JOIN fichas_tecnicas ft ON ft.produto_id = p.id AND ft.loja_id = vf.loja_id AND ft.deleted_at IS NULL AND ft.ativo = true
    WHERE ft.id IS NULL
    GROUP BY vf.item_nome, vf.item_sku, p.id, p.nome, p.sku_externo
)
SELECT
    produto_id,
    produto_nome,
    produto_sku,
    quantidade_vendida,
    0 as custo_unitario_teorico,
    0 as cmv_total,
    receita_total,
    receita_total as margem_bruta,
    100 as margem_bruta_percent,
    false as tem_ficha_tecnica
FROM produtos_sem_ficha
ORDER BY receita_total DESC LIMIT $5 OFFSET $6;