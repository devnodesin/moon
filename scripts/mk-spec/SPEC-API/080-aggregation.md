## Aggregation Operations

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

**Aggregation Endpoints:**

- Count Records: `GET /{collection_name}:count`
- Sum Numeric Field: `GET /{collection_name}:sum` (requires `?field=...`)
- Average Numeric Field: `GET /{collection_name}:avg` (requires `?field=...`)
- Minimum Value: `GET /{collection_name}:min` (requires `?field=...`)
- Maximum Value: `GET /{collection_name}:max` (requires `?field=...`)

**Note:**

- Replace `{collection_name}` with your collection name.
- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.

**Example Request:**

```sh
GET /products:sum?field=quantity
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 55
  }
}
```

**Aggregation with Filters:** Combine aggregation with query filters for calculations on specific subsets:

- `/products:count?quantity[gt]=10`
- `/products:sum?field=quantity&brand[eq]=Wow`
- `/products:max?field=quantity`

### Error Handling

**Error Response:** Follow [Standard Error Response](#standard-error-response) for any error handling
