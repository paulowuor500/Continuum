package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"continuum/api/internal/handler"
	"continuum/api/internal/repository"
	"continuum/api/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

func connectToDatabaseWithRetry(dataSourceName string) (*repository.Database, error) {
	var (
		db  *repository.Database
		err error
	)

	for attempt := 1; attempt <= 30; attempt++ {
		db, err = repository.NewDatabase(dataSourceName)
		if err == nil {
			return db, nil
		}

		log.Printf("⏳ Database connection attempt %d/30 failed: %v", attempt, err)
		if attempt == 30 {
			return nil, err
		}
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func applyProductionMigrations(db *repository.Database) error {
	if db == nil || db.Db == nil {
		return nil
	}

	migrationSQL, err := os.ReadFile("internal/repository/migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = db.Db.Exec(string(migrationSQL))
	if err != nil {
		return err
	}

	log.Println("✅ Database schema check completed.")
	return nil
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("ℹ️  No .env file found — using host environment.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/continuum?sslmode=disable"
	}

	log.Println("🔌 Connecting to database...")
	db, err := connectToDatabaseWithRetry(dbURL)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	if err := applyProductionMigrations(db); err != nil {
		log.Printf("⚠️  Database migration check failed: %v", err)
	}

	// Lightning service (falls back to mock when LND is not configured)
	lndConfig := &service.LNDConfig{
		Host:         os.Getenv("LND_HOST"),
		MacaroonPath: os.Getenv("LND_MACAROON_PATH"),
		TLSCertPath:  os.Getenv("LND_TLS_CERT_PATH"),
	}
	lightningService, err := service.NewLightningService(lndConfig)
	if err != nil {
		log.Printf("⚠️  LND init failed (%v) — running in mock mode.", err)
	}

	multisigService := service.NewMultisigServiceFromEnv()
	vaultService := service.NewVaultService(db, lightningService)
	schedulerService := service.NewScheduler(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.StartCheckLoop(ctx, 10*time.Second)
	log.Println("⏰ Dead-man switch scheduler running.")

	app := fiber.New(fiber.Config{
		AppName: "Continuum Inheritance Protocol v1.0",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PATCH", "OPTIONS"},
	}))

	// Wire all dependencies into handlers
	vaultHandler := handler.NewVaultHandler(db, vaultService, multisigService, lightningService)
	proofHandler := handler.NewProofHandler(db, lightningService)
	recoveryHandler := handler.NewRecoveryHandler(db)

	api := app.Group("/api")
	{
		api.Get("/health", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"status": "ok", "service": "continuum-api"})
		})

		// Vault CRUD
		api.Get("/vaults", vaultHandler.ListVaults)
		api.Post("/vaults", vaultHandler.CreateVault)
		api.Get("/vaults/:id", recoveryHandler.GetVaultStatus)

		// Proof-of-life flow
		api.Post("/vaults/:id/invoice", vaultHandler.RequestCheckInToken)
		// Frontend calls /check-in (with hyphen) — register both for compatibility
		api.Post("/vaults/:id/checkin", vaultHandler.ConfirmCheckIn)
		api.Post("/vaults/:id/check-in", vaultHandler.ConfirmCheckIn)

		// Beneficiary management
		api.Post("/vaults/:id/beneficiaries", vaultHandler.AddBeneficiary)

		// Timer update (frontend calls PATCH /vaults/:id/timer)
		api.Patch("/vaults/:id/timer", func(c fiber.Ctx) error {
			id := c.Params("id")
			var body struct {
				CheckInIntervalSeconds int `json:"check_in_interval_seconds"`
			}
			if err := c.Bind().Body(&body); err != nil || body.CheckInIntervalSeconds <= 0 {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid interval"})
			}
			if err := db.UpdateVaultInterval(c.Context(), id, body.CheckInIntervalSeconds); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(fiber.Map{"status": "SUCCESS", "message": "Timer updated."})
		})

		// Dev-only: time-warp a vault to trigger dormancy
		if os.Getenv("ALLOW_DEV_TIME_WARP") == "true" {
			log.Println("⚠️  Time-warp backdoor enabled (dev only).")
			api.Post("/vaults/:id/warp", proofHandler.SimulateTimeWarp)
		} else {
			log.Println("🔒 Time-warp endpoint disabled.")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Printf("🚀 Continuum API listening on :%s\n", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("⚠️  Server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gracefully...")
	cancel()
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("✨ Server closed cleanly.")
}
