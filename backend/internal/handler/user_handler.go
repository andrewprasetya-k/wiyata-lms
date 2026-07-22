package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var input dto.CreateUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	user := domain.User{
		FullName: input.FullName,
		Email:    input.Email,
		Password: input.Password,
	}

	actor := buildActorContext(c, domain.LogScopePlatform)
	if err := h.service.Create(actor, &user); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapToResponse(&user))
}

func (h *UserHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	search := c.Query("search")

	users, total, err := h.service.FindAll(search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := []dto.UserResponseDTO{}
	for _, u := range users {
		response = append(response, h.mapToResponse(u))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	paginatedResponse := dto.PaginatedResponse{
		Data:       response,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}
	c.JSON(http.StatusOK, paginatedResponse)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(user))
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if input.FullName != nil {
		user.FullName = *input.FullName
	}
	if input.Email != nil {
		user.Email = *input.Email
	}

	actor := buildActorContext(c, domain.LogScopePlatform)
	if err := h.service.Update(actor, user); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapToResponse(user))
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	actor := buildActorContext(c, domain.LogScopePlatform)
	if err := h.service.Delete(actor, id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	id := c.Param("id")
	var input dto.ChangePasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	actor := buildActorContext(c, domain.LogScopePlatform)
	if err := h.service.ChangePassword(actor, id, input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *UserHandler) mapToResponse(user *domain.User) dto.UserResponseDTO {
	return dto.UserResponseDTO{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive,
		CreatedAt: formatAPITime(user.CreatedAt),
	}
}
