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
	service             service.GradeService
	subjectClassService service.SubjectClassService
}

func NewGradeHandler(service service.GradeService, subjectClassService service.SubjectClassService) *GradeHandler {
	return &GradeHandler{service: service, subjectClassService: subjectClassService}
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
	schoolID, ok := getGradeActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	report, err := h.service.GetClassGradeReport(classID, subjectID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *GradeHandler) GetStudentGradeDetail(c *gin.Context) {
	classID := c.Param("classId")
	subjectID := c.Param("subjectId")
	studentID := c.Param("studentId")
	if classID == "" || subjectID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "classId, subjectId, and studentId are required"})
		return
	}

	schoolID, ok := getGradeActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	if !h.authorizeStudentGradeDetailAccess(c, schoolID, classID, subjectID, studentID) {
		return
	}

	detail, err := h.service.GetStudentGradeDetail(classID, subjectID, studentID, schoolID)
	if err != nil {
		if errors.Is(err, service.ErrStudentNotEnrolledInClass) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: student is not enrolled in this class"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, detail)
}

// authorizeStudentGradeDetailAccess allows: admin (school-scoped); teacher,
// only if they actually teach this class+subject (TeacherOwnsClassSubject,
// same ownership pattern used by assignment/material endpoints); student,
// only for their own studentID. This is intentionally stricter than
// GetClassGradeReport (which only checks the school-level "teacher" role) —
// see grade_handler.go's GetStudentGradeDetail for rationale.
func (h *GradeHandler) authorizeStudentGradeDetailAccess(c *gin.Context, schoolID, classID, subjectID, studentID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	roles := getGradeActiveRoles(c)
	for _, role := range roles {
		switch role {
		case "admin":
			return true
		case "student":
			if userID == studentID {
				return true
			}
		case "teacher":
			owns, err := h.subjectClassService.TeacherOwnsClassSubject(userID, schoolID, classID, subjectID)
			if err != nil {
				HandleError(c, err)
				return false
			}
			if owns {
				return true
			}
		}
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you cannot access this student's grade detail"})
	return false
}

func getGradeActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}

func (h *GradeHandler) GetStudentReport(c *gin.Context) {
	classID := c.Param("classId")
	studentID := c.Param("studentId")
	if classID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "classId and studentId are required"})
		return
	}

	schoolID, ok := getGradeActiveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	if !h.authorizeStudentReportAccess(c, schoolID, classID, studentID) {
		return
	}

	report, err := h.service.GetStudentReport(classID, studentID, schoolID)
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

// authorizeStudentReportAccess mirrors authorizeStudentGradeDetailAccess's
// pattern (admin / teacher-ownership / student-self), but the report spans
// every subject in the class rather than one, so "teacher" here means "owns
// at least one subject_class in this class" rather than one specific
// subject. That's checked by fetching the class's subject_classes once
// (independent of studentID) and running TeacherOwnsClassSubject per subject
// until one matches — bounded by the class's subject count, not the
// student's.
func (h *GradeHandler) authorizeStudentReportAccess(c *gin.Context, schoolID, classID, studentID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	roles := getGradeActiveRoles(c)
	for _, role := range roles {
		switch role {
		case "admin":
			return true
		case "student":
			if userID == studentID {
				return true
			}
		case "teacher":
			subjectClasses, err := h.subjectClassService.GetByClassInSchool(classID, schoolID)
			if err != nil {
				HandleError(c, err)
				return false
			}
			for _, sc := range subjectClasses {
				owns, err := h.subjectClassService.TeacherOwnsClassSubject(userID, schoolID, classID, sc.SubjectID)
				if err != nil {
					HandleError(c, err)
					return false
				}
				if owns {
					return true
				}
			}
		}
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you cannot access this student's report"})
	return false
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
