## Overview

- Enhance the existing `query.Builder` to robustly support advanced filtering operators beyond basic equality
- Ensure the builder safely handles operators like `LIKE`, `gt` (greater than), `lt` (less than), `gte`, `lte`, etc.
- Encapsulate all dialect-specific SQL generation logic within the query package to keep handlers clean
- The current `Condition` struct is already flexible with `Column`, `Operator`, and `Value` fields, but needs validation and safety improvements

## Requirements

- Support standard SQL comparison operators: `=`, `>`, `<`, `>=`, `<=`, `!=`
- Support `LIKE` operator for partial string matching with proper escaping
- Support `IN` operator for matching against multiple values
- Validate operators to prevent SQL injection through operator strings
- Ensure proper value escaping and parameterization for all operators
- Handle dialect-specific differences (SQLite, Postgres, MySQL) within the builder
- Add operator constant definitions to prevent magic strings
- Maintain backward compatibility with existing query builder usage
- Add comprehensive tests for all operators and edge cases

## Acceptance

- Query builder correctly generates parameterized SQL for all supported operators
- All dialect-specific logic is encapsulated in the query package (no `if dialect == ...` in handlers)
- Tests pass for SQLite, Postgres, and MySQL dialects
- No SQL injection vulnerabilities through operator or value manipulation
- Documentation updated with supported operators and examples
- Existing functionality remains unbroken
