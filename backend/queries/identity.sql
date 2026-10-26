-- name: CreateEmpresa :one
INSERT INTO empresas (cnpj, razao_social, nome_fantasia)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmpresa :one
SELECT * FROM empresas WHERE id = $1 AND deleted_at IS NULL;

-- name: GetEmpresaByCNPJ :one
SELECT * FROM empresas WHERE cnpj = $1 AND deleted_at IS NULL;

-- name: ListEmpresas :many
SELECT * FROM empresas WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateEmpresa :one
UPDATE empresas
SET razao_social = $2, nome_fantasia = $3, ativo = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEmpresa :exec
UPDATE empresas SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: CreateLoja :one
INSERT INTO lojas (empresa_id, codigo, nome, endereco, fuso_horario)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLoja :one
SELECT * FROM lojas WHERE id = $1 AND deleted_at IS NULL;

-- name: ListLojasByEmpresa :many
SELECT * FROM lojas WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateLoja :one
UPDATE lojas
SET codigo = $2, nome = $3, endereco = $4, fuso_horario = $5, ativo = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteLoja :exec
UPDATE lojas SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: CreateUsuario :one
INSERT INTO usuarios (empresa_id, email, nome, senha_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUsuario :one
SELECT * FROM usuarios WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUsuarioByEmail :one
SELECT * FROM usuarios WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUsuarioByEmailAndEmpresa :one
SELECT * FROM usuarios WHERE email = $1 AND empresa_id = $2 AND deleted_at IS NULL;

-- name: ListUsuariosByEmpresa :many
SELECT * FROM usuarios WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateUsuario :one
UPDATE usuarios
SET nome = $2, email = $3, ativo = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUsuarioSenha :exec
UPDATE usuarios SET senha_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateUsuarioUltimoLogin :exec
UPDATE usuarios SET ultimo_login = NOW(), updated_at = NOW() WHERE id = $1;

-- name: DeleteUsuario :exec
UPDATE usuarios SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: CreatePapel :one
INSERT INTO papeis (empresa_id, nome, descricao, sistema)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPapel :one
SELECT * FROM papeis WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPapeisByEmpresa :many
SELECT * FROM papeis WHERE empresa_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePapel :one
UPDATE papeis
SET nome = $2, descricao = $3, sistema = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeletePapel :exec
UPDATE papeis SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: CreatePermissao :one
INSERT INTO permissoes (chave, descricao)
VALUES ($1, $2)
RETURNING *;

-- name: GetPermissao :one
SELECT * FROM permissoes WHERE id = $1;

-- name: GetPermissaoByChave :one
SELECT * FROM permissoes WHERE chave = $1;

-- name: ListPermissoes :many
SELECT * FROM permissoes ORDER BY chave LIMIT $1 OFFSET $2;

-- name: CreateUsuarioPapel :one
INSERT INTO usuarios_papeis (usuario_id, papel_id, loja_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUsuarioPapel :one
SELECT * FROM usuarios_papeis WHERE id = $1;

-- name: ListUsuarioPapeisByUsuario :many
SELECT up.*, p.nome as papel_nome, l.nome as loja_nome
FROM usuarios_papeis up
JOIN papeis p ON up.papel_id = p.id
LEFT JOIN lojas l ON up.loja_id = l.id
WHERE up.usuario_id = $1;

-- name: DeleteUsuarioPapel :exec
DELETE FROM usuarios_papeis WHERE id = $1;

-- name: CreatePapelPermissao :exec
INSERT INTO papeis_permissoes (papel_id, permissao_id)
VALUES ($1, $2)
ON CONFLICT (papel_id, permissao_id) DO NOTHING;

-- name: DeletePapelPermissao :exec
DELETE FROM papeis_permissoes WHERE papel_id = $1 AND permissao_id = $2;

-- name: ListPermissoesByPapel :many
SELECT p.* FROM permissoes p
JOIN papeis_permissoes pp ON p.id = pp.permissao_id
WHERE pp.papel_id = $1;

-- name: ListPermissoesByUsuario :many
SELECT DISTINCT p.* FROM permissoes p
JOIN papeis_permissoes pp ON p.id = pp.permissao_id
JOIN usuarios_papeis up ON pp.papel_id = up.papel_id
WHERE up.usuario_id = $1 AND (up.loja_id IS NULL OR up.loja_id = $2);

-- name: CreateAuditoria :one
INSERT INTO auditoria (empresa_id, usuario_id, entidade, entidade_id, acao, antes, depois, ip, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAuditoria :many
SELECT * FROM auditoria
WHERE ($1::uuid IS NULL OR empresa_id = $1)
  AND ($2::uuid IS NULL OR usuario_id = $2)
  AND ($3::text IS NULL OR entidade = $3)
  AND ($4::text IS NULL OR acao = $4)
ORDER BY created_at DESC LIMIT $5 OFFSET $6;