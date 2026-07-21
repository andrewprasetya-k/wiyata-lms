package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	service service.InvitationService
}

func NewInvitationHandler(service service.InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) GetMetadata(c *gin.Context) {
	response, err := h.service.GetMetadata(c.Param("token"))
	if err != nil {
		handleInvitationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	var input dto.AcceptInvitationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	response, err := h.service.Accept(c.Param("token"), input)
	if err != nil {
		handleInvitationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InvitationHandler) AcceptAuthenticated(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.service.AcceptAuthenticated(c.Param("token"), userID)
	if err != nil {
		handleInvitationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func handleInvitationError(c *gin.Context, err error) {
	errStr := err.Error()
	switch {
	case errors.Is(err, repository.ErrInvitationInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation is invalid or expired"})
	case errors.Is(err, repository.ErrInvitationClassUnavailable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation class is no longer available"})
	case errors.Is(err, repository.ErrInvitationEmailMismatch):
		c.JSON(http.StatusForbidden, gin.H{"error": "Invitation email does not match the authenticated account"})
	case strings.Contains(errStr, "invitation"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
	default:
		HandleError(c, err)
	}
}
