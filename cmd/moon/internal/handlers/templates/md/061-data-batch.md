### Create Records (Batch)

```bash
curl -s -X POST "http://localhost:6006/products:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
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
  "data": [
    {
      "id": "01KJMQ40VQK0XXHWYDD0E8QZEP",
      "quantity": 5,
      "title": "Product 1"
    },
    {
      "id": "01KJMQ40WQX4JHRDTHMXZZVMM1",
      "quantity": 10,
      "title": "Product 2"
    },
    {
      "id": "01KJMQ40WW1GFHBXN5C41ZWR1C",
      "quantity": 20,
      "title": "Product 3"
    },
    {
      "id": "01KJMQ40X2ZFZYM7JRXJ9FVVDC",
      "quantity": 55,
      "title": "Product 4"
    },
    {
      "id": "01KJMQ40X78J1ZYNVR02KPYGFN",
      "quantity": 56,
      "title": "Product 5"
    },
    {
      "id": "01KJMQ40XCKBFYPJDWYY523EM7",
      "quantity": 5,
      "title": "Product 6"
    },
    {
      "id": "01KJMQ40XJAXJQT11WCNDWFMMG",
      "quantity": 12,
      "title": "Product 7"
    }
  ],
  "message": "7 record(s) created successfully",
  "meta": {
    "failed": 0,
    "succeeded": 7,
    "total": 7
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
            "id": "01KJMQ40VQK0XXHWYDD0E8QZEP",
            "quantity": 1200,
            "title": "Updated Product 1"
          },
          {
            "id": "01KJMQ40WQX4JHRDTHMXZZVMM1",
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
      "id": "01KJMQ40VQK0XXHWYDD0E8QZEP",
      "quantity": 1200,
      "title": "Updated Product 1"
    },
    {
      "id": "01KJMQ40WQX4JHRDTHMXZZVMM1",
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

### Destroy Records (Batch)

```bash
curl -s -X POST "http://localhost:6006/products:destroy" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": [
          "01KJMQ40VQK0XXHWYDD0E8QZEP",
          "01KJMQ40WQX4JHRDTHMXZZVMM1",
          "01KJMQ40WW1GFHBXN5C41ZWR1C",
          "01KJMQ40X2ZFZYM7JRXJ9FVVDC",
          "01KJMQ40X78J1ZYNVR02KPYGFN",
          "01KJMQ40XCKBFYPJDWYY523EM7"
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    "01KJMQ40VQK0XXHWYDD0E8QZEP",
    "01KJMQ40WQX4JHRDTHMXZZVMM1",
    "01KJMQ40WW1GFHBXN5C41ZWR1C",
    "01KJMQ40X2ZFZYM7JRXJ9FVVDC",
    "01KJMQ40X78J1ZYNVR02KPYGFN",
    "01KJMQ40XCKBFYPJDWYY523EM7"
  ],
  "message": "6 record(s) deleted successfully",
  "meta": {
    "failed": 0,
    "succeeded": 6,
    "total": 6
  }
}
```
