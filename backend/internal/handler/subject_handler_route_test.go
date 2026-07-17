package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type subjectRouteSchoolServiceStub struct{}

func (subjectRouteSchoolServiceStub) CreateSchool(*domain.School, string) (*domain.SchoolUser, error) {
	return nil, nil
}
func (subjectRouteSchoolServiceStub) GetSchools(string, string, int, int, string, string) ([]*domain.School, int64, error) {
	return nil, 0, nil
}
func (subjectRouteSchoolServiceStub) GetSchoolByCode(string) (*domain.School, error) { return nil, nil }
func (subjectRouteSchoolServiceStub) GetSchoolByID(string) (*domain.School, error)   { return nil, nil }
func (subjectRouteSchoolServiceStub) RestoreDeletedSchool(string) error              { return nil }
func (subjectRouteSchoolServiceStub) UpdateSchool(*domain.School) error              { return nil }
func (subjectRouteSchoolServiceStub) DeleteSchool(string) error                      { return nil }
func (subjectRouteSchoolServiceStub) HardDeleteSchool(string) error                  { return nil }
func (subjectRouteSchoolServiceStub) GetSchoolSummary() (*dto.SchoolSummaryDTO, error) {
	return nil, nil
}
func (subjectRouteSchoolServiceStub) CheckCodeAvailability(string) (bool, error) { return true, nil }
func (subjectRouteSchoolServiceStub) ConvertCodeToID(code string) (string, error) {
	return code, nil
}

type subjectRouteServiceStub struct {
	subjects map[string]*domain.Subject
}

func (s *subjectRouteServiceStub) Create(*domain.Subject) error { return nil }

func (s *subjectRouteServiceStub) FindAll(string, string, int, int) ([]*domain.Subject, int64, error) {
	return nil, 0, nil
}

func (s *subjectRouteServiceStub) GetBySchool(string) ([]*domain.Subject, error) { return nil, nil }

func (s *subjectRouteServiceStub) GetByID(id string) (*domain.Subject, error) {
	subject, ok := s.subjects[id]
	if !ok {
		return nil, errors.New("subject not found")
	}
	return subject, nil
}

func (s *subjectRouteServiceStub) GetByCode(string, string) (*domain.Subject, error) {
	return nil, nil
}

func (s *subjectRouteServiceStub) Update(*domain.Subject) error { return nil }
func (s *subjectRouteServiceStub) Delete(string) error           { return nil }

// TestSubjectGetByIDCrossTenantIsForbidden is a representative regression test for the
// Sprint 1 IDOR closure: a member of school A must not be able to read a subject that
// belongs to school B by guessing/iterating its ID, even though route middleware
// (RequireSchoolMember) alone would let the request through since it only validates
// membership of the school named in the header, not ownership of the requested resource.
func TestSubjectGetByIDCrossTenantIsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware.InitRBAC(&mediaRouteRBACStub{
		membership: map[string]map[string]bool{
			"user-1": {"school-a": true},
		},
	})

	subjectSvc := &subjectRouteServiceStub{
		subjects: map[string]*domain.Subject{
			"subject-in-school-a": {ID: "subject-in-school-a", SchoolID: "school-a"},
			"subject-in-school-b": {ID: "subject-in-school-b", SchoolID: "school-b"},
		},
	}
	subjectHandler := NewSubjectHandler(subjectSvc, subjectRouteSchoolServiceStub{})

	router := gin.New()
	router.GET(
		"/subjects/:id",
		func(c *gin.Context) {
			c.Set("user", jwt.MapClaims{"user_id": "user-1"})
			c.Next()
		},
		middleware.RequireSchoolMember(subjectRouteSchoolServiceStub{}),
		subjectHandler.GetByID,
	)

	tests := []struct {
		name       string
		subjectID  string
		wantStatus int
	}{
		{name: "own school subject is readable", subjectID: "subject-in-school-a", wantStatus: http.StatusOK},
		{name: "other school subject is forbidden", subjectID: "subject-in-school-b", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/subjects/"+tt.subjectID, nil)
			req.Header.Set("SchoolId", "school-a")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}
