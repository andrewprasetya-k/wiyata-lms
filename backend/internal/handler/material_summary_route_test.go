package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func TestMaterialSummaryRouteAllowsOnlyStudentActiveRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware.InitRBAC(&materialSummaryRBACStub{
		rolesByUserSchool: map[string][]string{
			"user-1:school-1": []string{"admin", "teacher", "student"},
		},
	})

	router := gin.New()
	materials := router.Group("/materials")
	materials.POST(
		"/:materialId/media/:mediaId/summary",
		func(c *gin.Context) {
			c.Set("user", jwt.MapClaims{"user_id": "user-1"})
			c.Next()
		},
		middleware.RequireSchoolMember(materialSummarySchoolServiceStub{}),
		middleware.RequireRole(materialSummarySchoolServiceStub{}, "student"),
		func(c *gin.Context) {
			c.String(http.StatusOK, "summary")
		},
	)

	tests := []struct {
		name       string
		activeRole string
		wantStatus int
	}{
		{name: "student allowed", activeRole: "student", wantStatus: http.StatusOK},
		{name: "teacher rejected", activeRole: "teacher", wantStatus: http.StatusForbidden},
		{name: "admin rejected", activeRole: "admin", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/materials/material-1/media/media-1/summary", nil)
			request.Header.Set("SchoolId", "school-1")
			request.Header.Set("Active-Role", tt.activeRole)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}

type materialSummarySchoolServiceStub struct{}

func (materialSummarySchoolServiceStub) ConvertCodeToID(code string) (string, error) {
	return code, nil
}

type materialSummaryRBACStub struct {
	rolesByUserSchool map[string][]string
}

func (s *materialSummaryRBACStub) CreateRole(*domain.Role) error { return nil }
func (s *materialSummaryRBACStub) GetRoleByID(string) (*domain.Role, error) {
	return nil, nil
}
func (s *materialSummaryRBACStub) GetAllRoles() ([]*domain.Role, error) { return nil, nil }
func (s *materialSummaryRBACStub) UpdateRole(*domain.Role) error        { return nil }
func (s *materialSummaryRBACStub) DeleteRole(string) error              { return nil }
func (s *materialSummaryRBACStub) CheckDuplicateRoleName(string, string) (bool, error) {
	return false, nil
}
func (s *materialSummaryRBACStub) AssignRole(*domain.UserRole) error { return nil }
func (s *materialSummaryRBACStub) RemoveRoleFromUser(string, string) error {
	return nil
}
func (s *materialSummaryRBACStub) GetUserRoles(string) ([]*domain.UserRole, error) {
	return nil, nil
}
func (s *materialSummaryRBACStub) SyncUserRoles(string, []string) error { return nil }
func (s *materialSummaryRBACStub) GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error) {
	return s.rolesByUserSchool[userID+":"+schoolID], nil
}
func (s *materialSummaryRBACStub) IsUserInSchool(userID, schoolID string) (bool, error) {
	return s.rolesByUserSchool[userID+":"+schoolID] != nil, nil
}
func (s *materialSummaryRBACStub) GetSchoolUserID(userID, schoolID string) (string, error) {
	return "school-user-1", nil
}
func (s *materialSummaryRBACStub) IsSuperAdmin(string) (bool, error) { return false, nil }
