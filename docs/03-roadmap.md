# Roadmap — IntelliGate ERP

## Fase 1 — Fundação Técnica (Semanas 1–3)
**Objetivo**: Base sólida para desenvolvimento seguro e rápido.

| Entregável | Detalhes |
|------------|----------|
| Projeto Go (Chi, pgx, sqlc, Goose) | Estrutura de pastas, config, healthcheck, graceful shutdown. |
| PostgreSQL + Docker Compose | Migrações iniciais (empresa, loja, usuário, papel, permissão). |
| Autenticação & Autorização | JWT access/refresh, RBAC, middleware Chi, testes. |
| Auditoria base | Tabela `auditoria_log`, trigger genérico, soft delete. |
| CI/CD local | `gofmt`, `go vet`, `golangci-lint`, `go test -race`, `goose up`. |
| Next.js painel | Scaffold, login, layout, tenant switcher, storybook. |
| Documentação OpenAPI | `swag` ou `oapi-codegen` integrado ao build. |

**Critério de aceite**: `go build ./...`, lint/test passam, migrações aplicam/rollback, login funciona no painel.

---

## Fase 2 — Modelo Canônico de Venda (Semanas 4–6)
**Objetivo**: Estrutura unificada para receber pedidos de qualquer PDV.

| Entregável | Detalhes |
|------------|----------|
| Entidades canônicas | `produto_canonico`, `pedido_canonico`, `item_pedido_canonico`, `pagamento_pedido`, `evento_ingestao`. |
| Normalizador genérico | Interface `Normalizer` + implementação stub; mapeamento campo-a-campo configurável por PDV. |
| API de ingestão | `POST /api/v1/ingestion/events` (idempotent, validação, resposta RFC 7807). |
| Outbox + worker | Tabela `outbox_events`, publisher worker, retry/backoff, DLQ. |
| Testes de contrato | Payloads Gula reais (sample) → modelo canônico → snapshot testing. |

**Critério de aceite**: Ingestão de 10k pedidos simulados < 2 min, 0 duplicatas, 0 perda, auditoria completa.

---

## Fase 3 — Conector Gula Menu (Semanas 7–9)
**Objetivo**: Integração real, homologada e monitorada com o Gula.

| Entregável | Detalhes |
|------------|----------|
| Cliente Gula (`adapter/external/gula`) | Auth (OAuth2/API Key), rate limit, circuit breaker, paginação, webhooks. |
| Normalizador Gula | Mapeamento completo: categorias, produtos, modificadores, pagamentos, cupons, clientes. |
| Sincronismo inicial | Backfill histórico (últimos 90 dias) com idempotência; relatório de divergências. |
| Webhooks Gula | `order.created`, `order.updated`, `order.cancelled`, `product.updated` — validação HMAC. |
| Painel de monitoramento | Dashboard de ingestão (sucesso/erro/latência), replay de eventos, alertas. |
| Documentação de onboarding | Guia para nova loja: credenciais, mapeamentos, validação, go-live. |

**Critério de aceite**: 3 lojas piloto em produção, 99.9% ingestão ok, SLA < 30s pedido→canônico.

---

## Fase 4 — Ficha Técnica & CMV (Semanas 10–13)
**Objetivo**: Custo real por produto vendido.

| Entregável | Detalhes |
|------------|----------|
| CRUD Insumo / Unidade / Conversão | Unidade base, fatores de conversão (kg↔g, cx↔un), validação circular. |
| Ficha Técnica versionada | Header + itens (insumo, qtd, unidade, fator correção), rendimento, modo preparo. |
| Cálculo de CMV | Custo médio ponderado por insumo × ficha × qtd vendida; snapshot por período/loja. |
| API de custo | `GET /api/v1/products/{id}/cost` (detalhado por insumo), `GET /api/v1/reports/cmv`. |
| Alerta de divergência | CMV teórico (ficha) vs. CMV real (estoque) > threshold → notificação. |
| Testes de precisão | Casos conhecidos (receitas padrão) → resultado centavo a centavo. |

**Critério de aceite**: CMV de 100 produtos conferido com planilha contábil (diferença < 0,5%).

---

## Fase 5 — Compras (Semanas 14–17)
**Objetivo**: Suprimento baseado em necessidade real (estoque + produção).

| Entregável | Detalhes |
|------------|----------|
| Fornecedor & Condições | Cadastro, lead time, mínimo pedido, tabela de preços vigência. |
| Solicitação de Compra | Origem: ponto de reposição, previsão produção, manual; workflow aprovação. |
| Cotação & Pedido de Compra | Múltiplos fornecedores, comparação, emissão PDF/e-mail, rastreamento. |
| Recebimento & Conferência | NF-e importação (XML), qtd/qualidade/preço vs. pedido, divergências → devolução/ajuste. |
| Integração Estoque | Recebimento aprovado → movimento entrada + atualização custo médio. |
| KPIs | Lead time real, fill rate, custo unitário médio por fornecedor. |

**Critério de aceite**: Ciclo completo (solicitação→recebimento) em < 48h para 80% dos itens; 0 entrada sem NF.

---

## Fase 6 — DRE & Financeiro Gerencial (Semanas 18–22)
**Objetivo**: Visão gerencial de resultado e caixa por loja/empresa.

| Entregável | Detalhes |
|------------|----------|
| Centro de Custo & Plano de Contas | Hierarquia, importação contábil, mapeamento automático (regras). |
| Lançamentos (Partida Dobrada) | API + importação OFX/CSV; conciliação bancária assistida por regras. |
| Contas a Pagar/Receber | Vencimentos, parcelas, juros/multa, baixa automática (PIX/boleto). |
| DRE Gerencial | Receita bruta, deduções, CMV, despesas por centro/categoria, EBITDA, líquido. |
| Fluxo de Caixa | Realizado + projetado (compras a pagar, vendas a receber, folha, impostos). |
| Relatórios & Export | PDF/Excel agendados, drill-down até lançamento, comparativo orçado x realizado. |

**Critério de aceite**: DRE mensal fechado até D+5; conciliação bancária > 95% automática.

---

## Fase 7 — Contabilidade Gerencial (Semanas 23–26)
**Objetivo**: Entregar dados prontos para contabilidade fiscal/gerencial.

| Entregável | Detalhes |
|------------|----------|
| Plano de Contas Contábil | Mapeamento 1:N (gerencial → contábil), contas de resultado/balanço. |
| Livro Diário & Razão | Geração automática a partir de lançamentos; validação saldo zero. |
| Balancete & Balanço | Por período, loja, consolidado; export SPED Contábil (layout ECD). |
| Apuração de Impostos | Simples Nacional / Lucro Presumido — base cálculo, DAS, DCTF. |
| Integração Contador | Portal/export para contador (XML, PDF, CSV), log de envio. |
| Auditoria Contábil | Rastreabilidade lançamento → origem (venda, compra, ajuste, rateio). |

**Critério de aceite**: Contador valida balancete sem ajustes manuais; SPED passa no validador oficial.

---

## Fase 8 — Agentes de IA (Semanas 27–32)
**Objetivo**: Inteligência acionável, não autônoma.

| Entregável | Detalhes |
|------------|----------|
| Plataforma de Agentes | Framework (LangChain/Go), RAG sobre docs + schema + dados anonimizados, sandbox. |
| Agente "Compras Inteligentes" | Sugere pedidos base: curva ABC, sazonalidade, lead time, promoções fornecedores. |
| Agente "Análise de CMV" | Detecta insumos com desvio > X%, sugere ficha técnica ou troca fornecedor. |
| Agente "DRE Insights" | Explica variações (mix, preço, volume), alerta margem negativa por categoria. |
| Agente "Conciliação" | Propõe matches bancários não óbvios, aprende com aprovações. |
| Workflow de Aprovação | `PropostaIA` → notificação → aprovação/rejeição com justificativa → execução (se aprovado). |
| Observabilidade IA | Latência, tokens, acurácia (feedback loop), custo por agente. |
| Governança | Políticas: dados sensíveis, ações permitidas, auditoria total de decisões. |

**Critério de aceite**: 3 agentes em produção com > 80% propostas aprovadas sem edição; custo < R$ 500/mês.

---

## Princípios Transversais (Todas as Fases)
1. **Testes primeiro** — unitários (usecase) + integração (Testcontainers) ≥ 80% cobertura.
2. **Migrações versionadas** — Goose, sem `DROP`/`TRUNCATE` destrutivo.
3. **Idempotência** — toda API mutante aceita `Idempotency-Key`.
4. **Observabilidade** — logs, métricas, traces desde a Fase 1.
5. **Documentação viva** — OpenAPI, ADRs, runbooks atualizados a cada fase.
6. **Segurança** — pentest anual, dependabot, secrets rotation.