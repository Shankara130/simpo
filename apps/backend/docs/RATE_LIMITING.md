# Rate Limiting

## Overview

The API implements JWT-based rate limiting to prevent abuse and ensure fair resource allocation. Rate limits are enforced per authenticated user (by user ID) or per IP address for unauthenticated requests.

## Configuration

Rate limiting is configured via environment variables or `config.yaml`:

### Environment Variables

```bash
RATELIMIT_ENABLED=true       # Enable/disable rate limiting
RATELIMIT_REQUESTS=100       # Maximum requests per window
RATELIMIT_WINDOW=1m          # Time window (e.g., 1m, 5m, 1h)
```

### Config File (config.yaml)

```yaml
ratelimit:
  enabled: true
  requests: 100
  window: "1m"
```

### Per-Environment Configuration

Different environments can have different rate limits:

```bash
# Production - stricter limits
RATELIMIT_REQUESTS=100
RATELIMIT_WINDOW=1m

# Development - more lenient
RATELIMIT_REQUESTS=1000
RATELIMIT_WINDOW=1m

# Testing - very permissive
RATELIMIT_REQUESTS=10000
RATELIMIT_WINDOW=1m
```

## Rate Limit Behavior

### JWT-Based Tracking (Authenticated Requests)

When a user is authenticated via JWT token, rate limiting tracks by user ID:

- **Key Format**: `user:{userID}`
- **Scope**: All requests from the same user share the same rate limit
- **Benefit**: Users cannot reset limits by changing IP address

Example:
```json
{
  "user": 123,
  "rate_limit_key": "user:123"
}
```

### IP-Based Tracking (Unauthenticated Requests)

For unauthenticated requests, rate limiting tracks by IP address:

- **Key Format**: `ip:{ipAddress}`
- **Fallback Order**: `X-Forwarded-For` → `X-Real-IP` → `ClientIP()` → `unknown`
- **Scope**: Requests from the same IP share the same rate limit

Example:
```json
{
  "ip": "192.168.1.100",
  "rate_limit_key": "ip:192.168.1.100"
}
```

### Algorithm

The rate limiter uses a **token-bucket algorithm** via `golang.org/x/time/rate`:

- **Rate (R)**: `requests / window` tokens per second
- **Burst**: Allows short spikes up to `requests` concurrent requests
- **Sliding Window**: Old requests' weight decreases over time

## Response Headers

All API responses include rate limit headers:

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Maximum requests in window | `100` |
| `X-RateLimit-Remaining` | Requests remaining in window | `95` |
| `X-RateLimit-Reset` | Unix timestamp when window resets | `1717039200` |

### Example Response

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1717039200
Content-Type: application/json
```

## Rate Limit Exceeded

When the rate limit is exceeded, the API returns:

### HTTP Status

```
429 Too Many Requests
```

### Response Headers

```http
Retry-After: 45
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1717039200
```

### Response Body (RFC 7807 Format)

```json
{
  "success": false,
  "error": {
    "type": "https://api.simpo.com/errors/TOO_MANY_REQUESTS",
    "title": "Rate Limit Exceeded",
    "status": 429,
    "detail": "Rate limit exceeded",
    "code": "TOO_MANY_REQUESTS",
    "details": "Too many requests. Please try again in 45 seconds.",
    "retry_after": 45
  }
}
```

## Best Practices for API Consumers

### 1. Implement Exponential Backoff

When receiving a 429 response, implement exponential backoff:

```javascript
async function makeRequest(url, options) {
  try {
    const response = await fetch(url, options);
    
    if (response.status === 429) {
      const retryAfter = response.headers.get('Retry-After');
      const waitTime = parseInt(retryAfter) * 1000;
      
      // Exponential backoff
      await new Promise(resolve => setTimeout(resolve, waitTime));
      return makeRequest(url, options);
    }
    
    return response.json();
  } catch (error) {
    console.error('Request failed:', error);
    throw error;
  }
}
```

### 2. Monitor Rate Limit Headers

Track rate limit headers to avoid hitting limits:

```javascript
function checkRateLimit(headers) {
  const remaining = parseInt(headers.get('X-RateLimit-Remaining'));
  const limit = parseInt(headers.get('X-RateLimit-Limit'));
  
  if (remaining < (limit * 0.1)) {
    console.warn('Rate limit nearly exceeded:', remaining, 'remaining');
  }
}
```

### 3. Use Batching for High-Volume Operations

For operations that require many requests, use batching endpoints if available:

```javascript
// Instead of multiple individual requests
for (const item of items) {
  await createItem(item);
}

// Use a batch endpoint
await createItemsBatch(items);
```

## Rate Limiting Strategy

### Global Middleware

Rate limiting is applied globally to all endpoints when enabled:

```go
// Applied in router.go
if rlCfg.Enabled {
    router.Use(middleware.NewRateLimitMiddleware(
        rlCfg.Window,
        rlCfg.Requests,
        keyFunc,
        nil,
    ))
}
```

### Strict Endpoints

Certain endpoints (e.g., staff registration) have stricter rate limits:

```go
// 5 requests per 15 minutes for staff registration
authStrictGroup.Use(middleware.NewRateLimitMiddleware(
    15*time.Minute,
    5,
    strictKeyFunc,
    nil,
))
```

### Excluded Endpoints

Health check endpoints are excluded from rate limiting:

```go
// These endpoints bypass rate limiting
router.GET("/health", healthHandler.Health)
router.GET("/health/live", healthHandler.Live)
router.GET("/health/ready", healthHandler.Ready)
```

## Testing Rate Limiting

### Unit Tests

```go
func TestRateLimitMiddleware_Basic(t *testing.T) {
    middleware := NewRateLimitMiddleware(time.Second, 2, keyFunc, store)
    
    router := gin.New()
    router.Use(middleware)
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })
    
    // First 2 requests should succeed
    for i := 0; i < 2; i++ {
        req := httptest.NewRequest("GET", "/test", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        assert.Equal(t, 200, w.Code)
    }
    
    // Third request should be blocked
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, 429, w.Code)
}
```

### Integration Tests

```bash
# Test rate limiting with curl
for i in {1..105}; do
  echo "Request $i:"
  curl -i http://localhost:8080/api/v1/transactions \
    -H "Authorization: Bearer $JWT_TOKEN"
done
```

## Performance Considerations

- **Memory**: Each rate limit entry uses ~100 bytes
- **Cache**: LRU cache with 5000 entry limit and 6-hour TTL
- **Speed**: O(1) operations per request (negligible overhead)

## Troubleshooting

### Issue: Rate Limits Too Strict

**Solution**: Increase limits for your environment:

```bash
RATELIMIT_REQUESTS=200
RATELIMIT_WINDOW=1m
```

### Issue: Legitimate Users Blocked

**Solution**: Implement user tiering with different limits:

```go
// Enterprise tier users
if user.IsEnterprise {
    return NewRateLimitMiddleware(time.Minute, 1000, keyFunc, store)
}
// Standard users
return NewRateLimitMiddleware(time.Minute, 100, keyFunc, store)
```

### Issue: Health Checks Blocked

**Solution**: Ensure health endpoints are registered before rate limit middleware:

```go
// Health endpoints first
router.GET("/health", healthHandler.Health)

// Then apply rate limiting
router.Use(rateLimitMiddleware)
```

## References

- [RFC 7807: Problem Details for HTTP APIs](https://tools.ietf.org/html/rfc7807)
- [RFC 6585: Additional HTTP Status Codes](https://tools.ietf.org/html/rfc6585)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
