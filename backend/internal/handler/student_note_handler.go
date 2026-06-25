package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudentNoteHandler struct {
	service service.StudentNoteService
}

func NewStudentNoteHandler(service service.StudentNoteService) *StudentNoteHandler {
	return &StudentNoteHandler{service: service}
}

func (h *StudentNoteHandler) GetMaterialNote(c *gin.Context) {
	materialID := c.Param("materialId")
	schoolID, userID, ok := getStudentNoteContext(c)
	if !ok {
		return
	}

	note, err := h.service.GetMaterialNote(materialID, schoolID, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.StudentNoteEnvelopeDTO{Note: mapStudentNoteResponse(note)})
}

func (h *StudentNoteHandler) GetSubjectClassNotes(c *gin.Context) {
	subjectClassID := c.Param("subjectClassId")
	schoolID, userID, ok := getStudentNoteContext(c)
	if !ok {
		return
	}

	notes, err := h.service.GetSubjectClassNotes(subjectClassID, schoolID, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := make([]dto.StudentNoteCollectionItemDTO, 0, len(notes))
	for _, note := range notes {
		response = append(response, dto.StudentNoteCollectionItemDTO{
			ID:            note.ID,
			MaterialID:    note.MaterialID,
			MaterialTitle: note.MaterialTitle,
			Content:       note.Content,
			CreatedAt:     note.CreatedAt,
			UpdatedAt:     note.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.StudentNoteCollectionDTO{Notes: response})
}

func (h *StudentNoteHandler) SaveMaterialNote(c *gin.Context) {
	materialID := c.Param("materialId")
	schoolID, userID, ok := getStudentNoteContext(c)
	if !ok {
		return
	}

	var input dto.SaveStudentNoteDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	note, err := h.service.SaveMaterialNote(materialID, schoolID, userID, input.Content)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.StudentNoteEnvelopeDTO{Note: mapStudentNoteResponse(note)})
}

func (h *StudentNoteHandler) DeleteMaterialNote(c *gin.Context) {
	materialID := c.Param("materialId")
	schoolID, userID, ok := getStudentNoteContext(c)
	if !ok {
		return
	}

	if err := h.service.DeleteMaterialNote(materialID, schoolID, userID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

func getStudentNoteContext(c *gin.Context) (string, string, bool) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", "", false
	}

	if rawSchoolID, exists := c.Get("school_id"); exists {
		if schoolID, ok := rawSchoolID.(string); ok && schoolID != "" {
			return schoolID, userID, true
		}
	}
	if schoolID := c.GetHeader("SchoolId"); schoolID != "" {
		return schoolID, userID, true
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "School context required"})
	return "", "", false
}

func mapStudentNoteResponse(note *domain.StudentNote) *dto.StudentNoteResponseDTO {
	if note == nil {
		return nil
	}
	return &dto.StudentNoteResponseDTO{
		ID:         note.ID,
		MaterialID: note.MaterialID,
		Content:    note.Content,
		CreatedAt:  note.CreatedAt,
		UpdatedAt:  note.UpdatedAt,
	}
}
