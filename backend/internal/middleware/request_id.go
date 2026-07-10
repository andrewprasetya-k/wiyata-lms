package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is both the incoming header this middleware honors (so a
// caller/load-balancer that already generated a request id has it preserved
// end-to-end) and the response header it always sets.
const RequestIDHeader = "X-Request-ID"

const requestIDContextKey = "request_id"

// RequestID assigns a request id to every request — reusing an incoming
// X-Request-ID header if the caller already set one, otherwise generating a
// new one — stores it in the gin context for handlers/logging to use, and
// echoes it back on the response header so the caller can correlate logs.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set(requestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID returns the request id assigned by RequestID, or "" if the
// middleware was not registered on this route.
func GetRequestID(c *gin.Context) string {
	if v, exists := c.Get(requestIDContextKey); exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
