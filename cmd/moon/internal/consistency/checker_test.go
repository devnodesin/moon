package consistency

import (
	"context"
	"testing"
	"time"

	"github.com/thalib/moon/cmd/moon/internal/config"
	"github.com/thalib/moon/cmd/moon/internal/database"
	"github.com/thalib/moon/cmd/moon/internal/registry"
)

func setupTest(t *testing.T) (database.Driver, *registry.SchemaRegistry, func()) {
	t.Helper()

	cfg := database.Config{
		ConnectionString: "sqlite://:memory:",
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Minute * 5,
	}

	driver, err := database.NewDriver(cfg)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	ctx := context.Background()
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	reg := registry.NewSchemaRegistry()

	cleanup := func() {
		driver.Close()
	}

	return driver, reg, cleanup
}

func TestChecker_ConsistentState(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a table and register it
	_, err := driver.Exec(ctx, "CREATE TABLE users (ulid TEXT PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	err = reg.Set(&registry.Collection{
		Name: "users",
		Columns: []registry.Column{
			{Name: "name", Type: registry.TypeString},
			{Name: "email", Type: registry.TypeString},
		},
	})
	if err != nil {
		t.Fatalf("failed to register collection: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if !result.Consistent {
		t.Errorf("Expected consistent state, got inconsistent")
	}

	if len(result.Issues) != 0 {
		t.Errorf("Expected no issues, got %d", len(result.Issues))
	}
}

func TestChecker_OrphanedRegistry(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Register a collection without creating the table
	err := reg.Set(&registry.Collection{
		Name: "products",
		Columns: []registry.Column{
			{Name: "title", Type: registry.TypeString},
		},
	})
	if err != nil {
		t.Fatalf("failed to register collection: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.Type != IssueOrphanedRegistry {
		t.Errorf("Expected issue type %s, got %s", IssueOrphanedRegistry, issue.Type)
	}

	if issue.Name != "products" {
		t.Errorf("Expected issue name 'products', got '%s'", issue.Name)
	}

	if issue.Repaired {
		t.Errorf("Expected issue not repaired, but it was")
	}
}

func TestChecker_OrphanedRegistry_Repair(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Register a collection without creating the table
	err := reg.Set(&registry.Collection{
		Name: "products",
		Columns: []registry.Column{
			{Name: "title", Type: registry.TypeString},
		},
	})
	if err != nil {
		t.Fatalf("failed to register collection: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   true,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if !issue.Repaired {
		t.Errorf("Expected issue repaired, but it wasn't")
	}

	// Verify the collection was removed from registry
	if reg.Exists("products") {
		t.Errorf("Expected collection removed from registry, but it still exists")
	}
}

func TestChecker_OrphanedTable(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a table without registering it
	_, err := driver.Exec(ctx, "CREATE TABLE orders (ulid TEXT PRIMARY KEY, total REAL)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.Type != IssueOrphanedTable {
		t.Errorf("Expected issue type %s, got %s", IssueOrphanedTable, issue.Type)
	}

	if issue.Name != "orders" {
		t.Errorf("Expected issue name 'orders', got '%s'", issue.Name)
	}

	if issue.Repaired {
		t.Errorf("Expected issue not repaired, but it was")
	}
}

func TestChecker_OrphanedTable_RepairByRegistering(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a table without registering it
	_, err := driver.Exec(ctx, "CREATE TABLE orders (ulid TEXT PRIMARY KEY, total REAL)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   true,
		DropOrphans:  false, // Register instead of drop
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if !issue.Repaired {
		t.Errorf("Expected issue repaired, but it wasn't")
	}

	// Verify the table was registered
	if !reg.Exists("orders") {
		t.Errorf("Expected table registered, but it wasn't")
	}

	// Verify the registered schema
	collection, exists := reg.Get("orders")
	if !exists {
		t.Fatalf("Collection not found in registry")
	}

	// Should have 1 column (total) - ulid is skipped as it's the primary key
	if len(collection.Columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(collection.Columns))
	}
}

func TestChecker_OrphanedTable_RepairByDropping(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a table without registering it
	_, err := driver.Exec(ctx, "CREATE TABLE temp_table (ulid TEXT PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   true,
		DropOrphans:  true, // Drop orphaned tables
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if !issue.Repaired {
		t.Errorf("Expected issue repaired, but it wasn't")
	}

	// Verify the table was dropped
	exists, err := driver.TableExists(ctx, "temp_table")
	if err != nil {
		t.Fatalf("TableExists() error = %v", err)
	}

	if exists {
		t.Errorf("Expected table dropped, but it still exists")
	}
}

func TestChecker_MultipleIssues(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create orphaned table
	_, err := driver.Exec(ctx, "CREATE TABLE orphaned_table (ulid TEXT PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("failed to create orphaned table: %v", err)
	}

	// Create orphaned registry entry
	err = reg.Set(&registry.Collection{
		Name: "orphaned_registry",
		Columns: []registry.Column{
			{Name: "field", Type: registry.TypeString},
		},
	})
	if err != nil {
		t.Fatalf("failed to register orphaned collection: %v", err)
	}

	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)
	result, err := checker.Check(ctx)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Consistent {
		t.Errorf("Expected inconsistent state, got consistent")
	}

	if len(result.Issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(result.Issues))
	}

	// Check both issue types are present
	hasOrphanedTable := false
	hasOrphanedRegistry := false

	for _, issue := range result.Issues {
		if issue.Type == IssueOrphanedTable {
			hasOrphanedTable = true
		}
		if issue.Type == IssueOrphanedRegistry {
			hasOrphanedRegistry = true
		}
	}

	if !hasOrphanedTable {
		t.Errorf("Expected orphaned table issue")
	}

	if !hasOrphanedRegistry {
		t.Errorf("Expected orphaned registry issue")
	}
}

func TestChecker_Timeout(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	// Use a very short timeout
	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 0, // Immediate timeout
	}

	checker := NewChecker(driver, reg, cfg)

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := checker.Check(ctx)

	// Should timeout
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result != nil && !result.TimedOut {
		t.Errorf("Expected timed out result")
	}
}

func TestChecker_GetStatus(t *testing.T) {
	driver, reg, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	cfg := &config.RecoveryConfig{
		AutoRepair:   false,
		DropOrphans:  false,
		CheckTimeout: 5,
	}

	checker := NewChecker(driver, reg, cfg)

	// Test consistent state
	status := checker.GetStatus(ctx)
	if status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", status)
	}

	// Add an orphaned registry entry
	err := reg.Set(&registry.Collection{
		Name: "test",
		Columns: []registry.Column{
			{Name: "field", Type: registry.TypeString},
		},
	})
	if err != nil {
		t.Fatalf("failed to register collection: %v", err)
	}

	// Test inconsistent state
	status = checker.GetStatus(ctx)
	if status != "inconsistent" {
		t.Errorf("Expected status 'inconsistent', got '%s'", status)
	}
}
