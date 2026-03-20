// Package main provides the API server entry point.
package main

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "telegram-api/docs"
	"telegram-api/internal/config"
	"telegram-api/internal/handler"
	"telegram-api/internal/middleware"
	"telegram-api/internal/repository/postgres"
	"telegram-api/internal/repository/redis"
	"telegram-api/internal/service"
	"telegram-api/internal/telegram"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	redisLib "github.com/redis/go-redis/v9"
)

// Version se inyecta en build time con -ldflags.
var Version = "0.1.1"

// @title Telegram API
// @version 0.1.0
// @description API REST para gestionar múltiples sesiones de Telegram via MTProto.
// @host localhost:7789
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization.
// setupDatabase initializes and connects to PostgreSQL database.
func setupDatabase(ctx context.Context, cfg *config.Config) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("PostgreSQL connection failed")
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("PostgreSQL ping failed")
	}
	logger.Info().Msg("PostgreSQL connected")

	if err := runMigrations(pool); err != nil {
		logger.Fatal().Err(err).Msg("Migrations failed")
	}

	return pool
}

// setupRedis initializes and connects to Redis.
func setupRedis(ctx context.Context, cfg *config.Config) *redisLib.Client {
	rdb := redisLib.NewClient(&redisLib.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("Redis ping failed")
	}
	logger.Info().Msg("Redis connected")

	return rdb
}

// setupRepositories initializes all repository instances.
func setupRepositories(pool *pgxpool.Pool, rdb *redisLib.Client) (
	userRepo *postgres.UserRepository,
	tokenRepo *postgres.RefreshTokenRepository,
	sessionRepo *postgres.SessionRepository,
	webhookRepo *postgres.WebhookRepository,
	cacheRepo *redis.CacheRepository,
) {
	userRepo = postgres.NewUserRepository(pool)
	tokenRepo = postgres.NewRefreshTokenRepository(pool)
	sessionRepo = postgres.NewSessionRepository(pool)
	webhookRepo = postgres.NewWebhookRepository(pool)
	cacheRepo = redis.NewCacheRepository(rdb)
	return
}

// setupTelegram initializes Telegram manager and session pool.
func setupTelegram(
	cfg *config.Config,
	sessionRepo *postgres.SessionRepository,
	webhookRepo *postgres.WebhookRepository,
) (*telegram.ClientManager, *telegram.SessionPool) {
	tgManager, err := telegram.NewManager(cfg, sessionRepo)
	if err != nil {
		logger.Fatal().Err(err).Msg("Telegram Manager failed")
	}

	sessionPool := telegram.NewSessionPool(tgManager, sessionRepo, webhookRepo)
	tgManager.SetPool(sessionPool)

	return tgManager, sessionPool
}

// setupServices initializes all service instances.
func setupServices(
	userRepo *postgres.UserRepository,
	tokenRepo *postgres.RefreshTokenRepository,
	sessionRepo *postgres.SessionRepository,
	_ *postgres.WebhookRepository,
	cacheRepo *redis.CacheRepository,
	tgManager *telegram.ClientManager,
	sessionPool *telegram.SessionPool,
	cfg *config.Config,
) (
	authService *service.AuthService,
	sessionService *service.SessionService,
	messageService *service.MessageService,
	chatService *service.ChatService,
) {
	authService = service.NewAuthService(userRepo, tokenRepo, cacheRepo, cfg)
	sessionService = service.NewSessionService(sessionRepo, userRepo, tgManager, cacheRepo, cfg)
	messageService = service.NewMessageService(sessionRepo, cacheRepo, tgManager, sessionPool)
	chatService = service.NewChatService(sessionRepo, cacheRepo, tgManager, cfg)
	return
}

// setupFiberApp creates and configures the Fiber application.
func setupFiberApp(sessionPool *telegram.SessionPool) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Telegram API v" + Version,
	})
	app.Use(recover.New())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestLogger())

	// Documentation
	app.Get("/docs/*", swagger.HandlerDefault)
	app.Get("/redoc", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/redoc.html")
	})

	// Health & Info
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":          "ok",
			"version":         Version,
			"active_sessions": sessionPool.ActiveCount(),
		})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Telegram API",
			"version": Version,
			"docs": fiber.Map{
				"swagger": "/docs/",
				"redoc":   "/redoc",
				"openapi": "/docs/swagger.json",
			},
		})
	})

	return app
}

// setupRoutes configures all application routes.
func setupRoutes(
	app *fiber.App,
	authService *service.AuthService,
	sessionService *service.SessionService,
	messageService *service.MessageService,
	chatService *service.ChatService,
	webhookRepo *postgres.WebhookRepository,
	sessionRepo *postgres.SessionRepository,
	sessionPool *telegram.SessionPool,
) {
	api := app.Group("/api/v1")

	authHandler := handler.NewAuthHandler(authService)
	authHandler.RegisterRoutes(api)

	protected := api.Group("/", middleware.JWTMiddleware(authService))

	sessionHandler := handler.NewSessionHandler(sessionService)
	sessionHandler.RegisterRoutes(protected)

	messageHandler := handler.NewMessageHandler(messageService)
	messageHandler.RegisterRoutes(protected)

	chatHandler := handler.NewChatHandler(chatService)
	chatHandler.RegisterRoutes(protected)

	webhookHandler := handler.NewWebhookHandler(webhookRepo, sessionRepo, sessionPool)
	webhookHandler.RegisterRoutes(protected)

	printRoutes(app)
}

// startServer starts the Fiber server.
func startServer(app *fiber.App) {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info().
		Str("port", port).
		Str("version", Version).
		Str("swagger", "http://localhost:"+port+"/docs/").
		Str("redoc", "http://localhost:"+port+"/redoc").
		Msg("🚀 Server started")

	if err := app.Listen(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic("config error: " + err.Error())
	}

	logger.Init(cfg.Log.Level)
	logger.Info().Str("version", Version).Msg("🚀 Telegram API starting...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := setupDatabase(ctx, cfg)
	defer func() {
		if err := pool.Close(); err != nil {
			logger.Error().Err(err).Msg("PostgreSQL close failed")
		}
	}()

	rdb := setupRedis(ctx, cfg)
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error().Err(err).Msg("Redis close failed")
		}
	}()

	userRepo, tokenRepo, sessionRepo, webhookRepo, cacheRepo := setupRepositories(pool, rdb)

	tgManager, sessionPool := setupTelegram(cfg, sessionRepo, webhookRepo)

	authService, sessionService, messageService, chatService := setupServices(
		userRepo, tokenRepo, sessionRepo, webhookRepo, cacheRepo, tgManager, sessionPool, cfg,
	)

	app := setupFiberApp(sessionPool)

	setupRoutes(app, authService, sessionService, messageService, chatService, webhookRepo, sessionRepo, sessionPool)

	startServer(app)
}

func runMigrations(pool *pgxpool.Pool) error {
	paths := []string{"db/migrations/*.sql", "/db/migrations/*.sql", "migrations/*.sql"}

	var files []string
	for _, p := range paths {
		if found, _ := filepath.Glob(p); len(found) > 0 {
			files = found
			break
		}
	}

	if len(files) == 0 {
		logger.Warn().Msg("No migration files found")
		return nil
	}

	sort.Strings(files)

	ctx := context.Background()
	for _, f := range files {
		// #nosec G304 -- Reading migration files from trusted directory
		schema, err := os.ReadFile(f)
		if err != nil {
			logger.Error().Err(err).Str("file", f).Msg("Error reading migration")
			return err
		}
		if _, err := pool.Exec(ctx, string(schema)); err != nil {
			logger.Error().Err(err).Str("file", f).Msg("Error executing migration")
			return err
		}
		logger.Info().Str("file", filepath.Base(f)).Msg("Migration applied")
	}
	return nil
}

func printRoutes(app *fiber.App) {
	logger.Info().Msg("📍 Routes registered:")
	valid := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true}
	seen := make(map[string]bool)
	for _, r := range app.GetRoutes() {
		key := r.Method + r.Path
		if valid[r.Method] && r.Path != "/" && r.Path != "/api/v1/" && !seen[key] {
			seen[key] = true
			logger.Info().Str("method", r.Method).Str("path", r.Path).Msg("")
		}
	}
}
