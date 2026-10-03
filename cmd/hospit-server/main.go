package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/connectors/bookingcom"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/internal/api"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/internal/orchestrator"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
)

func main() {
	log.Println("==================================================================")
	log.Println("  Hospit: Open Booking Integration Platform (Golang Engine)")
	log.Println("  Multi-Threaded Partner Aggregation & Real-Time ARI Sync")
	log.Println("==================================================================")

	// 1. Initialize Pluggable Connector Registry
	registry := connector.NewRegistry()

	// 2. Register Booking.com Connector (Demand API v3 & Connectivity Switch)
	bkConfig := bookingcom.Config{
		APIKey:      os.Getenv("BOOKINGCOM_API_KEY"),
		AffiliateID: os.Getenv("BOOKINGCOM_AFFILIATE_ID"),
		Sandbox:     os.Getenv("BOOKINGCOM_SANDBOX") != "false", // Default to sandbox/mock mode if not set
		Timeout:     5 * time.Second,
	}
	bkConnector := bookingcom.NewConnector(bkConfig)
	if err := registry.Register(bkConnector); err != nil {
		log.Fatalf("Failed to register Booking.com connector: %v", err)
	}
	log.Printf("[Registry] Registered supplier connector: %s (%s)", bkConnector.Name(), bkConnector.ID())

	// 3. Initialize Multi-Threaded Search Orchestrator (2.5s fan-out deadline)
	orch := orchestrator.NewOrchestrator(registry, 2500*time.Millisecond)

	// 4. Resolve Port
	port := 4000
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	// 5. Start HTTP API Server
	server := api.NewServer(port, orch, registry)

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	<-stopChan
	log.Println("\n[Hospit Engine] Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Hospit Engine] Shutdown error: %v", err)
	}

	log.Println("[Hospit Engine] Stopped.")
}
