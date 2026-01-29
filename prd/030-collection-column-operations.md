## Overview
- Extend the `/collections:update` endpoint to support removing and modifying columns
- Enable full column lifecycle management: add, remove, rename, and modify column types
- Maintain backward compatibility with existing `add_columns` functionality
- Provide safe column operations with validation and error handling across all database dialects

## Requirements
- Support `remove_columns` field in `POST /collections:update` request body
  - Accept array of column names to remove from collection
  - Validate column exists before removal
  - Prevent removal of system columns (`id`, `ulid`)
  - Generate and execute `ALTER TABLE DROP COLUMN` DDL per dialect
- Support `rename_columns` field in `POST /collections:update` request body
  - Accept array of objects with `old_name` and `new_name` fields
  - Validate old column exists and new name doesn't conflict
  - Prevent renaming system columns (`id`, `ulid`)
  - Generate and execute column rename DDL per dialect (varies by database)
- Support `modify_columns` field in `POST /collections:update` request body
  - Accept array of column definitions with `name` and new `type`/constraints
  - Validate type compatibility and data integrity
  - Generate and execute `ALTER TABLE MODIFY/ALTER COLUMN` DDL per dialect
  - Handle nullable/unique/default value changes
- All operations can be combined in single request (add + remove + rename + modify)
- Operations execute in order: rename → modify → add → remove
- Each operation updates In-Memory Registry atomically
- Rollback registry on DDL execution failure
- Cross-dialect compatibility for SQLite, PostgreSQL, and MySQL
- Comprehensive validation and error messages for each operation type

## Acceptance
- `remove_columns` successfully drops columns from table and registry
- `rename_columns` successfully renames columns in table and registry
- `modify_columns` successfully changes column types/constraints in table and registry
- Cannot remove or rename system columns (`id`, `ulid`)
- Invalid column names return descriptive errors
- Type changes validate compatibility (e.g., cannot change text to integer with existing data)
- Registry stays consistent with database after each operation
- SQLite workarounds for limited ALTER TABLE support (recreate table if needed)
- Unit tests cover all validation logic and SQL generation per dialect
- Integration tests verify end-to-end column operations across databases
- Test script updated with examples for remove, rename, and modify operations
- SPEC.md updated to reflect new API capabilities
- Documentation includes examples for all operation types
- Update script `samples\test_scripts\collection.sh` to include the curl commads for testing above new features. 
