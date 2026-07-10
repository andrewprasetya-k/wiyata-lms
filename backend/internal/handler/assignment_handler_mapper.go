package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
)

// Response-mapping helpers for AssignmentHandler. Split out of
// assignment_handler.go to keep the HTTP handler methods themselves shorter
// and easier to scan; behavior is unchanged.

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
