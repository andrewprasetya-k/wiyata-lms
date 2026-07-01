package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AssignmentHandler struct {
	service             service.AssignmentService
	schoolService       service.SchoolService
	subjectClassService service.SubjectClassService
}

func NewAssignmentHandler(service service.AssignmentService, schoolService service.SchoolService, subjectClassService service.SubjectClassService) *AssignmentHandler {
	return &AssignmentHandler{
		service:             service,
		schoolService:       schoolService,
		subjectClassService: subjectClassService,
	}
}

func (h *AssignmentHandler) CreateCategory(c *gin.Context) {
	var input dto.CreateAssignmentCategoryDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}
	if !h.validateRequestSchool(c, input.SchoolID) {
		return
	}

	cat := domain.AssignmentCategory{
		SchoolID: input.SchoolID,
		Name:     input.Name,
	}

	if err := h.service.CreateCategory(&cat); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created"})
}

func (h *AssignmentHandler) GetCategoriesBySchool(c *gin.Context) {
	schoolCode := c.Param("schoolCode")

	// 1. Get School Header
	school, err := h.schoolService.GetSchoolByCode(schoolCode)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Get Categories
	cats, err := h.service.GetCategoriesBySchool(school.ID)
	if err != nil {
		HandleError(c, err)
		return
	}

	var response []dto.AssignmentCategoryResponseDTO
	for _, cat := range cats {
		response = append(response, dto.AssignmentCategoryResponseDTO{
			ID:        cat.ID,
			SchoolID:  cat.SchoolID,
			Name:      cat.Name,
			CreatedAt: formatAPITime(cat.CreatedAt),
		})
	}

	c.JSON(http.StatusOK, dto.SchoolWithAssignmentCategoriesDTO{
		School: dto.SchoolHeaderDTO{
			ID:     school.ID,
			Name:   school.Name,
			Code:   school.Code,
			LogoID: school.LogoID,
		},
		Categories: response,
	})
}

func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	var input dto.CreateAssignmentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !h.validateRequestSchool(c, input.SchoolID) {
		return
	}
	if !h.authorizeTeacherForSubjectClass(c, input.SubjectClassID) {
		return
	}

	asg := domain.Assignment{
		SchoolID:            input.SchoolID,
		SubjectClassID:      input.SubjectClassID,
		CategoryID:          input.CategoryID,
		Title:               input.Title,
		Description:         input.Description,
		Deadline:            input.Deadline,
		AllowLateSubmission: input.AllowLateSubmission,
		CreatedBy:           userID,
	}

	if err := h.service.CreateAssignment(&asg, input.MediaIDs, userID, h.hasActiveRole(c, "admin")); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Assignment created"})
}

func (h *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateAssignmentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	// Get existing assignment
	existing, err := h.service.GetAssignmentByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeAssignmentMutation(c, existing) {
		return
	}

	// Update fields if provided
	if input.CategoryID != nil {
		existing.CategoryID = *input.CategoryID
	}
	if input.Title != nil {
		existing.Title = *input.Title
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Deadline != nil {
		existing.Deadline = input.Deadline
	}
	if input.AllowLateSubmission != nil {
		existing.AllowLateSubmission = *input.AllowLateSubmission
	}

	if err := h.service.UpdateAssignment(id, existing, input.MediaIDs, middleware.GetUserID(c), h.hasActiveRole(c, "admin"), input.CategoryID != nil); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment updated"})
}

func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.service.GetAssignmentByID(id)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeAssignmentMutation(c, existing) {
		return
	}

	if err := h.service.DeleteAssignment(id); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted"})
}

func (h *AssignmentHandler) GetBySubjectClass(c *gin.Context) {
	subjectClassID := c.Param("subjectClassId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	if !h.authorizeUserForSubjectClassAccess(c, subjectClassID) {
		return
	}

	// 1. Get SubjectClass Header
	subjectClassHeader, err := h.subjectClassService.GetByID(subjectClassID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 2. Get Assignments
	results, total, err := h.service.GetAssignmentsBySubjectClass(subjectClassID, search, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	var assignments []dto.AssignmentResponseDTO
	for _, r := range results {
		assignments = append(assignments, h.mapAsgToResponse(r))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.AssignmentPerSubjectClassResponseDTO{
		SubjectClass: dto.SubjectClassHeaderDTO{
			ID:           subjectClassHeader.ID,
			SubjectCode:  subjectClassHeader.Subject.Code,
			SubjectName:  subjectClassHeader.Subject.Name,
			SubjectColor: subjectClassHeader.Subject.Color,
			TeacherID:    subjectClassHeader.Teacher.ID,
			TeacherName:  subjectClassHeader.Teacher.User.FullName,
		},
		Data: dto.PaginatedResponse{
			Data:       assignments,
			TotalItems: total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(totalPages),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetStudentAssignmentDetail(c *gin.Context) {
	assignmentID := c.Param("assignmentId")
	assignment, err := h.service.GetAssignmentByID(assignmentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}
	if assignment.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	if !h.authorizeStudentForSubjectClass(c, assignment.SubjectClassID) {
		return
	}

	attachments := make([]dto.MediaResponseDTO, 0, len(assignment.Attachments))
	for _, attachment := range assignment.Attachments {
		if media, ok := mapAttachmentMedia(attachment, assignment.SchoolID); ok {
			attachments = append(attachments, media)
		}
	}

	c.JSON(http.StatusOK, dto.StudentAssignmentDetailDTO{
		ID:                  assignment.ID,
		SubjectClassID:      assignment.SubjectClassID,
		SubjectName:         assignment.SubjectClass.Subject.Name,
		SubjectCode:         assignment.SubjectClass.Subject.Code,
		SubjectColor:        assignment.SubjectClass.Subject.Color,
		Title:               assignment.Title,
		Description:         assignment.Description,
		CategoryName:        assignment.Category.Name,
		Deadline:            assignment.Deadline,
		AllowLateSubmission: assignment.AllowLateSubmission,
		CreatedAt:           formatAPITime(assignment.CreatedAt),
		UpdatedAt:           formatAPITime(assignment.UpdatedAt),
		Attachments:         attachments,
	})
}

func (h *AssignmentHandler) GetSubjectClassSubmissions(c *gin.Context) {
	subjectClassID := c.Param("subjectClassId")
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID := ""
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok {
			schoolID = value
		}
	}
	if schoolID == "" {
		schoolID = c.GetHeader("SchoolId")
	}
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	ownsSubjectClass, err := h.subjectClassService.TeacherOwnsSubjectClass(userID, schoolID, subjectClassID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !ownsSubjectClass {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: teacher does not teach this subject class"})
		return
	}

	subjectClassHeader, err := h.subjectClassService.GetByID(subjectClassID)
	if err != nil {
		HandleError(c, err)
		return
	}

	assignments, err := h.service.GetSubjectClassSubmissions(subjectClassID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := dto.SubjectClassSubmissionsResponseDTO{
		SubjectClass: dto.SubjectClassHeaderDTO{
			ID:           subjectClassHeader.ID,
			SubjectCode:  subjectClassHeader.Subject.Code,
			SubjectName:  subjectClassHeader.Subject.Name,
			SubjectColor: subjectClassHeader.Subject.Color,
			TeacherID:    subjectClassHeader.Teacher.ID,
			TeacherName:  subjectClassHeader.Teacher.User.FullName,
		},
		Assignments: make([]dto.AssignmentSubmissionGroupDTO, 0, len(assignments)),
	}

	for _, asg := range assignments {
		group := dto.AssignmentSubmissionGroupDTO{
			Assignment: dto.AssignmentHeaderDTO{
				ID:           asg.ID,
				Title:        asg.Title,
				SubjectName:  asg.SubjectClass.Subject.Name,
				CategoryName: asg.Category.Name,
				Deadline:     asg.Deadline,
			},
			Submissions: make([]dto.SubmissionResponseDTO, 0, len(asg.Submissions)),
		}

		response.Summary.AssignmentCount++

		for _, submission := range asg.Submissions {
			var assessmentDTO *dto.AssessmentResponseDTO
			if submission.Assessment != nil {
				assessmentDTO = &dto.AssessmentResponseDTO{
					Score:      submission.Assessment.Score,
					Feedback:   submission.Assessment.Feedback,
					Assessor:   submission.Assessment.Assessor.FullName,
					AssessedAt: formatAPITime(submission.Assessment.AssessedAt),
				}
				group.GradedCount++
				response.Summary.GradedCount++
			} else {
				group.PendingCount++
				response.Summary.PendingCount++
			}

			isLate := asg.Deadline != nil && submission.SubmittedAt.After(*asg.Deadline)
			if isLate {
				response.Summary.LateCount++
			}

			atts := make([]dto.MediaResponseDTO, 0, len(submission.Attachments))
			for _, a := range submission.Attachments {
				if attachment, ok := mapAttachmentMedia(a, submission.SchoolID); ok {
					atts = append(atts, attachment)
				}
			}

			group.Submissions = append(group.Submissions, dto.SubmissionResponseDTO{
				ID:          submission.ID,
				UserName:    submission.User.FullName,
				SubmittedAt: formatAPITime(submission.SubmittedAt),
				IsLate:      isLate,
				Attachments: atts,
				Assessment:  assessmentDTO,
			})
		}

		group.SubmissionCount = len(group.Submissions)
		response.Summary.SubmissionCount += group.SubmissionCount
		response.Assignments = append(response.Assignments, group)
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetTeacherSubmissionInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	response, err := h.service.GetTeacherSubmissionInbox(userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetTeacherAssignmentInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	response, err := h.service.GetTeacherAssignmentInbox(userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetStudentAssignmentInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	response, err := h.service.GetStudentAssignmentInbox(userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetSubmissionsByAssignment(c *gin.Context) {
	assignmentID := c.Param("assignmentId")
	asg, err := h.service.GetAssignmentWithSubmissions(assignmentID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeTeacherForSubjectClass(c, asg.SubjectClassID) {
		return
	}

	var submissionsDTO []dto.SubmissionResponseDTO
	for _, s := range asg.Submissions {
		var assessmentDTO *dto.AssessmentResponseDTO
		if s.Assessment != nil {
			assessmentDTO = &dto.AssessmentResponseDTO{
				Score:      s.Assessment.Score,
				Feedback:   s.Assessment.Feedback,
				Assessor:   s.Assessment.Assessor.FullName,
				AssessedAt: formatAPITime(s.Assessment.AssessedAt),
			}
		}

		atts := make([]dto.MediaResponseDTO, 0, len(s.Attachments))
		for _, a := range s.Attachments {
			if attachment, ok := mapAttachmentMedia(a, s.SchoolID); ok {
				atts = append(atts, attachment)
			}
		}

		submissionsDTO = append(submissionsDTO, dto.SubmissionResponseDTO{
			ID:          s.ID,
			UserName:    s.User.FullName,
			SubmittedAt: formatAPITime(s.SubmittedAt),
			IsLate:      asg.Deadline != nil && s.SubmittedAt.After(*asg.Deadline),
			Attachments: atts,
			Assessment:  assessmentDTO,
		})
	}

	response := dto.AssignmentWithSubmissionsDTO{
		Assignment: dto.AssignmentHeaderDTO{
			ID:           asg.ID,
			Title:        asg.Title,
			SubjectName:  asg.SubjectClass.Subject.Name,
			CategoryName: asg.Category.Name,
			Deadline:     asg.Deadline,
		},
		Submissions: submissionsDTO,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) GetAssignmentStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.service.GetAssignmentStatus(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *AssignmentHandler) GetMySubmissionByAssignment(c *gin.Context) {
	assignmentID := c.Param("assignmentId")
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	schoolID := ""
	if sid, exists := c.Get("school_id"); exists {
		if value, ok := sid.(string); ok {
			schoolID = value
		}
	}
	if schoolID == "" {
		schoolID = c.GetHeader("SchoolId")
	}
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	assignment, err := h.service.GetAssignmentByID(assignmentID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if assignment.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	if !h.authorizeStudentForSubjectClass(c, assignment.SubjectClassID) {
		return
	}

	submission, err := h.service.GetMySubmissionByAssignment(assignmentID, userID, schoolID)
	if err != nil {
		HandleError(c, err)
		return
	}

	if submission == nil {
		c.JSON(http.StatusOK, dto.MySubmissionResponseDTO{
			Status:     "not_submitted",
			Submission: nil,
		})
		return
	}

	status := "submitted"
	if submission.Assessment != nil {
		status = "graded"
	}

	c.JSON(http.StatusOK, dto.MySubmissionResponseDTO{
		Status:     status,
		Submission: h.mapMySubmissionToResponse(submission),
	})
}

func (h *AssignmentHandler) Submit(c *gin.Context) {
	var input dto.CreateSubmissionDTO
	var assignmentId = c.Param("assignmentId")
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}
	if !h.validateRequestSchool(c, input.SchoolID) {
		return
	}

	assignment, err := h.service.GetAssignmentByID(assignmentId)
	if err != nil {
		HandleError(c, err)
		return
	}
	if assignment.SchoolID != schoolID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: assignment does not belong to active school"})
		return
	}
	if !h.authorizeStudentForSubjectClass(c, assignment.SubjectClassID) {
		return
	}

	sbm := domain.Submission{
		SchoolID:     schoolID,
		AssignmentID: assignmentId,
		UserID:       userID,
	}

	if err := h.service.Submit(&sbm, input.MediaIDs, userID, h.hasActiveRole(c, "admin")); err != nil {
		if err.Error() == "submission past due" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot submit past deadline"})
			return
		}
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Submission received"})
}

func (h *AssignmentHandler) UpdateSubmission(c *gin.Context) {
	var input dto.CreateSubmissionDTO
	submissionId := c.Param("submissionId")
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}
	if !h.validateRequestSchool(c, input.SchoolID) {
		return
	}
	if !h.authorizeStudentForSubmission(c, submissionId) {
		return
	}
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.UpdateSubmission(submissionId, input.MediaIDs, userID, h.hasActiveRole(c, "admin")); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission updated"})
}

func (h *AssignmentHandler) DeleteSubmission(c *gin.Context) {
	submissionId := c.Param("submissionId")
	if !h.authorizeStudentForSubmission(c, submissionId) {
		return
	}

	if err := h.service.DeleteSubmission(submissionId); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission deleted"})
}

func (h *AssignmentHandler) GetSubmissionByID(c *gin.Context) {
	submissionId := c.Param("submissionId")

	submission, err := h.service.GetSubmissionByID(submissionId)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Get assignment for deadline check
	assignment, err := h.service.GetAssignmentByID(submission.AssignmentID)
	if err != nil {
		HandleError(c, err)
		return
	}
	if !h.authorizeTeacherForSubjectClass(c, assignment.SubjectClassID) {
		return
	}

	var assessmentDTO *dto.AssessmentResponseDTO
	if submission.Assessment != nil {
		assessmentDTO = &dto.AssessmentResponseDTO{
			Score:      submission.Assessment.Score,
			Feedback:   submission.Assessment.Feedback,
			Assessor:   submission.Assessment.Assessor.FullName,
			AssessedAt: formatAPITime(submission.Assessment.AssessedAt),
		}
	}

	atts := make([]dto.MediaResponseDTO, 0, len(submission.Attachments))
	for _, a := range submission.Attachments {
		if attachment, ok := mapAttachmentMedia(a, submission.SchoolID); ok {
			atts = append(atts, attachment)
		}
	}

	response := dto.SubmissionResponseDTO{
		ID:          submission.ID,
		UserName:    submission.User.FullName,
		SubmittedAt: formatAPITime(submission.SubmittedAt),
		IsLate:      assignment.Deadline != nil && submission.SubmittedAt.After(*assignment.Deadline),
		Attachments: atts,
		Assessment:  assessmentDTO,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AssignmentHandler) Assess(c *gin.Context) {
	var input dto.CreateAssessmentDTO
	var submissionId = c.Param("submissionId")
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if !h.authorizeTeacherForSubmission(c, submissionId) {
		return
	}

	asm := domain.Assessment{
		SubmissionID: submissionId,
		Score:        input.Score,
		Feedback:     input.Feedback,
		AssessedBy:   userID,
	}

	if err := h.service.Assess(&asm); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assessment recorded"})
}

func (h *AssignmentHandler) UpdateAssessment(c *gin.Context) {
	submissionId := c.Param("submissionId")
	var input dto.UpdateAssessmentDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}
	if !h.authorizeTeacherForSubmission(c, submissionId) {
		return
	}

	asm := &domain.Assessment{
		SubmissionID: submissionId,
	}

	if input.Score != nil {
		asm.Score = *input.Score
	}
	if input.Feedback != nil {
		asm.Feedback = *input.Feedback
	}

	if err := h.service.UpdateAssessment(submissionId, asm); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assessment updated"})
}

func (h *AssignmentHandler) DeleteAssessment(c *gin.Context) {
	submissionId := c.Param("submissionId")
	if !h.authorizeTeacherForSubmission(c, submissionId) {
		return
	}

	if err := h.service.DeleteAssessment(submissionId); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assessment deleted"})
}

func (h *AssignmentHandler) mapAsgToResponse(a *domain.Assignment) dto.AssignmentResponseDTO {
	atts := make([]dto.MediaResponseDTO, 0, len(a.Attachments))
	for _, att := range a.Attachments {
		if attachment, ok := mapAttachmentMedia(att, a.SchoolID); ok {
			atts = append(atts, attachment)
		}
	}

	return dto.AssignmentResponseDTO{
		ID:                  a.ID,
		Title:               a.Title,
		Description:         a.Description,
		CategoryName:        a.Category.Name,
		Deadline:            a.Deadline,
		AllowLateSubmission: a.AllowLateSubmission,
		CreatedAt:           formatAPITime(a.CreatedAt),
		Attachments:         atts,
	}
}

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

func (h *AssignmentHandler) mapMySubmissionToResponse(s *domain.Submission) *dto.MySubmissionDTO {
	atts := make([]dto.MediaResponseDTO, 0, len(s.Attachments))
	for _, a := range s.Attachments {
		if attachment, ok := mapAttachmentMedia(a, s.SchoolID); ok {
			atts = append(atts, attachment)
		}
	}

	var assessment *dto.MySubmissionAssessmentDTO
	if s.Assessment != nil {
		assessment = &dto.MySubmissionAssessmentDTO{
			ID:           s.Assessment.ID,
			Score:        s.Assessment.Score,
			Feedback:     s.Assessment.Feedback,
			AssessedAt:   formatAPITime(s.Assessment.AssessedAt),
			AssessorName: s.Assessment.Assessor.FullName,
		}
	}

	return &dto.MySubmissionDTO{
		ID:           s.ID,
		AssignmentID: s.AssignmentID,
		SubmittedAt:  formatAPITime(s.SubmittedAt),
		Attachments:  atts,
		Assessment:   assessment,
	}
}
