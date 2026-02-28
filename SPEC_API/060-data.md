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
      "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
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

### Create Records (Batch)

```bash
curl -s -X POST "http://localhost:6006/products:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          {
            "title": "Keyboard",
            "price": "49.99",
            "details": "Mechanical keyboard",
            "quantity": 5,
            "brand": "KeyPro"
          },
          {
            "title": "Monitor",
            "price": "199.99",
            "details": "24-inch FHD monitor",
            "quantity": 2,
            "brand": "ViewMax"
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
      "brand": "KeyPro",
      "details": "Mechanical keyboard",
      "id": "01KJHCX4QQTBQJPQ91GRMENRD6",
      "price": "49.99",
      "quantity": 5,
      "title": "Keyboard"
    },
    {
      "brand": "ViewMax",
      "details": "24-inch FHD monitor",
      "id": "01KJHCX4QXAM6XFS6Y8TV9MD6J",
      "price": "199.99",
      "quantity": 2,
      "title": "Monitor"
    }
  ],
  "message": "2 record(s) created successfully",
  "meta": {
    "failed": 0,
    "succeeded": 2,
    "total": 2
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
      "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "KeyPro",
      "details": "Mechanical keyboard",
      "id": "01KJHCX4QQTBQJPQ91GRMENRD6",
      "price": "49.99",
      "quantity": 5,
      "title": "Keyboard"
    },
    {
      "brand": "ViewMax",
      "details": "24-inch FHD monitor",
      "id": "01KJHCX4QXAM6XFS6Y8TV9MD6J",
      "price": "199.99",
      "quantity": 2,
      "title": "Monitor"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 15,
    "next": null,
    "prev": null,
    "total": 3
  }
}
```

### Get Single Record

```bash
curl -s -X GET "http://localhost:6006/products:get?id=01KJHCX4EF5SWJ7WGEJCQXTB87" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
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
            "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
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
      "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
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

### Update Records (Batch)

```bash
curl -s -X POST "http://localhost:6006/products:update" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          {
            "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
            "price": "100.00",
            "title": "Updated Product 1"
          },
          {
            "id": "01KJHCX4QQTBQJPQ91GRMENRD6",
            "price": "200.00",
            "title": "Updated Product 2"
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
      "id": "01KJHCX4EF5SWJ7WGEJCQXTB87",
      "price": "100.00",
      "title": "Updated Product 1"
    },
    {
      "id": "01KJHCX4QQTBQJPQ91GRMENRD6",
      "price": "200.00",
      "title": "Updated Product 2"
    }
  ],
  "message": "2 record(s) updated successfully",
  "meta": {
    "failed": 0,
    "succeeded": 2,
    "total": 2
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
          "01KJHCX4EF5SWJ7WGEJCQXTB87"
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    "01KJHCX4EF5SWJ7WGEJCQXTB87"
  ],
  "message": "1 record(s) deleted successfully",
  "meta": {
    "failed": 0,
    "succeeded": 1,
    "total": 1
  }
}
```

### Destroy Records (Batch)

```bash
curl -s -X POST "http://localhost:6006/products:destroy" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          "01KJHCX4QQTBQJPQ91GRMENRD6",
          "01KJHCX4QXAM6XFS6Y8TV9MD6J"
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    "01KJHCX4QQTBQJPQ91GRMENRD6",
    "01KJHCX4QXAM6XFS6Y8TV9MD6J"
  ],
  "message": "2 record(s) deleted successfully",
  "meta": {
    "failed": 0,
    "succeeded": 2,
    "total": 2
  }
}
```
