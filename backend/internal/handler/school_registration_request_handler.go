package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SchoolRegistrationRequestHandler struct {
	service service.SchoolRegistrationRequestService
}

func NewSchoolRegistrationRequestHandler(service service.SchoolRegistrationRequestService) *SchoolRegistrationRequestHandler {
	return &SchoolRegistrationRequestHandler{service: service}
}

func (h *SchoolRegistrationRequestHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.CreateSchoolRegistrationRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, err := h.service.Create(input, userID)
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *SchoolRegistrationRequestHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	status := c.Query("status")

	response, err := h.service.List(status, page, limit)
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchoolRegistrationRequestHandler) GetByID(c *gin.Context) {
	response, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchoolRegistrationRequestHandler) Reject(c *gin.Context) {
	var input dto.RejectSchoolRegistrationRequestDTO
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&input); err != nil && err != io.EOF {
			HandleBindingError(c, err)
			return
		}
	}

	response, err := h.service.Reject(c.Param("id"), middleware.GetUserID(c), input)
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "School registration request rejected",
		"request": response,
	})
}

func (h *SchoolRegistrationRequestHandler) Approve(c *gin.Context) {
	var input dto.ApproveSchoolRegistrationRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil && err != io.EOF {
		HandleBindingError(c, err)
		return
	}

	response, err := h.service.Approve(c.Param("id"), middleware.GetUserID(c), input)
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func handleSchoolRegistrationRequestError(c *gin.Context, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "pending duplicate"):
		c.JSON(http.StatusConflict, gin.H{"error": "A pending registration request already exists for this school or contact email"})
	case strings.Contains(errStr, "duplicate school code"):
		c.JSON(http.StatusConflict, gin.H{"error": "School code already exists"})
	case strings.Contains(errStr, "is not pending"):
		c.JSON(http.StatusConflict, gin.H{"error": "School registration request is not pending"})
	case strings.Contains(errStr, "school registration"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
	default:
		HandleError(c, err)
	}
}
