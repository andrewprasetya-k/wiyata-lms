package middleware

import (
	"backend/internal/repository"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

var rbacRepo repository.RBACRepository

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

		// kalau super admin, bypass cek membership sekolah
		isSuperAdmin, err := rbacRepo.IsSuperAdmin(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify permissions"})
			c.Abort()
			return
		}

		if !isSuperAdmin {
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

		roles, err := rbacRepo.GetUserRoleNamesInSchool(userID, schoolID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
			c.Abort()
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

// RequireSystemSuperAdmin checks whether the current user has super_admin role
// on the system school (sch_code = 0000), regardless of the active SchoolId header.
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

		systemSchoolID, err := schoolService.ConvertCodeToID("0000")
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
