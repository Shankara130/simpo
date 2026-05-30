# Logging Documentation

## Overview

The Simpo backend uses structured logging with Go's standard library `log/slog` (Go 1.21+). All logs are output in JSON format to stdout for Docker container compatibility and log aggregation.

**Architecture Note:** The system uses `log/slog` instead of `zap` as specified in the original architecture. This provides equivalent functionality with better maintainability (standard library, no external dependencies).

---

## Features

### 1. Structured JSON Logging

All logs are emitted in JSON format with the following fields:

- **level**: Log level (debug, info, warn, error)
- **timestamp**: ISO 8601 timestamp with timezone
- **msg**: Descriptive log message
- **caller**: (Optional) Source location (file:line) when `IncludeCaller` is enabled
- **context**: Request-specific context (request_id, user_id, etc.)

**Example Log Output:**
```json
{
  "time": "2026-05-30T11:24:15.680538+07:00",
  "level": "INFO",
  "msg": "HTTP Request",
  "request_id": "e6118d9b-4799-4f23-a888-7b96f15b1f34",
  "method": "GET",
  "path": "/api/products?page=1",
  "status": 200,
  "duration": "8.75ms",
  "client_ip": "192.0.2.1",
  "user_agent": "Mozilla/5.0...",
  "response_size": 1024
}
```

---

### 2. Sensitive Data Redaction

**Security First:** Sensitive data is automatically redacted from logs to prevent security breaches.

#### Redacted Fields

The following field patterns are redacted by default:
- `password`, `passwd`
- `token`
- `api_key`, `apikey`
- `secret`
- `authorization`
- `cookie`
- `session`

#### Redaction Examples

**Authorization Header:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
↓
authorization: "Bearer ****"
```

**Cookie Header:**
```
Cookie: session_id=abc123; user_token=xyz789
↓
cookie: "****"
```

**Query Parameters:**
```
GET /api/login?password=MySecret123&token=abc123&user=john
↓
path: "/api/login?password=****&token=****&user=john"
query_params: {"password": "****", "token": "****", "user": "john"}
```

#### Configuration

```bash
# Enable sensitive data redaction (default: true)
LOGGING_REDACT_ENABLED=true

# Custom redaction patterns (comma-separated)
LOGGING_REDACT_PATTERNS=password,token,secret,authorization,cookie
```

**⚠️ Security:** Never disable redaction in production environments.

---

### 3. Business Event Logging

For audit trail compliance, business events are logged with full context.

#### Event Types

**Transaction Events:**
```go
utils.LogTransactionEvent(ctx, logger, "transaction.completed", "TRX-20240530-0001", map[string]interface{}{
    "cashier_id": 123,
    "branch_id": 1,
    "total": 150000,
    "payment_method": "cash",
})
```

**Stock Change Events:**
```go
utils.LogStockChangeEvent(ctx, logger, productID, oldQty, newQty, "sale")
```

**User Action Events:**
```go
utils.LogUserActionEvent(ctx, logger, userID, "user.created", map[string]interface{}{
    "target_user_id": newUser.ID,
    "role": "cashier",
})
```

**System Events:**
```go
utils.LogSystemEvent(ctx, logger, "backup.completed", map[string]interface{}{
    "backup_file": "/backups/simpo_db_20240530.sql",
    "size_bytes": 1024000,
})
```

**Configuration:**
```bash
# Enable business event logging (default: true)
LOGGING_BUSINESS_EVENTS=true
```

---

### 4. Configurable Log Levels

Supported log levels:
- `debug` - Verbose logging for development
- `info` - Normal operational messages (default)
- `warn` - Warning messages
- `error` - Error messages only

**Configuration:**
```bash
# Set log level
LOGGING_LEVEL=info
```

**Log Level Selection by Status Code:**
- 2xx (Success) → `info`
- 4xx (Client Error) → `warn`
- 5xx (Server Error) → `error`

---

### 5. Docker-Ready Output

Logs are written to **stdout** in JSON format for Docker container compatibility.

**Advantages:**
- Compatible with Docker logging drivers
- No file rotation needed (handled at container level)
- Easy log aggregation with external tools
- Works with all container orchestration platforms

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LOGGING_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `LOGGING_REDACT_ENABLED` | Enable sensitive data redaction | `true` |
| `LOGGING_INCLUDE_CALLER` | Include caller information (file:line) | `false` |
| `LOGGING_BUSINESS_EVENTS` | Enable business event logging | `true` |
| `LOGGING_REDACT_PATTERNS` | Custom redaction patterns | `password,token,secret,authorization,cookie` |

### YAML Configuration

```yaml
# configs/config.yaml
logging:
  level: "info"
  redact_enabled: true
  include_caller: false
  business_events: true
  redact_patterns:
    - "password"
    - "token"
    - "secret"
    - "authorization"
    - "cookie"
```

---

## Docker Log Rotation

### Container-Level Rotation Strategy

The application does **not** implement file-based log rotation. Instead, log rotation is handled at the Docker/container level.

### Docker Compose Configuration

```yaml
# docker-compose.yml
services:
  backend:
    image: simpo-backend:latest
    logging:
      driver: "json-file"
      options:
        max-size: "10m"      # Maximum log file size
        max-file: "3"        # Maximum number of log files
```

**Configuration Options:**
- `max-size`: Maximum size of each log file before rotation (e.g., "10m", "100m")
- `max-file`: Maximum number of rotated log files to retain
- `compress`: Enable log compression (true/false)

### Production Recommendations

**For Small Deployments:**
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "50m"
    max-file: "5"
```

**For Large Deployments:**
```yaml
logging:
  driver: "syslog"  # or "journald"
  options:
    syslog-address: "tcp://192.168.0.42:514"
    tag: "simpo-backend"
```

### Kubernetes Logging

Use `LogRotation` feature in Kubernetes or external logging systems:

```yaml
# pods/backend-container.yaml
apiVersion: v1
kind: Pod
metadata:
  name: backend
spec:
  containers:
  - name: backend
    image: simpo-backend:latest
```

Logs are automatically collected by the cluster logging system (e.g., Elasticsearch, Loki).

---

## Log Aggregation

### Recommended Tools

**ELK Stack (Elasticsearch, Logstash, Kibana):**
- Industry-standard log aggregation
- Powerful search and visualization
- Supports JSON log parsing
- Requires significant infrastructure

**Grafana Loki:**
- Lightweight alternative to ELK
- Labels-based indexing
- Integrates with Grafana
- Lower resource requirements

**Cloudflare Logpush:**
- For cloud deployments
- Push logs to S3, R2, or other storage
- Cost-effective for large volumes

**Datadog/New Relic:**
- Managed observability platforms
- Built-in alerting and dashboards
- Higher cost but minimal setup

### Loki Example Configuration

```yaml
# promtail-config.yml
scrape_configs:
  - job_name: simpo-backend
    docker:
      host: unix:///var/run/docker.sock
    relabel_configs:
      - source_labels: ['container_name']
        regex: 'simpo-backend.*'
        action: keep
```

---

## Troubleshooting

### Issue: Logs Not Appearing

**Symptoms:** No log output in Docker logs

**Solutions:**
1. Check if logging is enabled:
   ```bash
   docker logs simpo-backend
   ```
2. Verify log level is not too restrictive:
   ```bash
   LOGGING_LEVEL=debug
   ```
3. Check if stdout is being captured:
   ```yaml
   # docker-compose.yml
   logging:
     driver: "json-file"
   ```

---

### Issue: Sensitive Data Leaking in Logs

**Symptoms:** Passwords or tokens visible in logs

**Solutions:**
1. Verify redaction is enabled:
   ```bash
   LOGGING_REDACT_ENABLED=true
   ```
2. Check redaction patterns cover all sensitive fields:
   ```bash
   LOGGING_REDACT_PATTERNS=password,token,api_key,secret,authorization,cookie,session
   ```
3. Add custom patterns for application-specific fields:
   ```bash
   LOGGING_REDACT_PATTERNS=password,token,custom_field
   ```

---

### Issue: Missing Caller Information

**Symptoms:** No file:line information in logs

**Solutions:**
1. Enable caller information:
   ```bash
   LOGGING_INCLUDE_CALLER=true
   ```
2. Note: Caller information adds ~1μs overhead per log entry

---

### Issue: Disk Space Exhaustion

**Symptoms:** Server disk full due to logs

**Solutions:**
1. **Immediate:** Clean up Docker logs:
   ```bash
   docker system prune -a
   ```
2. **Configure log rotation** in docker-compose.yml:
   ```yaml
   logging:
     options:
       max-size: "10m"
       max-file: "3"
   ```
3. **Long-term:** Implement centralized log aggregation (ELK, Loki)

---

### Issue: Business Events Not Logged

**Symptoms:** Missing audit trail entries

**Solutions:**
1. Verify business event logging is enabled:
   ```bash
   LOGGING_BUSINESS_EVENTS=true
   ```
2. Check if services are calling event logging functions:
   ```go
   utils.LogTransactionEvent(ctx, logger, "transaction.completed", trxID, details)
   ```
3. Verify logger is properly initialized in services

---

## Best Practices

### Development
- Use `LOGGING_LEVEL=debug` for verbose output
- Enable caller information for debugging: `LOGGING_INCLUDE_CALLER=true`
- Keep redaction enabled to build secure habits

### Production
- Use `LOGGING_LEVEL=info` or `warn`
- Disable caller information for performance: `LOGGING_INCLUDE_CALLER=false`
- Always enable redaction: `LOGGING_REDACT_ENABLED=true`
- Configure Docker log rotation limits

### Security
- **Never** log passwords, tokens, or API keys (even hashed)
- **Never** log full JWT tokens (log token ID only)
- **Never** log credit card numbers (log last 4 digits only)
- Redact PII (Personally Identifiable Information) when required

### Performance
- Structured logging has minimal overhead (<1ms per log entry)
- Caller information adds ~1μs per log entry (disable in production for max performance)
- Log level filtering prevents unnecessary processing

---

## References

- [Go log/slog Package Documentation](https://pkg.go.dev/log/slog)
- [Docker Logging Drivers](https://docs.docker.com/config/containers/logging/configure/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [ELK Stack Guide](https://www.elastic.co/guide/en/elastic-stack/current/index.html)

---

## Story References

- **Story 9.5:** Implement Structured Logging with Zap
- **Acceptance Criteria:**
  - AC1: JSON format with complete fields (level, timestamp, caller, message, context)
  - AC2: Docker-ready output (stdout, JSON)
  - AC3: Log rotation (container-level)
  - AC4: Sensitive data redaction
  - AC5: Configurable log levels
  - AC6: Business event logging
