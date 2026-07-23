package middleware

import (
	"time"
)

func NewGeneralAPIStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(20, 40, 10*time.Minute)
}

func NewChangePasswordAttemptStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(5.0/(15*60), 5, 20*time.Minute)
}

func NewRefreshTokenIPAttemptStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(30.0/(10*60), 30, 20*time.Minute)
}

func NewRefreshTokenFamilyAttemptStore() *InMemoryRateLimiterStore {
	return NewInMemoryRateLimiterStore(8.0/(10*60), 8, 20*time.Minute)
}
