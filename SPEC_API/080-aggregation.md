### Count Records

```bash
curl -s -X GET "http://localhost:6006/products:count" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 3
  }
}
```

### Sum Numeric Field

```bash
curl -s -X GET "http://localhost:6006/products:sum?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 85
  }
}
```

### Average Numeric Field

```bash
curl -s -X GET "http://localhost:6006/products:avg?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 28.333333333333332
  }
}
```

### Minimum Value

```bash
curl -s -X GET "http://localhost:6006/products:min?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 10
  }
}
```

### Maximum Value

```bash
curl -s -X GET "http://localhost:6006/products:max?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 55
  }
}
```
