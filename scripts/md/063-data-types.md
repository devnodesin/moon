### Create Collection with Typed Columns

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "name": "typed_items",
            "columns": [
              {
                "name": "string_test",
                "type": "string",
                "nullable": false
              },
              {
                "name": "integer_test",
                "type": "integer",
                "nullable": true
              },
              {
                "name": "decimal_test",
                "type": "decimal",
                "nullable": true
              },
              {
                "name": "boolean_test",
                "type": "boolean",
                "nullable": true
              },
              {
                "name": "date_test",
                "type": "datetime",
                "nullable": true
              },
              {
                "name": "json_test",
                "type": "json",
                "nullable": true
              }
            ]
          }
        ]
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "message": "Collection created successfully",
  "data": [
    {
      "columns": [
        {
          "name": "string_test",
          "nullable": false,
          "type": "string",
          "unique": false
        },
        {
          "name": "integer_test",
          "nullable": true,
          "type": "integer",
          "unique": false
        },
        {
          "name": "decimal_test",
          "nullable": true,
          "type": "decimal",
          "unique": false
        },
        {
          "name": "boolean_test",
          "nullable": true,
          "type": "boolean",
          "unique": false
        },
        {
          "name": "date_test",
          "nullable": true,
          "type": "datetime",
          "unique": false
        },
        {
          "name": "json_test",
          "nullable": true,
          "type": "json",
          "unique": false
        }
      ],
      "name": "typed_items"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Get Schema

Retrieve the schema for `typed_items`. Each field must report the correct Moon type: `id` for the primary key, `string`, `integer`, `decimal`, `boolean`, `datetime`, and `json` for the user-defined columns.

```bash
curl -s -X GET "http://localhost:6000/data/typed_items:schema" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Schema retrieved successfully",
  "data": [
    {
      "name": "typed_items",
      "fields": [
        {
          "name": "id",
          "type": "id",
          "nullable": true,
          "unique": false,
          "readonly": true
        },
        {
          "name": "string_test",
          "type": "string",
          "nullable": false,
          "unique": false,
          "readonly": false
        },
        {
          "name": "integer_test",
          "type": "integer",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "decimal_test",
          "type": "decimal",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "boolean_test",
          "type": "boolean",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "date_test",
          "type": "datetime",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "json_test",
          "type": "json",
          "nullable": true,
          "unique": false,
          "readonly": false
        }
      ]
    }
  ]
}
```

### Create Record

Insert one record that exercises every supported field type.

```bash
curl -s -X POST "http://localhost:6000/data/typed_items:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "string_test": "sample",
            "integer_test": 42,
            "decimal_test": "9.99",
            "boolean_test": true,
            "date_test": "2024-01-01T00:00:00Z",
            "json_test": {
              "key": "value"
            }
          }
        ]
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "message": "Resource created successfully",
  "data": [
    {
      "boolean_test": true,
      "date_test": "2024-01-01 00:00:00 +0000 UTC",
      "decimal_test": "9.99",
      "id": "01KK92XEWYF61XCZCQ7GFB20B3",
      "integer_test": 42,
      "json_test": {
        "key": "value"
      },
      "string_test": "sample"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Query Record

Fetch the records

```bash
curl -s -X GET "http://localhost:6000/data/typed_items:query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "boolean_test": true,
      "date_test": "2024-01-01 00:00:00 +0000 UTC",
      "decimal_test": "9.99",
      "id": "01KK92XEWYF61XCZCQ7GFB20B3",
      "integer_test": 42,
      "json_test": {
        "key": "value"
      },
      "string_test": "sample"
    }
  ],
  "meta": {
    "count": 1,
    "current_page": 1,
    "per_page": 15,
    "total": 1,
    "total_pages": 1
  },
  "links": {
    "first": "/data/typed_items:query?page=1&per_page=15",
    "last": "/data/typed_items:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```
