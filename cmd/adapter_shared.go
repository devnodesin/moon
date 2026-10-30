package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Shared SQL helpers (backend-agnostic)
//
// These helpers are used by every DatabaseAdapter implementation. They rely on
// the standard library database/sql package so both the SQLite adapter and the
// PostgreSQL adapter (via the pgx stdlib driver) can share them unchanged.
// Placeholder syntax is "?" which SQLite accepts directly. The PostgreSQL
// adapter rewrites "?" to PostgreSQL-native "$n" placeholders itself in
// pgQuestionToDollar before execution, because the pgx stdlib driver does not
// perform that conversion.
// ---------------------------------------------------------------------------

// quoteIdent wraps a SQL identifier in double quotes to prevent injection.
// Any embedded double quotes are escaped by doubling.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// filterOpSQL maps filter operator names to SQL operators.
var filterOpSQL = map[string]string{
	"eq":   "=",
	"ne":   "!=",
	"neq":  "!=",
	"gt":   ">",
	"gte":  ">=",
	"lt":   "<",
	"lte":  "<=",
	"like": "LIKE",
}

// buildWhereClause builds a WHERE clause from QueryOptions filters and
// search parameters. Returns the clause string (including " WHERE " prefix
// if non-empty) and the corresponding parameter values.
func buildWhereClause(opts QueryOptions) (string, []any) {
	var conditions []string
	var args []any

	for _, f := range opts.Filters {
		if f.Op == "in" {
			values, ok := f.Value.([]string)
			if !ok || len(values) == 0 {
				continue
			}
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = "?"
				args = append(args, v)
			}
			conditions = append(conditions,
				fmt.Sprintf("%s IN (%s)", quoteIdent(f.Field), strings.Join(placeholders, ", ")))
			continue
		}
		sqlOp, ok := filterOpSQL[f.Op]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s %s ?", quoteIdent(f.Field), sqlOp))
		args = append(args, f.Value)
	}

	if opts.Search != "" && len(opts.SearchFields) > 0 {
		var searchConds []string
		for _, sf := range opts.SearchFields {
			searchConds = append(searchConds, fmt.Sprintf("%s LIKE ?", quoteIdent(sf)))
			args = append(args, "%"+opts.Search+"%")
		}
		conditions = append(conditions, "("+strings.Join(searchConds, " OR ")+")")
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// scanRows reads all rows from a *sql.Rows into a slice of maps. Any time.Time
// value is normalized to an RFC3339 string so that the external API behavior is
// identical across backends (SPEC.md §9.3 requires datetime values to be
// returned as RFC3339 strings; SQLite stores them as TEXT while PostgreSQL uses
// a native TIMESTAMP that database/sql decodes into time.Time).
func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeScannedValue(values[i])
		}
		results = append(results, row)
	}
	return results, nil
}

// normalizeScannedValue converts driver-specific scanned values into the
// canonical representation used by the API layer. time.Time (produced by
// PostgreSQL TIMESTAMP columns) is rendered as an RFC3339 string. All other
// values are returned unchanged.
func normalizeScannedValue(v any) any {
	switch t := v.(type) {
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case *time.Time:
		if t == nil {
			return nil
		}
		return t.UTC().Format(time.RFC3339)
	default:
		return v
	}
}
