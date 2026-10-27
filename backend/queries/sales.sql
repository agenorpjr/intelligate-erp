-- PDV Origem
-- name: CreatePdvOrigem :one
INSERT INTO pdvs_origem (empresa_id, nome, codigo, tipo, configuracao)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPdvOrigem :one
SELECT * FROM pdvs_origem WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPdvOrigemByCodigo :one
SELECT * FROM pdvs_origem WHERE empresa_id = $1 AND codigo = $2 AND deleted_at IS NULL;

-- name: ListPdvsOrigemByEmpresa :many
SELECT * FROM pdvs_origem WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePdvOrigem :one
UPDATE pdvs_origem
SET nome = $2, codigo = $3, tipo = $4, configuracao = $5, ativo = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePdvOrigem :exec
UPDATE pdvs_origem SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Categoria Canônica
-- name: CreateCategoriaCanonica :one
INSERT INTO categorias_canonicas (empresa_id, nome, descricao, cor, icone)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCategoriaCanonica :one
SELECT * FROM categorias_canonicas WHERE id = $1 AND deleted_at IS NULL;

-- name: ListCategoriasCanonicasByEmpresa :many
SELECT * FROM categorias_canonicas WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateCategoriaCanonica :one
UPDATE categorias_canonicas
SET nome = $2, descricao = $3, cor = $4, icone = $5, ativo = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCategoriaCanonica :exec
UPDATE categorias_canonicas SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Produto Canônico
-- name: CreateProdutoCanonico :one
INSERT INTO produtos_canonicos (empresa_id, categoria_id, sku, nome, descricao, unidade, preco_base, ncm, cest, tributacao)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetProdutoCanonico :one
SELECT * FROM produtos_canonicos WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProdutoCanonicoBySKU :one
SELECT * FROM produtos_canonicos WHERE empresa_id = $1 AND sku = $2 AND deleted_at IS NULL;

-- name: ListProdutosCanonicosByEmpresa :many
SELECT * FROM produtos_canonicos WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateProdutoCanonico :one
UPDATE produtos_canonicos
SET categoria_id = $2, sku = $3, nome = $4, descricao = $5, unidade = $6, preco_base = $7, ncm = $8, cest = $9, tributacao = $10, ativo = $11, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProdutoCanonico :exec
UPDATE produtos_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Pedido Canônico
-- name: CreatePedidoCanonico :one
INSERT INTO pedidos_canonicos (empresa_id, loja_id, pdv_origem_id, numero_externo, data_hora, status, valor_total, valor_desconto, valor_acrescimo, idempotency_key, cliente_nome, cliente_documento, cliente_email, observacoes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetPedidoCanonico :one
SELECT * FROM pedidos_canonicos WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPedidoCanonicoByIdempotencyKey :one
SELECT * FROM pedidos_canonicos WHERE empresa_id = $1 AND idempotency_key = $2 AND deleted_at IS NULL;

-- name: ListPedidosCanonicos :many
SELECT * FROM pedidos_canonicos
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::timestamptz IS NULL OR data_hora >= $3)
  AND ($4::timestamptz IS NULL OR data_hora <= $4)
  AND ($5::text IS NULL OR status = $5)
ORDER BY data_hora DESC LIMIT $6 OFFSET $7;

-- name: UpdatePedidoCanonico :one
UPDATE pedidos_canonicos
SET loja_id = $2, pdv_origem_id = $3, numero_externo = $4, data_hora = $5, status = $6, valor_total = $7, valor_desconto = $8, valor_acrescimo = $9, cliente_nome = $10, cliente_documento = $11, cliente_email = $12, observacoes = $13, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePedidoCanonico :exec
UPDATE pedidos_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Pedido Item Canônico
-- name: CreatePedidoItemCanonico :one
INSERT INTO pedido_itens_canonicos (pedido_id, produto_canonico_id, sku_externo, nome, quantidade, preco_unitario, valor_total, desconto, modificadores, observacoes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetPedidoItemCanonico :one
SELECT * FROM pedido_itens_canonicos WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPedidoItensByPedido :many
SELECT * FROM pedido_itens_canonicos WHERE pedido_id = $1 AND deleted_at IS NULL ORDER BY created_at;

-- name: UpdatePedidoItemCanonico :one
UPDATE pedido_itens_canonicos
SET produto_canonico_id = $2, sku_externo = $3, nome = $4, quantidade = $5, preco_unitario = $6, valor_total = $7, desconto = $8, modificadores = $9, observacoes = $10, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePedidoItemCanonico :exec
UPDATE pedido_itens_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: DeletePedidoItensByPedido :exec
UPDATE pedido_itens_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE pedido_id = $1;

-- Pedido Pagamento Canônico
-- name: CreatePedidoPagamentoCanonico :one
INSERT INTO pedido_pagamentos_canonicos (pedido_id, forma_pagamento, valor, taxa, troco, referencia_externa, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPedidoPagamentoCanonico :one
SELECT * FROM pedido_pagamentos_canonicos WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPedidoPagamentosByPedido :many
SELECT * FROM pedido_pagamentos_canonicos WHERE pedido_id = $1 AND deleted_at IS NULL ORDER BY created_at;

-- name: UpdatePedidoPagamentoCanonico :one
UPDATE pedido_pagamentos_canonicos
SET forma_pagamento = $2, valor = $3, taxa = $4, troco = $5, referencia_externa = $6, status = $7, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePedidoPagamentoCanonico :exec
UPDATE pedido_pagamentos_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: DeletePedidoPagamentosByPedido :exec
UPDATE pedido_pagamentos_canonicos SET deleted_at = NOW(), updated_at = NOW() WHERE pedido_id = $1;

-- Evento Ingestão
-- name: CreateEventoIngestao :one
INSERT INTO eventos_ingestao (empresa_id, loja_id, pdv_origem_id, tipo_evento, payload_original, status, erro, tentativas, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetEventoIngestao :one
SELECT * FROM eventos_ingestao WHERE id = $1 AND deleted_at IS NULL;

-- name: GetEventoIngestaoByIdempotencyKey :one
SELECT * FROM eventos_ingestao WHERE empresa_id = $1 AND idempotency_key = $2 AND deleted_at IS NULL;

-- name: ListEventosIngestao :many
SELECT * FROM eventos_ingestao
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::uuid IS NULL OR pdv_origem_id = $3)
  AND ($4::text IS NULL OR tipo_evento = $4)
  AND ($5::text IS NULL OR status = $5)
ORDER BY created_at DESC LIMIT $6 OFFSET $7;

-- name: UpdateEventoIngestao :one
UPDATE eventos_ingestao
SET status = $2, erro = $3, tentativas = $4, processado_em = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEventoIngestao :exec
UPDATE eventos_ingestao SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Resumo de Vendas
-- name: GetVendasResumo :one
SELECT
    COALESCE(SUM(valor_total), 0) AS total_vendas,
    COUNT(*) AS total_pedidos,
    CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(valor_total), 0) / COUNT(*) ELSE 0 END AS ticket_medio,
    COALESCE(SUM(valor_desconto), 0) AS total_descontos
FROM pedidos_canonicos
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::uuid IS NULL OR loja_id = $2)
  AND ($3::timestamptz IS NULL OR data_hora >= $3)
  AND ($4::timestamptz IS NULL OR data_hora <= $4)
  AND status != 'cancelado';

-- PDV Integração
-- name: CreatePdvIntegracao :one
INSERT INTO pdv_integracoes (empresa_id, loja_id, provedor, ambiente, url_base, api_key_encrypted, client_id, client_secret_encrypted)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPdvIntegracao :one
SELECT * FROM pdv_integracoes WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPdvIntegracaoByLojaProvedor :one
SELECT * FROM pdv_integracoes WHERE loja_id = $1 AND provedor = $2 AND deleted_at IS NULL;

-- name: ListPdvIntegracoesByEmpresa :many
SELECT * FROM pdv_integracoes
WHERE empresa_id = $1
  AND deleted_at IS NULL
  AND ($2::text IS NULL OR provedor = $2)
  AND ($3::boolean IS NULL OR ativo = $3)
ORDER BY created_at DESC LIMIT $4 OFFSET $5;

-- name: UpdatePdvIntegracao :one
UPDATE pdv_integracoes
SET provedor = $2, ambiente = $3, url_base = $4, api_key_encrypted = $5, client_id = $6, client_secret_encrypted = $7, ativo = $8, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdatePdvIntegracaoUltimaSincronizacao :one
UPDATE pdv_integracoes
SET ultima_sincronizacao = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePdvIntegracao :exec
UPDATE pdv_integracoes SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;