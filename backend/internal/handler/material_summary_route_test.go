package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMaterialSummaryRouteDoesNotConflictWithProgressRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	materials := router.Group("/materials")
	{
		materials.POST("/:materialId/media/:mediaId/summary", func(c *gin.Context) {
			c.String(http.StatusOK, "summary")
		})
		materials.POST("/progress", func(c *gin.Context) {
			c.String(http.StatusOK, "progress")
		})
	}

	summaryRequest := httptest.NewRequest(http.MethodPost, "/materials/material-1/media/media-1/summary", nil)
	summaryRecorder := httptest.NewRecorder()
	router.ServeHTTP(summaryRecorder, summaryRequest)
	if summaryRecorder.Code != http.StatusOK || summaryRecorder.Body.String() != "summary" {
		t.Fatalf("summary route = %d %q, want 200 summary", summaryRecorder.Code, summaryRecorder.Body.String())
	}

	progressRequest := httptest.NewRequest(http.MethodPost, "/materials/progress", nil)
	progressRecorder := httptest.NewRecorder()
	router.ServeHTTP(progressRecorder, progressRequest)
	if progressRecorder.Code != http.StatusOK || progressRecorder.Body.String() != "progress" {
		t.Fatalf("progress route = %d %q, want 200 progress", progressRecorder.Code, progressRecorder.Body.String())
	}
}
