package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// CollectionHandler implements GET /collections:query and POST /collections:mutate.
type CollectionHandler struct {
	db       DatabaseAdapter
	registry *SchemaRegistry
	cfg      *AppConfig
	prefix   string
}

// NewCollectionHandler creates a CollectionHandler with the given dependencies.
func NewCollectionHandler(db DatabaseAdapter, registry *SchemaRegistry, cfg *AppConfig) *CollectionHandler {
	return &CollectionHandler{
		db:       db,
		registry: registry,
		cfg:      cfg,
		prefix:   strings.TrimRight(cfg.Server.Prefix, "/"),
	}
}

// ---------------------------------------------------------------------------
// GET /collections:query
// ---------------------------------------------------------------------------

// HandleQuery dispatches list-mode and get-one-mode collection queries.
func (h *CollectionHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name != "" {
		h.handleGetOne(w, r, name)
		return
	}
	h.handleList(w, r)
}

func (h *CollectionHandler) handleGetOne(w http.ResponseWriter, _ *http.Request, name string) {
	if strings.HasPrefix(name, "moon_") {
		WriteError(w, http.StatusBadRequest, "Collection name is reserved")
		return
	}

	col, ok := h.registry.Get(name)
	if !ok {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("Collection '%s' not found", name))
		return
	}

	count, err := h.db.CountRows(context.Background(), col.Name)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	item := map[string]any{"name": col.Name, "count": count, "system": col.System}
	WriteSuccess(w, http.StatusOK, "Collection retrieved successfully", []any{item})
}

func (h *CollectionHandler) handleList(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	allCollections := h.registry.List()
	total := len(allCollections)
	totalPages := 1
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageItems := allCollections[start:end]

	data := make([]any, 0, len(pageItems))
	for _, col := range pageItems {
		count, err := h.db.CountRows(context.Background(), col.Name)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		data = append(data, map[string]any{"name": col.Name, "count": count, "system": col.System})
	}

	basePath := h.prefix + "/collections:query"
	meta := map[string]any{
		"total":        total,
		"count":        len(data),
		"per_page":     perPage,
		"current_page": page,
		"total_pages":  totalPages,
	}
	links := buildPaginationLinks(basePath, page, perPage, totalPages)

	WriteSuccessFull(w, http.StatusOK, "Collections retrieved successfully", data, meta, links)
}

// ---------------------------------------------------------------------------
// POST /collections:mutate
// ---------------------------------------------------------------------------

// collectionMutateRequest is the JSON body for POST /collections:mutate.
type collectionMutateRequest struct {
	Op   string            `json:"op"`
	Data []json.RawMessage `json:"data"`
}

// collectionCreateItem is a single item in op=create.
type collectionCreateItem struct {
	Name    string             `json:"name"`
	Columns []collectionColumn `json:"columns"`
}

// collectionColumn is a column definition for create/add_columns.
type collectionColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable *bool  `json:"nullable,omitempty"`
	Unique   *bool  `json:"unique,omitempty"`
}

// collectionUpdateItem is a single item in op=update.
type collectionUpdateItem struct {
	Name          string             `json:"name"`
	AddColumns    []collectionColumn `json:"add_columns,omitempty"`
	RenameColumns []renameColumnSpec `json:"rename_columns,omitempty"`
	ModifyColumns []collectionColumn `json:"modify_columns,omitempty"`
	RemoveColumns []string           `json:"remove_columns,omitempty"`
}

// renameColumnSpec specifies a column rename.
type renameColumnSpec struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

// collectionDestroyItem is a single item in op=destroy.
type collectionDestroyItem struct {
	Name string `json:"name"`
}

// HandleMutate dispatches collection mutation operations.
func (h *CollectionHandler) HandleMutate(w http.ResponseWriter, r *http.Request) {
	identity, ok := GetAuthIdentity(r.Context())
	if !ok || identity.Role != "admin" {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req collectionMutateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	switch req.Op {
	case "create":
		h.handleCreate(w, req.Data)
	case "update":
		h.handleUpdate(w, req.Data)
	case "destroy":
		h.handleDestroy(w, req.Data)
	default:
		WriteError(w, http.StatusBadRequest, "Invalid operation")
		return
	}
}

// ---------------------------------------------------------------------------
// op=create
// ---------------------------------------------------------------------------

func (h *CollectionHandler) handleCreate(w http.ResponseWriter, rawItems []json.RawMessage) {
	if len(rawItems) == 0 {
		WriteError(w, http.StatusBadRequest, "Data must not be empty")
		return
	}

	var results []any
	for _, raw := range rawItems {
		var item collectionCreateItem
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid create item")
			return
		}

		if err := h.validateCreateItem(item); err != nil {
			writeCollectionError(w, err)
			return
		}

		ddl := h.buildCreateDDL(item)
		if err := h.db.ExecDDL(context.Background(), ddl); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if err := h.registry.Refresh(); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		cols := make([]map[string]any, 0, len(item.Columns))
		for _, c := range item.Columns {
			cols = append(cols, map[string]any{
				"name":     c.Name,
				"type":     c.Type,
				"nullable": boolVal(c.Nullable, false),
				"unique":   boolVal(c.Unique, false),
			})
		}
		results = append(results, map[string]any{
			"name":    item.Name,
			"columns": cols,
		})
	}

	meta := map[string]any{"success": len(results), "failed": 0}
	WriteSuccessFull(w, http.StatusCreated, "Collection created successfully", results, meta, nil)
}

func (h *CollectionHandler) validateCreateItem(item collectionCreateItem) *collectionError {
	if item.Name == "" {
		return &collectionError{Status: http.StatusBadRequest, Message: "Collection name is required"}
	}

	if strings.HasPrefix(item.Name, "moon_") {
		return &collectionError{Status: http.StatusBadRequest, Message: "Collection name is reserved"}
	}

	if item.Name == "users" || item.Name == "apikeys" {
		return &collectionError{Status: http.StatusForbidden, Message: "Forbidden"}
	}

	if !IsValidCollectionName(item.Name) {
		return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid collection name %q", item.Name)}
	}

	if _, exists := h.registry.Get(item.Name); exists {
		return &collectionError{Status: http.StatusConflict, Message: fmt.Sprintf("Collection '%s' already exists", item.Name)}
	}

	if len(item.Columns) == 0 {
		return &collectionError{Status: http.StatusBadRequest, Message: "Columns must not be empty"}
	}

	seen := make(map[string]bool)
	for _, col := range item.Columns {
		if col.Name == "id" {
			return &collectionError{Status: http.StatusBadRequest, Message: "Column 'id' is managed by the server"}
		}
		if !IsValidFieldName(col.Name) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column name %q", col.Name)}
		}
		if !isValidMoonType(col.Type) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column type %q", col.Type)}
		}
		if seen[col.Name] {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Duplicate column name %q", col.Name)}
		}
		seen[col.Name] = true
	}
	return nil
}

func (h *CollectionHandler) buildCreateDDL(item collectionCreateItem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (%s TEXT PRIMARY KEY", quoteIdent(item.Name), quoteIdent("id")))

	for _, col := range item.Columns {
		sb.WriteString(", ")
		sb.WriteString(quoteIdent(col.Name))
		sb.WriteString(" ")
		sb.WriteString(moonTypeToSQLite(col.Type))
		if !boolVal(col.Nullable, false) {
			sb.WriteString(" NOT NULL")
		}
		if boolVal(col.Unique, false) {
			sb.WriteString(" UNIQUE")
		}
	}
	sb.WriteString(")")
	return sb.String()
}

// ---------------------------------------------------------------------------
// op=update
// ---------------------------------------------------------------------------

func (h *CollectionHandler) handleUpdate(w http.ResponseWriter, rawItems []json.RawMessage) {
	if len(rawItems) == 0 {
		WriteError(w, http.StatusBadRequest, "Data must not be empty")
		return
	}

	var results []any
	for _, raw := range rawItems {
		var item collectionUpdateItem
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid update item")
			return
		}

		if err := h.validateUpdateItem(item); err != nil {
			writeCollectionError(w, err)
			return
		}

		if err := h.executeUpdate(item); err != nil {
			writeCollectionError(w, err)
			return
		}

		if err := h.registry.Refresh(); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		col, ok := h.registry.Get(item.Name)
		if !ok {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		apiFields := col.APIFields()
		cols := make([]map[string]any, 0, len(apiFields))
		for _, f := range apiFields {
			if f.Name == "id" {
				continue
			}
			cols = append(cols, map[string]any{
				"name":     f.Name,
				"type":     f.Type,
				"nullable": f.Nullable,
				"unique":   f.Unique,
			})
		}
		results = append(results, map[string]any{
			"name":    item.Name,
			"columns": cols,
		})
	}

	meta := map[string]any{"success": len(results), "failed": 0}
	WriteSuccessFull(w, http.StatusOK, "Collection updated successfully", results, meta, nil)
}

func (h *CollectionHandler) validateUpdateItem(item collectionUpdateItem) *collectionError {
	if item.Name == "" {
		return &collectionError{Status: http.StatusBadRequest, Message: "Collection name is required"}
	}
	if strings.HasPrefix(item.Name, "moon_") {
		return &collectionError{Status: http.StatusBadRequest, Message: "Collection name is reserved"}
	}
	if item.Name == "users" || item.Name == "apikeys" {
		return &collectionError{Status: http.StatusForbidden, Message: "Forbidden"}
	}

	if _, exists := h.registry.Get(item.Name); !exists {
		return &collectionError{Status: http.StatusNotFound, Message: fmt.Sprintf("Collection '%s' not found", item.Name)}
	}

	opCount := 0
	if len(item.AddColumns) > 0 {
		opCount++
	}
	if len(item.RenameColumns) > 0 {
		opCount++
	}
	if len(item.ModifyColumns) > 0 {
		opCount++
	}
	if len(item.RemoveColumns) > 0 {
		opCount++
	}
	if opCount == 0 {
		return &collectionError{Status: http.StatusBadRequest, Message: "Exactly one sub-operation is required"}
	}
	if opCount > 1 {
		return &collectionError{Status: http.StatusBadRequest, Message: "Exactly one sub-operation is required"}
	}

	return nil
}

func (h *CollectionHandler) executeUpdate(item collectionUpdateItem) *collectionError {
	ctx := context.Background()
	switch {
	case len(item.AddColumns) > 0:
		return h.executeAddColumns(ctx, item.Name, item.AddColumns)
	case len(item.RenameColumns) > 0:
		return h.executeRenameColumns(ctx, item.Name, item.RenameColumns)
	case len(item.ModifyColumns) > 0:
		return h.executeModifyColumns(ctx, item.Name, item.ModifyColumns)
	case len(item.RemoveColumns) > 0:
		return h.executeRemoveColumns(ctx, item.Name, item.RemoveColumns)
	}
	return nil
}

func (h *CollectionHandler) executeAddColumns(ctx context.Context, table string, cols []collectionColumn) *collectionError {
	col, _ := h.registry.Get(table)
	existing := make(map[string]bool)
	for _, f := range col.Fields {
		existing[f.Name] = true
	}

	for _, c := range cols {
		if c.Name == "id" {
			return &collectionError{Status: http.StatusBadRequest, Message: "Column 'id' is managed by the server"}
		}
		if !IsValidFieldName(c.Name) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column name %q", c.Name)}
		}
		if !isValidMoonType(c.Type) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column type %q", c.Type)}
		}
		if existing[c.Name] {
			return &collectionError{Status: http.StatusConflict, Message: fmt.Sprintf("Column '%s' already exists", c.Name)}
		}

		ddl := h.buildAddColumnDDL(table, c)
		if err := h.db.ExecDDL(ctx, ddl); err != nil {
			return &collectionError{Status: http.StatusInternalServerError, Message: "Internal server error"}
		}
		existing[c.Name] = true
	}
	return nil
}

func (h *CollectionHandler) buildAddColumnDDL(table string, c collectionColumn) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		quoteIdent(table), quoteIdent(c.Name), moonTypeToSQLite(c.Type)))

	nullable := boolVal(c.Nullable, false)
	if !nullable {
		// SQLite requires a DEFAULT for NOT NULL columns added via ALTER TABLE
		// to satisfy existing rows. This is a safety/compatibility choice.
		sb.WriteString(fmt.Sprintf(" NOT NULL DEFAULT %s", defaultForType(c.Type)))
	}
	if boolVal(c.Unique, false) {
		sb.WriteString(" UNIQUE")
	}
	return sb.String()
}

func (h *CollectionHandler) executeRenameColumns(ctx context.Context, table string, renames []renameColumnSpec) *collectionError {
	col, _ := h.registry.Get(table)
	existing := make(map[string]bool)
	for _, f := range col.Fields {
		existing[f.Name] = true
	}

	for _, r := range renames {
		if r.OldName == "id" || r.NewName == "id" {
			return &collectionError{Status: http.StatusBadRequest, Message: "Column 'id' is managed by the server"}
		}
		if !existing[r.OldName] {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Column '%s' does not exist", r.OldName)}
		}
		if !IsValidFieldName(r.NewName) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column name %q", r.NewName)}
		}

		ddl := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			quoteIdent(table), quoteIdent(r.OldName), quoteIdent(r.NewName))
		if err := h.db.ExecDDL(ctx, ddl); err != nil {
			return &collectionError{Status: http.StatusInternalServerError, Message: "Internal server error"}
		}

		delete(existing, r.OldName)
		existing[r.NewName] = true
	}
	return nil
}

func (h *CollectionHandler) executeModifyColumns(ctx context.Context, table string, cols []collectionColumn) *collectionError {
	col, _ := h.registry.Get(table)
	existing := make(map[string]bool)
	for _, f := range col.Fields {
		existing[f.Name] = true
	}

	for _, c := range cols {
		if c.Name == "id" {
			return &collectionError{Status: http.StatusBadRequest, Message: "Column 'id' is managed by the server"}
		}
		if !existing[c.Name] {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Column '%s' does not exist", c.Name)}
		}
		if !isValidMoonType(c.Type) {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid column type %q", c.Type)}
		}
	}

	// SQLite does not support ALTER COLUMN. Recreate the table with
	// modified column definitions, copy data, drop the original, and rename.
	if err := h.recreateTableWithModifications(ctx, table, col, cols); err != nil {
		return &collectionError{Status: http.StatusInternalServerError, Message: "Internal server error"}
	}
	return nil
}

func (h *CollectionHandler) recreateTableWithModifications(ctx context.Context, table string, col *Collection, mods []collectionColumn) error {
	modMap := make(map[string]collectionColumn)
	for _, m := range mods {
		modMap[m.Name] = m
	}

	tempTable := table + "_moon_tmp_" + GenerateULID()

	var colDefs []string
	var colNames []string

	for _, f := range col.Fields {
		if f.Name == "id" {
			colDefs = append(colDefs, fmt.Sprintf("%s TEXT PRIMARY KEY", quoteIdent("id")))
			colNames = append(colNames, quoteIdent("id"))
			continue
		}

		mod, isModified := modMap[f.Name]
		fieldType := f.Type
		nullable := f.Nullable
		unique := f.Unique

		if isModified {
			fieldType = mod.Type
			nullable = boolVal(mod.Nullable, false)
			unique = boolVal(mod.Unique, false)
		}

		def := fmt.Sprintf("%s %s", quoteIdent(f.Name), moonTypeToSQLite(fieldType))
		if !nullable {
			def += " NOT NULL"
		}
		if unique {
			def += " UNIQUE"
		}
		colDefs = append(colDefs, def)
		colNames = append(colNames, quoteIdent(f.Name))
	}

	colNameStr := strings.Join(colNames, ", ")

	steps := []string{
		fmt.Sprintf("CREATE TABLE %s (%s)", quoteIdent(tempTable), strings.Join(colDefs, ", ")),
		fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", quoteIdent(tempTable), colNameStr, colNameStr, quoteIdent(table)),
		fmt.Sprintf("DROP TABLE %s", quoteIdent(table)),
		fmt.Sprintf("ALTER TABLE %s RENAME TO %s", quoteIdent(tempTable), quoteIdent(table)),
	}

	for _, ddl := range steps {
		if err := h.db.ExecDDL(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

func (h *CollectionHandler) executeRemoveColumns(ctx context.Context, table string, colNames []string) *collectionError {
	col, _ := h.registry.Get(table)
	existing := make(map[string]bool)
	for _, f := range col.Fields {
		existing[f.Name] = true
	}

	for _, name := range colNames {
		if name == "id" {
			return &collectionError{Status: http.StatusBadRequest, Message: "Column 'id' is managed by the server"}
		}
		if !existing[name] {
			return &collectionError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Column '%s' does not exist", name)}
		}

		ddl := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", quoteIdent(table), quoteIdent(name))
		if err := h.db.ExecDDL(ctx, ddl); err != nil {
			return &collectionError{Status: http.StatusInternalServerError, Message: "Internal server error"}
		}
		delete(existing, name)
	}
	return nil
}

// ---------------------------------------------------------------------------
// op=destroy
// ---------------------------------------------------------------------------

func (h *CollectionHandler) handleDestroy(w http.ResponseWriter, rawItems []json.RawMessage) {
	if len(rawItems) == 0 {
		WriteError(w, http.StatusBadRequest, "Data must not be empty")
		return
	}

	var results []any
	for _, raw := range rawItems {
		var item collectionDestroyItem
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid destroy item")
			return
		}

		if item.Name == "" {
			WriteError(w, http.StatusBadRequest, "Collection name is required")
			return
		}
		if strings.HasPrefix(item.Name, "moon_") {
			WriteError(w, http.StatusBadRequest, "Collection name is reserved")
			return
		}
		if item.Name == "users" || item.Name == "apikeys" {
			WriteError(w, http.StatusForbidden, "Forbidden")
			return
		}
		if _, exists := h.registry.Get(item.Name); !exists {
			WriteError(w, http.StatusNotFound, fmt.Sprintf("Collection '%s' not found", item.Name))
			return
		}

		ddl := fmt.Sprintf("DROP TABLE %s", quoteIdent(item.Name))
		if err := h.db.ExecDDL(context.Background(), ddl); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if err := h.registry.Refresh(); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		results = append(results, map[string]any{"name": item.Name})
	}

	meta := map[string]any{"success": len(results), "failed": 0}
	WriteSuccessFull(w, http.StatusOK, "Collection destroyed successfully", results, meta, nil)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// collectionError carries an HTTP status and message for collection operations.
type collectionError struct {
	Status  int
	Message string
}

func writeCollectionError(w http.ResponseWriter, e *collectionError) {
	WriteError(w, e.Status, e.Message)
}

// isValidMoonType returns true if t is a recognized Moon field type for column definitions.
func isValidMoonType(t string) bool {
	switch t {
	case MoonFieldTypeString, MoonFieldTypeInteger, MoonFieldTypeDecimal,
		MoonFieldTypeBoolean, MoonFieldTypeDatetime, MoonFieldTypeJSON:
		return true
	}
	return false
}

// moonTypeToSQLite maps a Moon field type to the corresponding SQLite column type.
func moonTypeToSQLite(t string) string {
	switch t {
	case MoonFieldTypeString:
		return SQLiteTypeString
	case MoonFieldTypeInteger:
		return SQLiteTypeInteger
	case MoonFieldTypeDecimal:
		return SQLiteTypeDecimal
	case MoonFieldTypeBoolean:
		return SQLiteTypeBoolean
	case MoonFieldTypeDatetime:
		return SQLiteTypeDatetime
	case MoonFieldTypeJSON:
		return SQLiteTypeJSON
	default:
		return SQLiteTypeString
	}
}

// defaultForType returns a SQL default literal for ADD COLUMN NOT NULL in SQLite.
func defaultForType(t string) string {
	switch t {
	case MoonFieldTypeInteger, MoonFieldTypeDecimal, MoonFieldTypeBoolean:
		return "0"
	default:
		return "''"
	}
}

// boolVal returns the value pointed to by p, or the fallback if p is nil.
func boolVal(p *bool, fallback bool) bool {
	if p == nil {
		return fallback
	}
	return *p
}

// parsePagination extracts page and per_page from query parameters.
func parsePagination(r *http.Request) (page, perPage int) {
	page = 1
	perPage = DefaultPerPage

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			perPage = n
			if perPage > MaxPerPage {
				perPage = MaxPerPage
			}
		}
	}
	return page, perPage
}

// buildPaginationLinks builds the standard pagination links map.
func buildPaginationLinks(basePath string, page, perPage, totalPages int) map[string]any {
	linkURL := func(p int) string {
		return fmt.Sprintf("%s?page=%d&per_page=%d", basePath, p, perPage)
	}

	links := map[string]any{
		"first": linkURL(1),
		"last":  linkURL(totalPages),
	}

	if page > 1 {
		links["prev"] = linkURL(page - 1)
	} else {
		links["prev"] = nil
	}

	if page < totalPages {
		links["next"] = linkURL(page + 1)
	} else {
		links["next"] = nil
	}

	return links
}
