package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GradeHandler struct {
	service service.GradeService
}

func NewGradeHandler(service service.GradeService) *GradeHandler {
	return &GradeHandler{service: service}
}

func (h *GradeHandler) ConfigureWeights(c *gin.Context) {
	var req dto.ConfigureWeightsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleBindingError(c, err)
		return
	}
	schoolID, ok := getGradeActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
		return
	}

	if err := h.service.ConfigureWeights(&req, schoolID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Weights configured successfully"})
}

func (h *GradeHandler) GetWeightsBySubject(c *gin.Context) {
	subjectID := c.Param("subjectId")
	schoolID, ok := getGradeActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
		return
	}

	weights, err := h.service.GetWeightsBySubject(subjectID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, weights)
}

func getGradeActiveSchoolID(c *gin.Context) (string, bool) {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok && value != "" {
			return value, true
		}
	}
	if value := c.GetHeader("SchoolId"); value != "" {
		return value, true
	}
	return "", false
}

func (h *GradeHandler) GetClassGradeReport(c *gin.Context) {
	classID := c.Param("classId")
	subjectID := c.Param("subjectId")

	report, err := h.service.GetClassGradeReport(classID, subjectID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *GradeHandler) GetMyGradebookByClass(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID, ok := c.Get("school_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
		return
	}

	schoolIDString, ok := schoolID.(string)
	if !ok || schoolIDString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header or schoolCode param)"})
		return
	}

	classID := c.Param("classId")
	report, err := h.service.GetMyGradebookByClass(userID, schoolIDString, classID)
	if err != nil {
		if errors.Is(err, service.ErrStudentNotEnrolledInClass) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: student is not enrolled in this class"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}
