package main

// ---------------------------------------------------------------------------
// Backend-aware column type mapping
//
// Dynamic-collection DDL and ALTER statements must emit the correct native SQL
// column type for the selected backend (SPEC.md §9.3). The Moon field types are
// shared; only the physical storage type differs per backend.
// ---------------------------------------------------------------------------

// moonTypeToSQL maps a Moon field type to the corresponding SQL column type for
// the given database connection. It preserves the existing SQLite behavior and
// adds the PostgreSQL mapping required by SPEC.md §9.3.
func moonTypeToSQL(t string, connection string) string {
	switch connection {
	case DBConnectionPostgres:
		return moonTypeToPostgres(t)
	case DBConnectionMySQL:
		return moonTypeToMySQL(t)
	default:
		return moonTypeToSQLite(t)
	}
}

// moonTypeToPostgres maps a Moon field type to a PostgreSQL column type.
func moonTypeToPostgres(t string) string {
	switch t {
	case MoonFieldTypeString:
		return PGTypeString
	case MoonFieldTypeInteger:
		return PGTypeInteger
	case MoonFieldTypeDecimal:
		return PGTypeDecimal
	case MoonFieldTypeBoolean:
		return PGTypeBoolean
	case MoonFieldTypeDatetime:
		return PGTypeDatetime
	case MoonFieldTypeJSON:
		return PGTypeJSON
	default:
		return PGTypeString
	}
}

// moonTypeToMySQL maps a Moon field type to a MySQL column type. MySQL is not yet
// implemented, but the mapping is provided so DDL generation is backend-driven.
func moonTypeToMySQL(t string) string {
	switch t {
	case MoonFieldTypeString:
		return MySQLTypeString
	case MoonFieldTypeInteger:
		return MySQLTypeInteger
	case MoonFieldTypeDecimal:
		return MySQLTypeDecimal
	case MoonFieldTypeBoolean:
		return MySQLTypeBoolean
	case MoonFieldTypeDatetime:
		return MySQLTypeDatetime
	case MoonFieldTypeJSON:
		return MySQLTypeJSON
	default:
		return MySQLTypeString
	}
}

// moonTypeToSQLite maps a Moon field type to the corresponding SQLite column
// type. Kept for SQLite DDL generation and existing tests.
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

// defaultLiteralForType returns a SQL default literal used when adding a NOT
// NULL column to an existing table (SQLite ALTER TABLE requires a DEFAULT for
// NOT NULL columns). The literal must be a valid constant for the target
// backend. PostgreSQL rejects integer literals for BOOLEAN columns (it requires
// true/false) and requires a valid JSON document for JSON columns rather than an
// empty string, so backend-specific literals are emitted.
func defaultLiteralForType(t string, connection string) string {
	switch t {
	case MoonFieldTypeInteger, MoonFieldTypeDecimal:
		return "0"
	case MoonFieldTypeBoolean:
		if connection == DBConnectionPostgres {
			return "false"
		}
		return "0"
	case MoonFieldTypeJSON:
		if connection == DBConnectionPostgres {
			return "'\"{}\"'"
		}
		return "''"
	default:
		return "''"
	}
}

// ---------------------------------------------------------------------------
// Backend-aware column type mapping
