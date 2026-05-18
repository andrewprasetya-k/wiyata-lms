package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
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
			CreatedAt: cat.CreatedAt.Format("02-01-2006 15:04:05"),
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

	asg := domain.Assignment{
		SchoolID:            input.SchoolID,
		SubjectClassID:      input.SubjectClassID,
		CategoryID:          input.CategoryID,
		Title:               input.Title,
		Description:         input.Description,
		Deadline:            input.Deadline,
		AllowLateSubmission: input.AllowLateSubmission,
		CreatedBy:           input.CreatedBy,
	}

	if err := h.service.CreateAssignment(&asg, input.MediaIDs); err != nil {
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

	if err := h.service.UpdateAssignment(id, existing, input.MediaIDs); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment updated"})
}

func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	id := c.Param("id")

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
			ID:          subjectClassHeader.ID,
			SubjectCode: subjectClassHeader.Subject.Code,
			SubjectName: subjectClassHeader.Subject.Name,
			TeacherID:   subjectClassHeader.Teacher.ID,
			TeacherName: subjectClassHeader.Teacher.User.FullName,
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

func (h *AssignmentHandler) GetSubmissionsByAssignment(c *gin.Context) {
	assignmentID := c.Param("assignmentId")
	asg, err := h.service.GetAssignmentWithSubmissions(assignmentID)
	if err != nil {
		HandleError(c, err)
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
				AssessedAt: s.Assessment.AssessedAt.Format("02-01-2006 15:04:05"),
			}
		}

		var atts []dto.MediaResponseDTO
		for _, a := range s.Attachments {
			atts = append(atts, dto.MediaResponseDTO{
				ID:       a.Media.ID,
				Name:     a.Media.Name,
				FileURL:  a.Media.FileURL,
				MimeType: a.Media.MimeType,
			})
		}

		submissionsDTO = append(submissionsDTO, dto.SubmissionResponseDTO{
			ID:          s.ID,
			UserName:    s.User.FullName,
			SubmittedAt: s.SubmittedAt.Format("02-01-2006 15:04:05"),
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

func (h *AssignmentHandler) Submit(c *gin.Context) {
	var input dto.CreateSubmissionDTO
	var assignmentId = c.Param("assignmentId")
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	sbm := domain.Submission{
		SchoolID:     input.SchoolID,
		AssignmentID: assignmentId,
		UserID:       input.UserID,
	}

	if err := h.service.Submit(&sbm, input.MediaIDs); err != nil {
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

	if err := h.service.UpdateSubmission(submissionId, input.MediaIDs); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission updated"})
}

func (h *AssignmentHandler) DeleteSubmission(c *gin.Context) {
	submissionId := c.Param("submissionId")

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

	var assessmentDTO *dto.AssessmentResponseDTO
	if submission.Assessment != nil {
		assessmentDTO = &dto.AssessmentResponseDTO{
			Score:      submission.Assessment.Score,
			Feedback:   submission.Assessment.Feedback,
			Assessor:   submission.Assessment.Assessor.FullName,
			AssessedAt: submission.Assessment.AssessedAt.Format("02-01-2006 15:04:05"),
		}
	}

	var atts []dto.MediaResponseDTO
	for _, a := range submission.Attachments {
		atts = append(atts, dto.MediaResponseDTO{
			ID:       a.Media.ID,
			Name:     a.Media.Name,
			FileURL:  a.Media.FileURL,
			MimeType: a.Media.MimeType,
		})
	}

	response := dto.SubmissionResponseDTO{
		ID:          submission.ID,
		UserName:    submission.User.FullName,
		SubmittedAt: submission.SubmittedAt.Format("02-01-2006 15:04:05"),
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

	asm := domain.Assessment{
		SubmissionID: submissionId,
		Score:        input.Score,
		Feedback:     input.Feedback,
		AssessedBy:   input.AssessedBy,
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

	if err := h.service.DeleteAssessment(submissionId); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assessment deleted"})
}

func (h *AssignmentHandler) mapAsgToResponse(a *domain.Assignment) dto.AssignmentResponseDTO {
	var atts []dto.MediaResponseDTO
	for _, att := range a.Attachments {
		atts = append(atts, dto.MediaResponseDTO{
			ID:       att.Media.ID,
			Name:     att.Media.Name,
			FileURL:  att.Media.FileURL,
			MimeType: att.Media.MimeType,
		})
	}

	return dto.AssignmentResponseDTO{
		ID:                  a.ID,
		Title:               a.Title,
		Description:         a.Description,
		CategoryName:        a.Category.Name,
		Deadline:            a.Deadline,
		AllowLateSubmission: a.AllowLateSubmission,
		CreatedAt:           a.CreatedAt.Format("02-01-2006 15:04:05"),
		Attachments:         atts,
	}
}
