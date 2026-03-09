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
                "name": "label",
                "type": "string",
                "nullable": false
              },
              {
                "name": "count",
                "type": "integer",
                "nullable": true
              },
              {
                "name": "price",
                "type": "decimal",
                "nullable": true
              },
              {
                "name": "active",
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
          "name": "label",
          "nullable": false,
          "type": "string",
          "unique": false
        },
        {
          "name": "count",
          "nullable": true,
          "type": "integer",
          "unique": false
        },
        {
          "name": "price",
          "nullable": true,
          "type": "decimal",
          "unique": false
        },
        {
          "name": "active",
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
          "name": "label",
          "type": "string",
          "nullable": false,
          "unique": false,
          "readonly": false
        },
        {
          "name": "count",
          "type": "integer",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "price",
          "type": "decimal",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "active",
          "type": "boolean",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "date_test",
          "type": "string",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "json_test",
          "type": "string",
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
            "label": "sample",
            "count": 42,
            "price": "9.99",
            "active": true,
            "date_test": "2024-01-01T00:00:00Z",
            "json_test": {
              "key": "value"
            }
          }
        ]
      }
    ' | jq .
```

**Response (400 Bad Request):**

```json
{
  "message": "Invalid value for field 'json_test' of type 'string'"
}
```

### Query Record

Fetch the record by its ULID. The response must return `active` as a boolean, `count` as an integer, `price` as a decimal string, and `meta` as a JSON object.

```bash
curl -s -X GET "http://localhost:6000/data/typed_items:query?id=$ULID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (404 Not Found):**

```json
{
  "message": "Resource not found"
}
```
