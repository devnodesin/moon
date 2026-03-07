package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("c", DefaultConfigPath, "path to the YAML configuration file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("moon: listening on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
