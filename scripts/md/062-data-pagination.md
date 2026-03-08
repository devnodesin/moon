### First Page

Retrieve the first page (3 records per page). The response includes `meta` with pagination info and `links` for navigation.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?per_page=3&page=1" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KK7526EFZC27XMZ4SCPTXSYC",
      "quantity": 1,
      "title": "Product 1"
    },
    {
      "id": "01KK7526EF3YBYTEJGSFFAHM0R",
      "quantity": 2,
      "title": "Product 2"
    },
    {
      "id": "01KK7526EFBADT0GRNNESV62ZS",
      "quantity": 3,
      "title": "Product 3"
    }
  ],
  "meta": {
    "count": 3,
    "current_page": 1,
    "per_page": 3,
    "total": 7,
    "total_pages": 3
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=3",
    "last": "/data/products:query?page=3&per_page=3",
    "next": "/data/products:query?page=2&per_page=3",
    "prev": null
  }
}
```

### Second Page

Retrieve the second page.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?per_page=3&page=2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KK7526EFXT7H8GF3N6VYAVRJ",
      "quantity": 4,
      "title": "Product 4"
    },
    {
      "id": "01KK7526EFGS3AWQWNXQW0XHB7",
      "quantity": 5,
      "title": "Product 5"
    },
    {
      "id": "01KK7526EFBKVNSY730FYJ2819",
      "quantity": 6,
      "title": "Product 6"
    }
  ],
  "meta": {
    "count": 3,
    "current_page": 2,
    "per_page": 3,
    "total": 7,
    "total_pages": 3
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=3",
    "last": "/data/products:query?page=3&per_page=3",
    "next": "/data/products:query?page=3&per_page=3",
    "prev": "/data/products:query?page=1&per_page=3"
  }
}
```

### Last Page

Retrieve the last (third) page. Only one record is returned.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?per_page=3&page=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KK7526EGE8YCY2JNX656AZHG",
      "quantity": 7,
      "title": "Product 7"
    }
  ],
  "meta": {
    "count": 1,
    "current_page": 3,
    "per_page": 3,
    "total": 7,
    "total_pages": 3
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=3",
    "last": "/data/products:query?page=3&per_page=3",
    "next": null,
    "prev": "/data/products:query?page=2&per_page=3"
  }
}
```
