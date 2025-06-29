package main

import (
	"log"
	"net/http"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/api"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/config"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/routing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	router := routing.NewRouter(cfg)
	server := api.NewServer(router)

	addr := cfg.ServerAddress
	log.Printf("Starting API server at %s...", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server: %v", err)
	}
}