Replace hardcoded JSON Appendix with a dynamic, pretty-printed JSON string generated at render time from the runtime registry and config. No backward compatibility steps required.

## Files to modify

- cmd/moon/internal/handlers/doc.go
- cmd/moon/internal/handlers/templates/doc.md.tmpl
- cmd/moon/internal/handlers/doc_test.go

## High-level steps

- Add a new JSONAppendix string field to the template data passed to doc.md.tmpl.
- Implement a buildJSONAppendix(...) helper in DocHandler that:
  - Reads collections and fields from the SchemaRegistry and config flags (auth, base URL, prefix, version).
  - Assembles typed structs (not maps) for deterministic ordering.
  - Marshals to pretty JSON (MarshalIndent) and returns the string.
- Call the helper during Markdown/HTML generation and set DocData.JSONAppendix before executing the template (ensure caching integrates this output).
- Replace the hardcoded JSON Appendix section in doc.md.tmpl with a verbatim insertion of JSONAppendix inside a fenced JSON block.
- Log errors and render a small fallback JSON block if generation fails.

## Tests to add/modify

- Presence test: rendered markdown contains the JSON Appendix block and includes collection names/fields.
- Determinism test: different insertion orders produce identical appendix (sort collections/fields).
- Refresh test: after registry change + doc:refresh, generated appendix reflects updates.
- 404/behavior tests unchanged for doc endpoints (ensure content-type and caching preserved).

## Determinism & safety notes

- Use structs and sorted slices (collection and field names) to guarantee stable JSON output.
- Do not include secrets; only capability flags and header names.
- Keep appendix generation cheap; rely on existing caching to avoid runtime cost.

## Complexity & risk

- Complexity: low.
- Risk: small and isolated to documentation generation.
- Main care points: deterministic ordering, integrate with cache invalidation, and test updates.

## Estimated effort

- Basic implementation + tests: ~2–4 hours.
- More extensive appendix detail: up to 1–2 days.

## Acceptance criteria

- Markdown output includes a pretty JSON Appendix generated from registry/config.
- JSON contains service/version, base URL/prefix, auth capabilities, collections and fields (name, type, nullable).
- Output is deterministic and updates after doc:refresh when registry changes.
- Existing doc endpoints, content-type, caching, and public/no-auth behavior remain unchanged.

## Minimal PR checklist

- [ ] Add Appendix builder and DocData.JSONAppendix
- [ ] Integrate builder into doc generation (md/html) and caching
- [ ] Update template to render JSONAppendix
- [ ] Add/modify tests (presence, determinism, refresh)
- [ ] Run CI and merge

Would you like a concise diff and exact function/struct names next?

## Currnt for your refernce

```markdown
## JSON Appendix

This machine-readable appendix enables automated client generation and reduces ambiguity for AI coding agents.

```json
{
  "service": "moon",
  "version": "1.99",
  "document_version": "1.5.1",
  "base_url": "http://localhost:6006",
  "url_prefix": null,

  "authentication": {
    "modes": ["jwt", "api_key"],
    "headers": {
      "jwt": "Authorization: Bearer <token>",
      "api_key": "X-API-Key: <key>"
    },
    "rate_limits": {
      "jwt": "100 requests per minute per user",
      "api_key": "1000 requests per minute per key"
    },
    "rules": {
      "jwt_for": "user-facing apps with session management",
      "api_key_for": "server-to-server or backend services"
    }
  },

  "collections": {
    "terminology": {
      "collection": "table/database collection",
      "field": "column/table column",
      "record": "row/table row"
    },
    "naming": {
      "case": "snake_case",
      "lowercase": true,
      "pattern": "^[a-z][a-z0-9_]*$"
    },
    "constraints": {
      "joins_supported": false,
      "foreign_keys": false,
      "transactions": false,
      "triggers": false,
      "background_jobs": false
    }
  },

  "data_types": [
    {
      "name": "string",
      "description": "Text values of any length",
      "sql_mapping": "TEXT",
      "example": "Wireless Mouse"
    },
    {
      "name": "integer",
      "description": "64-bit whole numbers",
      "sql_mapping": "INTEGER",
      "example": 42
    },
    {
      "name": "boolean",
      "description": "true/false values",
      "sql_mapping": "BOOLEAN",
      "example": true,
    },
    {
      "name": "datetime",
      "description": "Date/time in RFC3339 or ISO 8601 format",
      "sql_mapping": "DATETIME",
      "example": "2023-01-31T13:45:00Z"
    },
    {
      "name": "json",
      "description": "Arbitrary JSON object or array",
      "sql_mapping": "JSON",
      "example": {"key": "value"}
    },
    {
      "name": "decimal",
      "description": "Decimal values with precision",
      "sql_mapping": "DECIMAL",
      "format": "string",
      "example": "199.99",
      "note": "API input/output uses strings, default 2 decimal places"
    }
  ],

  "endpoints": {
    "health": {
      "path": "/health",
      "method": "GET",
      "auth_required": false,
      "description": "Health check endpoint"
    },
    "authentication": {
      "login": {
        "path": "/auth:login",
        "method": "POST",
        "auth_required": false,
        "description": "Authenticate user, receive JWT tokens"
      },
      "logout": {
        "path": "/auth:logout",
        "method": "POST",
        "auth_required": true,
        "description": "Invalidate current session's refresh token"
      },
      "refresh": {
        "path": "/auth:refresh",
        "method": "POST",
        "auth_required": false,
        "description": "Exchange refresh token for new access token"
      },
      "me": {
        "path": "/auth:me",
        "methods": ["GET", "POST"],
        "auth_required": true,
        "description": "Get or update current user profile"
      }
    },
    "user_management": {
      "list": {
        "path": "/users:list",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin"
      },
      "get": {
        "path": "/users:get",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"]
      },
      "create": {
        "path": "/users:create",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin"
      },
      "update": {
        "path": "/users:update",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"],
        "actions": ["reset_password", "revoke_sessions"]
      },
      "destroy": {
        "path": "/users:destroy",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"]
      }
    },
    "apikey_management": {
      "list": {
        "path": "/apikeys:list",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin"
      },
      "get": {
        "path": "/apikeys:get",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"]
      },
      "create": {
        "path": "/apikeys:create",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin"
      },
      "update": {
        "path": "/apikeys:update",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"],
        "actions": ["rotate"]
      },
      "destroy": {
        "path": "/apikeys:destroy",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "params": ["id"]
      }
    },
    "collection_management": {
      "list": {
        "path": "/collections:list",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin"
      },
      "get": {
        "path": "/collections:get",
        "method": "GET",
        "auth_required": true,
        "role_required": "admin",
        "params": ["name"]
      },
      "create": {
        "path": "/collections:create",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin"
      },
      "update": {
        "path": "/collections:update",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "operations": ["add_columns", "rename_columns", "modify_columns", "remove_columns"]
      },
      "destroy": {
        "path": "/collections:destroy",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin"
      }
    },
    "data_access": {
      "list": {
        "path": "/{collection}:list",
        "method": "GET",
        "auth_required": true,
        "description": "List records in collection"
      },
      "get": {
        "path": "/{collection}:get",
        "method": "GET",
        "auth_required": true,
        "params": ["id"],
        "description": "Get single record by ID"
      },
      "create": {
        "path": "/{collection}:create",
        "method": "POST",
        "auth_required": true,
        "description": "Create new record"
      },
      "update": {
        "path": "/{collection}:update",
        "method": "POST",
        "auth_required": true,
        "description": "Update existing record"
      },
      "destroy": {
        "path": "/{collection}:destroy",
        "method": "POST",
        "auth_required": true,
        "description": "Delete record"
      }
    },
    "aggregation": {
      "count": {
        "path": "/{collection}:count",
        "method": "GET",
        "auth_required": true,
        "description": "Count records"
      },
      "sum": {
        "path": "/{collection}:sum",
        "method": "GET",
        "auth_required": true,
        "params": ["field"],
        "description": "Sum numeric field"
      },
      "avg": {
        "path": "/{collection}:avg",
        "method": "GET",
        "auth_required": true,
        "params": ["field"],
        "description": "Average numeric field"
      },
      "min": {
        "path": "/{collection}:min",
        "method": "GET",
        "auth_required": true,
        "params": ["field"],
        "description": "Minimum value"
      },
      "max": {
        "path": "/{collection}:max",
        "method": "GET",
        "auth_required": true,
        "params": ["field"],
        "description": "Maximum value"
      }
    },
    "documentation": {
      "html": {
        "path": "/doc/",
        "method": "GET",
        "auth_required": false,
        "description": "HTML documentation"
      },
      "markdown": {
        "path": "/doc/md",
        "method": "GET",
        "auth_required": false,
        "description": "Markdown documentation"
      },
      "refresh": {
        "path": "/doc:refresh",
        "method": "POST",
        "auth_required": true,
        "role_required": "admin",
        "description": "Refresh documentation cache"
      }
    }
  },

  "query": {
    "operators": ["eq", "ne", "gt", "lt", "gte", "lte", "like", "in"],
    "syntax": {
      "filter": "?column[operator]=value",
      "examples": [
        "?price[gte]=100",
        "?category[eq]=electronics",
        "?name[like]=%mouse%"
      ]
    },
    "sorting": {
      "syntax": "?sort={field1,-field2}",
      "ascending": "field",
      "descending": "-field",
      "example": "?sort=-price,name"
    },
    "pagination": {
      "cursor_param": "after",
      "limit_param": "limit",
      "example": "?limit=10&after=01ARZ3NDEKTSV4RRFFQ69G5FBX"
    },
    "search": {
      "full_text_param": "q",
      "description": "Searches across all text/string columns",
      "example": "?q=wireless"
    },
    "field_selection": {
      "param": "fields",
      "description": "Return only specified fields (id always included)",
      "example": "?fields=name,price"
    }
  },

  "aggregation": {
    "supported": ["count", "sum", "avg", "min", "max"],
    "numeric_types": ["integer", "decimal"],
    "note": "Aggregation functions work on integer and decimal field types only"
  },

  "http_status_codes": {
    "200": "OK - Successful GET request",
    "201": "Created - Successful POST request creating resource",
    "400": "Bad Request - Invalid input or parameters",
    "401": "Unauthorized - Missing or invalid authentication",
    "403": "Forbidden - Insufficient permissions",
    "404": "Not Found - Resource not found",
    "409": "Conflict - Resource already exists",
    "429": "Too Many Requests - Rate limit exceeded",
    "500": "Internal Server Error - Server error"
  },

  "rate_limiting": {
    "headers": {
      "limit": "X-RateLimit-Limit",
      "remaining": "X-RateLimit-Remaining",
      "reset": "X-RateLimit-Reset"
    }
  },

  "cors": {
    "allowed_methods": ["GET", "POST", "OPTIONS"],
    "allowed_headers": ["Authorization", "Content-Type", "X-API-Key"],
    "configurable": true,
    "config_file": "samples/moon.conf",
    "config_key": "cors.allowed_origins"
  },

  "guarantees": {
    "transactions": false,
    "joins": false,
    "foreign_keys": false,
    "triggers": false,
    "background_jobs": false
  },

  "aip_standards": {
    "custom_actions": "AIP-136",
    "pattern": "resource:action",
    "separator": ":",
    "description": "APIs use colon separator between resource and action for predictable interface"
  }
}
```
```