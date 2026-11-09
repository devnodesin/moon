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
      "id": "01M6Z8R5XSR84DTDKQF8C6C5CJ",
      "quantity": 5,
      "title": "Product 1"
    },
    {
      "id": "01M6Z8R5XTRSHA6B90KJTJS6JD",
      "quantity": 10,
      "title": "Product 2"
    },
    {
      "id": "01M6Z8R5XT61D43HW0FNZVWN9T",
      "quantity": 20,
      "title": "Product 3"
    },
    {
      "id": "01M6Z8R5XVG4RR8FA4Q2VWNFB9",
      "quantity": 55,
      "title": "Product 4"
    },
    {
      "id": "01M6Z8R5XVCD7SXEE1PGNAGEMB",
      "quantity": 56,
      "title": "Product 5"
    },
    {
      "id": "01M6Z8R5XVXSKT7JV8Q3KN4SQS",
      "quantity": 5,
      "title": "Product 6"
    },
    {
      "id": "01M6Z8R5XVFVF9GY18695BXB42",
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
            "id": "01M6Z8R5XSR84DTDKQF8C6C5CJ",
            "quantity": 1200,
            "title": "Updated Product 1"
          },
          {
            "id": "01M6Z8R5XTRSHA6B90KJTJS6JD",
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
      "id": "01M6Z8R5XSR84DTDKQF8C6C5CJ",
      "quantity": 1200,
      "title": "Updated Product 1"
    },
    {
      "id": "01M6Z8R5XTRSHA6B90KJTJS6JD",
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
            "id": "01M6Z8R5XSR84DTDKQF8C6C5CJ"
          },
          {
            "id": "01M6Z8R5XTRSHA6B90KJTJS6JD"
          },
          {
            "id": "01M6Z8R5XT61D43HW0FNZVWN9T"
          },
          {
            "id": "01M6Z8R5XVG4RR8FA4Q2VWNFB9"
          },
          {
            "id": "01M6Z8R5XVCD7SXEE1PGNAGEMB"
          },
          {
            "id": "01M6Z8R5XVXSKT7JV8Q3KN4SQS"
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
