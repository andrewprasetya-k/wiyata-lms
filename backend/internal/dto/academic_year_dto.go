package dto

type CreateAcademicYearDTO struct {
	SchoolID string `json:"schoolId" binding:"required,uuid"`
	Name     string `json:"academicYearName" binding:"required"`
}

type UpdateAcademicYearDTO struct {
	Name *string `json:"academicYearName"`
}

type AcademicYearResponseDTO struct {
	ID         string `json:"academicYearId"`
	SchoolID   string `json:"schoolId"`
	SchoolName string `json:"schoolName,omitempty"`
	SchoolCode string `json:"schoolCode,omitempty"`
	Name       string `json:"academicYearName"`
	IsActive   bool   `json:"isActive"`
	CreatedAt  string `json:"createdAt"`
}

type AcademicYearWithSchoolDTO struct {
	School SchoolHeaderDTO           `json:"school"`
	Data   []AcademicYearResponseDTO `json:"data"`
}
