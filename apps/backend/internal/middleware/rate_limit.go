package middleware

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/time/rate"

	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
)

// Storage abstracts the backing store for per-key limiters.
type Storage interface {
	Add(string, *rate.Limiter) bool
	Get(string) (*rate.Limiter, bool)
}

var (
	// Default LRU capacity and TTL for limiter entries.
	DefaultCacheSize = 5000
	DefaultTTL       = 6 * time.Hour
)

// Default in-memory store (LRU with TTL).
var defaultStore = expirable.NewLRU[string, *rate.Limiter](DefaultCacheSize, nil, DefaultTTL)

// NewRateLimitMiddleware installs a token-bucket rate limiter per key.
//
// The rate limiter uses a token-bucket algorithm with the following parameters:
//   - R = requests / window (tokens per second)
//   - Burst = requests (allows short spikes up to N concurrent requests)
//
// When the rate limit is exceeded, the middleware:
//   - Aborts the request with HTTP 429 Too Many Requests
//   - Sets Retry-After header indicating when to retry
//   - Returns RFC 7807 compliant error response
//
// Rate Limit Key Function:
//   - For authenticated requests: extracts user ID from JWT context ("user" key)
//   - For unauthenticated requests: falls back to IP address
//   - Key format: "user:{userID}" or "ip:{ipAddress}"
//
// Response Headers (all responses):
//   - X-RateLimit-Limit: Maximum requests allowed in window
//   - X-RateLimit-Remaining: Requests remaining in current window
//   - X-RateLimit-Reset: Unix timestamp when window resets
//
// Response Headers (on 429):
//   - Retry-After: Seconds until client should retry
//
// Rate Limit Response (429 Too Many Requests):
//   {
//     "success": false,
//     "error": {
//       "type": "https://api.simpo.com/errors/TOO_MANY_REQUESTS",
//       "title": "Rate Limit Exceeded",
//       "status": 429,
//       "detail": "Rate limit exceeded",
//       "code": "TOO_MANY_REQUESTS",
//       "details": "Too many requests. Please try again in {N} seconds.",
//       "retry_after": {N}
//     }
//   }
//
// Example usage:
//
//	router.Use(middleware.NewRateLimitMiddleware(
//	    time.Minute,        // 1 minute window
//	    100,                // 100 requests per window
//	    func(c *gin.Context) string {
//	        if userID, exists := c.Get("userID"); exists {
//	            return fmt.Sprintf("user:%v", userID)
//	        }
//	        return "ip:" + c.ClientIP()
//	    },
//	    nil,                // Use default LRU store
//	))
func NewRateLimitMiddleware(
	window time.Duration,
	requests int,
	keyFunc func(*gin.Context) string,
	store Storage,
) gin.HandlerFunc {

	if store == nil {
		store = defaultStore
	}

	r := rate.Limit(float64(requests) / window.Seconds())
	burst := requests

	return func(c *gin.Context) {
		key := keyFunc(c)

		lim, ok := store.Get(key)
		if !ok {
			lim = rate.NewLimiter(r, burst)
			store.Add(key, lim)
		}

		res := lim.Reserve()
		delay := res.Delay()

		if delay > 0 {
			res.Cancel()
			ra := int(math.Ceil(delay.Seconds()))
			resetAt := time.Now().Add(time.Duration(ra) * time.Second).Unix()

			c.Header("Retry-After", strconv.Itoa(ra))
			c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

			_ = c.Error(apiErrors.TooManyRequests(ra))
			c.Abort()
			return
		}

		remaining := lim.Tokens()
		resetAt := time.Now().Add(window).Unix()

		c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		c.Next()
	}
}
