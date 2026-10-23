# Visão do Produto — IntelliGate ERP

## Propósito
ERP food service multi-loja para centralizar a gestão de vendas, ficha técnica, estoque, compras, financeiro gerencial, contabilidade gerencial e agentes de IA.

## Escopo Inicial
- **Integração com PDVs externos** — começando pelo **Gula Menu**.
- **Não inclui PDV próprio** nesta fase.
- Foco em **back-office** e inteligência de dados.

## Domínios Funcionais
| Domínio | Descrição |
|---------|-----------|
| **Vendas** | Ingestão e normalização de pedidos de PDVs externos (modelo canônico). |
| **Ficha Técnica** | Composição de produtos, insumos, unidades, rendimento e CMV. |
| **Estoque** | Movimentação, contagem, transferência entre lojas, alertas de ruptura. |
| **Compras** | Solicitação, cotação, pedido de compra, recebimento e conferência. |
| **Financeiro Gerencial** | Contas a pagar/receber, fluxo de caixa, conciliação bancária, DRE gerencial. |
| **Contabilidade Gerencial** | Plano de contas, centros de custo, lançamentos, apuração de resultados. |
| **Agentes de IA** | Consultam APIs, analisam dados, propõem ações (não executam sem aprovação). |

## Princípios
1. **Fonte única da verdade** — PostgreSQL.
2. **Integração por eventos** — outbox pattern para confiabilidade.
3. **Idempotência** — toda ingestão externa tem chave de deduplicação.
4. **Auditoria total** — created_at, updated_at, deleted_at, histórico de aprovações.
5. **Multi-tenancy nativo** — empresa → lojas → usuários/papéis.