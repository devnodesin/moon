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
    "total": 0
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
        "data": {
          "title": "Wireless Mouse",
          "price": "29.99",
          "details": "Ergonomic wireless mouse",
          "quantity": 10,
          "brand": "Wow"
        }
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": {
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
    "price": "29.99",
    "quantity": 10,
    "title": "Wireless Mouse"
  },
  "message": "Record created successfully with id 01KHNM1NFT0ZX9ZQYHSZHW5KPG"
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

**Response (207 Multi-Status):**

```json
{
  "results": [
    {
      "index": 0,
      "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
      "status": "created",
      "data": {
        "brand": "KeyPro",
        "details": "Mechanical keyboard",
        "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
        "price": "49.99",
        "quantity": 5,
        "title": "Keyboard"
      }
    },
    {
      "index": 1,
      "id": "01KHNM1NN1FEKGPRVQ01RPRY6Q",
      "status": "created",
      "data": {
        "brand": "ViewMax",
        "details": "24-inch FHD monitor",
        "id": "01KHNM1NN1FEKGPRVQ01RPRY6Q",
        "price": "199.99",
        "quantity": 2,
        "title": "Monitor"
      }
    }
  ],
  "summary": {
    "total": 2,
    "succeeded": 2,
    "failed": 0
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
      "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "KeyPro",
      "details": "Mechanical keyboard",
      "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
      "price": "49.99",
      "quantity": 5,
      "title": "Keyboard"
    },
    {
      "brand": "ViewMax",
      "details": "24-inch FHD monitor",
      "id": "01KHNM1NN1FEKGPRVQ01RPRY6Q",
      "price": "199.99",
      "quantity": 2,
      "title": "Monitor"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Get Single Record

```bash
curl -s -X GET "http://localhost:6006/products:get?id=01KHNM1NFT0ZX9ZQYHSZHW5KPG" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
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
        "data": {
          "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
          "price": "6000.00"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
    "price": "6000.00"
  },
  "message": "Record 01KHNM1NFT0ZX9ZQYHSZHW5KPG updated successfully"
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
            "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
            "price": "100.00",
            "title": "Updated Product 1"
          },
          {
            "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
            "price": "200.00",
            "title": "Updated Product 2"
          }
        ]
      }
    ' | jq .
```

**Response (207 Multi-Status):**

```json
{
  "results": [
    {
      "index": 0,
      "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
      "status": "updated",
      "data": {
        "id": "01KHNM1NFT0ZX9ZQYHSZHW5KPG",
        "price": "100.00",
        "title": "Updated Product 1"
      }
    },
    {
      "index": 1,
      "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
      "status": "updated",
      "data": {
        "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
        "price": "200.00",
        "title": "Updated Product 2"
      }
    }
  ],
  "summary": {
    "total": 2,
    "succeeded": 2,
    "failed": 0
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
        "data": "01KHNM1NFT0ZX9ZQYHSZHW5KPG"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    "01KHNM1NFT0ZX9ZQYHSZHW5KPG"
  ],
  "message": "Record 01KHNM1NFT0ZX9ZQYHSZHW5KPG deleted successfully"
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
          "01KHNM1NMXTRYD4ZQ0MCF17J95",
          "01KHNM1NN1FEKGPRVQ01RPRY6Q"
        ]
      }
    ' | jq .
```

**Response (207 Multi-Status):**

```json
{
  "results": [
    {
      "index": 0,
      "id": "01KHNM1NMXTRYD4ZQ0MCF17J95",
      "status": "deleted"
    },
    {
      "index": 1,
      "id": "01KHNM1NN1FEKGPRVQ01RPRY6Q",
      "status": "deleted"
    }
  ],
  "summary": {
    "total": 2,
    "succeeded": 2,
    "failed": 0
  }
}
```
