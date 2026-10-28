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
	"github.com/intelligate/erp/backend/internal/adapter/db"
	httpadapter "github.com/intelligate/erp/backend/internal/adapter/http"
	"github.com/intelligate/erp/backend/internal/config"
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
	auditoriaRepo := db.NewAuditoriaRepository(dbpool)
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

	// Initialize crypto service
	cryptoService, err := crypto.NewCryptoServiceFromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize crypto service: INTELLIGATE_ENCRYPTION_KEY must be set and valid")
	}

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(usuarioRepo, cfg.JWTSecret, cfg.JWTExpiry)
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

	// Initialize handlers
	authHandler := httpadapter.NewAuthHandler(authUsecase)
	lojaHandler := httpadapter.NewLojaHandler(lojaUsecase)
	salesHandler := httpadapter.NewSalesHandler(salesUsecase)
	insumoHandler := httpadapter.NewInsumoHandler(insumoUsecase)
	produtoHandler := httpadapter.NewProdutoHandler(produtoUsecase)
	fichaTecnicaHandler := httpadapter.NewFichaTecnicaHandler(fichaTecnicaUsecase)
	cmvHandler := httpadapter.NewCMVHandler(fichaTecnicaUsecase)
	pdvIntegracaoHandler := httpadapter.NewPdvIntegracaoHandler(pdvIntegracaoUsecase)
	authMiddleware := httpadapter.NewAuthMiddleware(authUsecase)

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

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/me", authHandler.Me)

			// Lojas
			r.Post("/lojas", lojaHandler.Criar)
			r.Get("/lojas", lojaHandler.Listar)
			r.Get("/lojas/{id}", lojaHandler.BuscarPorID)
			r.Put("/lojas/{id}", lojaHandler.Atualizar)
			r.Delete("/lojas/{id}", lojaHandler.Deletar)

			// Insumos
			r.Post("/insumos", insumoHandler.Create)
			r.Get("/insumos", insumoHandler.List)
			r.Get("/insumos/{id}", insumoHandler.GetByID)
			r.Put("/insumos/{id}", insumoHandler.Update)
			r.Delete("/insumos/{id}", insumoHandler.Delete)
			r.Post("/insumos/{id}/custos", insumoHandler.CreateCusto)
			r.Get("/insumos/{id}/custos", insumoHandler.ListCustos)
			r.Get("/insumos/{id}/custos/vigente", insumoHandler.GetCustoVigente)
			r.Get("/insumos/{id}/custos/atual", insumoHandler.GetCustoAtual)

			// Produtos
			r.Post("/produtos", produtoHandler.Create)
			r.Get("/produtos", produtoHandler.List)
			r.Get("/produtos/{id}", produtoHandler.GetByID)
			r.Put("/produtos/{id}", produtoHandler.Update)
			r.Delete("/produtos/{id}", produtoHandler.Delete)

			// Fichas Técnicas
			r.Post("/produtos/{id}/fichas-tecnicas", fichaTecnicaHandler.Create)
			r.Get("/produtos/{id}/fichas-tecnicas", fichaTecnicaHandler.ListByProduto)
			r.Get("/fichas-tecnicas/{id}", fichaTecnicaHandler.GetByID)
			r.Put("/fichas-tecnicas/{id}", fichaTecnicaHandler.Update)
			r.Delete("/fichas-tecnicas/{id}", fichaTecnicaHandler.Delete)
			r.Get("/lojas/{id}/fichas-tecnicas", fichaTecnicaHandler.ListByLoja)
			r.Get("/fichas-tecnicas", fichaTecnicaHandler.ListByEmpresa)

			// CMV
			r.Get("/cmv/resumo", cmvHandler.GetCMVResumo)
			r.Get("/cmv/produtos", cmvHandler.GetCMVProdutosDetalhe)

			// Vendas / Ingestão
			r.Post("/ingest/sales-event", salesHandler.IngestSalesEvent)
			r.Get("/vendas/resumo", salesHandler.GetVendasResumo)
			r.Get("/vendas", salesHandler.ListPedidos)

			// Integrações PDV
			r.Post("/integracoes", pdvIntegracaoHandler.Criar)
			r.Get("/integracoes", pdvIntegracaoHandler.Listar)
			r.Get("/integracoes/{id}", pdvIntegracaoHandler.BuscarPorID)
			r.Put("/integracoes/{id}", pdvIntegracaoHandler.Atualizar)
			r.Delete("/integracoes/{id}", pdvIntegracaoHandler.Deletar)
			r.Post("/integracoes/{id}/sincronizar", pdvIntegracaoHandler.Sincronizar)

			// Compras - Fornecedores
			r.Post("/fornecedores", fornecedorHandler.Create)
			r.Get("/fornecedores", fornecedorHandler.List)
			r.Get("/fornecedores/{id}", fornecedorHandler.GetByID)
			r.Put("/fornecedores/{id}", fornecedorHandler.Update)
			r.Delete("/fornecedores/{id}", fornecedorHandler.Delete)

			// Compras - Ordens de Compra
			r.Post("/ordens-compra", ordemCompraHandler.Create)
			r.Get("/ordens-compra", ordemCompraHandler.List)
			r.Get("/ordens-compra/{id}", ordemCompraHandler.GetByID)
			r.Put("/ordens-compra/{id}", ordemCompraHandler.Update)
			r.Delete("/ordens-compra/{id}", ordemCompraHandler.Cancel)
			r.Post("/ordens-compra/{id}/itens", ordemCompraHandler.AddItem)
			r.Put("/ordens-compra/itens/{item_id}", ordemCompraHandler.UpdateItem)
			r.Delete("/ordens-compra/itens/{item_id}", ordemCompraHandler.RemoveItem)
			r.Get("/lojas/{loja_id}/ordens-compra", ordemCompraHandler.ListByLoja)

			// Compras - Recebimentos
			r.Post("/recebimentos-compra", recebimentoCompraHandler.Create)
			r.Get("/recebimentos-compra/{id}", recebimentoCompraHandler.GetByID)
			r.Get("/ordens-compra/{ordem_compra_id}/recebimentos", recebimentoCompraHandler.ListByOrdemCompra)
			r.Get("/recebimentos-compra", recebimentoCompraHandler.ListByFilter)

			// Compras - Contas a Pagar
			r.Post("/contas-pagar", contaPagarHandler.Create)
			r.Get("/contas-pagar", contaPagarHandler.List)
			r.Get("/contas-pagar/{id}", contaPagarHandler.GetByID)
			r.Put("/contas-pagar/{id}", contaPagarHandler.Update)
			r.Post("/contas-pagar/{id}/pagar", contaPagarHandler.Pagar)
			r.Delete("/contas-pagar/{id}", contaPagarHandler.Delete)
			r.Get("/contas-pagar/fornecedores/{fornecedor_id}", contaPagarHandler.ListByFornecedor)
			r.Get("/contas-pagar/status", contaPagarHandler.ListByStatus)

			// Financeiro (Etapa 10) - Plano de Contas Gerencial
			r.Post("/contas-gerenciais", contaGerencialHandler.Create)
			r.Get("/contas-gerenciais", contaGerencialHandler.List)
			r.Get("/contas-gerenciais/hierarquia", contaGerencialHandler.ListWithHierarchy)
			r.Get("/contas-gerenciais/{id}", contaGerencialHandler.GetByID)
			r.Put("/contas-gerenciais/{id}", contaGerencialHandler.Update)
			r.Delete("/contas-gerenciais/{id}", contaGerencialHandler.Delete)

			// Financeiro (Etapa 10) - Categorias Financeiras
			r.Post("/categorias-financeiras", categoriaFinanceiraHandler.Create)
			r.Get("/categorias-financeiras", categoriaFinanceiraHandler.List)
			r.Get("/categorias-financeiras/{id}", categoriaFinanceiraHandler.GetByID)
			r.Put("/categorias-financeiras/{id}", categoriaFinanceiraHandler.Update)
			r.Delete("/categorias-financeiras/{id}", categoriaFinanceiraHandler.Delete)

			// Financeiro (Etapa 10) - Movimentações de Caixa
			r.Post("/movimentacoes-caixa", movimentacaoCaixaHandler.Create)
			r.Get("/movimentacoes-caixa", movimentacaoCaixaHandler.List)
			r.Get("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.GetByID)
			r.Put("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.Update)
			r.Delete("/movimentacoes-caixa/{id}", movimentacaoCaixaHandler.Delete)

			// Financeiro (Etapa 10) - DRE Gerencial
			r.Get("/dre", dreHandler.GetDRE)

			// Financeiro (Etapa 10) - Fluxo de Caixa
			r.Get("/fluxo-caixa", fluxoCaixaHandler.GetFluxoCaixa)

			// Dashboard (Etapa 11) - Indicadores por loja e período
			r.Get("/dashboard", dashboardHandler.GetDashboard)

			// Estoque (Etapa 12) - Saldos e movimentações
			r.Get("/estoque/saldos", estoqueHandler.ListSaldos)
			r.Get("/estoque/saldos/{insumo_id}", estoqueHandler.GetSaldo)
			r.Post("/estoque/saidas", estoqueHandler.RegistrarSaida)
			r.Post("/estoque/ajustes", estoqueHandler.RegistrarAjuste)
			r.Post("/estoque/transferencias", estoqueHandler.Transferir)
			r.Get("/estoque/movimentacoes", estoqueHandler.ListMovimentacoes)
			r.Get("/estoque/perdas", estoqueHandler.GetResumoPerdas)
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
