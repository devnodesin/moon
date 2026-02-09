## Troubleshooting

### Common Issues

**Authentication Errors**
- Verify your JWT token or API key is valid
- Check that the token is included in the Authorization header
- Ensure the token hasn't expired

**Collection Not Found**
- Use `collections:list` to verify the collection exists
- Collection names are case-insensitive and stored in lowercase
- Check for typos in the collection name

**Invalid Field Types**
- Supported types: string, integer, boolean, datetime, json, decimal
- Use decimal for currency values, not float
- DateTime values must be in RFC3339 or ISO 8601 format

**Rate Limiting**
- JWT tokens: 100 requests per minute per user
- API keys: 1000 requests per minute per key
- Wait for the rate limit window to reset before retrying
