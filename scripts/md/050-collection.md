### Create Collection

Create a new collection named `products` with typed columns.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "name": "products",
            "columns": [
              {
                "name": "title",
                "type": "string",
                "nullable": false,
                "unique": true
              },
              {
                "name": "price",
                "type": "integer",
                "nullable": false
              },
              {
                "name": "description",
                "type": "string",
                "nullable": true
              },
              {
                "name": "category",
                "type": "string",
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
          "name": "title",
          "nullable": false,
          "type": "string",
          "unique": true
        },
        {
          "name": "price",
          "nullable": false,
          "type": "integer",
          "unique": false
        },
        {
          "name": "description",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "category",
          "nullable": true,
          "type": "string",
          "unique": false
        }
      ],
      "name": "products"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Create Collection

Create a new collection named `category` with typed columns.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "name": "category",
            "columns": [
              {
                "name": "title",
                "type": "string",
                "nullable": false,
                "unique": true
              },
              {
                "name": "description",
                "type": "string",
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
          "name": "title",
          "nullable": false,
          "type": "string",
          "unique": true
        },
        {
          "name": "description",
          "nullable": true,
          "type": "string",
          "unique": false
        }
      ],
      "name": "category"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### List Collections

Retrieve all user-defined collections.

```bash
curl -s -X GET "http://localhost:6000/collections:query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collections retrieved successfully",
  "data": [
    {
      "count": 0,
      "name": "apikeys",
      "system": true
    },
    {
      "count": 0,
      "name": "category",
      "system": false
    },
    {
      "count": 0,
      "name": "products",
      "system": false
    },
    {
      "count": 1,
      "name": "users",
      "system": true
    }
  ],
  "meta": {
    "count": 4,
    "current_page": 1,
    "per_page": 15,
    "total": 4,
    "total_pages": 1
  },
  "links": {
    "first": "/collections:query?page=1&per_page=15",
    "last": "/collections:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Get Collection

Retrieve metadata for a specific collection.

```bash
curl -s -X GET "http://localhost:6000/collections:query?name=products" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection retrieved successfully",
  "data": [
    {
      "count": 0,
      "name": "products",
      "system": false
    }
  ]
}
```

### Update Collection — Add Column

Add a new `stock` column to the `products` collection.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "name": "products",
            "add_columns": [
              {
                "name": "stock",
                "type": "integer",
                "nullable": true
              }
            ]
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection updated successfully",
  "data": [
    {
      "columns": [
        {
          "name": "title",
          "nullable": false,
          "type": "string",
          "unique": true
        },
        {
          "name": "price",
          "nullable": false,
          "type": "integer",
          "unique": false
        },
        {
          "name": "description",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "category",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "stock",
          "nullable": true,
          "type": "integer",
          "unique": false
        }
      ],
      "name": "products"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Update Collection — Rename Column

Rename `description` to `details`.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "name": "products",
            "rename_columns": [
              {
                "old_name": "description",
                "new_name": "details"
              }
            ]
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection updated successfully",
  "data": [
    {
      "columns": [
        {
          "name": "title",
          "nullable": false,
          "type": "string",
          "unique": true
        },
        {
          "name": "price",
          "nullable": false,
          "type": "integer",
          "unique": false
        },
        {
          "name": "details",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "category",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "stock",
          "nullable": true,
          "type": "integer",
          "unique": false
        }
      ],
      "name": "products"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Update Collection — Remove Column

Remove the `category` column.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "name": "products",
            "remove_columns": [
              "category"
            ]
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection updated successfully",
  "data": [
    {
      "columns": [
        {
          "name": "title",
          "nullable": false,
          "type": "string",
          "unique": true
        },
        {
          "name": "price",
          "nullable": false,
          "type": "integer",
          "unique": false
        },
        {
          "name": "details",
          "nullable": true,
          "type": "string",
          "unique": false
        },
        {
          "name": "stock",
          "nullable": true,
          "type": "integer",
          "unique": false
        }
      ],
      "name": "products"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Delete Collection

Permanently delete the `products` collection and all its data.

```bash
curl -s -X POST "http://localhost:6000/collections:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "destroy",
        "data": [
          {
            "name": "products"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection destroyed successfully",
  "data": [
    {
      "name": "products"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```
