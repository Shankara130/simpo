# CORS Configuration Guide

## Overview

Cross-Origin Resource Sharing (CORS) is configured for the simpo API to enable secure cross-origin requests from the web dashboard and mobile app. This implementation follows security best practices by using specific allowed origins instead of wildcards.

**Story:** 9.4 - Implement CORS Middleware for Cross-Origin Requests

## Security Considerations

### Critical Security Fixes

**BEFORE (Insecure):**
```go
corsConfig.AllowAllOrigins = true  // ❌ SECURITY VULNERABILITY
```

**AFTER (Secure):**
```go
corsConfig.AllowOrigins = cfg.Cors.AllowedOrigins  // ✅ Specific origins only
```

### Best Practices

✅ **DO:**
- Use specific, explicitly allowed origins
- Validate origins server-side (enforced by gin-contrib/cors)
- Use HTTPS origins in production environments
- Support credentials when needed (cookies, auth headers)

❌ **DON'T:**
- Never use `AllowAllOrigins = true` in production
- Never use `Access-Control-Allow-Origin: "*"` with `AllowCredentials = true`
- Never reflect Origin header without validation
- Never use overly permissive headers

## Configuration

### Environment Variables (.env)

```bash
# Enable/disable CORS middleware
CORS_ENABLED=true

# Allowed origins (comma-separated, no spaces)
# Development: localhost origins
# Production: HTTPS domains only
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006,https://admin.simpo.com

# Allow credentials (cookies, auth headers)
CORS_ALLOW_CREDENTIALS=true

# Allowed HTTP methods
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS

# Allowed request headers
CORS_ALLOWED_HEADERS=Authorization,Content-Type,X-Requested-With

# Pre-flight cache duration (seconds)
CORS_MAX_AGE=86400  # 24 hours
```

### YAML Configuration (config.yaml)

```yaml
cors:
  enabled: true
  allowed_origins:
    - "http://localhost:3000"       # Web admin (development)
    - "http://localhost:19006"      # Expo mobile (development)
    - "https://admin.simpo.com"     # Production web dashboard
  allow_credentials: true
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Authorization"
    - "Content-Type"
    - "X-Requested-With"
  max_age: 86400  # 24 hours in seconds
```

## Deployment Scenarios

### Development Environment

**Origins:**
- `http://localhost:3000` - Next.js web dashboard
- `http://localhost:19006` - Expo mobile app

**Configuration:**
```bash
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006
CORS_ALLOW_CREDENTIALS=true
```

### Production Environment

**Origins:**
- `https://admin.simpo.com` - Production web dashboard
- `https://simpo-pharmacy.com` - Customer pharmacy domains

**Configuration:**
```bash
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://admin.simpo.com,https://simpo-pharmacy.com
CORS_ALLOW_CREDENTIALS=true
```

**IMPORTANT:** Always use HTTPS origins in production. HTTP origins should never be used in production environments.

### Self-Hosted Deployment

For customers hosting their own instance:

1. **Identify web dashboard domain** (e.g., `https://pharmacy.example.com`)
2. **Add to CORS configuration:**
   ```bash
   CORS_ALLOWED_ORIGINS=https://pharmacy.example.com
   ```
3. **Restart the backend service**
4. **Verify CORS headers are present** in browser developer tools

## Testing CORS

### Manual Testing with curl

```bash
# Test pre-flight OPTIONS request
curl -X OPTIONS http://localhost:8080/api/v1/health \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -v

# Expected response headers:
# Access-Control-Allow-Origin: http://localhost:3000
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# Access-Control-Allow-Headers: Authorization, Content-Type, X-Requested-With
```

### Browser Testing

1. Open browser developer tools (F12)
2. Go to Network tab
3. Make a request from allowed origin
4. Check response headers for:
   - `Access-Control-Allow-Origin`
   - `Access-Control-Allow-Methods`
   - `Access-Control-Allow-Headers`
   - `Access-Control-Allow-Credentials`

### Automated Testing

Run the CORS test suite:

```bash
# Unit tests
go test ./internal/middleware/cors_test.go -v

# Integration tests
go test ./internal/middleware/cors_integration_test.go -v
```

## Troubleshooting

### Issue: CORS errors in browser console

**Symptoms:**
```
Access to XMLHttpRequest has been blocked by CORS policy
```

**Solutions:**
1. Verify origin is in `CORS_ALLOWED_ORIGINS`
2. Check exact match (case-sensitive, includes protocol and port)
3. Ensure `CORS_ENABLED=true`
4. Restart backend after configuration changes

### Issue: Credentials not sent

**Symptoms:**
- Cookies not included in requests
- Authorization header missing

**Solutions:**
1. Verify `CORS_ALLOW_CREDENTIALS=true`
2. Check client includes `credentials: 'include'` (fetch) or `withCredentials: true` (axios)
3. Ensure origin is not wildcard (`*`)

### Issue: Pre-flight requests failing

**Symptoms:**
- OPTIONS request returns 404 or 405
- POST/PUT/DELETE requests blocked

**Solutions:**
1. Verify CORS middleware runs before auth middleware
2. Check `CORS_ALLOWED_METHODS` includes required methods
3. Ensure `CORS_ALLOWED_HEADERS` includes required headers

## Common Configuration Examples

### Single Origin (Simple Deployment)

```bash
CORS_ALLOWED_ORIGINS=https://admin.example.com
CORS_ALLOW_CREDENTIALS=true
```

### Multiple Origins (Multi-tenant)

```bash
CORS_ALLOWED_ORIGINS=https://admin.example.com,https://pharmacy1.example.com,https://pharmacy2.example.com
CORS_ALLOW_CREDENTIALS=true
```

### Development + Production

```bash
# .env.development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006

# .env.production
CORS_ALLOWED_ORIGINS=https://admin.example.com
```

## References

- [MDN CORS Guide](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [gin-contrib/cors](https://github.com/gin-contrib/cors)
- [Story 9.4: CORS Middleware Implementation](../_bmad-output/implementation-artifacts/9-4-implement-cors-middleware-for-cross-origin-requests.md)
