### Get Schema

```bash
curl -s -X GET "http://localhost:6006/products:schema" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "collection": "products",
    "fields": [
      {
        "name": "id",
        "type": "string",
        "nullable": false,
        "readonly": true
      },
      {
        "name": "title",
        "type": "string",
        "nullable": false
      },
      {
        "name": "price",
        "type": "decimal",
        "nullable": false
      },
      {
        "name": "details",
        "type": "string",
        "nullable": true,
        "default": "''"
      },
      {
        "name": "quantity",
        "type": "integer",
        "nullable": true,
        "default": "0"
      },
      {
        "name": "brand",
        "type": "string",
        "nullable": true,
        "default": "''"
      }
    ],
    "total": 6
  }
}
```

### Create Record (Single)

```bash
curl -s -X POST "http://localhost:6006/products:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          {
            "title": "Wireless Mouse",
            "price": "29.99",
            "details": "Ergonomic wireless mouse",
            "quantity": 10,
            "brand": "Wow"
          }
        ]
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "message": "1 record(s) created successfully",
  "meta": {
    "failed": 0,
    "succeeded": 1,
    "total": 1
  }
}
```

### Get All Records

```bash
curl -s -X GET "http://localhost:6006/products:list" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 15,
    "next": null,
    "prev": null,
    "total": 1
  }
}
```

### Get Single Record

```bash
curl -s -X GET "http://localhost:6006/products:get?id=01KJMQ3XZF5H1P2DDNGWGVXB5T" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
    "price": "29.99",
    "quantity": 10,
    "title": "Wireless Mouse"
  }
}
```

### Update Existing Record (Single)

```bash
curl -s -X POST "http://localhost:6006/products:update" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          {
            "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
            "price": "6000.00"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "price": "6000.00"
    }
  ],
  "message": "1 record(s) updated successfully",
  "meta": {
    "failed": 0,
    "succeeded": 1,
    "total": 1
  }
}
```

### Delete Record

```bash
curl -s -X POST "http://localhost:6006/products:destroy" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          "01KJMQ3XZF5H1P2DDNGWGVXB5T"
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    "01KJMQ3XZF5H1P2DDNGWGVXB5T"
  ],
  "message": "1 record(s) deleted successfully",
  "meta": {
    "failed": 0,
    "succeeded": 1,
    "total": 1
  }
}
```
