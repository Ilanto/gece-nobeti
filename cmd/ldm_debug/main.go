package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"time"

	"github.com/burak/linux-dashboard/internal/ai"
	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/config"
	"github.com/burak/linux-dashboard/internal/controller"
	"github.com/burak/linux-dashboard/internal/event"
	"github.com/burak/linux-dashboard/internal/server"
	"github.com/burak/linux-dashboard/internal/storage"
)

var version = "dev"

func main() {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Config loaded. Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Monitoring interval: %v\n", cfg.Monitoring.Interval)

	// Protected processes from config
	protected := cfg.Controller.ProtectedProcesses
	ctrl := controller.NewController(protected)
	_ = ctrl

	// Event Emitter
	emitter := event.NewEmitter()
	fmt.Println("Emitter created")

	// Collector Manager
	col := collector.NewManager(emitter)
	fmt.Println("Collector manager created")

	// Start with config interval
	col.Start(cfg.Monitoring.Interval)
	fmt.Printf("Collector started with interval %v\n", cfg.Monitoring.Interval)

	// Storage (historyCap, procHistoryCap)
	store := storage.NewStore(300, 60)
	fmt.Println("Store created")

	// Subscribe to snapshot events and forward to storage
	emitter.On("metrics.snapshot", func(data any) {
		fmt.Printf("[EVENT] Got snapshot event, data type: %T\n", data)
		if snap, ok := data.(*collector.SystemSnapshot); ok {
			fmt.Printf("[EVENT] Process count: %d\n", len(snap.Processes))
			store.SetLatest(snap)
		} else {
			fmt.Printf("[EVENT] Unexpected data type: %T\n", data)
		}
	})

	// AI Advisor — only if enabled
	var advisor *ai.Advisor
	if cfg.AI.Enabled {
		aiCfg := ai.Config{
			Provider:    cfg.AI.Provider,
			APIKey:      cfg.AI.APIKey,
			Model:       cfg.AI.Model,
			Endpoint:    cfg.AI.Endpoint,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
		advisor = ai.NewAdvisor(aiCfg)
		log.Printf("AI Advisor enabled: %s / %s", cfg.AI.Provider, cfg.AI.Model)
	}

	// HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := server.New(addr, col, store, advisor, emitter)

	log.Printf("Linux Dashboard starting on %s", addr)

	// Run server in goroutine
	go func() {
		fmt.Println("Starting HTTP server...")
		if err := srv.Listen(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit and check if snapshot is collected
	fmt.Println("Waiting 5 seconds for first collection...")
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		snap := col.LatestSnapshot()
		if snap != nil {
			fmt.Printf("[CHECK] After %d seconds: snap NOT nil, procs=%d\n", i+1, len(snap.Processes))
		} else {
			fmt.Printf("[CHECK] After %d seconds: snap is nil\n", i+1)
		}
	}

	// Block forever
	select {}
}