package dto

type CreateMaterialDTO struct {
	SchoolID       string              `json:"schoolId" binding:"required,uuid"`
	SubjectClassID string              `json:"subjectClassId" binding:"required,uuid"`
	Title          string              `json:"materialTitle" binding:"required"`
	Description    string              `json:"materialDesc"`
	Type           string              `json:"materialType" binding:"required,oneof=video pdf ppt other"`
	CreatedBy      string              `json:"createdBy" binding:"required,uuid"`
	MediaIDs       []string            `json:"mediaIds"` // Existing media IDs
	Medias         []CreateMediaInline `json:"medias"`   // New media to create
}

type CreateMediaInline struct {
	Name         string `json:"name" binding:"required"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	FileURL      string `json:"fileUrl" binding:"required"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

type UpdateMaterialDTO struct {
	Title       *string  `json:"materialTitle"`
	Description *string  `json:"materialDesc"`
	Type        *string  `json:"materialType" binding:"omitempty,oneof=video pdf ppt other"`
	MediaIDs    []string `json:"mediaIds"`
}

type MaterialResponseDTO struct {
	ID             string             `json:"materialId"`
	SubjectClassID string             `json:"subjectClassId"`
	SubjectName    string             `json:"subjectName,omitempty"`
	Title          string             `json:"materialTitle"`
	Description    string             `json:"materialDesc"`
	Type           string             `json:"materialType"`
	CreatorName    string             `json:"creatorName,omitempty"`
	CreatedAt      string             `json:"createdAt"`
	Attachments    []MediaResponseDTO `json:"attachments,omitempty"`
}

type MaterialListWithSchoolDTO struct {
	School SchoolHeaderDTO   `json:"school,omitempty"`
	Data   PaginatedResponse `json:"data"`
}

type MaterialListWithSubjectDTO struct {
	SubjectClass SubjectClassHeaderDTO `json:"subjectClass"`
	Data         PaginatedResponse     `json:"data"`
}

type UpdateProgressDTO struct {
	UserID     string `json:"userId" binding:"required,uuid"`
	MaterialID string `json:"materialId" binding:"required,uuid"`
	Status     string `json:"status" binding:"required,oneof=not_started completed"`
}
