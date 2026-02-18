## Standard Response Pattern for `:destroy` Endpoints

Destroy endpoints permanently delete resources from the system.

**Applicable Endpoints:**

- Delete User: `POST /users:destroy?id={id}`
- Delete API Key: `POST /apikeys:destroy?id={id}`
- Delete Collection: `POST /collections:destroy?name={collection_name}`
- Delete Collection Record(s): `POST /{collection_name}:destroy`

### Response Structure

**For Users, API Keys, and Collections:**

```sh
POST /users:destroy?id=01KHCZGWWRBQBREMG0K23C6C5H
POST /apikeys:destroy?id=01KHCZKCR7MHB0Q69KM63D6AXF
POST /collections:destroy?name=products
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

**For Collection Records:**

Single record:

```json
{
  "data": ["01KHCZKMM0N808MKSHBNWF464F"]
}
```

Multiple records:

```json
{
  "data": [
    "01KHCZKMXYVC1NRHDZ83XMHY4N",
    "01KHCZKMY28ERJFPCVBQEKQ4SY",
    "01KHCZKMY28ERJFPCVBQEKQ4SZ"
  ]
}
```

**Response (200 OK) - All Succeeded:**

```json
{
  "data": [
    "01KHCZKMXYVC1NRHDZ83XMHY4N",
    "01KHCZKMY28ERJFPCVBQEKQ4SY",
    "01KHCZKMY28ERJFPCVBQEKQ4SZ"
  ],
  "meta": {
    "total": 3,
    "succeeded": 3,
    "failed": 0
  },
  "message": "3 record(s) deleted successfully"
}
```

**Response (200 OK) - Partial Success:**

```json
{
  "data": ["01KHCZKMXYVC1NRHDZ83XMHY4N", "01KHCZKMY28ERJFPCVBQEKQ4SY"],
  "meta": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  },
  "message": "2 of 3 record(s) deleted successfully"
}
```

### Parameters

| Parameter | Type   | Description                                          |
| --------- | ------ | ---------------------------------------------------- |
| `id`      | string | ULID of the resource (required for users, apikeys)   |
| `name`    | string | Name of the collection (required for collections)    |
| `data`    | array  | Array of record IDs to delete (required for records) |

### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single deletions.
- **Deleted IDs returned**: Response includes `data` array with IDs of successfully deleted records.
- **Partial success**: If some records fail to delete, the successfully deleted count is shown in `meta`.
- **Failed records**: Check `meta.failed` count to detect partial failures. Failed record IDs are excluded from the `data` array.
- **Status code**: Returns `200 OK` if at least one record was deleted successfully.
- **Message field**: Always includes a human-readable success message.

### Error Handling

**Error Response:** Follow [Standard Error Response](#standard-error-response) for any error handling
