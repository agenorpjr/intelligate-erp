package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permissao struct {
	ID        uuid.UUID `json:"id"`
	Chave     string    `json:"chave"`
	Descricao *string   `json:"descricao,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePermissaoInput struct {
	Chave     string  `json:"chave" validate:"required"`
	Descricao *string `json:"descricao,omitempty"`
}

// Constantes de permissões padronizadas no formato modulo.acao
const (
	// Usuários
	PermUsuariosCriar           = "usuarios.criar"
	PermUsuariosListar          = "usuarios.listar"
	PermUsuariosVer             = "usuarios.ver"
	PermUsuariosAtualizar       = "usuarios.atualizar"
	PermUsuariosDeletar         = "usuarios.deletar"
	PermUsuariosGerenciarPapeis = "usuarios.gerenciar_papeis"

	// Lojas
	PermLojasCriar     = "lojas.criar"
	PermLojasListar    = "lojas.listar"
	PermLojasVer       = "lojas.ver"
	PermLojasAtualizar = "lojas.atualizar"
	PermLojasDeletar   = "lojas.deletar"

	// Insumos
	PermInsumosCriar           = "insumos.criar"
	PermInsumosListar          = "insumos.listar"
	PermInsumosVer             = "insumos.ver"
	PermInsumosAtualizar       = "insumos.atualizar"
	PermInsumosDeletar         = "insumos.deletar"
	PermInsumosGerenciarCustos = "insumos.gerenciar_custos"

	// Produtos
	PermProdutosCriar     = "produtos.criar"
	PermProdutosListar    = "produtos.listar"
	PermProdutosVer       = "produtos.ver"
	PermProdutosAtualizar = "produtos.atualizar"
	PermProdutosDeletar   = "produtos.deletar"

	// Fichas Técnicas
	PermFichasTecnicasCriar     = "fichas_tecnicas.criar"
	PermFichasTecnicasListar    = "fichas_tecnicas.listar"
	PermFichasTecnicasVer       = "fichas_tecnicas.ver"
	PermFichasTecnicasAtualizar = "fichas_tecnicas.atualizar"
	PermFichasTecnicasDeletar   = "fichas_tecnicas.deletar"

	// CMV
	PermCMVVerResumo   = "cmv.ver_resumo"
	PermCMVVerDetalhe  = "cmv.ver_detalhe"
	PermCMVVerSemFicha = "cmv.ver_sem_ficha"

	// Vendas
	PermVendasIngest    = "vendas.ingest"
	PermVendasVerResumo = "vendas.ver_resumo"
	PermVendasListar    = "vendas.listar"

	// Integrações PDV
	PermIntegracoesCriar       = "integracoes.criar"
	PermIntegracoesListar      = "integracoes.listar"
	PermIntegracoesVer         = "integracoes.ver"
	PermIntegracoesAtualizar   = "integracoes.atualizar"
	PermIntegracoesDeletar     = "integracoes.deletar"
	PermIntegracoesSincronizar = "integracoes.sincronizar"

	// Fornecedores
	PermFornecedoresCriar     = "fornecedores.criar"
	PermFornecedoresListar    = "fornecedores.listar"
	PermFornecedoresVer       = "fornecedores.ver"
	PermFornecedoresAtualizar = "fornecedores.atualizar"
	PermFornecedoresDeletar   = "fornecedores.deletar"

	// Ordens de Compra
	PermOrdensCompraCriar          = "ordens_compra.criar"
	PermOrdensCompraListar         = "ordens_compra.listar"
	PermOrdensCompraVer            = "ordens_compra.ver"
	PermOrdensCompraAtualizar      = "ordens_compra.atualizar"
	PermOrdensCompraDeletar        = "ordens_compra.deletar"
	PermOrdensCompraGerenciarItens = "ordens_compra.gerenciar_itens"

	// Recebimentos
	PermRecebimentosCriar  = "recebimentos.criar"
	PermRecebimentosListar = "recebimentos.listar"
	PermRecebimentosVer    = "recebimentos.ver"

	// Contas a Pagar
	PermContasPagarCriar     = "contas_pagar.criar"
	PermContasPagarListar    = "contas_pagar.listar"
	PermContasPagarVer       = "contas_pagar.ver"
	PermContasPagarAtualizar = "contas_pagar.atualizar"
	PermContasPagarPagar     = "contas_pagar.pagar"
	PermContasPagarDeletar   = "contas_pagar.deletar"

	// Contas Gerenciais
	PermContasGerenciaisCriar     = "contas_gerenciais.criar"
	PermContasGerenciaisListar    = "contas_gerenciais.listar"
	PermContasGerenciaisVer       = "contas_gerenciais.ver"
	PermContasGerenciaisAtualizar = "contas_gerenciais.atualizar"
	PermContasGerenciaisDeletar   = "contas_gerenciais.deletar"

	// Categorias Financeiras
	PermCategoriasFinanceirasCriar     = "categorias_financeiras.criar"
	PermCategoriasFinanceirasListar    = "categorias_financeiras.listar"
	PermCategoriasFinanceirasVer       = "categorias_financeiras.ver"
	PermCategoriasFinanceirasAtualizar = "categorias_financeiras.atualizar"
	PermCategoriasFinanceirasDeletar   = "categorias_financeiras.deletar"

	// Movimentações de Caixa
	PermMovimentacoesCaixaCriar     = "movimentacoes_caixa.criar"
	PermMovimentacoesCaixaListar    = "movimentacoes_caixa.listar"
	PermMovimentacoesCaixaVer       = "movimentacoes_caixa.ver"
	PermMovimentacoesCaixaAtualizar = "movimentacoes_caixa.atualizar"
	PermMovimentacoesCaixaDeletar   = "movimentacoes_caixa.deletar"

	// DRE Gerencial
	PermDREVer = "dre.ver"

	// Fluxo de Caixa
	PermFluxoCaixaVer = "fluxo_caixa.ver"

	// Dashboard
	PermDashboardVer = "dashboard.ver"

	// Estoque
	PermEstoqueVerSaldos        = "estoque.ver_saldos"
	PermEstoqueRegistrarSaida   = "estoque.registrar_saida"
	PermEstoqueRegistrarAjuste  = "estoque.registrar_ajuste"
	PermEstoqueTransferir       = "estoque.transferir"
	PermEstoqueVerMovimentacoes = "estoque.ver_movimentacoes"
	PermEstoqueVerPerdas        = "estoque.ver_perdas"

	// Agente de IA
	PermAgenteCriarSessao   = "agente.criar_sessao"
	PermAgenteListarSessoes = "agente.listar_sessoes"
	PermAgenteVerSessao     = "agente.ver_sessao"
	PermAgentePerguntar     = "agente.perguntar"

	// Auditoria
	PermAuditoriaListar = "auditoria.listar"

	// Papéis
	PermPapeisCriar               = "papeis.criar"
	PermPapeisListar              = "papeis.listar"
	PermPapeisVer                 = "papeis.ver"
	PermPapeisAtualizar           = "papeis.atualizar"
	PermPapeisDeletar             = "papeis.deletar"
	PermPapeisGerenciarPermissoes = "papeis.gerenciar_permissoes"

	// Permissões
	PermPermissoesListar = "permissoes.listar"
	PermPermissoesVer    = "permissoes.ver"
)

// AllPermissions retorna todas as permissões do sistema
func AllPermissions() []string {
	return []string{
		PermUsuariosCriar, PermUsuariosListar, PermUsuariosVer, PermUsuariosAtualizar, PermUsuariosDeletar, PermUsuariosGerenciarPapeis,
		PermLojasCriar, PermLojasListar, PermLojasVer, PermLojasAtualizar, PermLojasDeletar,
		PermInsumosCriar, PermInsumosListar, PermInsumosVer, PermInsumosAtualizar, PermInsumosDeletar, PermInsumosGerenciarCustos,
		PermProdutosCriar, PermProdutosListar, PermProdutosVer, PermProdutosAtualizar, PermProdutosDeletar,
		PermFichasTecnicasCriar, PermFichasTecnicasListar, PermFichasTecnicasVer, PermFichasTecnicasAtualizar, PermFichasTecnicasDeletar,
		PermCMVVerResumo, PermCMVVerDetalhe, PermCMVVerSemFicha,
		PermVendasIngest, PermVendasVerResumo, PermVendasListar,
		PermIntegracoesCriar, PermIntegracoesListar, PermIntegracoesVer, PermIntegracoesAtualizar, PermIntegracoesDeletar, PermIntegracoesSincronizar,
		PermFornecedoresCriar, PermFornecedoresListar, PermFornecedoresVer, PermFornecedoresAtualizar, PermFornecedoresDeletar,
		PermOrdensCompraCriar, PermOrdensCompraListar, PermOrdensCompraVer, PermOrdensCompraAtualizar, PermOrdensCompraDeletar, PermOrdensCompraGerenciarItens,
		PermRecebimentosCriar, PermRecebimentosListar, PermRecebimentosVer,
		PermContasPagarCriar, PermContasPagarListar, PermContasPagarVer, PermContasPagarAtualizar, PermContasPagarPagar, PermContasPagarDeletar,
		PermContasGerenciaisCriar, PermContasGerenciaisListar, PermContasGerenciaisVer, PermContasGerenciaisAtualizar, PermContasGerenciaisDeletar,
		PermCategoriasFinanceirasCriar, PermCategoriasFinanceirasListar, PermCategoriasFinanceirasVer, PermCategoriasFinanceirasAtualizar, PermCategoriasFinanceirasDeletar,
		PermMovimentacoesCaixaCriar, PermMovimentacoesCaixaListar, PermMovimentacoesCaixaVer, PermMovimentacoesCaixaAtualizar, PermMovimentacoesCaixaDeletar,
		PermDREVer, PermFluxoCaixaVer, PermDashboardVer,
		PermEstoqueVerSaldos, PermEstoqueRegistrarSaida, PermEstoqueRegistrarAjuste, PermEstoqueTransferir, PermEstoqueVerMovimentacoes, PermEstoqueVerPerdas,
		PermAgenteCriarSessao, PermAgenteListarSessoes, PermAgenteVerSessao, PermAgentePerguntar,
		PermAuditoriaListar,
		PermPapeisCriar, PermPapeisListar, PermPapeisVer, PermPapeisAtualizar, PermPapeisDeletar, PermPapeisGerenciarPermissoes,
		PermPermissoesListar, PermPermissoesVer,
	}
}
