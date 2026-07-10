package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger replaces gin's default plain-text access log with
// structured (slog) request logging, tagged with the request id assigned by
// RequestID so every log line for a given request can be correlated.
//
// RequestID must run before this middleware in the chain so GetRequestID has
// a value to log.
func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"request_id", GetRequestID(c),
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if userID := GetUserID(c); userID != "" {
			attrs = append(attrs, "user_id", userID)
		}
		if schoolID := schoolIDForLog(c); schoolID != "" {
			attrs = append(attrs, "school_id", schoolID)
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		logger.Log(c.Request.Context(), level, "http_request", attrs...)
	}
}

// schoolIDForLog reads the school id resolved by RequireSchoolMember/
// RequireRole/RequireSystemSuperAdmin if the route went through one of them,
// falling back to the raw SchoolId header so requests rejected before that
// validation still get tagged for observability.
func schoolIDForLog(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if s, ok := sid.(string); ok && s != "" {
			return s
		}
	}
	return c.GetHeader("SchoolId")
}
