-- +goose Up
-- Initial schema for IntelliGate ERP
-- Creates core tables for multi-tenancy: empresa, loja, usuario, papel, permissao

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Empresa (tenant raiz)
CREATE TABLE empresa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    razao_social VARCHAR(255) NOT NULL,
    nome_fantasia VARCHAR(255),
    ativo BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_empresa_cnpj ON empresa(cnpj);
CREATE INDEX idx_empresa_ativo ON empresa(ativo) WHERE ativo = true;

-- Loja (unidade operacional)
CREATE TABLE loja (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresa(id) ON DELETE RESTRICT,
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

CREATE INDEX idx_loja_empresa_id ON loja(empresa_id);
CREATE INDEX idx_loja_ativo ON loja(ativo) WHERE ativo = true;

-- Papel (role)
CREATE TABLE papel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresa(id) ON DELETE RESTRICT,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    sistema BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (empresa_id, nome)
);

CREATE INDEX idx_papel_empresa_id ON papel(empresa_id);

-- Permissao (permission)
CREATE TABLE permissao (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chave VARCHAR(100) NOT NULL UNIQUE,
    descricao TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Papel_Permissao (many-to-many)
CREATE TABLE papel_permissao (
    papel_id UUID NOT NULL REFERENCES papel(id) ON DELETE CASCADE,
    permissao_id UUID NOT NULL REFERENCES permissao(id) ON DELETE CASCADE,
    PRIMARY KEY (papel_id, permissao_id)
);

-- Usuario
CREATE TABLE usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id UUID NOT NULL REFERENCES empresa(id) ON DELETE RESTRICT,
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

CREATE INDEX idx_usuario_empresa_id ON usuario(empresa_id);
CREATE INDEX idx_usuario_ativo ON usuario(ativo) WHERE ativo = true;

-- Usuario_Loja_Papel (user-store-role)
CREATE TABLE usuario_loja_papel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    loja_id UUID NOT NULL REFERENCES loja(id) ON DELETE CASCADE,
    papel_id UUID NOT NULL REFERENCES papel(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (usuario_id, loja_id, papel_id)
);

CREATE INDEX idx_ulp_usuario_id ON usuario_loja_papel(usuario_id);
CREATE INDEX idx_ulp_loja_id ON usuario_loja_papel(loja_id);

-- Auditoria log
CREATE TABLE auditoria_log (
    id BIGSERIAL PRIMARY KEY,
    empresa_id UUID REFERENCES empresa(id) ON DELETE SET NULL,
    usuario_id UUID REFERENCES usuario(id) ON DELETE SET NULL,
    entidade VARCHAR(100) NOT NULL,
    entidade_id UUID NOT NULL,
    acao VARCHAR(50) NOT NULL,
    antes JSONB,
    depois JSONB,
    ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auditoria_empresa_id ON auditoria_log(empresa_id);
CREATE INDEX idx_auditoria_entidade ON auditoria_log(entidade, entidade_id);
CREATE INDEX idx_auditoria_created_at ON auditoria_log(created_at DESC);

-- Trigger function for updated_at (using goose statement separator)
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
CREATE TRIGGER update_empresa_updated_at BEFORE UPDATE ON empresa
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_loja_updated_at BEFORE UPDATE ON loja
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_papel_updated_at BEFORE UPDATE ON papel
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usuario_updated_at BEFORE UPDATE ON usuario
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();