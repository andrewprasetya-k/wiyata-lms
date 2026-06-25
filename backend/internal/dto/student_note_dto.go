package dto

import "time"

type SaveStudentNoteDTO struct {
	Content string `json:"content" binding:"required,max=10000"`
}

type StudentNoteResponseDTO struct {
	ID         string    `json:"noteId"`
	MaterialID string    `json:"materialId"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type StudentNoteEnvelopeDTO struct {
	Note *StudentNoteResponseDTO `json:"note"`
}

type StudentNoteCollectionItemDTO struct {
	ID            string    `json:"noteId"`
	MaterialID    string    `json:"materialId"`
	MaterialTitle string    `json:"materialTitle"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type StudentNoteCollectionDTO struct {
	Notes []StudentNoteCollectionItemDTO `json:"notes"`
}
