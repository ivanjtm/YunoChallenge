package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	internalconfig "github.com/velamarket/refund-router/internal/config"
	"github.com/velamarket/refund-router/internal/handler"
	"github.com/velamarket/refund-router/internal/quota"
	"github.com/velamarket/refund-router/internal/router"
	"github.com/velamarket/refund-router/internal/testdata"
)

func main() {
	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Generate test data if it doesn't exist
	txnPath := "data/transactions.json"
	if _, err := os.Stat(txnPath); os.IsNotExist(err) {
		log.Println("Generating test transaction data...")
		if err := testdata.GenerateAndSave(txnPath, 200, time.Now()); err != nil {
			log.Fatalf("Failed to generate test data: %v", err)
		}
		log.Println("Generated 200 test transactions at", txnPath)
	}

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := internalconfig.LoadWithTransactions("config/processors.json", "config/rules.json", txnPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Loaded %d processors, %d rules, %d transactions", len(cfg.Processors), len(cfg.Rules), len(cfg.Transactions))

	// Create router engine
	routerEngine := router.NewRouter(cfg.Processors, cfg.Rules)

	// Create quota tracker
	quotaTracker := quota.NewTracker(cfg.Processors)

	// Create handlers
	healthH := &handler.HealthHandler{Config: cfg}
	refundH := &handler.RefundHandler{Router: routerEngine}
	batchH := &handler.BatchHandler{Router: routerEngine}
	quotaH := &handler.QuotaHandler{Tracker: quotaTracker}
	historicalH := &handler.HistoricalHandler{Router: routerEngine}

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", healthH.Handle)
	mux.HandleFunc("POST /api/v1/refund", refundH.Handle)
	mux.HandleFunc("POST /api/v1/refund/batch", batchH.Handle)
	mux.HandleFunc("POST /api/v1/simulation/quota", quotaH.Set)
	mux.HandleFunc("DELETE /api/v1/simulation/quota", quotaH.Reset)
	mux.HandleFunc("POST /api/v1/analysis/historical", historicalH.Handle)

	// Apply middleware
	srv := handler.Chain(mux,
		handler.RecoveryMiddleware,
		handler.LoggingMiddleware,
		handler.ContentTypeMiddleware,
	)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Refund Router Service on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  /api/v1/health")
	log.Printf("  POST /api/v1/refund")
	log.Printf("  POST /api/v1/refund/batch")
	log.Printf("  POST /api/v1/simulation/quota")
	log.Printf("  DELETE /api/v1/simulation/quota")
	log.Printf("  POST /api/v1/analysis/historical")

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
