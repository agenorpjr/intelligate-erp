-- +goose Up
-- Identity, Multi-loja, RBAC e Auditoria (novas tabelas com nomes em plural)

-- Empresas (tenant raiz)
CREATE TABLE empresas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    razao_social VARCHAR(255) NOT NULL,
    nome_fantasia VARCHAR(255),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_empresas_cnpj ON empresas(cnpj);
CREATE INDEX idx_empresas_ativo ON empresas(ativo) WHERE ativo = true;

-- Lojas (unidades operacionais)
CREATE TABLE lojas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    codigo VARCHAR(50) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    endereco JSONB,
    fuso_horario VARCHAR(50) NOT NULL DEFAULT 'America/Sao_Paulo',
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, codigo)
);

CREATE INDEX idx_lojas_empresa_id ON lojas(empresa_id);
CREATE INDEX idx_lojas_ativo ON lojas(ativo) WHERE ativo = true;

-- Usuários
CREATE TABLE usuarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    email VARCHAR(255) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT true,
    ultimo_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, email)
);

CREATE INDEX idx_usuarios_empresa_id ON usuarios(empresa_id);
CREATE INDEX idx_usuarios_ativo ON usuarios(ativo) WHERE ativo = true;
CREATE INDEX idx_usuarios_email ON usuarios(email);

-- Papéis (roles)
CREATE TABLE papeis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresas(id) ON DELETE RESTRICT,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    sistema BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, nome)
);

CREATE INDEX idx_papeis_empresa_id ON papeis(empresa_id);

-- Permissões
CREATE TABLE permissoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chave VARCHAR(100) NOT NULL UNIQUE,
    descricao TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Usuário-Papel (many-to-many com escopo de loja)
CREATE TABLE usuarios_papeis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    papel_id UUID NOT NULL REFERENCES papeis(id) ON DELETE CASCADE,
    loja_id UUID REFERENCES lojas(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (usuario_id, papel_id, loja_id)
);

CREATE INDEX idx_usuarios_papeis_usuario_id ON usuarios_papeis(usuario_id);
CREATE INDEX idx_usuarios_papeis_papel_id ON usuarios_papeis(papel_id);
CREATE INDEX idx_usuarios_papeis_loja_id ON usuarios_papeis(loja_id);

-- Papel-Permissão (many-to-many)
CREATE TABLE papeis_permissoes (
    papel_id UUID NOT NULL REFERENCES papeis(id) ON DELETE CASCADE,
    permissao_id UUID NOT NULL REFERENCES permissoes(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (papel_id, permissao_id)
);

-- Auditoria
CREATE TABLE auditoria (
    id BIGSERIAL PRIMARY KEY,
    empresa_id UUID REFERENCES empresas(id) ON DELETE SET NULL,
    usuario_id UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    entidade VARCHAR(100) NOT NULL,
    entidade_id UUID NOT NULL,
    acao VARCHAR(50) NOT NULL,
    antes JSONB,
    depois JSONB,
    ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auditoria_empresa_id_v2 ON auditoria(empresa_id);
CREATE INDEX idx_auditoria_usuario_id ON auditoria(usuario_id);
CREATE INDEX idx_auditoria_entidade_v2 ON auditoria(entidade, entidade_id);
CREATE INDEX idx_auditoria_acao ON auditoria(acao);
CREATE INDEX idx_auditoria_created_at_v2 ON auditoria(created_at DESC);

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
CREATE TRIGGER update_empresas_updated_at BEFORE UPDATE ON empresas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lojas_updated_at BEFORE UPDATE ON lojas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usuarios_updated_at BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_papeis_updated_at BEFORE UPDATE ON papeis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_papeis_updated_at ON papeis;
DROP TRIGGER IF EXISTS update_usuarios_updated_at ON usuarios;
DROP TRIGGER IF EXISTS update_lojas_updated_at ON lojas;
DROP TRIGGER IF EXISTS update_empresas_updated_at ON empresas;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS auditoria;
DROP TABLE IF EXISTS papeis_permissoes;
DROP TABLE IF EXISTS usuarios_papeis;
DROP TABLE IF EXISTS permissoes;
DROP TABLE IF EXISTS papeis;
DROP TABLE IF EXISTS usuarios;
DROP TABLE IF EXISTS lojas;
DROP TABLE IF EXISTS empresas;
