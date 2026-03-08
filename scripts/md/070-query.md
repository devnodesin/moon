### Filtering

Filter results using `?{field}[op]=value`. Supported operators: `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `in`.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?quantity[gt]=5&brand[eq]=Wow" \
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
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KK752CNPB98YQ8RJQPSSFQM3",
      "price": 199,
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Adjustable laptop stand",
      "id": "01KK752CNQ27A4QC6XK3CTVPTR",
      "price": 49,
      "quantity": 8,
      "title": "Laptop Stand"
    }
  ],
  "meta": {
    "count": 3,
    "current_page": 1,
    "per_page": 15,
    "total": 3,
    "total_pages": 1
  },
  "links": {
    "first": "/data/products:query?brand%5Beq%5D=Wow&page=1&per_page=15&quantity%5Bgt%5D=5",
    "last": "/data/products:query?brand%5Beq%5D=Wow&page=1&per_page=15&quantity%5Bgt%5D=5",
    "next": null,
    "prev": null
  }
}
```

### Sorting

Use `?sort={field1,-field2}` to sort. Prefix with `-` for descending. Multiple fields are comma-separated.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?sort=-quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KK752CNP351655QEGHA94QBS",
      "price": 19,
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KK752CNPB98YQ8RJQPSSFQM3",
      "price": 199,
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Adjustable laptop stand",
      "id": "01KK752CNQ27A4QC6XK3CTVPTR",
      "price": 49,
      "quantity": 8,
      "title": "Laptop Stand"
    },
    {
      "brand": "Orange",
      "details": "1080p webcam",
      "id": "01KK752CNQ154PYDD5QF0SWQYP",
      "price": 79,
      "quantity": 3,
      "title": "Webcam HD"
    }
  ],
  "meta": {
    "count": 5,
    "current_page": 1,
    "per_page": 15,
    "total": 5,
    "total_pages": 1
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=15&sort=-quantity%2Ctitle",
    "last": "/data/products:query?page=1&per_page=15&sort=-quantity%2Ctitle",
    "next": null,
    "prev": null
  }
}
```

### Full-Text Search

Use `?q={term}` to search across all string fields.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?q=mouse" \
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
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
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
    "first": "/data/products:query?page=1&per_page=15&q=mouse",
    "last": "/data/products:query?page=1&per_page=15&q=mouse",
    "next": null,
    "prev": null
  }
}
```

### Field Selection

Use `?fields={field1,field2}` to return only the specified columns (`id` is always included).

```bash
curl -s -X GET "http://localhost:6000/data/products:query?fields=quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "id": "01KK752CNP351655QEGHA94QBS",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "id": "01KK752CNPB98YQ8RJQPSSFQM3",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "id": "01KK752CNQ27A4QC6XK3CTVPTR",
      "quantity": 8,
      "title": "Laptop Stand"
    },
    {
      "id": "01KK752CNQ154PYDD5QF0SWQYP",
      "quantity": 3,
      "title": "Webcam HD"
    }
  ],
  "meta": {
    "count": 5,
    "current_page": 1,
    "per_page": 15,
    "total": 5,
    "total_pages": 1
  },
  "links": {
    "first": "/data/products:query?fields=quantity%2Ctitle&page=1&per_page=15",
    "last": "/data/products:query?fields=quantity%2Ctitle&page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Pagination

Use `?per_page={n}&page={p}` to paginate. Default `per_page` is 15; maximum is 200.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?per_page=2&page=1" \
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
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KK752CNP351655QEGHA94QBS",
      "price": 19,
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 2,
    "current_page": 1,
    "per_page": 2,
    "total": 5,
    "total_pages": 3
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=2",
    "last": "/data/products:query?page=3&per_page=2",
    "next": "/data/products:query?page=2&per_page=2",
    "prev": null
  }
}
```

### Combined Query Options

Combine filtering, sorting, field selection, and pagination in one request.

```bash
curl -s -X GET "http://localhost:6000/data/products:query?brand[eq]=Wow&sort=-price&fields=title,price,quantity&per_page=10&page=1" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KK752CNPB98YQ8RJQPSSFQM3",
      "price": 199,
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "id": "01KK752CNQ27A4QC6XK3CTVPTR",
      "price": 49,
      "quantity": 8,
      "title": "Laptop Stand"
    },
    {
      "id": "01KK752CNP8SQ6HJFMDVF7GXEN",
      "price": 29,
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "count": 3,
    "current_page": 1,
    "per_page": 10,
    "total": 3,
    "total_pages": 1
  },
  "links": {
    "first": "/data/products:query?brand%5Beq%5D=Wow&fields=title%2Cprice%2Cquantity&page=1&per_page=10&sort=-price",
    "last": "/data/products:query?brand%5Beq%5D=Wow&fields=title%2Cprice%2Cquantity&page=1&per_page=10&sort=-price",
    "next": null,
    "prev": null
  }
}
```
