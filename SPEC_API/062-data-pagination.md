### First Page

First page of results (no cursor). `meta.next` is captured as `$NEXT_CURSOR` and `meta.prev` as `$PREV_CURSOR` for subsequent requests.

```bash
curl -s -X GET "http://localhost:6006/products:list?limit=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJMQ43YWAVFE4Q6VVK85DHHJ",
      "quantity": 1,
      "title": "Product 1"
    },
    {
      "id": "01KJMQ43Z5Y9FVGHS8DEP2BWPA",
      "quantity": 2,
      "title": "Product 2"
    },
    {
      "id": "01KJMQ43ZA44WQ99QYMAE2P00W",
      "quantity": 3,
      "title": "Product 3"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 3,
    "next": "01KJMQ43ZA44WQ99QYMAE2P00W",
    "prev": null,
    "total": 7
  }
}
```

### Next Page (Forward)

Navigate forward using `$NEXT_CURSOR` (from `meta.next` of the previous response). Cursors are updated after each list response.

```bash
curl -s -X GET "http://localhost:6006/products:list?after=01KJMQ43ZA44WQ99QYMAE2P00W&limit=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJMQ43ZG3EE7D3PZ3Y2RH256",
      "quantity": 4,
      "title": "Product 4"
    },
    {
      "id": "01KJMQ43ZNEGYP1HHW60KNNGWE",
      "quantity": 5,
      "title": "Product 5"
    },
    {
      "id": "01KJMQ43ZVSN9MVW8GJ1WACN5N",
      "quantity": 6,
      "title": "Product 6"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 3,
    "next": "01KJMQ43ZVSN9MVW8GJ1WACN5N",
    "prev": null,
    "total": 7
  }
}
```

### Last Page (Forward)

Navigate to the last page using the updated `$NEXT_CURSOR`. On the last page `meta.next` is `null`; `$PREV_CURSOR` is available for backward navigation.

```bash
curl -s -X GET "http://localhost:6006/products:list?after=01KJMQ43ZVSN9MVW8GJ1WACN5N&limit=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJMQ4400CG26XG30KHTPP4AK",
      "quantity": 7,
      "title": "Product 7"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 3,
    "next": null,
    "prev": "01KJMQ43ZA44WQ99QYMAE2P00W",
    "total": 7
  }
}
```

### Previous Page (Backward)

Navigate backward using `$PREV_CURSOR` (from `meta.prev` of the last page), returning to the previous page of results.

```bash
curl -s -X GET "http://localhost:6006/products:list?after=01KJMQ43ZA44WQ99QYMAE2P00W&limit=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJMQ43ZG3EE7D3PZ3Y2RH256",
      "quantity": 4,
      "title": "Product 4"
    },
    {
      "id": "01KJMQ43ZNEGYP1HHW60KNNGWE",
      "quantity": 5,
      "title": "Product 5"
    },
    {
      "id": "01KJMQ43ZVSN9MVW8GJ1WACN5N",
      "quantity": 6,
      "title": "Product 6"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 3,
    "next": "01KJMQ43ZVSN9MVW8GJ1WACN5N",
    "prev": null,
    "total": 7
  }
}
```
