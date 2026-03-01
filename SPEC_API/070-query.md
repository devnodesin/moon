### Filtering

Filter results by column value using the syntax `?{column_name}[operator]=value`. You can combine multiple filters in a single request.

Supported filter operators: `eq` (equal to), `ne` (not equal to), `gt` (greater than), `lt` (less than), `gte` (greater than or equal to), `lte` (less than or equal to), `like` (pattern match, `%` is wildcard, e.g. `brand[like]=Wo%`), `in` (matches any value in a comma-separated list, e.g. `brand[in]=Wow,Orange`)

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
      "id": "01KJMQ46SPZXRYZ1WV87NP292N",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KJMQ47BYZE57GEWTE6C91MZ6",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 15,
    "next": null,
    "prev": null,
    "total": 2
  }
}
```

### Sorting

Use `?sort={field1,-field2,...}` to sort by one or more fields. Prefix a field name with `-` for descending order. Separate multiple fields with commas.

Sort by `field` (ascending) or `-field` (descending). Below sorts by `quantity` descending, then by `title` ascending.

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
      "id": "01KJMQ473FZH1HD427M2BMWEKN",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KJMQ47BYZE57GEWTE6C91MZ6",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KJMQ46SPZXRYZ1WV87NP292N",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
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
      "id": "01KJMQ46SPZXRYZ1WV87NP292N",
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
      "id": "01KJMQ46SPZXRYZ1WV87NP292N",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "id": "01KJMQ473FZH1HD427M2BMWEKN",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "id": "01KJMQ47BYZE57GEWTE6C91MZ6",
      "quantity": 20,
      "title": "Monitor 21 inch"
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

### Limit

**Query Option:** `?limit={limit}`

Use the query option `?limit={number}` to set the number of records returned per page. The default is 15; the maximum is 200.

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
      "id": "01KJMQ46SPZXRYZ1WV87NP292N",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KJMQ473FZH1HD427M2BMWEKN",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KJMQ473FZH1HD427M2BMWEKN",
    "prev": null,
    "total": 3
  }
}
```

### Pagination

**Query Option:** `?after={cursor}`

Response includes `next_cursor` when more results are available. (If `next_cursor` is present, use its value from the response to fetch subsequent pages.)

```bash
curl -s -X GET "http://localhost:6006/products:list?after=01KJMQ473FZH1HD427M2BMWEKN&limit=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KJMQ47BYZE57GEWTE6C91MZ6",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "SuperBrand",
      "details": "Mechanical keyboard",
      "id": "01KJMQ490T2SNMP317YKDDEEJS",
      "price": "49.99",
      "quantity": 5,
      "title": "Product 1"
    },
    {
      "brand": "SuperBrand",
      "details": "24-inch FHD monitor",
      "id": "01KJMQ4910F13Y86V9EFA6QKXX",
      "price": "199.99",
      "quantity": 2,
      "title": "Product 2"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 3,
    "next": "01KJMQ4910F13Y86V9EFA6QKXX",
    "prev": null,
    "total": 8
  }
}
```
