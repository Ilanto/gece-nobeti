//go:build linux

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
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
	cfgPath := flag.String("config", "", "Path to config file")
	v := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *v {
		fmt.Printf("Linux Dashboard %s\n", version)
		return
	}

	// Load config
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Single instance — use PID file
	pidFile := filepath.Join(os.Getenv("HOME"), ".config", "linux-dashboard", "linux-dashboard.pid")
	if err := os.MkdirAll(filepath.Dir(pidFile), 0755); err != nil {
		log.Fatalf("Cannot create config dir: %v", err)
	}
	if _, err := os.Stat(pidFile); err == nil {
		log.Fatal("Linux Dashboard is already running. (PID file exists)")
	}
	f, err := os.Create(pidFile)
	if err != nil {
		log.Fatalf("Cannot create PID file: %v", err)
	}
	fmt.Fprintf(f, "%d", os.Getpid())
	f.Close()
	defer os.Remove(pidFile)

	// Protected processes from config
	protected := cfg.Controller.ProtectedProcesses
	ctrl := controller.NewController(protected)
	_ = ctrl

	// Event Emitter
	emitter := event.NewEmitter()

	// Collector Manager
	col := collector.NewManager(emitter)
	col.Start(cfg.Monitoring.Interval)

	// Storage (historyCap, procHistoryCap)
	store := storage.NewStore(300, 60)

	// Subscribe to snapshot events and forward to storage
	emitter.On("metrics.snapshot", func(data any) {
		if snap, ok := data.(*collector.SystemSnapshot); ok {
			store.SetLatest(snap)
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

	// Open browser
	if cfg.Server.OpenBrowser {
		url := fmt.Sprintf("http://%s", addr)
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()
	}

	log.Printf("Linux Dashboard starting on %s", addr)
	log.Printf("Config: %s", cfg.ConfigPath())

	// Wait for shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Listen(); err != nil {
			log.Printf("Server error: %v", err)
		}
		cancel()
	}()

	select {
	case s := <-sig:
		log.Printf("Received signal: %v — shutting down", s)
	case <-ctx.Done():
	}

	col.Stop()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	cmd.Run()
}