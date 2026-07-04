package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/utils"
)

// visitor tracks a single user's token-bucket state.
type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// rateLimiter is a simple in-memory per-user token-bucket limiter. Suitable for
// a single-instance deployment; swap for Redis if horizontally scaled.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[uuid.UUID]*visitor
	rate     float64 // tokens added per second
	burst    float64 // max tokens (bucket size)
}

func newRateLimiter(ratePerMinute, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[uuid.UUID]*visitor),
		rate:     float64(ratePerMinute) / 60.0,
		burst:    float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(userID uuid.UUID) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[userID]
	if !ok {
		rl.visitors[userID] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// cleanup evicts stale visitors to bound memory usage.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for id, v := range rl.visitors {
			if time.Since(v.lastSeen) > 15*time.Minute {
				delete(rl.visitors, id)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitPerUser returns middleware that limits each authenticated user to
// `ratePerMinute` requests, allowing short bursts up to `burst`. Must run after
// AuthMiddleware (relies on the "userID" context key).
func RateLimitPerUser(ratePerMinute, burst int) gin.HandlerFunc {
	rl := newRateLimiter(ratePerMinute, burst)
	return func(c *gin.Context) {
		idVal, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}
		userID, ok := idVal.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}
		if !rl.allow(userID) {
			utils.TooManyRequests(c, "You're sending messages too quickly. Please slow down.")
			c.Abort()
			return
		}
		c.Next()
	}
}
