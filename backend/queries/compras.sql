-- name: CreateFornecedor :one
INSERT INTO fornecedores (
    empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at;

-- name: GetFornecedor :one
SELECT id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at
FROM fornecedores
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFornecedorByCNPJCPF :one
SELECT id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at
FROM fornecedores
WHERE empresa_id = $1 AND cnpj_cpf = $2 AND deleted_at IS NULL;

-- name: ListFornecedoresByEmpresa :many
SELECT id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at
FROM fornecedores
WHERE empresa_id = $1 AND ($2::boolean IS NULL OR ativo = $2) AND deleted_at IS NULL
ORDER BY nome
LIMIT $3 OFFSET $4;

-- name: ListFornecedoresByFilter :many
SELECT id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at
FROM fornecedores
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::boolean IS NULL OR ativo = $2)
  AND ($3::varchar IS NULL OR nome ILIKE '%' || $3 || '%')
  AND ($4::varchar IS NULL OR cnpj_cpf ILIKE '%' || $4 || '%')
ORDER BY nome
LIMIT $5 OFFSET $6;

-- name: UpdateFornecedor :one
UPDATE fornecedores
SET nome = $2, cnpj_cpf = $3, email = $4, telefone = $5, banco = $6, agencia = $7, conta = $8, tipo_conta = $9, pix_chave = $10, observacoes = $11, ativo = $12
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, nome, cnpj_cpf, email, telefone, banco, agencia, conta, tipo_conta, pix_chave, observacoes, ativo, created_at, updated_at, deleted_at;

-- name: DeleteFornecedor :exec
UPDATE fornecedores
SET deleted_at = NOW()
WHERE id = $1;

-- name: CreateOrdemCompra :one
INSERT INTO ordens_compra (
    empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at;

-- name: GetOrdemCompra :one
SELECT id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at
FROM ordens_compra
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOrdensCompraByEmpresa :many
SELECT id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at
FROM ordens_compra
WHERE empresa_id = $1 AND deleted_at IS NULL
ORDER BY data_compra DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdensCompraByLoja :many
SELECT id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at
FROM ordens_compra
WHERE loja_id = $1 AND deleted_at IS NULL
ORDER BY data_compra DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdensCompraByFilter :many
SELECT id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at
FROM ordens_compra
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::uuid IS NULL OR fornecedor_id = $3)
  AND ($4::varchar IS NULL OR status = $4)
  AND ($5::date IS NULL OR data_compra >= $5)
  AND ($6::date IS NULL OR data_compra <= $6)
ORDER BY data_compra DESC, created_at DESC
LIMIT $7 OFFSET $8;

-- name: UpdateOrdemCompra :one
UPDATE ordens_compra
SET loja_id = $2, fornecedor_id = $3, data_compra = $4, data_prevista_entrega = $5, status = $6, observacoes = $7, valor_total = $8
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at;

-- name: UpdateOrdemCompraStatus :one
UPDATE ordens_compra
SET status = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, loja_id, fornecedor_id, data_compra, data_prevista_entrega, status, observacoes, valor_total, created_at, updated_at, deleted_at;

-- name: DeleteOrdemCompra :exec
UPDATE ordens_compra
SET deleted_at = NOW()
WHERE id = $1;

-- name: CreateOrdemCompraItem :one
INSERT INTO ordem_compra_itens (
    ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario, quantidade_recebida, created_at, updated_at;

-- name: GetOrdemCompraItem :one
SELECT id, ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario, quantidade_recebida, created_at, updated_at
FROM ordem_compra_itens
WHERE id = $1;

-- name: ListOrdemCompraItensByOrdemCompra :many
SELECT id, ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario, quantidade_recebida, created_at, updated_at
FROM ordem_compra_itens
WHERE ordem_compra_id = $1
ORDER BY created_at;

-- name: UpdateOrdemCompraItem :one
UPDATE ordem_compra_itens
SET insumo_id = $2, quantidade = $3, unidade_medida = $4, preco_unitario = $5
WHERE id = $1
RETURNING id, ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario, quantidade_recebida, created_at, updated_at;

-- name: UpdateOrdemCompraItemQuantidadeRecebida :one
UPDATE ordem_compra_itens
SET quantidade_recebida = $2
WHERE id = $1
RETURNING id, ordem_compra_id, insumo_id, quantidade, unidade_medida, preco_unitario, quantidade_recebida, created_at, updated_at;

-- name: DeleteOrdemCompraItem :exec
DELETE FROM ordem_compra_itens
WHERE id = $1;

-- name: DeleteOrdemCompraItensByOrdemCompra :exec
DELETE FROM ordem_compra_itens
WHERE ordem_compra_id = $1;

-- name: CreateRecebimentoCompra :one
INSERT INTO recebimentos_compra (
    ordem_compra_id, loja_id, data_recebimento, observacoes, permitir_excedente
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, ordem_compra_id, loja_id, data_recebimento, observacoes, permitir_excedente, created_at, updated_at;

-- name: GetRecebimentoCompra :one
SELECT id, ordem_compra_id, loja_id, data_recebimento, observacoes, permitir_excedente, created_at, updated_at
FROM recebimentos_compra
WHERE id = $1;

-- name: ListRecebimentosCompraByOrdemCompra :many
SELECT id, ordem_compra_id, loja_id, data_recebimento, observacoes, permitir_excedente, created_at, updated_at
FROM recebimentos_compra
WHERE ordem_compra_id = $1
ORDER BY data_recebimento DESC;

-- name: ListRecebimentosCompraByFilter :many
SELECT id, ordem_compra_id, loja_id, data_recebimento, observacoes, permitir_excedente, created_at, updated_at
FROM recebimentos_compra
WHERE ($1::uuid IS NULL OR ordem_compra_id = $1)
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::timestamptz IS NULL OR data_recebimento >= $3)
  AND ($4::timestamptz IS NULL OR data_recebimento <= $4)
ORDER BY data_recebimento DESC
LIMIT $5 OFFSET $6;

-- name: CreateRecebimentoCompraItem :one
INSERT INTO recebimento_compra_itens (
    recebimento_compra_id, ordem_compra_item_id, insumo_id, quantidade_recebida, preco_unitario
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, recebimento_compra_id, ordem_compra_item_id, insumo_id, quantidade_recebida, preco_unitario, created_at;

-- name: ListRecebimentoCompraItensByRecebimento :many
SELECT id, recebimento_compra_id, ordem_compra_item_id, insumo_id, quantidade_recebida, preco_unitario, created_at
FROM recebimento_compra_itens
WHERE recebimento_compra_id = $1
ORDER BY created_at;

-- name: CreateContaPagar :one
INSERT INTO contas_pagar (
    empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, status, forma_pagamento, observacoes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at;

-- name: GetContaPagar :one
SELECT id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at
FROM contas_pagar
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListContasPagarByEmpresa :many
SELECT id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at
FROM contas_pagar
WHERE empresa_id = $1 AND deleted_at IS NULL
ORDER BY data_vencimento, created_at
LIMIT $2 OFFSET $3;

-- name: ListContasPagarByFornecedor :many
SELECT id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at
FROM contas_pagar
WHERE fornecedor_id = $1 AND deleted_at IS NULL
ORDER BY data_vencimento, created_at
LIMIT $2 OFFSET $3;

-- name: ListContasPagarByStatus :many
SELECT id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at
FROM contas_pagar
WHERE empresa_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY data_vencimento, created_at
LIMIT $3 OFFSET $4;

-- name: ListContasPagarByFilter :many
SELECT id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at
FROM contas_pagar
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::uuid IS NULL OR fornecedor_id = $3)
  AND ($4::varchar IS NULL OR status = $4)
  AND ($5::date IS NULL OR data_vencimento >= $5)
  AND ($6::date IS NULL OR data_vencimento <= $6)
  AND ($7::date IS NULL OR data_pagamento >= $7)
  AND ($8::date IS NULL OR data_pagamento <= $8)
ORDER BY data_vencimento, created_at
LIMIT $9 OFFSET $10;

-- name: UpdateContaPagar :one
UPDATE contas_pagar
SET descricao = $2, valor = $3, data_emissao = $4, data_vencimento = $5, data_pagamento = $6, status = $7, forma_pagamento = $8, observacoes = $9
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at;

-- name: UpdateContaPagarStatus :one
UPDATE contas_pagar
SET status = $2, data_pagamento = $3, forma_pagamento = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, loja_id, fornecedor_id, ordem_compra_id, recebimento_compra_id, descricao, valor, data_emissao, data_vencimento, data_pagamento, status, forma_pagamento, observacoes, created_at, updated_at, deleted_at;

-- name: DeleteContaPagar :exec
UPDATE contas_pagar
SET deleted_at = NOW()
WHERE id = $1;

-- name: GetOrdemCompraWithItens :many
SELECT
    oc.id, oc.empresa_id, oc.loja_id, oc.fornecedor_id, oc.data_compra, oc.data_prevista_entrega, oc.status, oc.observacoes, oc.valor_total, oc.created_at, oc.updated_at, oc.deleted_at,
    oci.id as item_id, oci.ordem_compra_id, oci.insumo_id, oci.quantidade, oci.unidade_medida, oci.preco_unitario, oci.quantidade_recebida, oci.created_at as item_created_at, oci.updated_at as item_updated_at
FROM ordens_compra oc
LEFT JOIN ordem_compra_itens oci ON oci.ordem_compra_id = oc.id
WHERE oc.id = $1 AND oc.deleted_at IS NULL
ORDER BY oci.created_at;

-- name: GetInsumoForUpdate :one
SELECT id, empresa_id, nome, unidade_medida, custo_unitario_atual, fornecedor, ativo, created_at, updated_at, deleted_at
FROM insumos
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateInsumoCustoUnitario :one
UPDATE insumos
SET custo_unitario_atual = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, empresa_id, nome, unidade_medida, custo_unitario_atual, fornecedor, ativo, created_at, updated_at, deleted_at;