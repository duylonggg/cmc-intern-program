package main

import (
	"log"
	"net/http"
	"os"

	"mini-asm/internal/config"
	"mini-asm/internal/handler"
	"mini-asm/internal/service"
	"mini-asm/internal/storage/postgres"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	log.Println("🚀 Starting Mini ASM Server (Session 5 - EASM Scanning)...")

	// ============================================
	// CONFIGURATION - Load from environment
	// ============================================
	config, err := config.LoadPostgresConfig(".env")
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}

	// ============================================
	// DEPENDENCY INJECTION - Wire up all layers
	// ============================================

	// Storage layer
	store, err := postgres.NewPostgresStorageFromConfig(config)
	if err != nil {
		log.Fatal("❌ Failed to initialize storage:", err)
	}
	log.Println("✅ Storage initialized: PostgreSQL")

	// Service layer (asset service from Session 4)
	assetService := service.NewAssetService(store)
	log.Println("✅ Service initialized: AssetService with Validator")

	// Scan service (new in Session 5)
	scanService, err := service.NewScanService(store, store)
	if err != nil {
		log.Fatal("❌ Failed to initialize scan service:", err)
	}
	log.Println("✅ Service initialized: ScanService with DNS, WHOIS, Subdomain scanners")

	// Handler layer
	assetHandler := handler.NewAssetHandler(assetService)
	scanHandler := handler.NewScanHandler(scanService)
	healthHandler := handler.NewHealthHandler()
	log.Println("✅ Handlers initialized")

	// ============================================
	// ROUTING - Register HTTP endpoints
	// ============================================

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.Check)

	// Asset CRUD operations (Session 2-4)
	mux.HandleFunc("POST /assets", assetHandler.CreateAsset)
	mux.HandleFunc("GET /assets", assetHandler.ListAssets)
	mux.HandleFunc("GET /assets/{id}", assetHandler.GetAsset)
	mux.HandleFunc("PUT /assets/{id}", assetHandler.UpdateAsset)
	mux.HandleFunc("DELETE /assets/{id}", assetHandler.DeleteAsset)

	// Scan operations (Session 5 - NEW!)
	mux.HandleFunc("POST /assets/{id}/scan", scanHandler.StartScan)
	mux.HandleFunc("POST /assets/{id}/scan/demo", scanHandler.DemoSyncVsAsync) // Demo: Sync vs Async
	mux.HandleFunc("GET /assets/{id}/scans", scanHandler.ListScanJobs)
	mux.HandleFunc("GET /scan-jobs/{id}", scanHandler.GetScanJob)
	mux.HandleFunc("GET /scan-jobs/{id}/results", scanHandler.GetScanResults)

	// Scan results by asset (Session 5 - NEW!)
	mux.HandleFunc("GET /assets/{id}/results", scanHandler.GetAssetAllResults)    // All results (NEW!)
	mux.HandleFunc("GET /assets/{id}/subdomains", scanHandler.GetAssetSubdomains) // Subdomains only
	mux.HandleFunc("GET /assets/{id}/dns", scanHandler.GetAssetDNS)               // DNS only
	mux.HandleFunc("GET /assets/{id}/whois", scanHandler.GetAssetWHOIS)           // WHOIS only

	log.Println("✅ Routes registered:")
	log.Println("   === Health ===")
	log.Println("   GET    /health")
	log.Println()
	log.Println("   === Assets (Session 2-4) ===")
	log.Println("   POST   /assets")
	log.Println("   GET    /assets         (📋 pagination, 🔍 filters, 🔄 sort)")
	log.Println("   GET    /assets/{id}")
	log.Println("   PUT    /assets/{id}")
	log.Println("   DELETE /assets/{id}")
	log.Println()
	log.Println("   === Scanning (Session 5 - NEW!) ===")
	log.Println("   POST   /assets/{id}/scan          🔍 Start scan (async)")
	log.Println("   POST   /assets/{id}/scan/demo     🎓 Demo: Sync vs Async comparison")
	log.Println("   GET    /assets/{id}/scans         📋 List all scans")
	log.Println("   GET    /scan-jobs/{id}            📊 Check scan status")
	log.Println("   GET    /scan-jobs/{id}/results    📄 Get scan results")
	log.Println()
	log.Println("   === Scan Results by Asset ===")
	log.Println("   GET    /assets/{id}/results       📦 ALL results (DNS + WHOIS + Subdomains)")
	log.Println("   GET    /assets/{id}/subdomains    🌐 Subdomains only")
	log.Println("   GET    /assets/{id}/dns           📡 DNS records only")
	log.Println("   GET    /assets/{id}/whois         📇 WHOIS info only")

	// ============================================
	// START SERVER
	// ============================================

	port := getEnv("SERVER_PORT", "8080")
	addr := ":" + port

	log.Printf("🌐 Server listening on http://localhost%s\n", addr)
	log.Println()
	log.Println("🆕 Session 5 Features:")
	log.Println("   ✓ WHOIS Scanning - Domain registration info")
	log.Println("   ✓ DNS Scanning - A, AAAA, MX, NS, TXT, CNAME records")
	log.Println("   ✓ Subdomain Enumeration - DNS bruteforce discovery")
	log.Println("   ✓ Async Job Pattern - Background processing")
	log.Println("   ✓ Result Tracking - Status polling")
	log.Println()
	log.Println("📝 Example: Start a DNS scan")
	log.Println("   1. Create asset:")
	log.Println("      curl -X POST http://localhost:8080/assets \\")
	log.Println("        -H \"Content-Type: application/json\" \\")
	log.Println("        -d '{\"name\":\"example.com\",\"type\":\"domain\"}'")
	log.Println()
	log.Println("   2. Start DNS scan:")
	log.Println("      curl -X POST http://localhost:8080/assets/{asset-id}/scan \\")
	log.Println("        -H \"Content-Type: application/json\" \\")
	log.Println("        -d '{\"scan_type\":\"dns\"}'")
	log.Println()
	log.Println("   3. Check scan status:")
	log.Println("      curl http://localhost:8080/scan-jobs/{job-id}")
	log.Println()
	log.Println("   4. Get results:")
	log.Println("      curl http://localhost:8080/scan-jobs/{job-id}/results")
	log.Println()
	log.Println("🎓 Scan Types: dns, whois, subdomain")
	log.Println("📚 See SCANNING_ARCHITECTURE.md for detailed docs")
	log.Println()
	log.Println("Press Ctrl+C to stop")

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

/*
🎓 TEACHING NOTES:

=== SESSION 4 HIGHLIGHTS ===

1. No Changes to Main Structure:
   - Same dependency injection pattern
   - Same layered architecture
   - Only enhancement: service now has validator

2. Key Differences from Session 3:
   Session 3:
     mux.HandleFunc("GET /assets", assetHandler.ListAssets)
     → Returns all assets, no pagination

   Session 4:
     mux.HandleFunc("GET /assets", assetHandler.ListAssets)
     → SAME CODE, but now supports:
       - Pagination (?page=1&page_size=20)
       - Filtering (?type=domain&status=active)
       - Search (?search=example)
       - Sorting (?sort_by=name&sort_order=asc)

3. Clean Architecture Benefit:
   - Main.go barely changed
   - All complexity in appropriate layers:
     * Validation → validator package
     * Query parsing → handler
     * Business logic → service
     * Database queries → storage

4. Startup Messages:
   - Show new features clearly
   - Provide example curl commands
   - Help students test immediately

CONFIGURATION:

Environment variables (same as Session 3):
- DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
- SERVER_PORT (default 8080)

Connection pool settings:
- MaxOpenConns: 25 (max concurrent connections)
- MaxIdleConns: 5 (keep idle for reuse)
- ConnMaxLifetime: 5 minutes

DEMO FLOW:

1. Start database:
   docker-compose up -d

2. Start server:
   go run cmd/server/main.go

3. Create some test data:
   curl -X POST http://localhost:8080/assets \
     -H "Content-Type: application/json" \
     -d '{"name":"example.com","type":"domain"}'

   curl -X POST http://localhost:8080/assets \
     -H "Content-Type: application/json" \
     -d '{"name":"test.com","type":"domain"}'

   curl -X POST http://localhost:8080/assets \
     -H "Content-Type: application/json" \
     -d '{"name":"192.168.1.1","type":"ip"}'

4. Test pagination:
   curl "http://localhost:8080/assets?page=1&page_size=2"
   → Should return 2 assets with pagination metadata

5. Test filtering:
   curl "http://localhost:8080/assets?type=domain"
   → Should return only domains

6. Test search:
   curl "http://localhost:8080/assets?search=example"
   → Should return only "example.com"

7. Test sorting:
   curl "http://localhost:8080/assets?sort_by=name&sort_order=asc"
   → Should return assets sorted alphabetically

8. Test validation (should fail):
   curl -X POST http://localhost:8080/assets \
     -H "Content-Type: application/json" \
     -d '{"name":"invalid..domain","type":"domain"}'
   → Should return 400 with validation error

9. Test combined query:
   curl "http://localhost:8080/assets?type=domain&search=example&page=1&sort_by=name"
   → All features working together!

COMPARISON WITH PREVIOUS SESSIONS:

Session 1: Basic HTTP server, health check
Session 2: CRUD API with in-memory storage
Session 3: PostgreSQL integration, persistence
Session 4: ⭐ Advanced features:
  - Pagination
  - Filtering
  - Search
  - Sorting
  - Comprehensive validation

ARCHITECTURE VISUALIZATION:

Request Flow:
1. HTTP Request with query params
   ↓
2. Handler parses params into QueryParams struct
   ↓
3. Service validates params
   ↓
4. Storage builds dynamic SQL query
   ↓
5. Database executes query
   ↓
6. Storage returns PaginatedResult
   ↓
7. Service returns to handler
   ↓
8. Handler sends JSON response

PRODUCTION CONSIDERATIONS:

1. Rate Limiting:
   - Add middleware to prevent abuse
   - Example: max 100 requests/minute per IP

2. Caching:
   - Cache common queries
   - Example: Redis for frequently accessed pages

3. Logging:
   - Log all queries with parameters
   - Monitor slow queries
   - Track validation failures (security)

4. Metrics:
   - API response times
   - Most common query patterns
   - Error rates

5. Security:
   - HTTPS in production
   - API key authentication
   - CORS configuration
   - Input sanitization (already done!)

HOMEWORK IDEAS:

1. Add more filters:
   - created_after, created_before (date range)
   - updated_recently (last 24 hours)

2. Add bulk operations:
   - POST /assets/bulk (create multiple)
   - DELETE /assets/bulk (delete by filter)

3. Add export functionality:
   - GET /assets/export?format=csv
   - GET /assets/export?format=json

4. Add statistics endpoint:
   - GET /assets/stats
   - Return count by type, status, etc.

5. Add field selection:
   - GET /assets?fields=id,name
   - Return only requested fields

NEXT SESSION PREVIEW:

Session 5: Testing
- Unit tests with mocks
- Integration tests
- Table-driven tests
- Coverage reports

All the validation we added makes testing much easier! 🎉
*/
