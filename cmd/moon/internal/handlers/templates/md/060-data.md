### Get Schema

```bash
curl -s -X GET "http://localhost:6006/products:schema" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
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
    "id": "01KHCZKMM0N808MKSHBNWF464F",
    "price": "29.99",
    "quantity": 10,
    "title": "Wireless Mouse"
  },
  "message": "Record created successfully with id 01KHCZKMM0N808MKSHBNWF464F"
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
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "status": "created",
      "data": {
        "brand": "KeyPro",
        "details": "Mechanical keyboard",
        "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
        "price": "49.99",
        "quantity": 5,
        "title": "Keyboard"
      }
    },
    {
      "index": 1,
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "status": "created",
      "data": {
        "brand": "ViewMax",
        "details": "24-inch FHD monitor",
        "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
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
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "KeyPro",
      "details": "Mechanical keyboard",
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "price": "49.99",
      "quantity": 5,
      "title": "Keyboard"
    },
    {
      "brand": "ViewMax",
      "details": "24-inch FHD monitor",
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "price": "199.99",
      "quantity": 2,
      "title": "Monitor"
    }
  ],
  "total": 3,
  "next_cursor": null,
  "limit": 15
}
```

### Get Single Record

```bash
curl -s -X GET "http://localhost:6006/products:get?id=01KHCZKMM0N808MKSHBNWF464F" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "id": "01KHCZKMM0N808MKSHBNWF464F",
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
        "id": "01KHCZKMM0N808MKSHBNWF464F",
        "data": {
          "price": "6000.00"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZKMM0N808MKSHBNWF464F",
    "price": "6000.00"
  },
  "message": "Record 01KHCZKMM0N808MKSHBNWF464F updated successfully"
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
            "id": "01KHCZKMM0N808MKSHBNWF464F",
            "price": "100.00",
            "title": "Updated Product 1"
          },
          {
            "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
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
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "status": "updated",
      "data": {
        "id": "01KHCZKMM0N808MKSHBNWF464F",
        "price": "100.00",
        "title": "Updated Product 1"
      }
    },
    {
      "index": 1,
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "status": "updated",
      "data": {
        "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
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
        "id": "01KHCZKMM0N808MKSHBNWF464F"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Record 01KHCZKMM0N808MKSHBNWF464F deleted successfully"
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
          "01KHCZKMXYVC1NRHDZ83XMHY4N",
          "01KHCZKMY28ERJFPCVBQEKQ4SY"
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
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "status": "deleted"
    },
    {
      "index": 1,
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
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
