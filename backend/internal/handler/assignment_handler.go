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

	response := []dto.AssignmentCategoryResponseDTO{}
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
	if limit <= 0 || limit > 100 {
		limit = 100
	}
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

	assignments := []dto.AssignmentResponseDTO{}
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

	submissionsDTO := []dto.SubmissionResponseDTO{}
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
	schoolID := h.getSchoolContext(c)
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School context required (SchoolId header)"})
		return
	}

	status, err := h.service.GetAssignmentStatus(id, schoolID)
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
		if !h.handleSubmissionMutationError(c, err) {
			HandleError(c, err)
		}
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
		if !h.handleSubmissionMutationError(c, err) {
			HandleError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission deleted"})
}

// handleSubmissionMutationError maps the shared submission-mutation business
// errors (see assignmentService.validateSubmissionMutable) to 400 responses,
// consistent with the "submission past due" handling in Submit. Returns true
// if the error was recognized and a response was written.
func (h *AssignmentHandler) handleSubmissionMutationError(c *gin.Context, err error) bool {
	switch err.Error() {
	case "submission already graded":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot modify a graded submission"})
		return true
	case "assignment closed":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot modify submission after the assignment is closed"})
		return true
	default:
		return false
	}
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

	actor := buildActorContext(c, domain.LogScopeSchool)
	if err := h.service.Assess(actor, &asm); err != nil {
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

	actor := buildActorContext(c, domain.LogScopeSchool)
	if err := h.service.UpdateAssessment(actor, submissionId, asm); err != nil {
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

	actor := buildActorContext(c, domain.LogScopeSchool)
	if err := h.service.DeleteAssessment(actor, submissionId); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assessment deleted"})
}
