package main

import (
"context"
"flag"
"fmt"
"os"

"github.com/thalib/moon/cmd/moon/internal/auth"
"github.com/thalib/moon/cmd/moon/internal/database"
)

func main() {
dbPath := flag.String("db", "/tmp/moon_test/data/sqlite.db", "Path to SQLite database")
username := flag.String("username", "admin", "Admin username")
email := flag.String("email", "admin@example.com", "Admin email")
password := flag.String("password", "moonadmin12#", "Admin password")
flag.Parse()

ctx := context.Background()

// Connect to database
dbConfig := database.Config{
ConnectionString: fmt.Sprintf("sqlite://%s", *dbPath),
}

driver, err := database.NewDriver(dbConfig)
if err != nil {
fmt.Fprintf(os.Stderr, "Failed to create database driver: %v\n", err)
os.Exit(1)
}

if err := driver.Connect(ctx); err != nil {
fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
os.Exit(1)
}
defer driver.Close()

// Bootstrap auth with admin user
bootstrapCfg := &auth.BootstrapConfig{
Username: *username,
Email:    *email,
Password: *password,
}

if err := auth.Bootstrap(ctx, driver, bootstrapCfg); err != nil {
fmt.Fprintf(os.Stderr, "Failed to bootstrap: %v\n", err)
os.Exit(1)
}

fmt.Println("Bootstrap completed successfully")
}
