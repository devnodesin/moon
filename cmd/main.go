package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	configPath := flag.String("c", DefaultConfigPath, "path to the YAML configuration file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}

	logger, err := InitLogger(cfg.Server.Logpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	adapter, err := NewDatabaseAdapter(cfg.Database, logger)
	if err != nil {
		logger.Error("database init failed", "error", err)
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}
	defer adapter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := adapter.Ping(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		logger.Error("system tables init failed", "error", err)
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}

	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		logger.Error("bootstrap admin init failed", "error", err)
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("moon: listening on %s:%d\n", cfg.Server.Host, cfg.Server.Port)

	if err := StartServer(cfg, logger, adapter); err != nil {
		logger.Error("server error", "error", err)
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
