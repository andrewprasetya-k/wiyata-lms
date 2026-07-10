package middleware

import (
	"backend/internal/domain"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type rbacTestRepoStub struct {
	// key: userID -> set of schoolIDs the user belongs to
	membership map[string]map[string]bool
	// key: "userID:schoolID" -> role names
	roles map[string][]string
	// key: userID -> is super admin
	superAdmins map[string]bool
}

func newRBACTestRepoStub() *rbacTestRepoStub {
	return &rbacTestRepoStub{
		membership:  map[string]map[string]bool{},
		roles:       map[string][]string{},
		superAdmins: map[string]bool{},
	}
}

func (s *rbacTestRepoStub) CreateRole(*domain.Role) error                          { return nil }
func (s *rbacTestRepoStub) GetRoleByID(string) (*domain.Role, error)               { return nil, nil }
func (s *rbacTestRepoStub) GetAllRoles() ([]*domain.Role, error)                   { return nil, nil }
func (s *rbacTestRepoStub) UpdateRole(*domain.Role) error                          { return nil }
func (s *rbacTestRepoStub) DeleteRole(string) error                                { return nil }
func (s *rbacTestRepoStub) CheckDuplicateRoleName(string, string) (bool, error)    { return false, nil }
func (s *rbacTestRepoStub) AssignRole(*domain.UserRole) error                      { return nil }
func (s *rbacTestRepoStub) RemoveRoleFromUser(string, string) error                { return nil }
func (s *rbacTestRepoStub) GetUserRoles(string) ([]*domain.UserRole, error)        { return nil, nil }
func (s *rbacTestRepoStub) SyncUserRoles(string, []string) error                   { return nil }

func (s *rbacTestRepoStub) GetUserRoleNamesInSchool(userID, schoolID string) ([]string, error) {
	return s.roles[userID+":"+schoolID], nil
}

func (s *rbacTestRepoStub) IsUserInSchool(userID, schoolID string) (bool, error) {
	return s.membership[userID][schoolID], nil
}

func (s *rbacTestRepoStub) GetSchoolUserID(userID, schoolID string) (string, error) {
	if !s.membership[userID][schoolID] {
		return "", errors.New("not a member of this school")
	}
	return "school-user-" + userID, nil
}

func (s *rbacTestRepoStub) IsSuperAdmin(userID string) (bool, error) {
	return s.superAdmins[userID], nil
}

type rbacTestSchoolServiceStub struct{}

func (rbacTestSchoolServiceStub) ConvertCodeToID(code string) (string, error) {
	return code, nil
}

func withAuthenticatedUser(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", jwt.MapClaims{"user_id": userID})
		c.Next()
	}
}

func newRBACTestRouter(t *testing.T, repo *rbacTestRepoStub, mw gin.HandlerFunc, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	InitRBAC(repo)

	router := gin.New()
	router.GET("/protected", withAuthenticatedUser(userID), mw, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return router
}

func doRequest(router *gin.Engine, schoolID string, activeRole string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if schoolID != "" {
		req.Header.Set("SchoolId", schoolID)
	}
	if activeRole != "" {
		req.Header.Set("Active-Role", activeRole)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestRequireSchoolMember(t *testing.T) {
	repo := newRBACTestRepoStub()
	repo.membership["user-1"] = map[string]bool{"school-a": true}

	router := newRBACTestRouter(t, repo, RequireSchoolMember(rbacTestSchoolServiceStub{}), "user-1")

	t.Run("allow: member of the requested school", func(t *testing.T) {
		rec := doRequest(router, "school-a", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %q", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject: wrong school (not a member)", func(t *testing.T) {
		rec := doRequest(router, "school-b", "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403, body = %q", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject: missing school context", func(t *testing.T) {
		rec := doRequest(router, "", "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400, body = %q", rec.Code, rec.Body.String())
		}
	})
}

func TestRequireSchoolMemberSuperAdminBypass(t *testing.T) {
	repo := newRBACTestRepoStub()
	repo.superAdmins["super-1"] = true
	// deliberately NOT a member of "school-a" to prove the bypass

	router := newRBACTestRouter(t, repo, RequireSchoolMember(rbacTestSchoolServiceStub{}), "super-1")

	t.Run("allow: super admin bypasses membership check", func(t *testing.T) {
		rec := doRequest(router, "school-a", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %q", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject: super admin with explicit active role still needs real membership", func(t *testing.T) {
		rec := doRequest(router, "school-a", "admin")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403, body = %q", rec.Code, rec.Body.String())
		}
	})
}

func TestRequireRole(t *testing.T) {
	repo := newRBACTestRepoStub()
	repo.membership["user-1"] = map[string]bool{"school-a": true}
	repo.roles["user-1:school-a"] = []string{"teacher"}

	router := newRBACTestRouter(t, repo, RequireRole(rbacTestSchoolServiceStub{}, "admin", "teacher"), "user-1")

	t.Run("allow: user has one of the allowed roles", func(t *testing.T) {
		rec := doRequest(router, "school-a", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %q", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject: missing role in this school", func(t *testing.T) {
		rec := doRequest(router, "school-b", "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403, body = %q", rec.Code, rec.Body.String())
		}
	})
}

func TestRequireRoleWrongRole(t *testing.T) {
	repo := newRBACTestRepoStub()
	repo.membership["user-1"] = map[string]bool{"school-a": true}
	repo.roles["user-1:school-a"] = []string{"student"}

	router := newRBACTestRouter(t, repo, RequireRole(rbacTestSchoolServiceStub{}, "admin", "teacher"), "user-1")

	t.Run("reject: role assigned but not allowed for this route", func(t *testing.T) {
		rec := doRequest(router, "school-a", "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403, body = %q", rec.Code, rec.Body.String())
		}
	})
}

func TestRequireSystemSuperAdmin(t *testing.T) {
	repo := newRBACTestRepoStub()
	repo.roles["super-1:"+SystemSchoolCode] = []string{"super_admin"}
	repo.roles["user-1:"+SystemSchoolCode] = []string{}

	router := newRBACTestRouter(t, repo, RequireSystemSuperAdmin(rbacTestSchoolServiceStub{}), "super-1")
	otherRouter := newRBACTestRouter(t, repo, RequireSystemSuperAdmin(rbacTestSchoolServiceStub{}), "user-1")

	t.Run("allow: user has super_admin role in system school", func(t *testing.T) {
		rec := doRequest(router, "", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %q", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject: user without super_admin role in system school", func(t *testing.T) {
		rec := doRequest(otherRouter, "", "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403, body = %q", rec.Code, rec.Body.String())
		}
	})
}
