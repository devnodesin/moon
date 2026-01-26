package main

import (
	"fmt"
	"os"

	"github.com/thalib/moon/internal/config"
)

func main() {
	fmt.Println("Moon - Dynamic Headless Engine")
	fmt.Println("Starting server...")

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server will start on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s\n", cfg.Database.ConnectionString)

	// TODO: Initialize database, router, and start server
	os.Exit(0)
}
