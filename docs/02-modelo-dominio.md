# Modelo de Domínio — IntelliGate ERP

## Identidade & Acesso
| Entidade | Descrição |
|----------|-----------|
| **Empresa** | Tenant raiz. CNPJ, razão social, nome fantasia, configurações globais. |
| **Loja** | Unidade operacional da empresa. Endereço, fuso, configurações de integração (PDV). |
| **Usuário** | Pessoa física com login. Vinculado a uma empresa. |
| **Papel** | Conjunto de permissões (ex.: `gerente`, `comprador`, `contador`). |
| **Permissão** | Ação granular (ex.: `purchase_order:create`, `dre:view`). |
| **UserStoreRole** | Vincula usuário × loja × papel (multi-loja, multi-papel). |

## Vendas (Modelo Canônico)
| Entidade | Descrição |
|----------|-----------|
| **PdvOrigem** | Cadastro do PDV externo (Gula, etc.). Chave API, webhook secret, mapeamentos. |
| **ProdutoCanonico** | Produto normalizado da empresa. SKU interno, nome, categoria, NCM, CEST, tributação. |
| **PedidoCanonico** | Pedido normalizado (independente do PDV). Loja, cliente, status, total, timestamps. |
| **ItemPedidoCanonico** | Linha do pedido. Produto canônico, qtd, preço unitário, descontos, impostos. |
| **PagamentoPedido** | Forma, valor, status, transaction_id do PDV, conciliação. |
| **EventoIngestao** | Log imutável de cada payload recebido do PDV (raw, status, tentativas, erro). |

## Ficha Técnica & Estoque
| Entidade | Descrição |
|----------|-----------|
| **Insumo** | Matéria-prima ou item de revenda. Nome, unidade base, NCM, custo médio/último. |
| **Unidade** | Unidade de medida (kg, g, L, ml, un, cx). Conversões entre unidades. |
| **FichaTecnica** | Receita de um produto canônico. Rendimento, modo de preparo, versão. |
| **FichaTecnicaItem** | Insumo × qtd × unidade na ficha. Fator de correção (perda, refugo). |
| **Estoque** | Saldo por loja × insumo/produto. Qtd, custo médio, lote, validade. |
| **MovimentoEstoque** | Entrada/saída/transferência/ajuste. Origem (compra, venda, produção, contagem). |
| **ContagemEstoque** | Inventário físico. Status, divergências, aprovação. |

## Compras
| Entidade | Descrição |
|----------|-----------|
| **Fornecedor** | Cadastro. CNPJ, contatos, condições comerciais, lead time. |
| **SolicitacaoCompra** | Necessidade detectada (estoque mínimo, produção). Itens, loja, solicitante. |
| **CotacaoCompra** | Solicitação × fornecedores. Itens, prazos, condições. |
| **PedidoCompra** | Emitido ao fornecedor. Itens, preços, entrega, status. |
| **RecebimentoCompra** | Conferência física × nota fiscal. Qtd recebida, qualidade, divergências. |

## Financeiro & Contabilidade Gerencial
| Entidade | Descrição |
|----------|-----------|
| **CentroCusto** | Hierárquico. Loja, setor, projeto. Rateio de despesas/receitas. |
| **PlanoContas** | Estrutura contábil (ativo, passivo, receita, despesa). Código, nível, tipo. |
| **Lancamento** | Partida dobrada. Data, valor, débito/crédito, centro de custo, conta, histórico. |
| **DreGerencial** | Apuração por período/loja. Receitas, CMV, despesas, resultado. |
| **ContaBancaria** | Banco, agência, conta, conciliação automática (OFX/PIX). |
| **ConciliaçãoBancaria** | Extrato × lançamentos. Status, regras de match. |

## Auditoria & Aprovação
| Entidade | Descrição |
|----------|-----------|
| **AuditoriaLog** | Quem, o quê, quando, antes/depois (JSONB). Tabela única, particionada por tempo. |
| **Aprovacao** | Workflow genérico. Entidade alvo, solicitante, aprovadores, status, SLA, decisão. |
| **PropostaIA** | Ação sugerida por agente. Contexto, justificativa, risco, status (`pending`, `approved`, `rejected`). |

## Convenções Transversais
- **Todas as entidades**: `id` (UUID), `empresa_id`, `created_at`, `updated_at`, `deleted_at` (soft delete), `versao` (optimistic lock).
- **Chaves naturais** únicas onde aplicável (ex.: `sku` por empresa, `cnpj` fornecedor).
- **Enumerações** como `text` com `CHECK constraint` (não `enum` nativo do PG).
- **JSONB** para payloads flexíveis (eventos, metadados, configurações).