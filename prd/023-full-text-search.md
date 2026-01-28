## Overview

- Add full-text search capability to the data list endpoint
- Enable clients to search across string columns in collections
- Start with basic `LIKE` pattern matching, with option to enhance later with FTS5
- Support a dedicated search query parameter (e.g., `?q=laptop`)
- Search should work across multiple text columns in a collection

## Requirements

- Accept a `q` query parameter for search terms
- Search across all text/string columns in the collection schema
- Use `LIKE %term%` pattern matching for basic search (case-insensitive where supported)
- Combine search conditions with OR logic across columns
- Allow search to work alongside filters and sorting
- Handle special characters in search terms safely (escape wildcards)
- Return empty results when search term matches nothing
- Validate search term length (e.g., minimum 1 character)
- Add tests for search functionality with various inputs
- Document search behavior and limitations

## Acceptance

- Clients can search collections using the `?q=term` parameter
- Search works across all string columns in the schema
- Search is case-insensitive where database supports it
- Special characters in search terms are properly escaped
- Search can be combined with filters and sorting
- Empty or very short search terms are handled gracefully
- Tests cover basic search, special characters, and combined operations
- API documentation includes search examples and behavior notes
