### Create Records (Batch)

Create multiple records in a single request.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "title": "Product 1",
            "quantity": 5
          },
          {
            "title": "Product 2",
            "quantity": 10
          },
          {
            "title": "Product 3",
            "quantity": 20
          },
          {
            "title": "Product 4",
            "quantity": 55
          },
          {
            "title": "Product 5",
            "quantity": 56
          },
          {
            "title": "Product 6",
            "quantity": 5
          },
          {
            "title": "Product 7",
            "quantity": 12
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
      "id": "01KK7522G6DX66VDCTZ5JYTQJH",
      "quantity": 5,
      "title": "Product 1"
    },
    {
      "id": "01KK7522G6CAC4V9BS0Z26WXG3",
      "quantity": 10,
      "title": "Product 2"
    },
    {
      "id": "01KK7522G662T7P7EAKP8409J4",
      "quantity": 20,
      "title": "Product 3"
    },
    {
      "id": "01KK7522G7WKQ9CCCGK2Z19JBM",
      "quantity": 55,
      "title": "Product 4"
    },
    {
      "id": "01KK7522G76KDCYQFF7QBPFZCS",
      "quantity": 56,
      "title": "Product 5"
    },
    {
      "id": "01KK7522G7KHAEGAV2Q1DG85T0",
      "quantity": 5,
      "title": "Product 6"
    },
    {
      "id": "01KK7522G7F8KSFE27C24WN8QX",
      "quantity": 12,
      "title": "Product 7"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 7
  }
}
```

### Update Records (Batch)

Update multiple records in a single request using numbered ULID placeholders.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "id": "01KK7522G6DX66VDCTZ5JYTQJH",
            "quantity": 1200,
            "title": "Updated Product 1"
          },
          {
            "id": "01KK7522G6CAC4V9BS0Z26WXG3",
            "title": "Updated Product 2"
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
      "id": "01KK7522G6DX66VDCTZ5JYTQJH",
      "quantity": 1200,
      "title": "Updated Product 1"
    },
    {
      "id": "01KK7522G6CAC4V9BS0Z26WXG3",
      "quantity": 10,
      "title": "Updated Product 2"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 2
  }
}
```

### Destroy Records (Batch)

Delete multiple records in a single request.

```bash
curl -s -X POST "http://localhost:6000/data/products:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "destroy",
        "data": [
          {
            "id": "01KK7522G6DX66VDCTZ5JYTQJH"
          },
          {
            "id": "01KK7522G6CAC4V9BS0Z26WXG3"
          },
          {
            "id": "01KK7522G662T7P7EAKP8409J4"
          },
          {
            "id": "01KK7522G7WKQ9CCCGK2Z19JBM"
          },
          {
            "id": "01KK7522G76KDCYQFF7QBPFZCS"
          },
          {
            "id": "01KK7522G7KHAEGAV2Q1DG85T0"
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
    "success": 6
  }
}
```
