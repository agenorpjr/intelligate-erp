# IntelliGate ERP

ERP food service multi-loja com integração a PDVs externos e agentes de IA.

## Stack Tecnológica

- **Backend**: Go 1.23 + Chi + PostgreSQL + pgx v5 + sqlc + Goose
- **Frontend**: Next.js + TypeScript (futuro)
- **Ambiente**: WSL2/Ubuntu + Docker Compose
- **Testes**: Go test + Testcontainers
- **Arquitetura**: Monolito modular / Hexagonal

## Estrutura do Projeto

```
intelligate-erp/
├── backend/
│   ├── cmd/api/              # Entry point da API
│   ├── internal/
│   │   ├── config/           # Configuração (viper/env)
│   │   ├── adapter/
│   │   │   ├── http/         # Handlers Chi, middlewares, DTOs
│   │   │   └── db/           # Implementações sqlc, repositórios (gerado)
│   │   ├── domain/           # Entidades, value objects, interfaces
│   │   └── usecase/          # Casos de uso (application services)
│   ├── migrations/           # Migrations Goose (versionadas)
│   ├── queries/              # Queries SQL para sqlc
│   ├── sqlc.yaml             # Configuração sqlc
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile
│   └── .env.example
├── docs/                     # Documentação do projeto
├── docker-compose.yml        # PostgreSQL apenas
├── .github/workflows/ci.yml  # CI/CD GitHub Actions
└── README.md
```

## Pré-requisitos

- Go 1.23+
- Docker e Docker Compose
- Make
- golangci-lint (para lint local)
- sqlc (para geração de código)
- goose (para migrations)

## Configuração Inicial

### 1. Subir o PostgreSQL

```bash
docker-compose up -d
```

Verificar se está saudável:
```bash
docker-compose ps
```

### 2. Configurar variáveis de ambiente

```bash
cd /home/agenor/intelligate-erp/backend
cp .env.example .env
# Edite .env se necessário (DATABASE_URL, APP_PORT, etc.)
```

### 3. Instalar dependências

```bash
cd /home/agenor/intelligate-erp/backend
make deps
```

### 4. Rodar migrations

```bash
cd /home/agenor/intelligate-erp/backend
make migrate-up
```

### 5. Gerar código sqlc (opcional, já incluso no generate)

```bash
cd /home/agenor/intelligate-erp/backend
make generate
```

### 6. Rodar a API

```bash
cd /home/agenor/intelligate-erp/backend
make run
```

A API estará disponível em `http://localhost:8080`

## Endpoints

### Health Check
```
GET /health
```

Resposta:
```json
{
  "status": "ok",
  "version": "0.1.0",
  "environment": "development",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### API v1 (placeholder)
```
GET /api/v1/
```

## Comandos Makefile

| Comando | Descrição |
|---------|-----------|
| `make run` | Inicia o servidor API |
| `make migrate-up` | Aplica todas as migrations pendentes |
| `make migrate-down` | Reverte a última migration |
| `make generate` | Gera código Go a partir das queries SQL (sqlc) |
| `make test` | Executa testes com race detector |
| `make lint` | Executa golangci-lint |
| `make fmt` | Formata código com gofmt |
| `make deps` | Baixa e organiza dependências |
| `make build` | Compila o binário |
| `make check` | Executa fmt, vet, lint e test |

## Testes

```bash
# Todos os testes
cd /home/agenor/intelligate-erp/backend
make test

# Testes com cobertura
cd /home/agenor/intelligate-erp/backend
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Lint e Formatação

```bash
# Formatar
cd /home/agenor/intelligate-erp/backend
make fmt

# Verificar (fmt + vet + lint + test)
cd /home/agenor/intelligate-erp/backend
make check
```

## Migrations

> **Importante**: Todos os comandos de migration devem ser executados a partir do diretório `backend/`

### Criar nova migration
```bash
cd /home/agenor/intelligate-erp/backend
goose -dir migrations create nome_da_migration sql
```

Isso cria dois arquivos:
- `migrations/XXXXXX_nome_da_migration.up.sql`
- `migrations/XXXXXX_nome_da_migration.down.sql`

### Aplicar migrations
```bash
cd /home/agenor/intelligate-erp/backend
make migrate-up
# ou diretamente:
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5433/intelligate?sslmode=disable" up
```

### Reverter última migration
```bash
cd /home/agenor/intelligate-erp/backend
make migrate-down
```

## Geração de Código (sqlc)

As queries SQL ficam em `backend/queries/`. Para gerar o código Go:

```bash
cd /home/agenor/intelligate-erp/backend
make generate
```

O código é gerado em `backend/internal/adapter/db/`.

## CI/CD

O pipeline GitHub Actions (`.github/workflows/ci.yml`) executa em todo push/PR:

1. **gofmt** - Verifica formatação
2. **go vet** - Análise estática
3. **golangci-lint** - Linting completo
4. **Migrations** - Aplica migrations no PostgreSQL de teste
5. **Testes** - Executa testes com race detector
6. **Build** - Compila o binário

## Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `DATABASE_URL` | URL de conexão PostgreSQL | `postgres://postgres:postgres@localhost:5433/intelligate?sslmode=disable` |
| `APP_ENV` | Ambiente (development/production) | `development` |
| `APP_PORT` | Porta do servidor HTTP | `8080` |
| `APP_VERSION` | Versão da aplicação | `0.1.0` |
| `HTTP_TIMEOUT` | Timeout do servidor HTTP | `30s` |

## Documentação

- [Visão do Produto](docs/00-visao-produto.md)
- [Arquitetura](docs/01-arquitetura.md)
- [Modelo de Domínio](docs/02-modelo-dominio.md)
- [Roadmap](docs/03-roadmap.md)
- [Regras do Projeto (AGENTS.md)](AGENTS.md)

## Licença

Proprietário - IntelliGate ERP