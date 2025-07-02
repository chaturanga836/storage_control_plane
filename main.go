package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// ServiceConfig represents configuration for the entire system
type ServiceConfig struct {
	Services map[string]ServiceInfo `json:"services"`
	Database DatabaseConfig         `json:"database"`
	Redis    RedisConfig           `json:"redis"`
}

type ServiceInfo struct {
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Enabled  bool   `json:"enabled"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all services
	serviceManager := NewServiceManager(config)
	
	if err := serviceManager.StartAll(ctx); err != nil {
		log.Fatalf("Failed to start services: %v", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown(cancel, serviceManager)

	log.Println("🚀 Storage Control Plane started successfully!")
	log.Println("📊 Health check: http://localhost:8090/health")
	
	// Wait for context cancellation
	<-ctx.Done()
	log.Println("👋 Storage Control Plane shutting down...")
}

func loadConfig() (*ServiceConfig, error) {
	config := &ServiceConfig{
		Services: map[string]ServiceInfo{
			"auth_gateway": {
				Port:    8090,
				Name:    "Auth Gateway",
				Version: "1.0.0",
				Enabled: true,
			},
			"tenant_node": {
				Port:    8000,
				Name:    "Tenant Node",
				Version: "1.0.0",
				Enabled: true,
			},
			"operation_node": {
				Port:    8081,
				Name:    "Operation Node",
				Version: "1.0.0",
				Enabled: true,
			},
			"cbo_engine": {
				Port:    8082,
				Name:    "Cost-Based Optimizer",
				Version: "1.0.0",
				Enabled: true,
			},
			"metadata_catalog": {
				Port:    8083,
				Name:    "Metadata Catalog",
				Version: "1.0.0",
				Enabled: true,
			},
			"monitoring": {
				Port:    8084,
				Name:    "Monitoring",
				Version: "1.0.0",
				Enabled: true,
			},
			"query_interpreter": {
				Port:    8085,
				Name:    "Query Interpreter",
				Version: "1.0.0",
				Enabled: true,
			},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Database: getEnv("DB_NAME", "storage_control"),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}

	return config, nil
}

func setupGracefulShutdown(cancel context.CancelFunc, serviceManager *ServiceManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Received shutdown signal...")
		
		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := serviceManager.StopAll(shutdownCtx); err != nil {
			log.Printf("❌ Error during shutdown: %v", err)
		}
		
		cancel()
	}()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

// ServiceManager manages all microservices in the monolith
type ServiceManager struct {
	config   *ServiceConfig
	services map[string]*http.Server
	mu       sync.RWMutex
}

func NewServiceManager(config *ServiceConfig) *ServiceManager {
	return &ServiceManager{
		config:   config,
		services: make(map[string]*http.Server),
	}
}

func (sm *ServiceManager) StartAll(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for serviceName, serviceInfo := range sm.config.Services {
		if !serviceInfo.Enabled {
			log.Printf("⏭️  Skipping disabled service: %s", serviceName)
			continue
		}

		mux := http.NewServeMux()
		
		// Setup routes for each service
		switch serviceName {
		case "auth_gateway":
			setupAuthGatewayRoutes(mux)
		case "tenant_node":
			setupTenantNodeRoutes(mux)
		case "operation_node":
			setupOperationNodeRoutes(mux)
		case "cbo_engine":
			setupCBOEngineRoutes(mux)
		case "metadata_catalog":
			setupMetadataCatalogRoutes(mux)
		case "monitoring":
			setupMonitoringRoutes(mux)
		case "query_interpreter":
			setupQueryInterpreterRoutes(mux)
		}

		// Add common routes
		setupCommonRoutes(mux, serviceInfo)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", serviceInfo.Port),
			Handler: mux,
		}

		sm.services[serviceName] = server

		// Start server in goroutine
		go func(name string, srv *http.Server) {
			log.Printf("🚀 Starting %s on port %s", serviceInfo.Name, srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("❌ %s failed: %v", name, err)
			}
		}(serviceName, server)
	}

	return nil
}

func (sm *ServiceManager) StopAll(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var wg sync.WaitGroup
	for serviceName, server := range sm.services {
		wg.Add(1)
		go func(name string, srv *http.Server) {
			defer wg.Done()
			log.Printf("🛑 Stopping %s...", name)
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("❌ Error stopping %s: %v", name, err)
			} else {
				log.Printf("✅ %s stopped gracefully", name)
			}
		}(serviceName, server)
	}

	wg.Wait()
	return nil
}

func setupCommonRoutes(mux *http.ServeMux, serviceInfo ServiceInfo) {
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "healthy",
			"service": serviceInfo.Name,
			"version": serviceInfo.Version,
			"time":    time.Now().UTC(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service": serviceInfo.Name,
			"version": serviceInfo.Version,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}
