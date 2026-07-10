package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authorization and request-context helpers for AssignmentHandler. Split out
// of assignment_handler.go to keep the HTTP handler methods themselves
// shorter and easier to scan; behavior is unchanged.

func (h *AssignmentHandler) getSchoolContext(c *gin.Context) string {
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok {
			return value
		}
	}
	return c.GetHeader("SchoolId")
}

func (h *AssignmentHandler) getActiveRoles(c *gin.Context) []string {
	if raw, exists := c.Get("user_roles"); exists {
		if roles, ok := raw.([]string); ok {
			return roles
		}
	}
	return nil
}

func (h *AssignmentHandler) hasActiveRole(c *gin.Context, role string) bool {
	for _, activeRole := range h.getActiveRoles(c) {
		if activeRole == role {
			return true
		}
	}
	return false
}

func (h *AssignmentHandler) validateRequestSchool(c *gin.Context, requestSchoolID string) bool {
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}
	if requestSchoolID != "" && requestSchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: schoolId does not match active school"})
		return false
	}
	return true
}

func (h *AssignmentHandler) authorizeUserForSubjectClassAccess(c *gin.Context, subjectClassID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	allowed, err := h.subjectClassService.UserCanAccessSubjectClass(userID, schoolID, subjectClassID, h.getActiveRoles(c))
	if err != nil {
		HandleError(c, err)
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you cannot access this subject class"})
		return false
	}
	return true
}

func (h *AssignmentHandler) authorizeStudentForSubjectClass(c *gin.Context, subjectClassID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	allowed, err := h.subjectClassService.UserCanAccessSubjectClass(userID, schoolID, subjectClassID, []string{"student"})
	if err != nil {
		HandleError(c, err)
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: student is not enrolled in this class"})
		return false
	}
	return true
}

func (h *AssignmentHandler) authorizeAssignmentMutation(c *gin.Context, assignment *domain.Assignment) bool {
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}
	if assignment.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: assignment does not belong to active school"})
		return false
	}

	if h.hasActiveRole(c, "admin") {
		return true
	}

	if h.hasActiveRole(c, "teacher") {
		return h.authorizeTeacherForSubjectClass(c, assignment.SubjectClassID)
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
	return false
}

func (h *AssignmentHandler) authorizeStudentForSubmission(c *gin.Context, submissionID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	submission, err := h.service.GetSubmissionByID(submissionID)
	if err != nil {
		HandleError(c, err)
		return false
	}
	if submission.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: submission does not belong to active school"})
		return false
	}
	if submission.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: submission belongs to another user"})
		return false
	}

	assignment, err := h.service.GetAssignmentByID(submission.AssignmentID)
	if err != nil {
		HandleError(c, err)
		return false
	}
	if assignment.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: assignment does not belong to active school"})
		return false
	}
	return h.authorizeStudentForSubjectClass(c, assignment.SubjectClassID)
}

func (h *AssignmentHandler) authorizeTeacherForSubmission(c *gin.Context, submissionID string) bool {
	submission, err := h.service.GetSubmissionByID(submissionID)
	if err != nil {
		HandleError(c, err)
		return false
	}

	assignment, err := h.service.GetAssignmentByID(submission.AssignmentID)
	if err != nil {
		HandleError(c, err)
		return false
	}

	return h.authorizeTeacherForSubjectClass(c, assignment.SubjectClassID)
}

func (h *AssignmentHandler) authorizeTeacherForSubjectClass(c *gin.Context, subjectClassID string) bool {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return false
	}

	ownsSubjectClass, err := h.subjectClassService.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
	if err != nil {
		HandleError(c, err)
		return false
	}
	if !ownsSubjectClass {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: teacher does not teach this subject class"})
		return false
	}
	return true
}
