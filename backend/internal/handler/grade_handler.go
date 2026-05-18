package handler

import (
	"backend/internal/dto"
	"backend/internal/service"
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

	if err := h.service.ConfigureWeights(&req); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Weights configured successfully"})
}

func (h *GradeHandler) GetWeightsBySubject(c *gin.Context) {
	subjectID := c.Param("subjectId")

	weights, err := h.service.GetWeightsBySubject(subjectID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, weights)
}

func (h *GradeHandler) GetStudentGrade(c *gin.Context) {
	userID := c.Param("userId")
	subjectID := c.Param("subjectId")

	grade, err := h.service.CalculateFinalGrade(userID, subjectID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, grade)
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
