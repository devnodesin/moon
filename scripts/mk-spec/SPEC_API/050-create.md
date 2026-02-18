## Standard Response Pattern for `:create` Endpoints

Create endpoints add new resources to the system.

**Applicable Endpoints:**

- Create User: `POST /users:create`
- Create API Key: `POST /apikeys:create`
- Create Collection: `POST /collections:create`
- Create Collection Record(s): `POST /{collection_name}:create`

### Response Structure

**For Users, API Keys, and Collections:**

```sh
POST /users:create
POST /apikeys:create
POST /collections:create
```

**Request Body:**

```json
{
  "data": {
    "username": "moonuser",
    "email": "moonuser@example.com",
    "password": "UserPass123#"
  }
}
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01KHCZK95DPBAT04EH0WWDZYR7",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "created_at": "2026-02-14T02:27:38Z"
  },
  "message": "User created successfully"
}
```

**For Collection Records:**

Single record:

```json
{
  "data": [{ "title": "Wireless Mouse", "price": "29.99" }]
}
```

Multiple records:

```json
{
  "data": [
    { "title": "Keyboard", "price": "49.99" },
    { "title": "Monitor", "price": "199.99" },
    { "title": "Keyboard", "price": "39.99" }
  ]
}
```

**Response (201 Created) - All Succeeded:**

```json
{
  "data": [
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "title": "Keyboard",
      "price": "49.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "title": "Monitor",
      "price": "199.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SZ",
      "title": "Keyboard",
      "price": "39.99"
    }
  ],
  "meta": {
    "total": 3,
    "succeeded": 3,
    "failed": 0
  },
  "message": "3 record(s) created successfully"
}
```

**Response (201 Created) - Partial Success:**

```json
{
  "data": [
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "title": "Keyboard",
      "price": "49.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "title": "Monitor",
      "price": "199.99"
    }
  ],
  "meta": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  },
  "message": "2 of 3 record(s) created successfully"
}
```

**For Collection Creation:**

**Request Body:**

```json
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
      "type": "integer",
      "nullable": false
    }
  ]
}
```

**Response (201 Created):**

```json
{
  "data": {
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
        "type": "integer",
        "nullable": false,
        "unique": false
      }
    ]
  },
  "message": "Collection 'products' created successfully"
}
```

### Important Notes

- **ID field**: The `id` field is system-generated and read-only. Do not include it in create requests.
- **Array format**: Collection records must always be sent as an array in `data`, even for single records.
- **Partial success**: If some records fail validation, successfully created records are returned in `data`.
- **Failed records**: Failed records are excluded from the `data` array. Check `meta.failed` count to detect partial failures.
- **Status code**: Always returns `201 Created` if at least one record was created successfully.
- **Consistent wrapper**: All `:create` endpoints use the `data` field for created resource(s).
- **Message field**: Always includes a human-readable success message.
- **API Key security**: The `key` field appears in `data` only once during creation.

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling
