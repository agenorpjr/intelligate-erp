package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/intelligate/erp/backend/internal/adapter/ai"
	"github.com/intelligate/erp/backend/internal/adapter/db"
	httpadapter "github.com/intelligate/erp/backend/internal/adapter/http"
	"github.com/intelligate/erp/backend/internal/config"
	"github.com/intelligate/erp/backend/internal/domain"
	"github.com/intelligate/erp/backend/internal/usecase"
	"github.com/intelligate/erp/backend/pkg/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database connection pool
	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database connection pool")
	}
	defer dbpool.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Database connection established")

	// Initialize repositories
	empresaRepo := db.NewEmpresaRepository(dbpool)
	lojaRepo := db.NewLojaRepository(dbpool)
	usuarioRepo := db.NewUsuarioRepository(dbpool)
	usuarioPapelRepo := db.NewUsuarioPapelRepository(dbpool)
	papelRepo := db.NewPapelRepository(dbpool)
	papelPermissaoRepo := db.NewPapelPermissaoRepository(dbpool)
	permissaoRepo := db.NewPermissaoRepository(dbpool)
	auditoriaRepo := db.NewAuditoriaRepository(dbpool)
	senhaRedefinicaoRepo := db.NewSenhaRedefinicaoRepository(dbpool)
	sessaoRevogadaRepo := db.NewSessaoRevogadaRepository(dbpool)
	pdvOrigemRepo := db.NewPdvOrigemRepository(dbpool)
	pdvIntegracaoRepo := db.NewPdvIntegracaoRepository(dbpool)
	produtoCanonicoRepo := db.NewProdutoCanonicoRepository(dbpool)
	pedidoCanonicoRepo := db.NewPedidoCanonicoRepository(dbpool)
	pedidoItemCanonicoRepo := db.NewPedidoItemCanonicoRepository(dbpool)
	pedidoPagamentoCanonicoRepo := db.NewPedidoPagamentoCanonicoRepository(dbpool)
	eventoIngestaoRepo := db.NewEventoIngestaoRepository(dbpool)
	vendasResumoRepo := db.NewVendasResumoRepository(dbpool)
	salesTxRepo := db.NewSalesTxRepository(dbpool)
	insumoRepo := db.NewInsumoRepository(dbpool)
	produtoRepo := db.NewProdutoRepository(dbpool)
	fichaTecnicaRepo := db.NewFichaTecnicaRepository(dbpool)

	// Compras repositories
	fornecedorRepo := db.NewFornecedorRepository(dbpool)
	ordemCompraRepo := db.NewOrdemCompraRepository(dbpool)
	recebimentoRepo := db.NewRecebimentoCompraRepository(dbpool)
	contaPagarRepo := db.NewContaPagarRepository(dbpool)

	// Financeiro (Etapa 10) repositories
	contaGerencialRepo := db.NewContaGerencialRepository(dbpool)
	categoriaFinanceiraRepo := db.NewCategoriaFinanceiraRepository(dbpool)
	movimentacaoCaixaRepo := db.NewMovimentacaoCaixaRepository(dbpool)

	// Dashboard (Etapa 11) repository
	dashboardRepo := db.NewDashboardRepository(dbpool)

	// Estoque (Etapa 12) repository
	estoqueRepo := db.NewEstoqueRepository(dbpool)

	// Agente de IA (Etapa 13) repository
	agenteRepo := db.NewAgenteRepository(dbpool)

	// Initialize crypto service
	cryptoService, err := crypto.NewCryptoServiceFromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize crypto service: INTELLIGATE_ENCRYPTION_KEY must be set and valid")
	}

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(usuarioRepo, usuarioPapelRepo, papelRepo, papelPermissaoRepo, cfg.JWTSecret, cfg.JWTExpiry)
	senhaSessaoUsecase := usecase.NewSenhaSessaoUsecase(
		usuarioRepo,
		senhaRedefinicaoRepo,
		sessaoRevogadaRepo,
		auditoriaRepo,
		cfg.AppEnv,
		cfg.AppEnv == "development",
	)
	lojaUsecase := usecase.NewLojaUsecase(lojaRepo, empresaRepo, auditoriaRepo)
	salesUsecase := usecase.NewSalesUsecase(
		pdvOrigemRepo,
		produtoCanonicoRepo,
		pedidoCanonicoRepo,
		pedidoItemCanonicoRepo,
		pedidoPagamentoCanonicoRepo,
		eventoIngestaoRepo,
		lojaRepo,
		vendasResumoRepo,
		salesTxRepo,
	)
	pdvIntegracaoUsecase := usecase.NewPdvIntegracaoUsecase(
		pdvIntegracaoRepo,
		lojaRepo,
		pdvOrigemRepo,
		salesUsecase,
		auditoriaRepo,
		cryptoService,
	)
	insumoUsecase := usecase.NewInsumoUsecase(insumoRepo, auditoriaRepo)
	produtoUsecase := usecase.NewProdutoUsecase(produtoRepo, auditoriaRepo)
	fichaTecnicaUsecase := usecase.NewFichaTecnicaUsecase(
		fichaTecnicaRepo,
		produtoRepo,
		lojaRepo,
		insumoRepo,
		auditoriaRepo,
	)

	// Estoque (Etapa 12) usecase
	estoqueUsecase := usecase.NewEstoqueUsecase(estoqueRepo, lojaRepo, insumoRepo, auditoriaRepo)

	// Compras usecases
	fornecedorUsecase := usecase.NewFornecedorUsecase(fornecedorRepo, auditoriaRepo)
	ordemCompraUsecase := usecase.NewOrdemCompraUsecase(ordemCompraRepo, fornecedorRepo, insumoRepo, auditoriaRepo)
	recebimentoCompraUsecase := usecase.NewRecebimentoCompraUsecase(recebimentoRepo, ordemCompraRepo, fornecedorRepo, insumoRepo, contaPagarRepo, auditoriaRepo, estoqueUsecase)
	contaPagarUsecase := usecase.NewContaPagarUsecase(contaPagarRepo, auditoriaRepo)

	// Financeiro (Etapa 10) usecases
	contaGerencialUsecase := usecase.NewContaGerencialUsecase(contaGerencialRepo, auditoriaRepo)
	categoriaFinanceiraUsecase := usecase.NewCategoriaFinanceiraUsecase(categoriaFinanceiraRepo, contaGerencialRepo, auditoriaRepo)
	movimentacaoCaixaUsecase := usecase.NewMovimentacaoCaixaUsecase(movimentacaoCaixaRepo, lojaRepo, categoriaFinanceiraRepo, contaGerencialRepo, auditoriaRepo)
	dreUsecase := usecase.NewDREGerencialUsecase(movimentacaoCaixaRepo, contaGerencialRepo)
	fluxoCaixaUsecase := usecase.NewFluxoCaixaUsecase(movimentacaoCaixaRepo, contaGerencialRepo, lojaRepo)

	// Dashboard (Etapa 11) usecase
	dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo, lojaRepo)

	// Agente de IA (Etapa 13) - cliente LLM e usecase
	llmClient, err := ai.NewLLMClient(ai.Config{
		Provider: cfg.AIProvider,
		APIKey:   cfg.AIAPIKey,
		Model:    cfg.AIModel,
		BaseURL:  cfg.AIBaseURL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize AI provider")
	}
	log.Info().Str("provider", llmClient.Provider()).Msg("AI provider initialized")

	agenteUsecase := usecase.NewAgenteUsecase(
		agenteRepo,
		dashboardRepo,
		fichaTecnicaRepo,
		estoqueRepo,
		lojaRepo,
		auditoriaRepo,
		llmClient,
	)

	// Initialize usecases
	papelPermissaoUsecase := usecase.NewPapelPermissaoUsecase(
		papelRepo,
		permissaoRepo,
		usuarioRepo,
		usuarioPapelRepo,
		papelPermissaoRepo,
		auditoriaRepo,
	)

	// Governança de acesso (Etapa 14) - leitura de auditoria
	auditoriaUsecase := usecase.NewAuditoriaUsecase(auditoriaRepo)

	// Usuários (Etapa 15)
	usuarioUsecase := usecase.NewUsuarioUsecase(
		usuarioRepo,
		usuarioPapelRepo,
		papelRepo,
		papelPermissaoRepo,
		lojaRepo,
		auditoriaRepo,
	)

	// Initialize handlers
	authHandler := httpadapter.NewAuthHandler(authUsecase, senhaSessaoUsecase)
	lojaHandler := httpadapter.NewLojaHandler(lojaUsecase)
	usuarioHandler := httpadapter.NewUsuarioHandler(usuarioUsecase)
	salesHandler := httpadapter.NewSalesHandler(salesUsecase)
	insumoHandler := httpadapter.NewInsumoHandler(insumoUsecase)
	produtoHandler := httpadapter.NewProdutoHandler(produtoUsecase)
	fichaTecnicaHandler := httpadapter.NewFichaTecnicaHandler(fichaTecnicaUsecase)
	cmvHandler := httpadapter.NewCMVHandler(fichaTecnicaUsecase)
	pdvIntegracaoHandler := httpadapter.NewPdvIntegracaoHandler(pdvIntegracaoUsecase)
	authMiddleware := httpadapter.NewAuthMiddleware(authUsecase, senhaSessaoUsecase)
	rbacMiddleware := httpadapter.NewRBACMiddleware(authUsecase)
	papelPermissaoHandler := httpadapter.NewPapelPermissaoHandler(papelPermissaoUsecase)
	auditoriaHandler := httpadapter.NewAuditoriaHandler(auditoriaUsecase)

	// Compras handlers
	fornecedorHandler := httpadapter.NewFornecedorHandler(fornecedorUsecase)
	ordemCompraHandler := httpadapter.NewOrdemCompraHandler(ordemCompraUsecase)
	recebimentoCompraHandler := httpadapter.NewRecebimentoCompraHandler(recebimentoCompraUsecase)
	contaPagarHandler := httpadapter.NewContaPagarHandler(contaPagarUsecase)

	// Financeiro (Etapa 10) handlers
	contaGerencialHandler := httpadapter.NewContaGerencialHandler(contaGerencialUsecase)
	categoriaFinanceiraHandler := httpadapter.NewCategoriaFinanceiraHandler(categoriaFinanceiraUsecase)
	movimentacaoCaixaHandler := httpadapter.NewMovimentacaoCaixaHandler(movimentacaoCaixaUsecase)
	dreHandler := httpadapter.NewDREGerencialHandler(dreUsecase)
	fluxoCaixaHandler := httpadapter.NewFluxoCaixaHandler(fluxoCaixaUsecase)

	// Dashboard (Etapa 11) handler
	dashboardHandler := httpadapter.NewDashboardHandler(dashboardUsecase)

	// Estoque (Etapa 12) handler
	estoqueHandler := httpadapter.NewEstoqueHandler(estoqueUsecase)

	// Agente de IA (Etapa 13) handler
	agenteHandler := httpadapter.NewAgenteHandler(agenteUsecase)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.HTTPTimeout))

	// Health routes
	httpadapter.RegisterHealthRoutes(r, cfg)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/esqueci-senha", authHandler.EsqueciSenha)
		r.Post("/auth/redefinir-senha", authHandler.RedefinirSenha)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/me", authHandler.Me)
			r.Get("/me/permissoes", authHandler.MinhasPermissoes)
			r.Post("/auth/alterar-senha", authHandler.AlterarSenha)
			r.Post("/auth/logout", authHandler.Logout)

			// Lojas
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermLojasCriar))
				r.Post("/", lojaHandler.Criar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermLojasListar))
				r.Get("/", lojaHandler.Listar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermLojasVer))
				r.Get("/{id}", lojaHandler.BuscarPorID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermLojasAtualizar))
				r.Put("/{id}", lojaHandler.Atualizar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermLojasDeletar))
				r.Delete("/{id}", lojaHandler.Deletar)
			})

			// Papéis
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisCriar))
				r.Post("/papeis", papelPermissaoHandler.CriarPapel)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisListar))
				r.Get("/papeis", papelPermissaoHandler.ListarPapeis)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisVer))
				r.Get("/papeis/{id}", papelPermissaoHandler.BuscarPapel)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisAtualizar))
				r.Put("/papeis/{id}", papelPermissaoHandler.AtualizarPapel)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisDeletar))
				r.Delete("/papeis/{id}", papelPermissaoHandler.DeletarPapel)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPapeisGerenciarPermissoes))
				r.Post("/papeis/{papel_id}/permissoes", papelPermissaoHandler.AdicionarPermissaoAoPapel)
				r.Delete("/papeis/{papel_id}/permissoes/{permissao_id}", papelPermissaoHandler.RemoverPermissaoDoPapel)
				r.Get("/papeis/{id}/permissoes", papelPermissaoHandler.ListarPermissoesDoPapel)
			})

			// Permissões
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPermissoesListar))
				r.Get("/permissoes", papelPermissaoHandler.ListarPermissoes)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermPermissoesVer))
				r.Get("/permissoes/{id}", papelPermissaoHandler.BuscarPermissao)
			})

			// Auditoria (governança de acesso)
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermAuditoriaListar))
				r.Get("/auditoria", auditoriaHandler.Listar)
			})

			// Usuário-Papel (legacy - mantido para compatibilidade)
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosGerenciarPapeis))
				r.Post("/usuarios/{id}/papeis", papelPermissaoHandler.AssociarUsuarioPapel)
				r.Delete("/usuarios/{id}/papeis/{papel_id}", papelPermissaoHandler.RemoverUsuarioPapel)
				r.Get("/usuarios/{id}/papeis", papelPermissaoHandler.ListarPapeisDoUsuario)
			})

			// Usuários (Etapa 15)
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosCriar))
				r.Post("/usuarios", usuarioHandler.Criar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosListar))
				r.Get("/usuarios", usuarioHandler.Listar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosVer))
				r.Get("/usuarios/{id}", usuarioHandler.BuscarPorID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosAtualizar))
				r.Put("/usuarios/{id}", usuarioHandler.Atualizar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosDeletar))
				r.Delete("/usuarios/{id}", usuarioHandler.Desativar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosAtualizar))
				r.Post("/usuarios/{id}/reativar", usuarioHandler.Reativar)
			})

			// Usuário-Papel (novos endpoints)
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosGerenciarPapeis))
				r.Post("/usuarios/{id}/papeis", usuarioHandler.AssociarPapel)
				r.Delete("/usuarios/{id}/papeis/{papel_id}", usuarioHandler.RemoverPapel)
				r.Get("/usuarios/{id}/papeis", usuarioHandler.ListarPapeis)
				r.Get("/usuarios/{id}/permissoes", usuarioHandler.ListarPermissoesEfetivas)
			})

			// Usuário-Loja
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermUsuariosAtualizar))
				r.Post("/usuarios/{id}/lojas", usuarioHandler.AdicionarLoja)
				r.Delete("/usuarios/{id}/lojas/{loja_id}", usuarioHandler.RemoverLoja)
				r.Get("/usuarios/{id}/lojas", usuarioHandler.ListarLojas)
			})

			// Insumos
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosCriar))
				r.Post("/", insumoHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosListar))
				r.Get("/", insumoHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosVer))
				r.Get("/{id}", insumoHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosAtualizar))
				r.Put("/{id}", insumoHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosDeletar))
				r.Delete("/{id}", insumoHandler.Delete)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermInsumosGerenciarCustos))
				r.Post("/{id}/custos", insumoHandler.CreateCusto)
				r.Get("/{id}/custos", insumoHandler.ListCustos)
				r.Get("/{id}/custos/vigente", insumoHandler.GetCustoVigente)
				r.Get("/{id}/custos/atual", insumoHandler.GetCustoAtual)
			})

			// Produtos
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermProdutosCriar))
				r.Post("/", produtoHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermProdutosListar))
				r.Get("/", produtoHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermProdutosVer))
				r.Get("/{id}", produtoHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermProdutosAtualizar))
				r.Put("/{id}", produtoHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermProdutosDeletar))
				r.Delete("/{id}", produtoHandler.Delete)
			})

			// Fichas Técnicas
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasCriar))
				r.Post("/produtos/{id}/fichas-tecnicas", fichaTecnicaHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasListar))
				r.Get("/produtos/{id}/fichas-tecnicas", fichaTecnicaHandler.ListByProduto)
				r.Get("/fichas-tecnicas", fichaTecnicaHandler.ListByEmpresa)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasVer))
				r.Get("/fichas-tecnicas/{id}", fichaTecnicaHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasVer))
				r.Use(rbacMiddleware.RequireLojaAccess("id"))
				r.Get("/lojas/{id}/fichas-tecnicas", fichaTecnicaHandler.ListByLoja)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasAtualizar))
				r.Put("/fichas-tecnicas/{id}", fichaTecnicaHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFichasTecnicasDeletar))
				r.Delete("/fichas-tecnicas/{id}", fichaTecnicaHandler.Delete)
			})

			// CMV
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCMVVerResumo))
				r.Get("/cmv/resumo", cmvHandler.GetCMVResumo)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCMVVerDetalhe))
				r.Get("/cmv/produtos", cmvHandler.GetCMVProdutosDetalhe)
			})

			// Vendas / Ingestão
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermVendasIngest))
				r.Post("/ingest/sales-event", salesHandler.IngestSalesEvent)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermVendasVerResumo))
				r.Get("/vendas/resumo", salesHandler.GetVendasResumo)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermVendasListar))
				r.Get("/vendas", salesHandler.ListPedidos)
			})

			// Integrações PDV
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesCriar))
				r.Post("/integracoes", pdvIntegracaoHandler.Criar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesListar))
				r.Get("/integracoes", pdvIntegracaoHandler.Listar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesVer))
				r.Get("/integracoes/{id}", pdvIntegracaoHandler.BuscarPorID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesAtualizar))
				r.Put("/integracoes/{id}", pdvIntegracaoHandler.Atualizar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesDeletar))
				r.Delete("/integracoes/{id}", pdvIntegracaoHandler.Deletar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermIntegracoesSincronizar))
				r.Post("/integracoes/{id}/sincronizar", pdvIntegracaoHandler.Sincronizar)
			})

			// Compras - Fornecedores
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFornecedoresCriar))
				r.Post("/fornecedores", fornecedorHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFornecedoresListar))
				r.Get("/fornecedores", fornecedorHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFornecedoresVer))
				r.Get("/fornecedores/{id}", fornecedorHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFornecedoresAtualizar))
				r.Put("/fornecedores/{id}", fornecedorHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFornecedoresDeletar))
				r.Delete("/fornecedores/{id}", fornecedorHandler.Delete)
			})

			// Compras - Ordens de Compra
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraCriar))
				r.Post("/ordens-compra", ordemCompraHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraListar))
				r.Get("/ordens-compra", ordemCompraHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraVer))
				r.Get("/ordens-compra/{id}", ordemCompraHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraAtualizar))
				r.Put("/ordens-compra/{id}", ordemCompraHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraDeletar))
				r.Delete("/ordens-compra/{id}", ordemCompraHandler.Cancel)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraGerenciarItens))
				r.Post("/ordens-compra/{id}/itens", ordemCompraHandler.AddItem)
				r.Put("/ordens-compra/itens/{item_id}", ordemCompraHandler.UpdateItem)
				r.Delete("/ordens-compra/itens/{item_id}", ordemCompraHandler.RemoveItem)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermOrdensCompraListar))
				r.Use(rbacMiddleware.RequireLojaAccess("loja_id"))
				r.Get("/lojas/{loja_id}/ordens-compra", ordemCompraHandler.ListByLoja)
			})

			// Compras - Recebimentos
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermRecebimentosCriar))
				r.Post("/recebimentos-compra", recebimentoCompraHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermRecebimentosVer))
				r.Get("/recebimentos-compra/{id}", recebimentoCompraHandler.GetByID)
				r.Get("/ordens-compra/{ordem_compra_id}/recebimentos", recebimentoCompraHandler.ListByOrdemCompra)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermRecebimentosListar))
				r.Get("/recebimentos-compra", recebimentoCompraHandler.ListByFilter)
			})

			// Compras - Contas a Pagar
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarCriar))
				r.Post("/contas-pagar", contaPagarHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarListar))
				r.Get("/contas-pagar", contaPagarHandler.List)
				r.Get("/contas-pagar/fornecedores/{fornecedor_id}", contaPagarHandler.ListByFornecedor)
				r.Get("/contas-pagar/status", contaPagarHandler.ListByStatus)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarVer))
				r.Get("/contas-pagar/{id}", contaPagarHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarAtualizar))
				r.Put("/contas-pagar/{id}", contaPagarHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarPagar))
				r.Post("/contas-pagar/{id}/pagar", contaPagarHandler.Pagar)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasPagarDeletar))
				r.Delete("/contas-pagar/{id}", contaPagarHandler.Delete)
			})

			// Financeiro (Etapa 10) - Plano de Contas Gerencial
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasGerenciaisCriar))
				r.Post("/contas-gerenciais", contaGerencialHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasGerenciaisListar))
				r.Get("/contas-gerenciais", contaGerencialHandler.List)
				r.Get("/contas-gerenciais/hierarquia", contaGerencialHandler.ListWithHierarchy)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasGerenciaisVer))
				r.Get("/contas-gerenciais/{id}", contaGerencialHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasGerenciaisAtualizar))
				r.Put("/contas-gerenciais/{id}", contaGerencialHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermContasGerenciaisDeletar))
				r.Delete("/contas-gerenciais/{id}", contaGerencialHandler.Delete)
			})

			// Financeiro (Etapa 10) - Categorias Financeiras
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCategoriasFinanceirasCriar))
				r.Post("/categorias-financeiras", categoriaFinanceiraHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCategoriasFinanceirasListar))
				r.Get("/categorias-financeiras", categoriaFinanceiraHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCategoriasFinanceirasVer))
				r.Get("/categorias-financeiras/{id}", categoriaFinanceiraHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCategoriasFinanceirasAtualizar))
				r.Put("/categorias-financeiras/{id}", categoriaFinanceiraHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermCategoriasFinanceirasDeletar))
				r.Delete("/categorias-financeiras/{id}", categoriaFinanceiraHandler.Delete)
			})

			// Financeiro (Etapa 10) - Movimentações de Caixa
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermMovimentacoesCaixaCriar))
				r.Post("/movimentacoes-caixa", movimentacaoCaixaHandler.Create)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermMovimentacoesCaixaListar))
				r.Get("/movimentacoes-caixa", movimentacaoCaixaHandler.List)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermMovimentacoesCaixaVer))
				r.Get("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.GetByID)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermMovimentacoesCaixaAtualizar))
				r.Put("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.Update)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermMovimentacoesCaixaDeletar))
				r.Delete("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.Delete)
			})

			// Financeiro (Etapa 10) - DRE Gerencial
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermDREVer))
				r.Get("/dre", dreHandler.GetDRE)
			})

			// Financeiro (Etapa 10) - Fluxo de Caixa
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermFluxoCaixaVer))
				r.Get("/fluxo-caixa", fluxoCaixaHandler.GetFluxoCaixa)
			})

			// Dashboard (Etapa 11) - Indicadores por loja e período
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermDashboardVer))
				r.Get("/dashboard", dashboardHandler.GetDashboard)
			})

			// Estoque (Etapa 12) - Saldos e movimentações
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueVerSaldos))
				r.Get("/estoque/saldos", estoqueHandler.ListSaldos)
				r.Get("/estoque/saldos/{insumo_id}", estoqueHandler.GetSaldo)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueRegistrarSaida))
				r.Use(rbacMiddleware.RequireLojaAccessBody("loja_id"))
				r.Post("/estoque/saidas", estoqueHandler.RegistrarSaida)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueRegistrarAjuste))
				r.Use(rbacMiddleware.RequireLojaAccessBody("loja_id"))
				r.Post("/estoque/ajustes", estoqueHandler.RegistrarAjuste)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueTransferir))
				r.Use(rbacMiddleware.RequireLojaAccessBody("loja_origem_id"))
				r.Use(rbacMiddleware.RequireLojaAccessBody("loja_destino_id"))
				r.Post("/estoque/transferencias", estoqueHandler.Transferir)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueVerMovimentacoes))
				r.Get("/estoque/movimentacoes", estoqueHandler.ListMovimentacoes)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermEstoqueVerPerdas))
				r.Get("/estoque/perdas", estoqueHandler.GetResumoPerdas)
			})

			// Agente de IA (Etapa 13) - Análise gerencial
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermAgenteCriarSessao))
				r.Post("/agente/sessoes", agenteHandler.CriarSessao)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermAgenteListarSessoes))
				r.Get("/agente/sessoes", agenteHandler.ListarSessoes)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermAgenteVerSessao))
				r.Get("/agente/sessoes/{id}", agenteHandler.ObterSessao)
			})
			r.Group(func(r chi.Router) {
				r.Use(rbacMiddleware.RequirePermission(domain.PermAgentePerguntar))
				r.Post("/agente/sessoes/{id}/perguntas", agenteHandler.Perguntar)
			})
		})
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.AppPort).Str("env", cfg.AppEnv).Msg("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
