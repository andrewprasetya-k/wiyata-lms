package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubjectClassHandler struct {
	service      service.SubjectClassService
	classService service.ClassService
}

func NewSubjectClassHandler(service service.SubjectClassService, classService service.ClassService) *SubjectClassHandler {
	return &SubjectClassHandler{service: service, classService: classService}
}

func (h *SubjectClassHandler) Assign(c *gin.Context) {
	var input dto.CreateSubjectClassDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	scl := domain.SubjectClass{
		ClassID:      input.ClassID,
		SubjectID:    input.SubjectID,
		SchoolUserID: input.SchoolUserID,
	}

	if err := h.service.Assign(&scl); err != nil {
		if err.Error() == "this subject is already assigned to the class with the same teacher" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Subject and teacher assigned to class successfully"})
}

func (h *SubjectClassHandler) GetByClass(c *gin.Context) {
	classID := c.Param("classId")
	//get subject classes
	results, err := h.service.GetByClass(classID)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.SubjectClassResponseDTO
	for _, r := range results {
		response = append(response, h.mapToResponse(r))
	}

	//get class info
	classInfo, err := h.classService.GetByID(classID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response2 := dto.SubjectPerClassDTO{
		Class: dto.ClassHeaderDTO{
			ID:    classInfo.ID,
			Title: classInfo.Title,
			Code:  classInfo.Code,
		},
		Subjects: response,
	}

	c.JSON(http.StatusOK, response2)
}

func (h *SubjectClassHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.mapToResponse(result))
}

func (h *SubjectClassHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateSubjectClassDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	scl, err := h.service.GetByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if input.SubjectID != nil {
		scl.SubjectID = *input.SubjectID
	}
	if input.SchoolUserID != nil {
		scl.SchoolUserID = *input.SchoolUserID
	}

	if err := h.service.Update(scl); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment updated successfully"})
}

func (h *SubjectClassHandler) Unassign(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Unassign(id); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assignment removed successfully"})
}

func (h *SubjectClassHandler) mapToResponse(scl *domain.SubjectClass) dto.SubjectClassResponseDTO {
	return dto.SubjectClassResponseDTO{
		ID:          scl.ID,
		SubjectID:   scl.SubjectID,
		SubjectName: scl.Subject.Name,
		SubjectCode: scl.Subject.Code,
		TeacherID:   scl.SchoolUserID,
		TeacherName: scl.Teacher.User.FullName,
	}
}
