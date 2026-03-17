package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"himura-queue/internal/config"
	"himura-queue/internal/server"
)

func main() {
	configPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	srvCfg := server.Config{
		TCPPort:          cfg.Server.TCPPort,
		HTTPPort:         cfg.Server.HTTPPort,
		MinWorkers:       cfg.Worker.MinWorkers,
		MaxWorkers:       cfg.Worker.MaxWorkers,
		ShardCount:       cfg.Queue.ShardCount,
		SnapshotPath:     cfg.Snapshot.Path,
		SnapshotInterval: cfg.SnapshotInterval(),
	}

	srv, err := server.NewServer(srvCfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := srv.Stop(); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}
}
