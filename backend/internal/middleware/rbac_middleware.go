package middleware

import (
	"backend/internal/repository"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var rbacRepo repository.RBACRepository

const SystemSchoolCode = "000000"

var supportedActiveSchoolRoles = map[string]bool{
	"admin":   true,
	"teacher": true,
	"student": true,
}

// InitRBAC initializes RBAC middleware with repository
func InitRBAC(repo repository.RBACRepository) {
	rbacRepo = repo
}

// RequireSchoolMember checks if user belongs to the school
// Priority: SchoolId header > schoolCode URL param
// Note: super_admin bypasses school membership check
func RequireSchoolMember(schoolService interface {
	ConvertCodeToID(code string) (string, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if rbacRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RBAC middleware is not initialized"})
			c.Abort()
			return
		}

		var schoolID string
		var err error

		// cek school id di header
		schoolID = c.GetHeader("SchoolId")

		// cek school code di url param
		if schoolID == "" {
			schoolCode := c.Param("schoolCode")
			if schoolCode == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
				c.Abort()
				return
			}
			schoolID, err = schoolService.ConvertCodeToID(schoolCode)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "School not found"})
				c.Abort()
				return
			}
		}

		activeRole, hasActiveRole, ok := parseActiveRoleHeader(c)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported Active-Role"})
			c.Abort()
			return
		}

		// kalau super admin, bypass cek membership sekolah when no active role is selected.
		isSuperAdmin, err := rbacRepo.IsSuperAdmin(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify permissions"})
			c.Abort()
			return
		}

		if !isSuperAdmin || hasActiveRole {
			// check school membership kalau bukan super admin
			isMember, err := rbacRepo.IsUserInSchool(userID, schoolID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify school access"})
				c.Abort()
				return
			}

			if !isMember {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: not a member of this school"})
				c.Abort()
				return
			}

			schoolUserID, err := rbacRepo.GetSchoolUserID(userID, schoolID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: not a member of this school"})
				c.Abort()
				return
			}
			c.Set("school_user_id", schoolUserID)
		}

		if hasActiveRole {
			roles, err := rbacRepo.GetUserRoleNamesInSchool(userID, schoolID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
				c.Abort()
				return
			}
			// Cache the role names so a subsequent RequireRole in the same chain
			// can reuse them instead of querying the same data again.
			c.Set("school_role_names", roles)
			if !slices.Contains(roles, activeRole) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: active role is not assigned in this school"})
				c.Abort()
				return
			}
			c.Set("active_role", activeRole)
			c.Set("user_roles", []string{activeRole})
		}

		c.Set("school_id", schoolID)
		c.Next()
	}
}

// cek role tertentu (bisa multi-role)
func RequireRole(schoolService interface {
	ConvertCodeToID(code string) (string, error)
}, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if rbacRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RBAC middleware is not initialized"})
			c.Abort()
			return
		}

		var schoolID string
		var err error

		//cek school_id context apakah sudah ada (bisa di-set oleh RequireSchoolMember)
		if sid, exists := c.Get("school_id"); exists {
			schoolID = sid.(string)
		} else {
			// Cek SchoolId header
			schoolID = c.GetHeader("SchoolId")

			// cek schoolCode in URL param
			if schoolID == "" {
				schoolCode := c.Param("schoolCode")
				if schoolCode != "" {
					schoolID, err = schoolService.ConvertCodeToID(schoolCode)
					if err != nil {
						c.JSON(http.StatusNotFound, gin.H{"error": "School not found"})
						c.Abort()
						return
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
					c.Abort()
					return
				}
			}
		}

		activeRole, hasActiveRole, ok := activeRoleFromContextOrHeader(c)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported Active-Role"})
			c.Abort()
			return
		}

		// Reuse the role names already fetched by RequireSchoolMember earlier in
		// the chain, if present, instead of querying the same data again.
		var roles []string
		haveCachedRoles := false
		if cached, exists := c.Get("school_role_names"); exists {
			if cachedRoles, ok := cached.([]string); ok {
				roles = cachedRoles
				haveCachedRoles = true
			}
		}
		if !haveCachedRoles {
			var err error
			roles, err = rbacRepo.GetUserRoleNamesInSchool(userID, schoolID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
				c.Abort()
				return
			}
		}

		if hasActiveRole {
			if !slices.Contains(roles, activeRole) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: active role is not assigned in this school"})
				c.Abort()
				return
			}
			if !slices.Contains(allowedRoles, activeRole) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: active role is not allowed for this route"})
				c.Abort()
				return
			}
			c.Set("active_role", activeRole)
			c.Set("user_roles", []string{activeRole})
			c.Next()
			return
		}

		//cek apakah user memiliki role yang diizinkan
		hasRole := false
		for _, userRole := range roles {
			for _, allowedRole := range allowedRoles {
				if userRole == allowedRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("user_roles", roles)
		c.Next()
	}
}

func parseActiveRoleHeader(c *gin.Context) (string, bool, bool) {
	role := strings.ToLower(strings.TrimSpace(c.GetHeader("Active-Role")))
	if role == "" {
		return "", false, true
	}
	if !supportedActiveSchoolRoles[role] {
		return "", true, false
	}
	return role, true, true
}

func activeRoleFromContextOrHeader(c *gin.Context) (string, bool, bool) {
	if raw, exists := c.Get("active_role"); exists {
		if role, ok := raw.(string); ok && role != "" {
			return role, true, true
		}
	}
	return parseActiveRoleHeader(c)
}

// RequireSystemSuperAdmin checks whether the current user has super_admin role
// on the system school, regardless of the active SchoolId header.
func RequireSystemSuperAdmin(schoolService interface {
	ConvertCodeToID(code string) (string, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if rbacRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RBAC middleware is not initialized"})
			c.Abort()
			return
		}

		systemSchoolID, err := schoolService.ConvertCodeToID(SystemSchoolCode)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: system super admin context not available"})
			c.Abort()
			return
		}

		roles, err := rbacRepo.GetUserRoleNamesInSchool(userID, systemSchoolID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
			c.Abort()
			return
		}

		if slices.Contains(roles, "super_admin") {
			c.Set("school_id", systemSchoolID)
			c.Set("user_roles", roles)
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}
