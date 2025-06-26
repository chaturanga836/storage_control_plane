// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/duck"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/logger"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/registry"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/router"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/shutdown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go shutdown.HandleShutdown(cancel)

	// ✅ Initialize DuckDB
	if err := duck.InitDuckDB("data/duck.db"); err != nil {
		log.Fatalf("❌ DuckDB init failed: %v", err)
	}

	// ✅ Load users + tenants into in-memory registries
	registry.LoadUserRegistry()
	go registry.WatchUserFileChanges(ctx)

	registry.LoadTenantRegistry() // ✅ LOAD TENANTS
	go registry.WatchTenantFileChanges(ctx) // ✅ WATCH TENANTS

	// ✅ Setup router
	r := router.SetupRoutes()

	// ✅ Setup CORS middleware for frontend integration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(logger.LogRequest(r))

	log.Println("🚀 Starting Go Storage Control Plane on port 8081")

	if err := http.ListenAndServe(":8081", corsHandler); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

