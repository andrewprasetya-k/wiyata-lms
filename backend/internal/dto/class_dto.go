package dto

type CreateClassDTO struct {
	SchoolID    string `json:"schoolId" binding:"required,uuid"`
	TermID      string `json:"termId" binding:"required,uuid"`
	Code        string `json:"classCode" binding:"required"`
	Title       string `json:"classTitle" binding:"required"`
	Description string `json:"classDesc"`
	CreatedBy   string `json:"createdBy" binding:"required,uuid"`
}

type UpdateClassDTO struct {
	Title       *string `json:"classTitle"`
	Description *string `json:"classDesc"`
	IsActive    *bool   `json:"isActive"`
}

type ClassResponseDTO struct {
	ID               string `json:"classId"`
	SchoolID         string `json:"schoolId"`
	SchoolName       string `json:"schoolName,omitempty"`
	TermID           string `json:"termId"`
	TermName         string `json:"termName,omitempty"`
	AcademicYearName string `json:"academicYearName,omitempty"`
	Code             string `json:"classCode"`
	Title            string `json:"classTitle"`
	Description      string `json:"classDesc"`
	CreatedBy        string `json:"createdBy"`
	CreatorName      string `json:"creatorName,omitempty"`
	IsActive         bool   `json:"isActive"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type ClassListWithSchoolDTO struct {
	School SchoolHeaderDTO   `json:"school,omitempty"`
	Data   PaginatedResponse `json:"data"`
}

type ClassHeaderDTO struct {
	ID    string `json:"classId"`
	Title string `json:"classTitle"`
	Code  string `json:"classCode"`
}
