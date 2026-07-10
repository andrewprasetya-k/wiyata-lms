package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestRequestIDGeneratesWhenMissingAndEchoesWhenPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/x", func(c *gin.Context) {
		if GetRequestID(c) == "" {
			t.Error("expected GetRequestID to return a non-empty value inside the handler")
		}
		c.String(http.StatusOK, "ok")
	})

	t.Run("generates one when absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Header().Get(RequestIDHeader) == "" {
			t.Fatal("expected X-Request-ID response header to be set")
		}
	})

	t.Run("echoes an incoming request id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set(RequestIDHeader, "client-supplied-id")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if got := rec.Header().Get(RequestIDHeader); got != "client-supplied-id" {
			t.Fatalf("X-Request-ID = %q, want %q", got, "client-supplied-id")
		}
	})
}

func TestStructuredLoggerIncludesExpectedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	router := gin.New()
	router.Use(RequestID())
	router.Use(StructuredLogger(logger))
	router.GET("/x", func(c *gin.Context) {
		c.Set("user", jwt.MapClaims{"user_id": "user-1"})
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("SchoolId", "school-1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	for _, key := range []string{"request_id", "method", "path", "status", "latency_ms", "user_id", "school_id"} {
		if _, ok := entry[key]; !ok {
			t.Errorf("log entry missing key %q, got: %v", key, entry)
		}
	}
	if entry["method"] != "GET" {
		t.Errorf("method = %v, want GET", entry["method"])
	}
	if entry["path"] != "/x" {
		t.Errorf("path = %v, want /x", entry["path"])
	}
	if entry["user_id"] != "user-1" {
		t.Errorf("user_id = %v, want user-1", entry["user_id"])
	}
	if entry["school_id"] != "school-1" {
		t.Errorf("school_id = %v, want school-1", entry["school_id"])
	}
}
