-- ============================================================
-- Agente de IA para análise gerencial (Etapa 13)
-- ============================================================
-- Sessões de conversa e mensagens. O contexto usado é persistido em JSONB
-- para auditoria/rastreabilidade. Nenhum segredo é armazenado.

-- name: CreateAgenteSessao :one
INSERT INTO agente_sessoes (
    empresa_id, loja_id, usuario_id, titulo, status
) VALUES (
    sqlc.arg(empresa_id), sqlc.arg(loja_id), sqlc.arg(usuario_id),
    sqlc.arg(titulo), sqlc.arg(status)
)
RETURNING id, empresa_id, loja_id, usuario_id, titulo, status, created_at, updated_at;

-- name: GetAgenteSessao :one
SELECT id, empresa_id, loja_id, usuario_id, titulo, status, created_at, updated_at
FROM agente_sessoes
WHERE id = sqlc.arg(id);

-- name: ListAgenteSessoes :many
SELECT id, empresa_id, loja_id, usuario_id, titulo, status, created_at, updated_at
FROM agente_sessoes
WHERE empresa_id = sqlc.arg(empresa_id)
  AND (sqlc.narg(loja_id)::uuid IS NULL OR loja_id = sqlc.narg(loja_id))
  AND (sqlc.narg(usuario_id)::uuid IS NULL OR usuario_id = sqlc.narg(usuario_id))
  AND (sqlc.narg(status)::varchar IS NULL OR status = sqlc.narg(status))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpdateAgenteSessaoStatus :exec
UPDATE agente_sessoes
SET status = sqlc.arg(status),
    updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: CreateAgenteMensagem :one
INSERT INTO agente_mensagens (
    sessao_id, papel, conteudo, contexto_usado, tokens_estimados
) VALUES (
    sqlc.arg(sessao_id), sqlc.arg(papel), sqlc.arg(conteudo),
    sqlc.narg(contexto_usado), sqlc.arg(tokens_estimados)
)
RETURNING id, sessao_id, papel, conteudo, contexto_usado, tokens_estimados, created_at;

-- name: ListAgenteMensagensBySessao :many
SELECT id, sessao_id, papel, conteudo, contexto_usado, tokens_estimados, created_at
FROM agente_mensagens
WHERE sessao_id = sqlc.arg(sessao_id)
ORDER BY created_at ASC;
