package dto

type CreateSchoolDTO struct {
	Name    string  `json:"schoolName" binding:"required"`
	Code    string  `json:"schoolCode"`
	LogoID  *string `json:"schoolLogo,omitempty"`
	Address string  `json:"schoolAddress" binding:"required"`
	Email   string  `json:"schoolEmail" binding:"required,email"`
	Phone   string  `json:"schoolPhone" binding:"required,numeric,min=10"`
	Website *string `json:"schoolWebsite,omitempty" binding:"omitempty,url"`
}

type UpdateSchoolDTO struct {
	Name    *string `json:"schoolName"`
	Code    *string `json:"schoolCode"`
	LogoID  *string `json:"schoolLogo,omitempty"`
	Address *string `json:"schoolAddress"`
	Email   *string `json:"schoolEmail" binding:"omitempty,email"`
	Phone   *string `json:"schoolPhone" binding:"omitempty,numeric,min=10"`
	Website *string `json:"schoolWebsite,omitempty" binding:"omitempty,url"`
}

type SchoolResponseDTO struct {
	ID        string  `json:"schoolId"`
	Name      string  `json:"schoolName"`
	Code      string  `json:"schoolCode"`
	LogoID    *string `json:"schoolLogo,omitempty"`
	Address   string  `json:"schoolAddress"`
	Email     string  `json:"schoolEmail"`
	Phone     string  `json:"schoolPhone"`
	Website   *string `json:"schoolWebsite,omitempty"`
	IsDeleted bool    `json:"isDeleted"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type SchoolHeaderDTO struct {
	ID     string  `json:"schoolId"`
	Name   string  `json:"schoolName"`
	Code   string  `json:"schoolCode"`
	LogoID *string `json:"schoolLogo,omitempty"`
}

type SchoolSummaryDTO struct {
	TotalActive  int64 `json:"totalActive"`
	TotalDeleted int64 `json:"totalDeleted"`
	TotalSchools int64 `json:"totalSchools"`
}
