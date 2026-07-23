package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiterStore interface {
	Allow(key string) bool
	// Reset clears any tracked state for key, restoring it to full budget.
	Reset(key string)
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type InMemoryRateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rateLimiterEntry
	rps      rate.Limit
	burst    int
	ttl      time.Duration
}

func NewInMemoryRateLimiterStore(rps float64, burst int, ttl time.Duration) *InMemoryRateLimiterStore {
	store := &InMemoryRateLimiterStore{
		limiters: make(map[string]*rateLimiterEntry),
		rps:      rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
	go store.sweepLoop()
	return store
}

func (s *InMemoryRateLimiterStore) Allow(key string) bool {
	s.mu.Lock()
	entry, exists := s.limiters[key]
	if !exists {
		entry = &rateLimiterEntry{limiter: rate.NewLimiter(s.rps, s.burst)}
		s.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	limiter := entry.limiter
	s.mu.Unlock()

	return limiter.Allow()
}

// Reset deletes key's tracked entry entirely, so the next Allow call starts
// a fresh limiter at full burst — used to clear a lockout on success (e.g.
// a correct password after prior failed attempts).
func (s *InMemoryRateLimiterStore) Reset(key string) {
	s.mu.Lock()
	delete(s.limiters, key)
	s.mu.Unlock()
}

func (s *InMemoryRateLimiterStore) sweepLoop() {
	interval := s.ttl
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-s.ttl)
		s.mu.Lock()
		for key, entry := range s.limiters {
			if entry.lastSeen.Before(cutoff) {
				delete(s.limiters, key)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimitPerTenant limits requests per school (tenant) rather than per global IP
func RateLimitPerTenant(store RateLimiterStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !store.Allow(tenantRateLimitKey(c)) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak permintaan, coba lagi sebentar lagi."})
			c.Abort()
			return
		}
		c.Next()
	}
}

func tenantRateLimitKey(c *gin.Context) string {
	if schoolID := c.GetHeader("SchoolId"); schoolID != "" {
		return "school:" + schoolID
	}
	return "ip:" + c.ClientIP()
}

// RateLimitByIP is a coarse, IP-only pre-check 
func RateLimitByIP(store RateLimiterStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !store.Allow("ip:" + c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak permintaan, coba lagi sebentar lagi."})
			c.Abort()
			return
		}
		c.Next()
	}
}
