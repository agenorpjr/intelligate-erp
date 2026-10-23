# Arquitetura — IntelliGate ERP

## Estilo Arquitetural
- **Monolito modular** em Go (Chi + pgx + sqlc).
- **Hexagonal** (ports & adapters) dentro do monolito.
- **PostgreSQL** como única fonte da verdade (single source of truth).
- **Next.js (TypeScript)** apenas como painel web (BFF opcional).

## Camadas (Backend)
```
cmd/                    # entrypoints (api, worker, migrate)
internal/
  domain/               # entidades, value objects, interfaces de repositório
  usecase/              # casos de uso (application services)
  adapter/
    http/               # handlers Chi, middlewares, DTOs
    db/                 # implementações sqlc, repositórios
    external/           # clientes HTTP (Gula Menu, etc.)
  config/               # viper/env
pkg/                    # código compartilhado reutilizável
migrations/             # Goose versionado
```

## Comunicação entre Módulos
- **Chamadas diretas** (in-process) entre usecases do mesmo domínio.
- **Eventos de domínio** para desacoplamento entre domínios distintos.
- **Outbox pattern** — tabela `outbox_events` transacional com a entidade de origem; worker publica no message broker (futuro) ou processa localmente.

## Integrações Externas (ex.: Gula Menu)
- **Adapter `external`** implementa cliente HTTP com retry, circuit breaker, idempotency key.
- **Ingestão** → normalização → modelo canônico → persistência → evento de domínio `OrderIngested`.
- **Webhooks** → validação de assinatura → idempotência → enfileiramento.

## Agentes de IA
- **Somente leitura** nas APIs internas (GET).
- **Propõem ações** via `Proposal` (entidade com status `pending_approval`).
- **Nunca alteram dados** diretamente — execução requer aprovação humana ou workflow configurado.
- Contexto via RAG sobre documentação, schema e dados anonimizados.

## Observabilidade
- Logs estruturados (JSON), correlation ID propagado.
- Métricas Prometheus (latência, erros, filas).
- Traces OpenTelemetry (opcional, fase posterior).

## Segurança
- JWT (access + refresh) com rotação.
- RBAC baseado em papel/permissão por empresa/loja.
- Rate limiting por tenant/IP.
- Secrets via Vault/env (não em código).