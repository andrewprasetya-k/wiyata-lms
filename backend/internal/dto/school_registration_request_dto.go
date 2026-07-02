package dto

type CreateSchoolRegistrationRequestDTO struct {
	SchoolName string  `json:"schoolName" binding:"required"`
	NPSN       *string `json:"npsn"`
	PICName    string  `json:"picName" binding:"required"`
	PICEmail   string  `json:"picEmail" binding:"required,email"`
	PICPhone   *string `json:"picPhone"`
	PICRole    *string `json:"picRole"`
	Message    *string `json:"message"`
}

type SchoolRegistrationRequestSummaryDTO struct {
	RequestID  string `json:"requestId"`
	SchoolName string `json:"schoolName"`
	PICName    string `json:"picName"`
	PICEmail   string `json:"picEmail"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type CreateSchoolRegistrationRequestResponseDTO struct {
	Message string                              `json:"message"`
	Request SchoolRegistrationRequestSummaryDTO `json:"request"`
}
