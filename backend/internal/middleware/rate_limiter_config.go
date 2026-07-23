package middleware

import (
	"time"
)

// NewGeneralAPIStore is the shared per-tenant limiter applied to
// login/register/forgot-password and every protected route: 20 req/s
// sustained, burst of 40, keyed by the SchoolId header (falls back to IP
// when there's no school context yet, e.g. pre-auth routes).
func NewGeneralAPIStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(20, 40, 10*time.Minute)
}

// NewChangePasswordAttemptStore backs the self-service change-password
// lockout: 5 failed current-password attempts, then locked out from
// further attempts for 15 minutes. burst=5 gives the 5 attempts; rps is
// tuned so the token bucket fully refills (back to 5 fresh attempts) after
// exactly 15 minutes of no further failed attempts. ttl (idle-entry sweep)
// is kept comfortably longer than the lock window so an entry can never be
// swept away mid-lock, which would otherwise shorten the lock early.
func NewChangePasswordAttemptStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(5.0/(15*60), 5, 20*time.Minute)
}

// NewRefreshTokenAttemptStore backs both the IP-scoped pre-check on
// POST /refresh-token and the per-family_id check inside
// AuthService.Refresh. Sized generously on purpose — legitimate rotation
// traffic (every ~15 min per active user) shouldn't compete with any
// general API burst budget. burst=8 tolerates a handful of rapid
// legitimate retries (e.g. a slow network causing a client-side retry, or
// concurrent tabs racing); rps refills gradually over 10 minutes.
func NewRefreshTokenAttemptStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(8.0/(10*60), 8, 20*time.Minute)
}
