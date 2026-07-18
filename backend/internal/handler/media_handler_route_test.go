package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type mediaRouteServiceStub struct {
	medias map[string]*domain.Media
}

func (s *mediaRouteServiceStub) RecordMetadata(*domain.Media) error { return nil }

func (s *mediaRouteServiceStub) UploadAndRecord(context.Context, *domain.Media, io.Reader) error {
	return nil
}

func (s *mediaRouteServiceStub) GetByID(id string) (*domain.Media, error) {
	media, ok := s.medias[id]
	if !ok {
		return nil, errors.New("media not found")
	}
	return media, nil
}

func (s *mediaRouteServiceStub) GetByOwner(string, string) ([]*domain.Media, error) {
	return nil, nil
}

func (s *mediaRouteServiceStub) Delete(context.Context, string) error { return nil }

type mediaRouteSchoolServiceStub struct{}

func (mediaRouteSchoolServiceStub) ConvertCodeToID(code string) (string, error) {
	return code, nil
}

type mediaRouteRBACStub struct {
	membership map[string]map[string]bool
}

func (s *mediaRouteRBACStub) CreateRole(*domain.Role) error                       { return nil }
func (s *mediaRouteRBACStub) GetRoleByID(string) (*domain.Role, error)            { return nil, nil }
func (s *mediaRouteRBACStub) GetRoleByName(string) (*domain.Role, error)          { return nil, nil }
func (s *mediaRouteRBACStub) WithTx(tx *gorm.DB) repository.RBACRepository        { return s }
func (s *mediaRouteRBACStub) GetAllRoles() ([]*domain.Role, error)                { return nil, nil }
func (s *mediaRouteRBACStub) UpdateRole(*domain.Role) error                       { return nil }
func (s *mediaRouteRBACStub) DeleteRole(string) error                             { return nil }
func (s *mediaRouteRBACStub) CheckDuplicateRoleName(string, string) (bool, error) { return false, nil }
func (s *mediaRouteRBACStub) AssignRole(*domain.UserRole) error                   { return nil }
func (s *mediaRouteRBACStub) RemoveRoleFromUser(string, string) error             { return nil }
func (s *mediaRouteRBACStub) GetUserRoles(string) ([]*domain.UserRole, error)     { return nil, nil }
func (s *mediaRouteRBACStub) SyncUserRoles(string, []string) error                { return nil }
func (s *mediaRouteRBACStub) GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error) {
	return nil, nil
}
func (s *mediaRouteRBACStub) IsUserInSchool(userID, schoolID string) (bool, error) {
	return s.membership[userID][schoolID], nil
}
func (s *mediaRouteRBACStub) GetSchoolUserID(userID, schoolID string) (string, error) {
	return "school-user-" + userID, nil
}
func (s *mediaRouteRBACStub) IsSuperAdmin(string) (bool, error) { return false, nil }

// TestMediaGetByIDCrossTenantIsForbidden is a regression test for the IDOR found in
// GET /medias/:id: a member of school A must not be able to read media that belongs
// to school B, even though they are a legitimate member of their own active school.
func TestMediaGetByIDCrossTenantIsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware.InitRBAC(&mediaRouteRBACStub{
		membership: map[string]map[string]bool{
			"user-1": {"school-a": true},
		},
	})

	mediaSvc := &mediaRouteServiceStub{
		medias: map[string]*domain.Media{
			"media-in-school-a": {ID: "media-in-school-a", SchoolID: "school-a"},
			"media-in-school-b": {ID: "media-in-school-b", SchoolID: "school-b"},
		},
	}
	mediaHandler := NewMediaHandler(mediaSvc)

	router := gin.New()
	router.GET(
		"/medias/:id",
		func(c *gin.Context) {
			c.Set("user", jwt.MapClaims{"user_id": "user-1"})
			c.Next()
		},
		middleware.RequireSchoolMember(mediaRouteSchoolServiceStub{}),
		mediaHandler.GetByID,
	)

	tests := []struct {
		name       string
		mediaID    string
		wantStatus int
	}{
		{name: "own school media is readable", mediaID: "media-in-school-a", wantStatus: http.StatusOK},
		{name: "other school media is forbidden", mediaID: "media-in-school-b", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/medias/"+tt.mediaID, nil)
			req.Header.Set("SchoolId", "school-a")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}
