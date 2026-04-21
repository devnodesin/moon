### Get Schema

Retrieve the field schema for the `products` collection.

```bash
curl -s -X GET "http://localhost:6000/data/products:schema" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Schema retrieved successfully",
  "data": [
    {
      "name": "products",
      "fields": [
        {
          "name": "id",
          "type": "id",
          "nullable": true,
          "unique": false,
          "readonly": true
        },
        {
          "name": "title",
          "type": "string",
          "nullable": false,
          "unique": false,
          "readonly": false
        },
        {
          "name": "price",
          "type": "integer",
          "nullable": false,
          "unique": false,
          "readonly": false
        },
        {
          "name": "details",
          "type": "string",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "quantity",
          "type": "integer",
          "nullable": true,
          "unique": false,
          "readonly": false
        },
        {
          "name": "brand",
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

Create a single record in the `products` collection.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "title": "Wireless Mouse",
            "price": 29,
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
  "message": "Resource created successfully",
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KPR8T0YFNYKNZ265XBGS8XR3",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### List Records

Retrieve all records from the `products` collection.

```bash
curl -s -X GET "http://localhost:6000/data/products:query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KPR8T0YFNYKNZ265XBGS8XR3",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
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
    "first": "/data/products:query?page=1&per_page=15",
    "last": "/data/products:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Get Record by ID

Retrieve a single record by its ULID.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?id=01KPR8T0YFNYKNZ265XBGS8XR3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KPR8T0YFNYKNZ265XBGS8XR3",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ]
}
```

### Update Record

Update fields of an existing record.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "id": "01KPR8T0YFNYKNZ265XBGS8XR3",
            "price": 6000
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource updated successfully",
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KPR8T0YFNYKNZ265XBGS8XR3",
      "price": 6000,
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Delete Record

Delete a record by its ULID.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "destroy",
        "data": [
          {
            "id": "01KPR8T0YFNYKNZ265XBGS8XR3"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource destroyed successfully",
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```
