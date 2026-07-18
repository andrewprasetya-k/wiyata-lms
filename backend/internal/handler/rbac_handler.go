package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	service           service.RBACService
	schoolUserService service.SchoolUserService
}

func NewRBACHandler(service service.RBACService, schoolUserService service.SchoolUserService) *RBACHandler {
	return &RBACHandler{service: service, schoolUserService: schoolUserService}
}

func getRBACSchoolID(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

// Role Handlers
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var input dto.CreateRoleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	role := domain.Role{
		Name: input.Name,
	}

	if err := h.service.CreateRole(&role); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapRoleToResponse(&role))
}

func (h *RBACHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.service.GetAllRoles()
	if err != nil {
		HandleError(c, err)
		return
	}

	response := []dto.RoleResponseDTO{}
	for _, r := range roles {
		response = append(response, h.mapRoleToResponse(r))
	}

	c.JSON(http.StatusOK, response)
}

func (h *RBACHandler) GetRoleByID(c *gin.Context) {
	id := c.Param("id")
	role, err := h.service.GetRoleByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapRoleToResponse(role))
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateRoleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	role, err := h.service.GetRoleByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if input.Name != nil {
		role.Name = *input.Name
	}

	if err := h.service.UpdateRole(role); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapRoleToResponse(role))
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRole(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// User-Role Handlers
func (h *RBACHandler) AssignRole(c *gin.Context) {
	var input dto.AssignRoleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := h.service.AssignRoleToUser(input.SchoolUserID, input.RoleID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role assigned successfully"})
}

func (h *RBACHandler) RemoveRole(c *gin.Context) {
	schoolUserID := c.Query("schoolUserId")
	roleID := c.Query("roleId")

	if err := h.service.RemoveRoleFromUser(schoolUserID, roleID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	schoolUserID := c.Param("schoolUserId")
	schoolID := getRBACSchoolID(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	belongs, err := h.schoolUserService.BelongsToSchool(schoolUserID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !belongs {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: school user does not belong to active school"})
		return
	}

	userRoles, err := h.service.GetUserRoles(schoolUserID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := []dto.RoleResponseDTO{}
	for _, ur := range userRoles {
		response = append(response, h.mapRoleToResponse(&ur.Role))
	}

	c.JSON(http.StatusOK, response)
}

func (h *RBACHandler) UpdateUserRoles(c *gin.Context) {
	schoolUserID := c.Param("schoolUserId")
	var input dto.SyncUserRolesDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := h.service.SyncUserRoles(schoolUserID, input.RoleIDs); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User roles updated successfully"})
}

func (h *RBACHandler) CreateSuperAdmin(c *gin.Context) {
	var input dto.CreateUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	if err := h.service.CreateSuperAdmin(input.FullName, input.Email, input.Password); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Super admin created successfully"})
}

// Helpers
func (h *RBACHandler) mapRoleToResponse(role *domain.Role) dto.RoleResponseDTO {
	return dto.RoleResponseDTO{
		ID:        role.ID,
		Name:      role.Name,
		CreatedAt: formatAPITime(role.CreatedAt),
	}
}
