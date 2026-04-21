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
      "id": "01KPR8T3SVZXSKS028C8TA420F",
      "quantity": 5,
      "title": "Product 1"
    },
    {
      "id": "01KPR8T3SV69F89Z2E4K7K4PXM",
      "quantity": 10,
      "title": "Product 2"
    },
    {
      "id": "01KPR8T3SVC9T082BH23RDJA6T",
      "quantity": 20,
      "title": "Product 3"
    },
    {
      "id": "01KPR8T3SWFY4XTVPZFH0WTS7A",
      "quantity": 55,
      "title": "Product 4"
    },
    {
      "id": "01KPR8T3SW9M9PV21JDSVDSQNM",
      "quantity": 56,
      "title": "Product 5"
    },
    {
      "id": "01KPR8T3SWV90NG8K3DXTV8EAW",
      "quantity": 5,
      "title": "Product 6"
    },
    {
      "id": "01KPR8T3SWGG60G31N0FYVBG2D",
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
            "id": "01KPR8T3SVZXSKS028C8TA420F",
            "quantity": 1200,
            "title": "Updated Product 1"
          },
          {
            "id": "01KPR8T3SV69F89Z2E4K7K4PXM",
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
      "id": "01KPR8T3SVZXSKS028C8TA420F",
      "quantity": 1200,
      "title": "Updated Product 1"
    },
    {
      "id": "01KPR8T3SV69F89Z2E4K7K4PXM",
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
            "id": "01KPR8T3SVZXSKS028C8TA420F"
          },
          {
            "id": "01KPR8T3SV69F89Z2E4K7K4PXM"
          },
          {
            "id": "01KPR8T3SVC9T082BH23RDJA6T"
          },
          {
            "id": "01KPR8T3SWFY4XTVPZFH0WTS7A"
          },
          {
            "id": "01KPR8T3SW9M9PV21JDSVDSQNM"
          },
          {
            "id": "01KPR8T3SWV90NG8K3DXTV8EAW"
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
