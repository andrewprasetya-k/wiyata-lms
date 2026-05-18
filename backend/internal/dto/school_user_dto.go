package dto

type AddSchoolUserDTO struct {
	UserID   string `json:"userId" binding:"required,uuid"`
	SchoolID string `json:"schoolId" binding:"required,uuid"`
}

type SchoolUserResponseDTO struct {
	ID         string   `json:"schoolUserId"`
	UserID     string   `json:"userId"`
	FullName   string   `json:"fullName,omitempty"`
	Email      string   `json:"email,omitempty"`
	SchoolID   string   `json:"schoolId"`
	SchoolName string   `json:"schoolName,omitempty"`
	SchoolCode string   `json:"schoolCode,omitempty"`
	Roles      []string `json:"roles,omitempty"`
	CreatedAt  string   `json:"createdAt"`
}

type SchoolWithMembersDTO struct {
	School  SchoolHeaderDTO   `json:"school"`
	Members PaginatedResponse `json:"members"`
}
