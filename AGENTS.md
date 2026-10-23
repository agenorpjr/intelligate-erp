# AGENTS.md — Regras do Projeto IntelliGate ERP

## Visão Geral do Projeto
**IntelliGate ERP** — ERP food service multi-loja, com integração a PDVs externos e agentes de IA.

## Stack Tecnológica
- **Backend**: Go + Chi + PostgreSQL + pgx v5 + sqlc + Goose
- **Frontend web**: Next.js + TypeScript
- **Ambiente**: WSL2/Ubuntu + Docker Compose
- **Testes**: Go test e Testcontainers
- **Arquitetura**: monolito modular / hexagonal
- **Integração inicial**: Gula Menu

---

## Regras Obrigatórias

### Banco de Dados & Migrations
- ❌ **Não usar ORM pesado** — usar `pgx` + `sqlc` para queries tipadas.
- ❌ **Não alterar migrations antigas** — migrations são imutáveis após aplicadas.
- ✅ **Toda mudança de schema exige nova migration versionada** (Goose).
- ❌ **Não usar `DROP`, `TRUNCATE` ou `ALTER` destrutivo** sem aprovação explícita por escrito.

### Arquivos Sensíveis
- ❌ **Não alterar** `.env`, `docker-compose.yml`, CI ou `AGENTS.md` sem aprovação.

### API & Validação
- ✅ **Toda rota deve**:
  1. Validar entrada (request body, query params, path params)
  2. Retornar erro estruturado (código, mensagem, detalhes)
  3. Ter teste automatizado

### Entidades & Auditoria
- ✅ **Toda entidade deve ter** `created_at` e `updated_at` (timestamps com timezone).
- ❌ **Nunca apagar dados fisicamente** — usar `status`, `deleted_at` (soft delete) ou tabela de histórico.

### Integrações Externas
- ✅ **Toda integração deve ter idempotência** (chave de idempotência, deduplicação).

### Qualidade de Código (Pré-commit / CI Local)
Antes de finalizar qualquer alteração no backend, rodar:
```bash
gofmt ./...
go vet ./...
golangci-lint run
go test ./...
```

### Comunicação de Mudanças
Sempre informar ao final:
- 📁 **Arquivos criados/alterados**
- ⚠️ **Riscos / impacto**
- 🧪 **Como testar** (comandos, endpoints, cenários)

---

## Convenções de Código Go
- Seguir *Effective Go* e *Go Code Review Comments*.
- Usar `context.Context` como primeiro parâmetro em funções de I/O.
- Erros: usar `fmt.Errorf` com `%w` para wrap; sentinel errors para casos esperados.
- Interfaces pequenas, definidas no pacote consumidor (Dependency Inversion).

## Estrutura de Pastas (Backend)
```
cmd/           # entrypoints (api, worker, migrate)
internal/
  domain/      # entidades, value objects, interfaces de repositório
  usecase/     # casos de uso (application services)
  adapter/
    http/      # handlers Chi, middlewares, DTOs
    db/        # implementações sqlc, repositórios
    external/  # clientes HTTP (Gula Menu, etc.)
  config/      # carregamento de config (viper/env)
pkg/           # código compartilhado reutilizável
migrations/    # arquivos .sql versionados (Goose)
```

## Convenções de API
- RESTful com JSON.
- Versionamento no path: `/api/v1/...`
- Paginação: `?page=1&limit=20` (máx 100).
- Filtros: `?filter[field]=value`
- Erros: RFC 7807 (Problem Details) — `type`, `title`, `status`, `detail`, `instance`.

## Testes
- Unitários: `*_test.go` ao lado do código.
- Integração: Testcontainers (PostgreSQL real).
- Cobertura mínima alvo: 80% em `usecase/` e `adapter/`.

## Git & Commits
- Commits semânticos (Conventional Commits): `feat:`, `fix:`, `refactor:`, `test:`, `chore:`, `docs:`.
- Branch por issue/PR: `feat/short-description`, `fix/short-description`.
- PRs exigem: descrição, testes passando, lint passando, revisão de 1+ aprovador.

---

## Checklist de Entrega (Definition of Done)
- [ ] Código compila (`go build ./...`)
- [ ] `gofmt`, `go vet`, `golangci-lint` passam
- [ ] Testes passam (`go test ./... -race`)
- [ ] Migration criada (se houver mudança de schema)
- [ ] Documentação atualizada (README, OpenAPI/Swagger se aplicável)
- [ ] Arquivos alterados listados, riscos descritos, instruções de teste fornecidas