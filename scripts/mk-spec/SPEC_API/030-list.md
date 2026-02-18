## Standard Response Pattern for `:list` Endpoints

List endpoints return paginated collections of resources. All list endpoints share a consistent request/response pattern described in this section.

**Applicable Endpoints:**

- List Users: `GET /users:list`
- List API Keys: `GET /apikeys:list`
- List Collections: `GET /collections:list`
- List Collection Records: `GET /{collection_name}:list`

**Response Structure:**
Every list endpoint returns a JSON object with two top-level keys: `data` and `meta`.

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "title": "Wireless Mouse",
      "price": "29.99"
    }
  ],
  "meta": {
    "count": 15,
    "limit": 15,
    "next": "01KHCZKMM0N808MKSHBNWF464F",
    "prev": "01KHCZFXAFJPS9SKSFKNBMHTP5"
  }
}
```

- `data` - An array of resource objects. Each record always includes an `id` field (ULID), except collections which use `name` as the identifier.
- `meta` - Pagination metadata for the current page.
  - `count` (integer): Number of records returned in this response
  - `limit` (integer): The page size limit that was applied. Default is 15; maximum allowed is 100.
  - `next` (string | null): Cursor pointing to the last record on the current page. Pass to ?after to get the next page. null on the last page.
  - `prev` (string | null): Cursor pointing to the record before the current page. Pass to ?after to return to the previous page. null on the first page.

---

The `:list` endpoint supports the following query parameters: `limit`, `after`, `sort`, `filter`, `search`, and `field selection`.

### Pagination

For pagination use parameter `?after={cursor}` to return records after the specified ULID cursor. Omit this parameter to start from the first page.

This API uses cursor-based pagination. Each response includes `meta.next` and `meta.prev` cursors, both of which are used with the `?after` parameter.

```sh
# First page (no cursor needed)
GET /products:list

# Next page — use meta.next, meta.prev, or any valid record id from the previous response
GET /products:list?after=01KHCZKMM0N808MKSHBNWF464F
```

**Notes:**

- `meta.prev` is `null` on the first page and `meta.next` is `null` on the last page.
- Records are always returned in chronological order (by ULID/creation time).
- For `?after={cursor}`, the cursor must always be a record's id (ULID). It can be:
  - A valid id of an existing record,
  - The value of `meta.prev` from the current response,
  - The value of `meta.next` from the current response.
- When `?after={cursor}` is used, only records that follow the specified id (ULID) are returned; the record matching the cursor is excluded from the results.
- If an invalid or non-existent cursor is provided, return an error response as specified in the [Standard Error Response](#standard-error-response) section.

### Limit

Use the query option `?limit={number}` to set the number of records returned per page. The default is 15; the maximum is 100.

```sh
GET /products:list?limit=2
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHCZKSPHB01TBEWKYQDKG5KS",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KHCZKSPHB01TBEWKYQDKG5KS",
    "prev": null
  }
}
```

### Filtering

Filter results by column value using the syntax `?{column_name}[operator]=value`. You can combine multiple filters in a single request.

**Supported operators:**

- `eq`: Equal to
- `ne`: Not equal to
- `gt`: Greater than
- `lt`: Less than
- `gte`: Greater than or equal to
- `lte`: Less than or equal to
- `like`: Pattern match. Use `%` as a wildcard, e.g. `brand[like]=Wo%`
- `in`: Matches any value in a comma-separated list, e.g. `brand[in]=Wow,Orange`

**Example:**

```sh
GET /products:list?quantity[gt]=5&brand[eq]=Wow
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHCZKT086EEB3EKM3PZ3N2Q0",
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

Use `?sort={field1,-field2,...}` to sort by one or more fields. Prefix a field name with `-` for descending order. Separate multiple fields with commas.

```sh
GET /products:list?sort=-quantity,title
```

Above sorts by `quantity` descending, then by `title` ascending.

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHCZKSPHB01TBEWKYQDKG5KS",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHCZKT086EEB3EKM3PZ3N2Q0",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
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

Use `?q` to search across all string and text fields in the collection.

```sh
GET /products:list?q=mouse
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
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

Return only the fields you need. `id` is always included.

```sh
GET /products:list?fields=quantity,title
```

**Response (200 OK):**

```json
{
  "data": [
    { "id": "01KHCZKSBQV1KH69AA6PVS12MM", "quantity": 10, "title": "Wireless Mouse" },
    { "id": "01KHCZKSPHB01TBEWKYQDKG5KS", "quantity": 55, "title": "USB Keyboard" },
    { "id": "01KHCZKT086EEB3EKM3PZ3N2Q0", "quantity": 20, "title": "Monitor 21 inch" }
  ],
  "meta": {
    "count": 3,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Combined Examples

All query parameters can be combined in a single request.

```sh
# Filter by price range, sort descending, limit results
GET /products:list?quantity[gte]=10&price[lt]=100&sort=-price&limit=5

# Full-text search with a brand filter, returning only select fields
GET /products:list?q=laptop&brand[eq]=Wow&fields=title,price,quantity

# Multi-filter with pagination
GET /products:list?price[gte]=100&quantity[gt]=0&sort=-price&limit=10&after=01KHCZKMM0N808MKSHBNWF464F
```

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling

