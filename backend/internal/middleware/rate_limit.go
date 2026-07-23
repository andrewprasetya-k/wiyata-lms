package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterStore is the minimal interface a rate-limit backend must
// satisfy. Kept small and swappable on purpose: today it's backed by an
// in-memory token bucket per key, but a Redis-backed (or any distributed)
// implementation can replace it later without touching the middleware or
// any call site, as long as it satisfies this interface.
type RateLimiterStore interface {
	Allow(key string) bool
	// Reset clears any tracked state for key, restoring it to full budget.
	Reset(key string)
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// InMemoryRateLimiterStore is a simple per-key token bucket limiter. It is
// intentionally in-process only (no Redis) — good enough for a single API
// instance; if the app is ever scaled to multiple instances behind a load
// balancer, swap the RateLimiterStore implementation for a shared backend.
type InMemoryRateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rateLimiterEntry
	rps      rate.Limit
	burst    int
	ttl      time.Duration
}

// NewInMemoryRateLimiterStore creates a store where each distinct key is
// allowed `rps` requests per second on average, with bursts up to `burst`.
// Entries idle for longer than `ttl` are swept periodically so the map
// doesn't grow unbounded over the life of the process.
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

// RateLimitPerTenant limits requests per school (tenant) rather than per
// global IP. The key is resolved as follows:
//
//  1. The "SchoolId" header, if present — this is the same header
//     RequireSchoolMember/RequireRole read to resolve the active school, so
//     it is available before any per-route authorization middleware runs.
//     Using the raw header (rather than the validated "school_id" context
//     value set later by RequireSchoolMember) is safe here: rate limiting is
//     an abuse-prevention concern, not an authorization decision — a caller
//     cannot use a forged SchoolId to affect another tenant's bucket, they
//     can only ever throttle themselves, and RequireSchoolMember still
//     independently rejects any request whose header doesn't correspond to
//     real membership regardless of what the rate limiter decided.
//  2. The client IP, for requests with no school context at all (the public,
//     pre-auth endpoints: login, register, school registration).
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
