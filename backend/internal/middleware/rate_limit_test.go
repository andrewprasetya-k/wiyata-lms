package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitPerTenantAllowsWithinBurstAndBlocksAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryRateLimiterStore(1, 2, time.Minute)

	router := gin.New()
	router.GET("/protected", RateLimitPerTenant(store), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	doRequest := func(schoolID string) int {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if schoolID != "" {
			req.Header.Set("SchoolId", schoolID)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Code
	}

	// Burst of 2 allowed immediately for the same tenant.
	if code := doRequest("school-a"); code != http.StatusOK {
		t.Fatalf("request 1: status = %d, want 200", code)
	}
	if code := doRequest("school-a"); code != http.StatusOK {
		t.Fatalf("request 2: status = %d, want 200", code)
	}
	// Third immediate request exceeds the burst.
	if code := doRequest("school-a"); code != http.StatusTooManyRequests {
		t.Fatalf("request 3: status = %d, want 429", code)
	}

	// A different tenant has its own independent bucket.
	if code := doRequest("school-b"); code != http.StatusOK {
		t.Fatalf("other tenant request: status = %d, want 200", code)
	}
}

func TestRateLimitPerTenantFallsBackToIPWithoutSchoolHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryRateLimiterStore(1, 1, time.Minute)

	router := gin.New()
	router.GET("/public", RateLimitPerTenant(store), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/public", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/public", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request (same IP, no school header): status = %d, want 429", rec2.Code)
	}
}

func TestInMemoryRateLimiterStoreSweepsStaleEntries(t *testing.T) {
	store := NewInMemoryRateLimiterStore(1, 1, 20*time.Millisecond)

	store.Allow("key-1")
	store.mu.Lock()
	initialCount := len(store.limiters)
	store.mu.Unlock()
	if initialCount != 1 {
		t.Fatalf("expected 1 tracked key, got %d", initialCount)
	}

	time.Sleep(80 * time.Millisecond)

	store.mu.Lock()
	remaining := len(store.limiters)
	store.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("expected stale entry to be swept, %d entries remain", remaining)
	}
}
