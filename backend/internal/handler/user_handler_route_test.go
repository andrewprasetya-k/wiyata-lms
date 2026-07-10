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

type userRouteServiceStub struct{}

func (s *userRouteServiceStub) Create(*domain.User) error { return nil }

func (s *userRouteServiceStub) FindAll(string, int, int) ([]*domain.User, int64, error) {
	return []*domain.User{{ID: "user-1", FullName: "Test User", Email: "test@example.com"}}, 1, nil
}

func (s *userRouteServiceStub) GetByID(string) (*domain.User, error)    { return nil, nil }
func (s *userRouteServiceStub) GetByEmail(string) (*domain.User, error) { return nil, nil }
func (s *userRouteServiceStub) Update(*domain.User) error               { return nil }
func (s *userRouteServiceStub) Delete(string) error                     { return nil }
func (s *userRouteServiceStub) ChangePassword(string, string, string) error {
	return nil
}

// userRouteRBACStub lets each test control exactly which role (if any) a user
// holds in the system school ("000000"), which is what RequireSystemSuperAdmin
// checks against.
type userRouteRBACStub struct {
	// key: "userID:schoolID" -> role names
	rolesByUserSchool map[string][]string
}

func (s *userRouteRBACStub) CreateRole(*domain.Role) error                       { return nil }
func (s *userRouteRBACStub) GetRoleByID(string) (*domain.Role, error)            { return nil, nil }
func (s *userRouteRBACStub) GetAllRoles() ([]*domain.Role, error)                { return nil, nil }
func (s *userRouteRBACStub) UpdateRole(*domain.Role) error                       { return nil }
func (s *userRouteRBACStub) DeleteRole(string) error                             { return nil }
func (s *userRouteRBACStub) CheckDuplicateRoleName(string, string) (bool, error) { return false, nil }
func (s *userRouteRBACStub) AssignRole(*domain.UserRole) error                   { return nil }
func (s *userRouteRBACStub) RemoveRoleFromUser(string, string) error             { return nil }
func (s *userRouteRBACStub) GetUserRoles(string) ([]*domain.UserRole, error)     { return nil, nil }
func (s *userRouteRBACStub) SyncUserRoles(string, []string) error                { return nil }
func (s *userRouteRBACStub) GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error) {
	return s.rolesByUserSchool[userID+":"+schoolID], nil
}
func (s *userRouteRBACStub) IsUserInSchool(userID, schoolID string) (bool, error) {
	return s.rolesByUserSchool[userID+":"+schoolID] != nil, nil
}
func (s *userRouteRBACStub) GetSchoolUserID(userID, schoolID string) (string, error) {
	return "school-user-" + userID, nil
}
func (s *userRouteRBACStub) IsSuperAdmin(userID string) (bool, error) {
	for _, role := range s.rolesByUserSchool[userID+":"+middleware.SystemSchoolCode] {
		if role == "super_admin" {
			return true, nil
		}
	}
	return false, nil
}

// TestUserListRequiresSystemSuperAdmin is a regression test for Sprint 6: GET
// /users used to accept any tenant "admin" role (via RequireRole), which let a
// school admin from any single school read the full cross-tenant user
// directory (full name + email of every user on the platform). It must now
// require system super admin, exactly like every other /users operation.
func TestUserListRequiresSystemSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// RequireSystemSuperAdmin always checks the role a user holds in the
	// system school (middleware.SystemSchoolCode), regardless of any
	// SchoolId header the caller sends. school-admin-1/teacher-1/student-1
	// each hold a real, legitimate role in a regular tenant school
	// ("school-a") — included here to make explicit that having ANY role in
	// ANY ordinary school must not grant access to this endpoint.
	repo := &userRouteRBACStub{
		rolesByUserSchool: map[string][]string{
			"super-admin-1:" + middleware.SystemSchoolCode: {"super_admin"},
			"school-admin-1:school-a":                       {"admin"},
			"teacher-1:school-a":                            {"teacher"},
			"student-1:school-a":                            {"student"},
		},
	}
	middleware.InitRBAC(repo)

	userHandler := NewUserHandler(&userRouteServiceStub{})

	router := gin.New()
	router.GET(
		"/users",
		func(c *gin.Context) {
			userID := c.GetHeader("X-Test-User-ID")
			c.Set("user", jwt.MapClaims{"user_id": userID})
			c.Next()
		},
		middleware.RequireSystemSuperAdmin(userRouteSchoolServiceStub{}),
		userHandler.FindAll,
	)

	tests := []struct {
		name       string
		userID     string
		wantStatus int
	}{
		{name: "system super admin is allowed", userID: "super-admin-1", wantStatus: http.StatusOK},
		{name: "school admin (tenant role only) is rejected", userID: "school-admin-1", wantStatus: http.StatusForbidden},
		{name: "teacher is rejected", userID: "teacher-1", wantStatus: http.StatusForbidden},
		{name: "student is rejected", userID: "student-1", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			req.Header.Set("X-Test-User-ID", tt.userID)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}

type userRouteSchoolServiceStub struct{}

func (userRouteSchoolServiceStub) ConvertCodeToID(code string) (string, error) {
	return code, nil
}
