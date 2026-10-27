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
	produtoCanonicoRepo := db.NewProdutoCanonicoRepository(dbpool)
	pedidoCanonicoRepo := db.NewPedidoCanonicoRepository(dbpool)
	pedidoItemCanonicoRepo := db.NewPedidoItemCanonicoRepository(dbpool)
	pedidoPagamentoCanonicoRepo := db.NewPedidoPagamentoCanonicoRepository(dbpool)
	eventoIngestaoRepo := db.NewEventoIngestaoRepository(dbpool)
	vendasResumoRepo := db.NewVendasResumoRepository(dbpool)
	salesTxRepo := db.NewSalesTxRepository(dbpool)

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

	// Initialize handlers
	authHandler := httpadapter.NewAuthHandler(authUsecase)
	lojaHandler := httpadapter.NewLojaHandler(lojaUsecase)
	salesHandler := httpadapter.NewSalesHandler(salesUsecase)
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

			// Vendas / Ingestão
			r.Post("/ingest/sales-event", salesHandler.IngestSalesEvent)
			r.Get("/vendas/resumo", salesHandler.GetVendasResumo)
			r.Get("/vendas", salesHandler.ListPedidos)
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
