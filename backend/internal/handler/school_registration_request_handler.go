package handler

import (
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
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
	var input dto.CreateSchoolRegistrationRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, err := h.service.Create(input)
	if err != nil {
		handleSchoolRegistrationRequestError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func handleSchoolRegistrationRequestError(c *gin.Context, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "pending duplicate"):
		c.JSON(http.StatusConflict, gin.H{"error": "A pending registration request already exists for this school or contact email"})
	case strings.Contains(errStr, "school registration"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
	default:
		HandleError(c, err)
	}
}
