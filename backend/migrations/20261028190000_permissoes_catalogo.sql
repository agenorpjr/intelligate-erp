-- +goose Up
-- Catálogo de permissões padronizado (formato modulo.acao)
-- Módulos: usuarios, lojas, insumos, produtos, fichas_tecnicas, cmv, vendas, integracoes,
--          fornecedores, ordens_compra, recebimentos, contas_pagar, contas_gerenciais,
--          categorias_financeiras, movimentacoes_caixa, dre, fluxo_caixa, dashboard,
--          estoque, agente, auditoria, papeis, permissoes

-- Inserir permissões padrão do sistema
INSERT INTO permissoes (chave, descricao) VALUES
-- Usuários
('usuarios.criar', 'Criar usuários'),
('usuarios.listar', 'Listar usuários'),
('usuarios.ver', 'Visualizar detalhes de usuário'),
('usuarios.atualizar', 'Atualizar usuário'),
('usuarios.deletar', 'Desativar/Excluir usuário'),
('usuarios.gerenciar_papeis', 'Gerenciar papéis de usuários'),

-- Lojas
('lojas.criar', 'Criar lojas'),
('lojas.listar', 'Listar lojas'),
('lojas.ver', 'Visualizar detalhes de loja'),
('lojas.atualizar', 'Atualizar loja'),
('lojas.deletar', 'Desativar/Excluir loja'),

-- Insumos
('insumos.criar', 'Criar insumos'),
('insumos.listar', 'Listar insumos'),
('insumos.ver', 'Visualizar detalhes de insumo'),
('insumos.atualizar', 'Atualizar insumo'),
('insumos.deletar', 'Desativar/Excluir insumo'),
('insumos.gerenciar_custos', 'Gerenciar custos de insumos'),

-- Produtos
('produtos.criar', 'Criar produtos'),
('produtos.listar', 'Listar produtos'),
('produtos.ver', 'Visualizar detalhes de produto'),
('produtos.atualizar', 'Atualizar produto'),
('produtos.deletar', 'Desativar/Excluir produto'),

-- Fichas Técnicas
('fichas_tecnicas.criar', 'Criar fichas técnicas'),
('fichas_tecnicas.listar', 'Listar fichas técnicas'),
('fichas_tecnicas.ver', 'Visualizar detalhes de ficha técnica'),
('fichas_tecnicas.atualizar', 'Atualizar ficha técnica'),
('fichas_tecnicas.deletar', 'Desativar/Excluir ficha técnica'),

-- CMV
('cmv.ver_resumo', 'Visualizar resumo de CMV'),
('cmv.ver_detalhe', 'Visualizar detalhe de CMV por produto'),
('cmv.ver_sem_ficha', 'Visualizar produtos sem ficha técnica'),

-- Vendas
('vendas.ingest', 'Ingerir eventos de vendas'),
('vendas.ver_resumo', 'Visualizar resumo de vendas'),
('vendas.listar', 'Listar pedidos de vendas'),

-- Integrações PDV
('integracoes.criar', 'Criar integrações PDV'),
('integracoes.listar', 'Listar integrações PDV'),
('integracoes.ver', 'Visualizar detalhes de integração'),
('integracoes.atualizar', 'Atualizar integração'),
('integracoes.deletar', 'Excluir integração'),
('integracoes.sincronizar', 'Sincronizar integração'),

-- Fornecedores
('fornecedores.criar', 'Criar fornecedores'),
('fornecedores.listar', 'Listar fornecedores'),
('fornecedores.ver', 'Visualizar detalhes de fornecedor'),
('fornecedores.atualizar', 'Atualizar fornecedor'),
('fornecedores.deletar', 'Desativar/Excluir fornecedor'),

-- Ordens de Compra
('ordens_compra.criar', 'Criar ordens de compra'),
('ordens_compra.listar', 'Listar ordens de compra'),
('ordens_compra.ver', 'Visualizar detalhes de ordem de compra'),
('ordens_compra.atualizar', 'Atualizar ordem de compra'),
('ordens_compra.deletar', 'Cancelar/Excluir ordem de compra'),
('ordens_compra.gerenciar_itens', 'Gerenciar itens da ordem de compra'),

-- Recebimentos
('recebimentos.criar', 'Criar recebimentos de compra'),
('recebimentos.listar', 'Listar recebimentos'),
('recebimentos.ver', 'Visualizar detalhes de recebimento'),

-- Contas a Pagar
('contas_pagar.criar', 'Criar contas a pagar'),
('contas_pagar.listar', 'Listar contas a pagar'),
('contas_pagar.ver', 'Visualizar detalhes de conta a pagar'),
('contas_pagar.atualizar', 'Atualizar conta a pagar'),
('contas_pagar.pagar', 'Registrar pagamento de conta'),
('contas_pagar.deletar', 'Excluir conta a pagar'),

-- Contas Gerenciais
('contas_gerenciais.criar', 'Criar contas gerenciais'),
('contas_gerenciais.listar', 'Listar contas gerenciais'),
('contas_gerenciais.ver', 'Visualizar detalhes de conta gerencial'),
('contas_gerenciais.atualizar', 'Atualizar conta gerencial'),
('contas_gerenciais.deletar', 'Excluir conta gerencial'),

-- Categorias Financeiras
('categorias_financeiras.criar', 'Criar categorias financeiras'),
('categorias_financeiras.listar', 'Listar categorias financeiras'),
('categorias_financeiras.ver', 'Visualizar detalhes de categoria financeira'),
('categorias_financeiras.atualizar', 'Atualizar categoria financeira'),
('categorias_financeiras.deletar', 'Excluir categoria financeira'),

-- Movimentações de Caixa
('movimentacoes_caixa.criar', 'Criar movimentações de caixa'),
('movimentacoes_caixa.listar', 'Listar movimentações de caixa'),
('movimentacoes_caixa.ver', 'Visualizar detalhes de movimentação'),
('movimentacoes_caixa.atualizar', 'Atualizar movimentação de caixa'),
('movimentacoes_caixa.deletar', 'Excluir movimentação de caixa'),

-- DRE Gerencial
('dre.ver', 'Visualizar DRE Gerencial'),

-- Fluxo de Caixa
('fluxo_caixa.ver', 'Visualizar Fluxo de Caixa'),

-- Dashboard
('dashboard.ver', 'Visualizar Dashboard'),

-- Estoque
('estoque.ver_saldos', 'Visualizar saldos de estoque'),
('estoque.registrar_saida', 'Registrar saída de estoque'),
('estoque.registrar_ajuste', 'Registrar ajuste de estoque'),
('estoque.transferir', 'Transferir estoque entre lojas'),
('estoque.ver_movimentacoes', 'Visualizar movimentações de estoque'),
('estoque.ver_perdas', 'Visualizar resumo de perdas'),

-- Agente de IA
('agente.criar_sessao', 'Criar sessão do agente'),
('agente.listar_sessoes', 'Listar sessões do agente'),
('agente.ver_sessao', 'Visualizar sessão do agente'),
('agente.perguntar', 'Fazer perguntas ao agente'),

-- Auditoria
('auditoria.listar', 'Listar logs de auditoria'),

-- Papéis
('papeis.criar', 'Criar papéis'),
('papeis.listar', 'Listar papéis'),
('papeis.ver', 'Visualizar detalhes de papel'),
('papeis.atualizar', 'Atualizar papel'),
('papeis.deletar', 'Excluir papel'),
('papeis.gerenciar_permissoes', 'Gerenciar permissões do papel'),

-- Permissões
('permissoes.listar', 'Listar permissões'),
('permissoes.ver', 'Visualizar detalhes de permissão');

-- +goose Down
DELETE FROM permissoes WHERE chave IN (
    'usuarios.criar', 'usuarios.listar', 'usuarios.ver', 'usuarios.atualizar', 'usuarios.deletar', 'usuarios.gerenciar_papeis',
    'lojas.criar', 'lojas.listar', 'lojas.ver', 'lojas.atualizar', 'lojas.deletar',
    'insumos.criar', 'insumos.listar', 'insumos.ver', 'insumos.atualizar', 'insumos.deletar', 'insumos.gerenciar_custos',
    'produtos.criar', 'produtos.listar', 'produtos.ver', 'produtos.atualizar', 'produtos.deletar',
    'fichas_tecnicas.criar', 'fichas_tecnicas.listar', 'fichas_tecnicas.ver', 'fichas_tecnicas.atualizar', 'fichas_tecnicas.deletar',
    'cmv.ver_resumo', 'cmv.ver_detalhe', 'cmv.ver_sem_ficha',
    'vendas.ingest', 'vendas.ver_resumo', 'vendas.listar',
    'integracoes.criar', 'integracoes.listar', 'integracoes.ver', 'integracoes.atualizar', 'integracoes.deletar', 'integracoes.sincronizar',
    'fornecedores.criar', 'fornecedores.listar', 'fornecedores.ver', 'fornecedores.atualizar', 'fornecedores.deletar',
    'ordens_compra.criar', 'ordens_compra.listar', 'ordens_compra.ver', 'ordens_compra.atualizar', 'ordens_compra.deletar', 'ordens_compra.gerenciar_itens',
    'recebimentos.criar', 'recebimentos.listar', 'recebimentos.ver',
    'contas_pagar.criar', 'contas_pagar.listar', 'contas_pagar.ver', 'contas_pagar.atualizar', 'contas_pagar.pagar', 'contas_pagar.deletar',
    'contas_gerenciais.criar', 'contas_gerenciais.listar', 'contas_gerenciais.ver', 'contas_gerenciais.atualizar', 'contas_gerenciais.deletar',
    'categorias_financeiras.criar', 'categorias_financeiras.listar', 'categorias_financeiras.ver', 'categorias_financeiras.atualizar', 'categorias_financeiras.deletar',
    'movimentacoes_caixa.criar', 'movimentacoes_caixa.listar', 'movimentacoes_caixa.ver', 'movimentacoes_caixa.atualizar', 'movimentacoes_caixa.deletar',
    'dre.ver', 'fluxo_caixa.ver', 'dashboard.ver',
    'estoque.ver_saldos', 'estoque.registrar_saida', 'estoque.registrar_ajuste', 'estoque.transferir', 'estoque.ver_movimentacoes', 'estoque.ver_perdas',
    'agente.criar_sessao', 'agente.listar_sessoes', 'agente.ver_sessao', 'agente.perguntar',
    'auditoria.listar',
    'papeis.criar', 'papeis.listar', 'papeis.ver', 'papeis.atualizar', 'papeis.deletar', 'papeis.gerenciar_permissoes',
    'permissoes.listar', 'permissoes.ver'
);