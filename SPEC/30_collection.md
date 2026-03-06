The `/collections:{query, mutate}` endpoint is strictly for listing, creating, and managing collection schemas.

- `users` and `apikeys` are system collections. `/collections:mutate` operations are not allowed on these collections.
- Do not allow batch schema changes in a single request (e.g., adding and removing columns together).

New collections can be created and managed using the `/collections:query` and `/collections:mutate` endpoints.

See [Standard Error Response (200 OK):](10_error.md) for any error handling

### `GET /collections:query`

`/collections:query` lists all available database tables.

```json
{
  "data": [
    // Mandatory, always an array
    { "name": "users", "count": 5 },
    { "name": "apikeys", "count": 2 },
    { "name": "products", "count": 55 }
  ],
  "meta": {
    "total": 3,
    "count": 3, // total records available for this request
    "per_page": 15, // number of records per page
    "current_page": 1, // current page number
    "total_pages": 1 // total number of pages available
  }
}
```

### `POST /collections:mutate`

- `op=create`: create collection(s) with `name` and `columns`
- `op=update`: schema changes using:
  - `add_columns`
  - `rename_columns`
  - `modify_columns`
  - `remove_columns`
- `op=destroy`: delete collection(s) by `name`
- **Important:** All operations are atomic; only one operation is allowed at a time.
- If `nullable` and `unique` are not specified, the field defaults to `nullable: false` and `unique: false`.

#### Create Collection `POST /collections:mutate`

Request

```json
{
  "op": "create",
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "title", "type": "string", "unique": true },
        { "name": "price", "type": "decimal", "nullable": true }
      ]
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products",
      "columns": [
        {
          "name": "title",
          "type": "string",
          "nullable": false,
          "unique": true
        },
        {
          "name": "price",
          "type": "decimal",
          "nullable": true,
          "unique": false
        }
      ]
    }
  ],
  "message": "Collection 'products' created successfully"
}
```

#### Add Columns `POST /collections:mutate`

Request

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "add_columns": [
        { "name": "description", "type": "string", "nullable": true },
        { "name": "sku", "type": "string", "unique": true }
      ]
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products",
      "columns": [
        {
          "name": "title",
          "type": "string",
          "nullable": false,
          "unique": true
        },
        {
          "name": "price",
          "type": "decimal",
          "nullable": true,
          "unique": false
        },
        {
          "name": "description",
          "type": "string",
          "nullable": true,
          "unique": false
        },
        { "name": "sku", "type": "string", "nullable": false, "unique": true }
      ]
    }
  ],
  "message": "Columns added to collection 'products' successfully"
}
```

#### Rename Columns `POST /collections:mutate`

Request

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "rename_columns": [
        { "old_name": "title", "new_name": "name" },
        { "old_name": "sku", "new_name": "product_code" }
      ]
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "name", "type": "string", "nullable": false, "unique": true },
        {
          "name": "price",
          "type": "decimal",
          "nullable": true,
          "unique": false
        },
        {
          "name": "description",
          "type": "string",
          "nullable": true,
          "unique": false
        },
        {
          "name": "product_code",
          "type": "string",
          "nullable": false,
          "unique": true
        }
      ]
    }
  ],
  "message": "Columns renamed in collection 'products' successfully"
}
```

#### Modify Columns `POST /collections:mutate`

Request

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "modify_columns": [
        { "name": "price", "type": "decimal", "nullable": false },
        { "name": "description", "type": "string", "nullable": false }
      ]
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "name", "type": "string", "nullable": false, "unique": true },
        {
          "name": "price",
          "type": "decimal",
          "nullable": false,
          "unique": false
        },
        {
          "name": "description",
          "type": "string",
          "nullable": false,
          "unique": false
        },
        {
          "name": "product_code",
          "type": "string",
          "nullable": false,
          "unique": true
        }
      ]
    }
  ],
  "message": "Columns modified in collection 'products' successfully"
}
```

#### Remove Columns `POST /collections:mutate`

Request

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "remove_columns": ["description"]
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "name", "type": "string", "nullable": false, "unique": true },
        {
          "name": "price",
          "type": "decimal",
          "nullable": false,
          "unique": false
        },
        {
          "name": "product_code",
          "type": "string",
          "nullable": false,
          "unique": true
        }
      ]
    }
  ],
  "message": "Columns removed from collection 'products' successfully"
}
```

#### Destroy Collection `POST /collections:mutate`

Request

```json
{
  "op": "destroy",
  "data": [
    {
      "name": "products"
    }
  ]
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "name": "products"
    }
  ],
  "message": "Collection 'products' destroyed successfully"
}
```
