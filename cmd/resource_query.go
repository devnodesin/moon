package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ResourceQueryHandler implements GET /data/{resource}:query.
type ResourceQueryHandler struct {
	db       DatabaseAdapter
	registry *SchemaRegistry
	cfg      *AppConfig
	prefix   string
}

// NewResourceQueryHandler creates a ResourceQueryHandler with the given dependencies.
func NewResourceQueryHandler(db DatabaseAdapter, registry *SchemaRegistry, cfg *AppConfig) *ResourceQueryHandler {
	return &ResourceQueryHandler{
		db:       db,
		registry: registry,
		cfg:      cfg,
		prefix:   strings.TrimRight(cfg.Server.Prefix, "/"),
	}
}

// HandleQuery handles GET /data/{resource}:query requests.
func (h *ResourceQueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}

	col, ok := h.registry.Get(resource)
	if !ok {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("Resource '%s' not found", resource))
		return
	}

	q := r.URL.Query()

	if err := h.validateQueryParams(q, col); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if id := q.Get("id"); id != "" {
		h.handleGetOne(w, r, resource, col, id)
		return
	}
	h.handleList(w, r, resource, col)
}

// ---------------------------------------------------------------------------
// Get-one mode
// ---------------------------------------------------------------------------

func (h *ResourceQueryHandler) handleGetOne(w http.ResponseWriter, _ *http.Request, resource string, col *Collection, id string) {
	opts := QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
		Page:    1,
		PerPage: 1,
	}

	rows, _, err := h.db.QueryRows(context.Background(), resource, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(rows) == 0 {
		WriteError(w, http.StatusNotFound, "Resource not found")
		return
	}

	record := formatRecord(rows[0], col)
	record = filterHiddenFields(resource, record)

	WriteSuccess(w, http.StatusOK, "Resource retrieved successfully", []any{record})
}

// ---------------------------------------------------------------------------
// List mode
// ---------------------------------------------------------------------------

func (h *ResourceQueryHandler) handleList(w http.ResponseWriter, r *http.Request, resource string, col *Collection) {
	q := r.URL.Query()
	page, perPage := parsePagination(r)

	opts := QueryOptions{
		Page:    page,
		PerPage: perPage,
	}

	// Sort
	if sortParam := q.Get("sort"); sortParam != "" {
		sortFields, err := parseSortParam(sortParam, col)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		opts.Sort = sortFields
	}

	// Fields projection
	if fieldsParam := q.Get("fields"); fieldsParam != "" {
		projFields, err := parseFieldsParam(fieldsParam, col)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		opts.Fields = projFields
	}

	// Full-text search
	if search := q.Get("q"); search != "" {
		opts.Search = search
		opts.SearchFields = getStringFields(col)
	}

	// Filters
	filters, err := parseFilterParams(q, col)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	opts.Filters = filters

	rows, total, err := h.db.QueryRows(context.Background(), resource, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	data := make([]any, 0, len(rows))
	for _, row := range rows {
		record := formatRecord(row, col)
		record = filterHiddenFields(resource, record)
		data = append(data, record)
	}

	totalPages := 1
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	meta := map[string]any{
		"total":        total,
		"count":        len(data),
		"per_page":     perPage,
		"current_page": page,
		"total_pages":  totalPages,
	}

	basePath := fmt.Sprintf("%s/data/%s:query", h.prefix, resource)
	links := buildResourcePaginationLinks(basePath, page, perPage, totalPages, q)

	WriteSuccessFull(w, http.StatusOK, "Resources retrieved successfully", data, meta, links)
}

// ---------------------------------------------------------------------------
// Query parameter validation
// ---------------------------------------------------------------------------

// knownQueryParams lists the recognized top-level query parameter names.
var knownQueryParams = map[string]bool{
	"page":     true,
	"per_page": true,
	"sort":     true,
	"q":        true,
	"fields":   true,
	"id":       true,
}

// filterParamPattern matches filter parameters like field[op].
var filterParamPattern = regexp.MustCompile(`^([a-z][a-z0-9_]*)\[([a-z]+)\]$`)

// validateQueryParams rejects unknown query parameters.
func (h *ResourceQueryHandler) validateQueryParams(q url.Values, col *Collection) error {
	for key := range q {
		if knownQueryParams[key] {
			continue
		}
		if filterParamPattern.MatchString(key) {
			continue
		}
		return fmt.Errorf("Unknown query parameter %q", key)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Sort parsing
// ---------------------------------------------------------------------------

func parseSortParam(sortParam string, col *Collection) ([]SortField, error) {
	fieldMap := buildFieldMap(col)
	parts := strings.Split(sortParam, ",")
	result := make([]SortField, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		desc := false
		fieldName := p
		if strings.HasPrefix(p, "-") {
			desc = true
			fieldName = p[1:]
		}
		if _, ok := fieldMap[fieldName]; !ok {
			return nil, fmt.Errorf("Unknown sort field %q", fieldName)
		}
		result = append(result, SortField{Field: fieldName, Desc: desc})
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Fields parsing
// ---------------------------------------------------------------------------

func parseFieldsParam(fieldsParam string, col *Collection) ([]string, error) {
	fieldMap := buildFieldMap(col)
	parts := strings.Split(fieldsParam, ",")
	seen := make(map[string]bool)
	result := []string{"id"}
	seen["id"] = true
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := fieldMap[p]; !ok {
			return nil, fmt.Errorf("Unknown field %q", p)
		}
		if !seen[p] {
			result = append(result, p)
			seen[p] = true
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Filter parsing
// ---------------------------------------------------------------------------

// validFilterOps lists all recognized filter operators.
var validFilterOps = map[string]bool{
	"eq": true, "ne": true, "gt": true, "lt": true,
	"gte": true, "lte": true, "like": true, "in": true,
}

// opsForType maps Moon field types to the set of valid filter operators.
var opsForType = map[string]map[string]bool{
	MoonFieldTypeID:       {"eq": true, "ne": true, "in": true},
	MoonFieldTypeString:   {"eq": true, "ne": true, "like": true, "in": true},
	MoonFieldTypeInteger:  {"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true, "in": true},
	MoonFieldTypeDecimal:  {"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true, "in": true},
	MoonFieldTypeDatetime: {"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true, "in": true},
	MoonFieldTypeBoolean:  {"eq": true, "ne": true},
	MoonFieldTypeJSON:     {"eq": true, "ne": true},
}

func parseFilterParams(q url.Values, col *Collection) ([]Filter, error) {
	fieldMap := buildFieldMap(col)
	var filters []Filter

	for key, values := range q {
		matches := filterParamPattern.FindStringSubmatch(key)
		if matches == nil {
			continue
		}
		fieldName := matches[1]
		op := matches[2]

		if !validFilterOps[op] {
			return nil, fmt.Errorf("Unknown filter operator %q", op)
		}

		f, ok := fieldMap[fieldName]
		if !ok {
			return nil, fmt.Errorf("Unknown filter field %q", fieldName)
		}

		allowed := opsForType[f.Type]
		if !allowed[op] {
			return nil, fmt.Errorf("Operator %q is not valid for field %q of type %q", op, fieldName, f.Type)
		}

		value := values[0]

		if op == "in" {
			inValues := strings.Split(value, ",")
			filters = append(filters, Filter{Field: fieldName, Op: "in", Value: inValues})
		} else if op == "ne" {
			filters = append(filters, Filter{Field: fieldName, Op: "ne", Value: value})
		} else if op == "like" {
			filters = append(filters, Filter{Field: fieldName, Op: "like", Value: "%" + value + "%"})
		} else {
			filters = append(filters, Filter{Field: fieldName, Op: op, Value: value})
		}
	}
	return filters, nil
}

// ---------------------------------------------------------------------------
// Field helpers
// ---------------------------------------------------------------------------

func buildFieldMap(col *Collection) map[string]Field {
	m := make(map[string]Field, len(col.Fields))
	for _, f := range col.Fields {
		m[f.Name] = f
	}
	return m
}

func getStringFields(col *Collection) []string {
	var result []string
	for _, f := range col.Fields {
		if f.Type == MoonFieldTypeString {
			result = append(result, f.Name)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Record formatting
// ---------------------------------------------------------------------------

// formatRecord converts raw DB values to Moon type representations.
func formatRecord(row map[string]any, col *Collection) map[string]any {
	fieldMap := buildFieldMap(col)
	result := make(map[string]any, len(row))
	for k, v := range row {
		f, ok := fieldMap[k]
		if !ok {
			result[k] = v
			continue
		}
		result[k] = convertToMoonType(v, f.Type)
	}
	return result
}

// convertToMoonType converts a raw database value to the appropriate Moon
// JSON representation based on the field type.
func convertToMoonType(value any, fieldType string) any {
	if value == nil {
		return nil
	}
	switch fieldType {
	case MoonFieldTypeBoolean:
		return toBool(value)
	case MoonFieldTypeInteger:
		return toInteger(value)
	case MoonFieldTypeDecimal:
		return toDecimalString(value)
	case MoonFieldTypeJSON:
		return toJSONValue(value)
	case MoonFieldTypeDatetime:
		return toString(value)
	case MoonFieldTypeID:
		return toString(value)
	default:
		return toString(value)
	}
}

func toInteger(v any) any {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case string:
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i
		}
		return n
	case []byte:
		s := string(n)
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		return s
	default:
		return v
	}
}

func toDecimalString(v any) any {
	switch n := v.(type) {
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(n, 10)
	case int:
		return strconv.Itoa(n)
	case string:
		return n
	case []byte:
		return string(n)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toJSONValue(v any) any {
	switch s := v.(type) {
	case string:
		var parsed any
		if err := json.Unmarshal([]byte(s), &parsed); err == nil {
			return parsed
		}
		return s
	case []byte:
		var parsed any
		if err := json.Unmarshal(s, &parsed); err == nil {
			return parsed
		}
		return string(s)
	default:
		return v
	}
}

func toString(v any) any {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ---------------------------------------------------------------------------
// Hidden field filtering
// ---------------------------------------------------------------------------

// filterHiddenFields removes fields that must not appear in API responses
// for system collections (e.g., password_hash for users, key_hash for apikeys).
func filterHiddenFields(resource string, record map[string]any) map[string]any {
	hidden, ok := hiddenSystemFields[resource]
	if !ok {
		return record
	}
	for field := range hidden {
		delete(record, field)
	}
	return record
}

// ---------------------------------------------------------------------------
// Pagination links with query params
// ---------------------------------------------------------------------------

// buildResourcePaginationLinks builds pagination links that preserve all
// active query parameters (sort, filter, q, fields, etc.).
func buildResourcePaginationLinks(basePath string, page, perPage, totalPages int, q url.Values) map[string]any {
	linkURL := func(p int) string {
		params := url.Values{}
		params.Set("page", strconv.Itoa(p))
		params.Set("per_page", strconv.Itoa(perPage))

		// Preserve other query params
		for key, vals := range q {
			if key == "page" || key == "per_page" {
				continue
			}
			for _, v := range vals {
				params.Add(key, v)
			}
		}
		return basePath + "?" + params.Encode()
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
