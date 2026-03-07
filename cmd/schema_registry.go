package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// Moon field type constants
// ---------------------------------------------------------------------------

const (
	MoonFieldTypeID       = "id"
	MoonFieldTypeString   = "string"
	MoonFieldTypeInteger  = "integer"
	MoonFieldTypeDecimal  = "decimal"
	MoonFieldTypeBoolean  = "boolean"
	MoonFieldTypeDatetime = "datetime"
	MoonFieldTypeJSON     = "json"
)

// ---------------------------------------------------------------------------
// Collection and field naming constraints (SPEC §9.5)
// ---------------------------------------------------------------------------

const (
	MinCollectionNameLen = 2
	MaxCollectionNameLen = 63
	MinFieldNameLen      = 3
	MaxFieldNameLen      = 63
)

// namePattern matches lowercase snake_case identifiers starting with a letter.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// reservedCollectionNames are exact names that cannot be used for dynamic
// collections. System collections (users, apikeys) are reserved but still
// appear in the registry for read access.
var reservedCollectionNames = map[string]bool{
	"users":       true,
	"apikeys":     true,
	"collections": true,
	"auth":        true,
	"doc":         true,
	"health":      true,
}

// sqlReservedKeywords lists SQL keywords that cannot be used as collection
// or field names. Based on the SQLite reserved word list plus common SQL
// keywords shared across backends.
var sqlReservedKeywords = map[string]bool{
	"abort": true, "action": true, "add": true, "after": true, "all": true,
	"alter": true, "always": true, "analyze": true, "and": true, "as": true,
	"asc": true, "attach": true, "autoincrement": true,
	"before": true, "begin": true, "between": true, "by": true,
	"cascade": true, "case": true, "cast": true, "check": true,
	"collate": true, "column": true, "commit": true, "conflict": true,
	"constraint": true, "create": true, "cross": true, "current": true,
	"current_date": true, "current_time": true, "current_timestamp": true,
	"database": true, "default": true, "deferrable": true, "deferred": true,
	"delete": true, "desc": true, "detach": true, "distinct": true,
	"do": true, "drop": true,
	"each": true, "else": true, "end": true, "escape": true, "except": true,
	"exclude": true, "exclusive": true, "exists": true, "explain": true,
	"fail": true, "filter": true, "first": true, "following": true,
	"for": true, "foreign": true, "from": true, "full": true,
	"generated": true, "glob": true, "group": true, "groups": true,
	"having": true,
	"if":     true, "ignore": true, "immediate": true, "in": true,
	"index": true, "indexed": true, "initially": true, "inner": true,
	"insert": true, "instead": true, "intersect": true, "into": true,
	"is": true, "isnull": true,
	"join": true, "key": true,
	"last": true, "left": true, "like": true, "limit": true,
	"match": true, "materialized": true,
	"natural": true, "no": true, "not": true, "nothing": true,
	"notnull": true, "null": true, "nulls": true,
	"of": true, "offset": true, "on": true, "or": true, "order": true,
	"others": true, "outer": true, "over": true,
	"partition": true, "plan": true, "pragma": true, "preceding": true,
	"primary": true,
	"query":   true,
	"raise":   true, "range": true, "recursive": true, "references": true,
	"regexp": true, "reindex": true, "release": true, "rename": true,
	"replace": true, "restrict": true, "returning": true, "right": true,
	"rollback": true, "row": true, "rows": true,
	"savepoint": true, "select": true, "set": true,
	"table": true, "temp": true, "temporary": true, "then": true,
	"ties": true, "to": true, "transaction": true, "trigger": true,
	"unbounded": true, "union": true, "unique": true, "update": true,
	"using":  true,
	"vacuum": true, "values": true, "view": true, "virtual": true,
	"when": true, "where": true, "window": true, "with": true,
	"without": true,
}

// ---------------------------------------------------------------------------
// System collection metadata
// ---------------------------------------------------------------------------

// systemReadOnlyFields maps system collection names to fields that are
// server-owned and must not be modified by API clients.
var systemReadOnlyFields = map[string]map[string]bool{
	"users": {
		"id": true, "password_hash": true,
		"created_at": true, "updated_at": true, "last_login_at": true,
	},
	"apikeys": {
		"id": true, "key_hash": true,
		"created_at": true, "updated_at": true, "last_used_at": true,
	},
}

// hiddenSystemFields maps system collection names to fields that must
// not appear in API schema responses.
var hiddenSystemFields = map[string]map[string]bool{
	"users":   {"password_hash": true},
	"apikeys": {"key_hash": true},
}

// ---------------------------------------------------------------------------
// Physical-to-Moon type mapping
// ---------------------------------------------------------------------------

// physicalTypeMap maps uppercase SQL type strings to Moon field types.
// Covers SQLite, PostgreSQL, and MySQL declared types.
var physicalTypeMap = map[string]string{
	"TEXT":      MoonFieldTypeString,
	"INTEGER":   MoonFieldTypeInteger,
	"BIGINT":    MoonFieldTypeInteger,
	"NUMERIC":   MoonFieldTypeDecimal,
	"BOOLEAN":   MoonFieldTypeBoolean,
	"JSON":      MoonFieldTypeJSON,
	"JSONB":     MoonFieldTypeJSON,
	"TIMESTAMP": MoonFieldTypeDatetime,
}

// ---------------------------------------------------------------------------
// Collection and Field types
// ---------------------------------------------------------------------------

// Collection represents an API-visible collection schema.
type Collection struct {
	Name   string
	Fields []Field
}

// APIFields returns only fields that should be visible in API schema
// responses, filtering out hidden system fields.
func (c *Collection) APIFields() []Field {
	hidden := hiddenSystemFields[c.Name]
	if len(hidden) == 0 {
		result := make([]Field, len(c.Fields))
		copy(result, c.Fields)
		return result
	}
	result := make([]Field, 0, len(c.Fields))
	for _, f := range c.Fields {
		if !hidden[f.Name] {
			result = append(result, f)
		}
	}
	return result
}

// Field represents a single field descriptor in a collection.
type Field struct {
	Name     string
	Type     string
	Nullable bool
	Unique   bool
	ReadOnly bool
}

// ---------------------------------------------------------------------------
// SchemaRegistry
// ---------------------------------------------------------------------------

// SchemaRegistry is the in-memory, concurrency-safe store of all API-visible
// collection schemas. Every validation and query-planning operation reads
// from it.
type SchemaRegistry struct {
	mu          sync.RWMutex
	collections map[string]*Collection
	order       []string // sorted collection names for stable iteration
	db          DatabaseAdapter
}

// NewSchemaRegistry creates a new registry and populates it from the
// physical database schema. Returns an error if any API-visible table
// contains unmappable columns.
func NewSchemaRegistry(db DatabaseAdapter) (*SchemaRegistry, error) {
	r := &SchemaRegistry{
		collections: make(map[string]*Collection),
		db:          db,
	}
	if err := r.populate(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

// Get returns a single collection by name. The second return value
// indicates whether the collection exists.
func (r *SchemaRegistry) Get(name string) (*Collection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.collections[name]
	return c, ok
}

// List returns all API-visible collections in alphabetical order.
func (r *SchemaRegistry) List() []*Collection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Collection, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.collections[name])
	}
	return result
}

// Refresh rebuilds the registry from the physical database schema.
// If the rebuild fails, the previous state is preserved and the error
// is returned. Readers always see either the old or the new complete
// state, never a partial update.
func (r *SchemaRegistry) Refresh() error {
	newCollections, newOrder, err := r.buildFromDB(context.Background())
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.collections = newCollections
	r.order = newOrder
	r.mu.Unlock()
	return nil
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

// IsValidCollectionName validates a name for use as a new dynamic
// collection. It checks length, pattern, reserved names, the moon_
// prefix, and SQL keywords.
func IsValidCollectionName(name string) bool {
	n := len(name)
	if n < MinCollectionNameLen || n > MaxCollectionNameLen {
		return false
	}
	if !namePattern.MatchString(name) {
		return false
	}
	if strings.HasPrefix(name, "moon_") {
		return false
	}
	if reservedCollectionNames[name] {
		return false
	}
	if sqlReservedKeywords[name] {
		return false
	}
	return true
}

// IsValidFieldName validates a name for use as a collection field.
// It checks length, pattern, and SQL keywords.
func IsValidFieldName(name string) bool {
	n := len(name)
	if n < MinFieldNameLen || n > MaxFieldNameLen {
		return false
	}
	if !namePattern.MatchString(name) {
		return false
	}
	if sqlReservedKeywords[name] {
		return false
	}
	return true
}

// IsSQLKeyword returns true if name is a SQL reserved keyword.
func IsSQLKeyword(name string) bool {
	return sqlReservedKeywords[strings.ToLower(name)]
}

// ---------------------------------------------------------------------------
// Internal population logic
// ---------------------------------------------------------------------------

// populate performs the initial registry build without holding the lock.
// Called only from NewSchemaRegistry before the registry is shared.
func (r *SchemaRegistry) populate(ctx context.Context) error {
	collections, order, err := r.buildFromDB(ctx)
	if err != nil {
		return err
	}
	r.collections = collections
	r.order = order
	return nil
}

// buildFromDB discovers all API-visible tables, maps their columns to
// Moon field types, and returns the new registry state. It does not
// modify the registry.
func (r *SchemaRegistry) buildFromDB(ctx context.Context) (map[string]*Collection, []string, error) {
	tables, err := r.db.ListTables(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("schema registry: list tables: %w", err)
	}

	collections := make(map[string]*Collection)
	var order []string

	for _, table := range tables {
		if strings.HasPrefix(table, "moon_") {
			continue
		}
		if !matchesCollectionPattern(table) {
			continue
		}

		columns, err := r.db.DescribeTable(ctx, table)
		if err != nil {
			return nil, nil, fmt.Errorf("schema registry: describe %q: %w", table, err)
		}

		fields, err := mapColumnsToFields(table, columns)
		if err != nil {
			return nil, nil, err
		}

		fields = ensureIDFirst(fields)
		collections[table] = &Collection{Name: table, Fields: fields}
		order = append(order, table)
	}

	sort.Strings(order)
	return collections, order, nil
}

// matchesCollectionPattern checks whether a table name matches the
// naming pattern for API-visible collections (length + snake_case).
// It does NOT check reserved names or SQL keywords, because system
// collections like "users" and "apikeys" are valid in the registry.
func matchesCollectionPattern(name string) bool {
	n := len(name)
	if n < MinCollectionNameLen || n > MaxCollectionNameLen {
		return false
	}
	return namePattern.MatchString(name)
}

// mapColumnsToFields converts database columns to Moon fields for the
// given table. Returns an error if any column has an unmappable type.
func mapColumnsToFields(table string, columns []ColumnInfo) ([]Field, error) {
	fields := make([]Field, 0, len(columns))
	for _, col := range columns {
		moonType, err := physicalToMoonType(col)
		if err != nil {
			return nil, fmt.Errorf("table %q column %q: %w", table, col.Name, err)
		}
		field := Field{
			Name:     col.Name,
			Type:     moonType,
			Nullable: col.Nullable,
			Unique:   col.Unique,
			ReadOnly: isReadOnlyField(table, col.Name, col.PK),
		}
		fields = append(fields, field)
	}
	return fields, nil
}

// physicalToMoonType maps a physical column to a Moon field type.
// The id column (PK named "id") is mapped to the "id" type.
// All other columns are mapped by their declared SQL type.
func physicalToMoonType(col ColumnInfo) (string, error) {
	if col.PK && col.Name == "id" {
		return MoonFieldTypeID, nil
	}

	upper := strings.ToUpper(col.Type)

	// Direct lookup first.
	if moonType, ok := physicalTypeMap[upper]; ok {
		return moonType, nil
	}

	// Handle parameterized types like NUMERIC(19,2) or DECIMAL(19,2).
	if strings.HasPrefix(upper, "NUMERIC(") || strings.HasPrefix(upper, "DECIMAL(") {
		return MoonFieldTypeDecimal, nil
	}

	// Handle TIMESTAMP WITH TIME ZONE and similar variants.
	if strings.HasPrefix(upper, "TIMESTAMP") {
		return MoonFieldTypeDatetime, nil
	}

	return "", fmt.Errorf("unmappable SQL type %q", col.Type)
}

// isReadOnlyField determines whether a field should be marked as read-only.
// The id field is always read-only. For system collections, additional
// server-owned fields are also read-only.
func isReadOnlyField(table, field string, isPK bool) bool {
	if isPK && field == "id" {
		return true
	}
	if sysFields, ok := systemReadOnlyFields[table]; ok {
		return sysFields[field]
	}
	return false
}

// ensureIDFirst returns a new slice with the "id" field moved to index 0,
// preserving the relative order of all other fields.
func ensureIDFirst(fields []Field) []Field {
	idIdx := -1
	for i, f := range fields {
		if f.Name == "id" {
			idIdx = i
			break
		}
	}
	if idIdx <= 0 {
		return fields
	}
	result := make([]Field, 0, len(fields))
	result = append(result, fields[idIdx])
	result = append(result, fields[:idIdx]...)
	result = append(result, fields[idIdx+1:]...)
	return result
}
