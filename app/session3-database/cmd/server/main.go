package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"mini-asm/internal/database"
	"mini-asm/internal/handler"
	"mini-asm/internal/service"
	"mini-asm/internal/storage/postgres"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	log.Println("🚀 Starting Mini ASM Server (Session 3 - Database)...")

	// ============================================
	// CONFIGURATION - Load from environment
	// ============================================

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "secops")
	dbPassword := getEnv("DB_PASSWORD", "secops123")
	dbName := getEnv("DB_NAME", "mini_asm")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)

	log.Printf("📊 Connecting to database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	// ============================================
	// DATABASE CONNECTION WITH RETRY
	// ============================================

	db, err := database.ConnectWithRetry(connStr, 5)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Optional: Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// ============================================
	// DEPENDENCY INJECTION - Wire up all layers
	// ============================================

	store := postgres.NewPostgresStorage(db)
	log.Println("✅ Storage initialized: PostgreSQL")

	assetService := service.NewAssetService(store)
	log.Println("✅ Service initialized: AssetService")

	assetHandler := handler.NewAssetHandler(assetService)
	healthHandler := handler.NewHealthHandler(db)
	log.Println("✅ Handlers initialized")

	// ============================================
	// ROUTING - Register HTTP endpoints
	// ============================================

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.Check)

	// Statistics (registered before /assets/{id} to avoid ambiguity)
	mux.HandleFunc("GET /assets/stats", assetHandler.GetStats)
	mux.HandleFunc("GET /assets/count", assetHandler.CountAssets)
	mux.HandleFunc("GET /assets/search", assetHandler.SearchAssets)

	// Batch operations
	mux.HandleFunc("POST /assets/batch", assetHandler.BatchCreate)
	mux.HandleFunc("DELETE /assets/batch", assetHandler.BatchDelete)

	// Asset CRUD operations
	mux.HandleFunc("POST /assets", assetHandler.CreateAsset)
	mux.HandleFunc("GET /assets", assetHandler.ListAssets)
	mux.HandleFunc("GET /assets/{id}", assetHandler.GetAsset)
	mux.HandleFunc("PUT /assets/{id}", assetHandler.UpdateAsset)
	mux.HandleFunc("DELETE /assets/{id}", assetHandler.DeleteAsset)

	log.Println("✅ Routes registered:")
	log.Println("   GET    /health")
	log.Println("   GET    /assets/stats")
	log.Println("   GET    /assets/count")
	log.Println("   GET    /assets/search")
	log.Println("   POST   /assets/batch")
	log.Println("   DELETE /assets/batch")
	log.Println("   POST   /assets")
	log.Println("   GET    /assets")
	log.Println("   GET    /assets/{id}")
	log.Println("   PUT    /assets/{id}")
	log.Println("   DELETE /assets/{id}")

	// ============================================
	// START SERVER
	// ============================================

	port := getEnv("SERVER_PORT", "8080")
	addr := ":" + port

	log.Printf("🌐 Server listening on http://localhost%s\n", addr)
	log.Println("📖 API Documentation: see docs/api.yml")
	log.Println("🗄️  Database: PostgreSQL (persistent storage)")
	log.Println("Press Ctrl+C to stop")
	log.Println()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

// getEnv retrieves an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
