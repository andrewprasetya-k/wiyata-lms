package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
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
	schoolID, ok := getSubjectClassActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	scl := domain.SubjectClass{
		ClassID:      input.ClassID,
		SubjectID:    input.SubjectID,
		SchoolUserID: input.SchoolUserID,
	}

	if err := h.service.AssignInSchool(&scl, schoolID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Subject and teacher assigned to class successfully"})
}

func (h *SubjectClassHandler) GetByClass(c *gin.Context) {
	classID := c.Param("classId")
	schoolID, ok := getSubjectClassActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}
	//get subject classes
	results, err := h.service.GetByClassInSchool(classID, schoolID)
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

func (h *SubjectClassHandler) GetMyTeaching(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID, _ := c.Get("school_id")
	schoolIDString, ok := schoolID.(string)
	if !ok || schoolIDString == "" {
		schoolIDString = c.GetHeader("SchoolId")
	}
	if schoolIDString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	results, err := h.service.GetTeachingByUserAndSchool(userID, schoolIDString)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := make([]dto.TeacherSubjectClassDTO, 0, len(results))
	for _, item := range results {
		response = append(response, dto.TeacherSubjectClassDTO{
			SubjectClassID:     item.SubjectClassID,
			ClassID:            item.ClassID,
			ClassName:          item.ClassName,
			ClassCode:          item.ClassCode,
			SubjectID:          item.SubjectID,
			SubjectName:        item.SubjectName,
			SubjectCode:        item.SubjectCode,
			StudentCount:       item.StudentCount,
			MaterialCount:      item.MaterialCount,
			AssignmentCount:    item.AssignmentCount,
			PendingSubmissions: item.PendingSubmissions,
		})
	}

	c.JSON(http.StatusOK, dto.TeacherSubjectClassesResponseDTO{Data: response})
}

func (h *SubjectClassHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getSubjectClassActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	result, err := h.service.GetByIDInSchool(id, schoolID)
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
	schoolID, ok := getSubjectClassActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	scl, err := h.service.GetByIDInSchool(id, schoolID)
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

	if err := h.service.UpdateInSchool(scl, schoolID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment updated successfully"})
}

func (h *SubjectClassHandler) Unassign(c *gin.Context) {
	id := c.Param("id")
	schoolID, ok := getSubjectClassActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
		return
	}

	if err := h.service.UnassignInSchool(id, schoolID); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil unassign teacher"})
}

func getSubjectClassActiveSchoolID(c *gin.Context) (string, bool) {
	value, exists := c.Get("school_id")
	if !exists {
		return "", false
	}
	schoolID, ok := value.(string)
	return schoolID, ok && schoolID != ""
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
