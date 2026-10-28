-- name: CreateContaGerencial :one
INSERT INTO contas_gerenciais (
    empresa_id, codigo, nome, tipo, grupo, conta_pai_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at;

-- name: GetContaGerencial :one
SELECT id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at
FROM contas_gerenciais
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetContaGerencialByCodigo :one
SELECT id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at
FROM contas_gerenciais
WHERE empresa_id = $1 AND codigo = $2 AND deleted_at IS NULL;

-- name: ListContasGerenciaisByEmpresa :many
SELECT id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at
FROM contas_gerenciais
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::varchar IS NULL OR tipo = $2)
  AND ($3::varchar IS NULL OR grupo = $3)
  AND ($4::boolean IS NULL OR ativo = $4)
ORDER BY codigo
LIMIT $5 OFFSET $6;

-- name: ListContasGerenciaisByEmpresaHierarquia :many
SELECT id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at
FROM contas_gerenciais
WHERE empresa_id = $1 AND deleted_at IS NULL
ORDER BY codigo;

-- name: UpdateContaGerencial :one
UPDATE contas_gerenciais
SET codigo = $2, nome = $3, tipo = $4, grupo = $5, conta_pai_id = $6, ativo = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, codigo, nome, tipo, grupo, conta_pai_id, ativo, created_at, updated_at, deleted_at;

-- name: DeleteContaGerencial :exec
UPDATE contas_gerenciais
SET deleted_at = NOW()
WHERE id = $1;

-- name: CheckContaGerencialHasMovimentacoes :one
SELECT EXISTS (
    SELECT 1 FROM movimentacoes_caixa WHERE movimentacoes_caixa.conta_gerencial_id = $1 AND deleted_at IS NULL
    UNION ALL
    SELECT 1 FROM contas_pagar WHERE contas_pagar.conta_gerencial_id = $1 AND deleted_at IS NULL
    UNION ALL
    SELECT 1 FROM categorias_financeiras WHERE categorias_financeiras.conta_gerencial_id = $1 AND deleted_at IS NULL
) as has_movimentacoes;

-- ============================================================
-- Categorias Financeiras
-- ============================================================

-- name: CreateCategoriaFinanceira :one
INSERT INTO categorias_financeiras (
    empresa_id, nome, descricao, conta_gerencial_id
) VALUES (
    $1, $2, $3, $4
) RETURNING id, empresa_id, nome, descricao, conta_gerencial_id, ativo, created_at, updated_at, deleted_at;

-- name: GetCategoriaFinanceira :one
SELECT id, empresa_id, nome, descricao, conta_gerencial_id, ativo, created_at, updated_at, deleted_at
FROM categorias_financeiras
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListCategoriasFinanceirasByEmpresa :many
SELECT id, empresa_id, nome, descricao, conta_gerencial_id, ativo, created_at, updated_at, deleted_at
FROM categorias_financeiras
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR conta_gerencial_id = $2)
  AND ($3::boolean IS NULL OR ativo = $3)
ORDER BY nome
LIMIT $4 OFFSET $5;

-- name: UpdateCategoriaFinanceira :one
UPDATE categorias_financeiras
SET nome = $2, descricao = $3, conta_gerencial_id = $4, ativo = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, nome, descricao, conta_gerencial_id, ativo, created_at, updated_at, deleted_at;

-- name: DeleteCategoriaFinanceira :exec
UPDATE categorias_financeiras
SET deleted_at = NOW()
WHERE id = $1;

-- ============================================================
-- Movimentações de Caixa
-- ============================================================

-- name: CreateMovimentacaoCaixa :one
INSERT INTO movimentacoes_caixa (
    empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes, created_at, updated_at, deleted_at;

-- name: GetMovimentacaoCaixa :one
SELECT id, empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes, created_at, updated_at, deleted_at
FROM movimentacoes_caixa
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListMovimentacoesCaixaByEmpresa :many
SELECT id, empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes, created_at, updated_at, deleted_at
FROM movimentacoes_caixa
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::varchar IS NULL OR tipo = $3)
  AND ($4::uuid IS NULL OR categoria_financeira_id = $4)
  AND ($5::uuid IS NULL OR conta_gerencial_id = $5)
  AND ($6::varchar IS NULL OR origem = $6)
  AND ($7::timestamptz IS NULL OR data >= $7)
  AND ($8::timestamptz IS NULL OR data <= $8)
ORDER BY data DESC, created_at DESC
LIMIT $9 OFFSET $10;

-- name: ListMovimentacoesCaixaByLoja :many
SELECT id, empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes, created_at, updated_at, deleted_at
FROM movimentacoes_caixa
WHERE loja_id = $1
  AND deleted_at IS NULL
  AND ($2::varchar IS NULL OR tipo = $2)
  AND ($3::uuid IS NULL OR categoria_financeira_id = $3)
  AND ($4::uuid IS NULL OR conta_gerencial_id = $4)
  AND ($5::varchar IS NULL OR origem = $5)
  AND ($6::timestamptz IS NULL OR data >= $6)
  AND ($7::timestamptz IS NULL OR data <= $7)
ORDER BY data DESC, created_at DESC
LIMIT $8 OFFSET $9;

-- name: UpdateMovimentacaoCaixa :one
UPDATE movimentacoes_caixa
SET data = $2, tipo = $3, valor = $4, descricao = $5, categoria_financeira_id = $6, conta_gerencial_id = $7, origem = $8, origem_id = $9, observacoes = $10
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, loja_id, data, tipo, valor, descricao, categoria_financeira_id, conta_gerencial_id, origem, origem_id, observacoes, created_at, updated_at, deleted_at;

-- name: DeleteMovimentacaoCaixa :exec
UPDATE movimentacoes_caixa
SET deleted_at = NOW()
WHERE id = $1;

-- name: GetSaldoPeriodo :one
SELECT COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE -valor END), 0) as saldo
FROM movimentacoes_caixa
WHERE empresa_id = $1
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND data >= $3
  AND data <= $4
  AND deleted_at IS NULL;

-- name: GetTotaisPorContaGerencial :many
SELECT conta_gerencial_id, COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE -valor END), 0) as total
FROM movimentacoes_caixa
WHERE empresa_id = $1
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND data >= $3
  AND data <= $4
  AND deleted_at IS NULL
  AND conta_gerencial_id IS NOT NULL
GROUP BY conta_gerencial_id;

-- name: GetTotaisPorCategoriaFinanceira :many
SELECT categoria_financeira_id, COALESCE(SUM(CASE WHEN tipo = 'entrada' THEN valor ELSE -valor END), 0) as total
FROM movimentacoes_caixa
WHERE empresa_id = $1
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND data >= $3
  AND data <= $4
  AND deleted_at IS NULL
  AND categoria_financeira_id IS NOT NULL
GROUP BY categoria_financeira_id;

-- ============================================================
-- DRE Gerencial
-- ============================================================

-- name: GetDREGerencialReceitas :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome, cg.tipo,
       COALESCE(SUM(CASE WHEN mc.tipo = 'entrada' THEN mc.valor ELSE -mc.valor END), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.tipo = 'receita'
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome, cg.tipo
ORDER BY cg.codigo;

-- name: GetDREGerencialCustos :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome, cg.tipo,
       COALESCE(SUM(CASE WHEN mc.tipo = 'saida' THEN mc.valor ELSE -mc.valor END), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.tipo = 'custo'
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome, cg.tipo
ORDER BY cg.codigo;

-- name: GetDREGerencialDespesas :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome, cg.tipo,
       COALESCE(SUM(CASE WHEN mc.tipo = 'saida' THEN mc.valor ELSE -mc.valor END), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.tipo = 'despesa'
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome, cg.tipo
ORDER BY cg.codigo;

-- name: GetDREGerencialInvestimentos :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome, cg.tipo,
       COALESCE(SUM(CASE WHEN mc.tipo = 'saida' THEN mc.valor ELSE -mc.valor END), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.tipo = 'investimento'
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome, cg.tipo
ORDER BY cg.codigo;

-- ============================================================
-- Fluxo de Caixa
-- ============================================================

-- name: GetFluxoCaixaEntradas :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome,
       COALESCE(SUM(mc.valor), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.tipo = 'entrada'
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome
ORDER BY cg.codigo;

-- name: GetFluxoCaixaSaidas :many
SELECT cg.id as conta_gerencial_id, cg.codigo, cg.nome,
       COALESCE(SUM(mc.valor), 0) as valor
FROM contas_gerenciais cg
LEFT JOIN movimentacoes_caixa mc ON mc.conta_gerencial_id = cg.id
    AND mc.empresa_id = $1
    AND ($2::uuid IS NULL OR mc.loja_id = $2)
    AND mc.data >= $3
    AND mc.data <= $4
    AND mc.tipo = 'saida'
    AND mc.deleted_at IS NULL
WHERE cg.empresa_id = $1
  AND cg.deleted_at IS NULL
GROUP BY cg.id, cg.codigo, cg.nome
ORDER BY cg.codigo;
