### Filtering

**Query Option:** `?column[operator]=value`

**Operators:** eq, ne, gt, lt, gte, lte, like, in

```bash
curl -s -X GET "http://localhost:6006/products:list?quantity[gt]=5&brand[eq]=Wow" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHNM1QV3XBT15NAY0MBGB2TF",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHNM1R4SCJWN4V3RK5GV48R2",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Sorting

**Query Option:** `?sort={-field1,field2}`

Sort by `field` (ascending) or `-field` (descending).

```bash
curl -s -X GET "http://localhost:6006/products:list?sort=-quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHNM1R07TDVJH8JJZKWV36NT",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHNM1R4SCJWN4V3RK5GV48R2",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHNM1QV3XBT15NAY0MBGB2TF",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
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

### Full-Text Search

**Query Option:** `?q={search_term}` (across all text columns)

Searches across all string/text fields in the collection.

```bash
curl -s -X GET "http://localhost:6006/products:list?q=mouse" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHNM1QV3XBT15NAY0MBGB2TF",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Field Selection

**Query Option:** `?fields={field1,field2}`

Returns only the specified fields (plus `id` which is always included).

```bash
curl -s -X GET "http://localhost:6006/products:list?fields=quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KHNM1QV3XBT15NAY0MBGB2TF",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "id": "01KHNM1R07TDVJH8JJZKWV36NT",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "id": "01KHNM1R4SCJWN4V3RK5GV48R2",
      "quantity": 20,
      "title": "Monitor 21 inch"
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

### Limit

**Query Option:** `?limit={limit}`

```bash
curl -s -X GET "http://localhost:6006/products:list?limit=2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHNM1QV3XBT15NAY0MBGB2TF",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHNM1R07TDVJH8JJZKWV36NT",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KHNM1R07TDVJH8JJZKWV36NT",
    "prev": null
  }
}
```

### Pagination

**Query Option:** `?after={cursor}`

 (Response includes `next_cursor` when more results are available.)

```bash
curl -s -X GET "http://localhost:6006/products:list?after=01KHNM1QV3XBT15NAY0MBGB2TF&limit=1" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHNM1R07TDVJH8JJZKWV36NT",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 1,
    "next": "01KHNM1R07TDVJH8JJZKWV36NT",
    "prev": null
  }
}
```
