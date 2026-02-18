## Standard Response Pattern for `:schema` Endpoints

Retrieve the schema definition for a collection, including all fields, their types, constraints, and defaults.

`GET /{collection_name}:schema`

**Example Request:**

```sh
GET /products:schema
```

**Response (200 OK):**

```json
{
  "data": {
    "collection": "products",
    "fields": [
      {
        "name": "id",
        "type": "string",
        "nullable": false,
        "readonly": true
      },
      {
        "name": "title",
        "type": "string",
        "nullable": false
      },
      {
        "name": "price",
        "type": "decimal",
        "nullable": false
      },
      {
        "name": "details",
        "type": "string",
        "nullable": true,
        "default": "''"
      },
      {
        "name": "quantity",
        "type": "integer",
        "nullable": true,
        "default": "0"
      },
      {
        "name": "brand",
        "type": "string",
        "nullable": true,
        "default": "''"
      }
    ],
    "total": 6
  }
}
```

**Field Properties**

- `name`: The field's name.
- `type`: The data type (`string`, `integer`, `decimal`, `boolean`, `timestamp`, etc., as defined in the specification).
- `nullable`: Indicates if the field can be omitted or set to `null` in API requests.
- `readonly`: Indicates if the field is system-generated and cannot be modified (e.g., `id`).
- `default`: The default value assigned when the field is not provided.
- `unique`: Specifies whether the field must have unique values (optional).

### Important Notes

- **System fields**: The `id` and `default` field is automatically included in every collection and is readonly
- **Total count**: Represents the total number of fields in the collection schema
- **Schema introspection**: Use this endpoint to dynamically discover collection structure
- **Validation**: Schema information helps clients validate data before submission
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling
