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
      "id": "01M6Z8R6AS0RCVKD22C1ZB3011",
      "quantity": 1,
      "title": "Product 1"
    },
    {
      "id": "01M6Z8R6ASGQXR7EEX2KYF5SJJ",
      "quantity": 2,
      "title": "Product 2"
    },
    {
      "id": "01M6Z8R6AS9VM934RRPJSSB21J",
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
      "id": "01M6Z8R6ATY265DP16NS5YJ8G9",
      "quantity": 4,
      "title": "Product 4"
    },
    {
      "id": "01M6Z8R6ATD2PR06ZW2WD06M0H",
      "quantity": 5,
      "title": "Product 5"
    },
    {
      "id": "01M6Z8R6ATXRESKC5CD3VCVVCK",
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
      "id": "01M6Z8R6ATAXFKYS144YR5JQR2",
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
