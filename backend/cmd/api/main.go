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
