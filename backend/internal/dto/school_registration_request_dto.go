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

type SchoolRegistrationRequestDetailDTO struct {
	RequestID       string  `json:"requestId"`
	RequesterUserID string  `json:"requesterUserId,omitempty"`
	SchoolName      string  `json:"schoolName"`
	NPSN            *string `json:"npsn,omitempty"`
	PICName         string  `json:"picName"`
	PICEmail        string  `json:"picEmail"`
	PICPhone        *string `json:"picPhone,omitempty"`
	PICRole         *string `json:"picRole,omitempty"`
	Message         *string `json:"message,omitempty"`
	Status          string  `json:"status"`
	ReviewedBy      *string `json:"reviewedBy,omitempty"`
	ReviewedAt      *string `json:"reviewedAt,omitempty"`
	ReviewNote      *string `json:"reviewNote,omitempty"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type CreateSchoolRegistrationRequestResponseDTO struct {
	Message string                              `json:"message"`
	Request SchoolRegistrationRequestSummaryDTO `json:"request"`
}

type RejectSchoolRegistrationRequestDTO struct {
	Reason *string `json:"reason"`
}

type ApproveSchoolRegistrationRequestDTO struct {
	SchoolCode string  `json:"schoolCode"`
	SchoolName *string `json:"schoolName"`
	Note       *string `json:"note"`
}

type SchoolRegistrationRequestListResponseDTO struct {
	Data       []SchoolRegistrationRequestDetailDTO `json:"data"`
	TotalItems int64                                `json:"totalItems"`
	Page       int                                  `json:"page"`
	Limit      int                                  `json:"limit"`
	TotalPages int                                  `json:"totalPages"`
}

type ApprovedSchoolDTO struct {
	SchoolID   string `json:"schoolId"`
	SchoolCode string `json:"schoolCode"`
	SchoolName string `json:"schoolName"`
}

type ApprovedAdminDTO struct {
	UserID       string `json:"userId"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	SchoolUserID string `json:"schoolUserId"`
	Role         string `json:"role"`
}

type ApproveSchoolRegistrationRequestResponseDTO struct {
	Message string                             `json:"message"`
	Request SchoolRegistrationRequestDetailDTO `json:"request"`
	School  ApprovedSchoolDTO                  `json:"school"`
	Admin   ApprovedAdminDTO                   `json:"admin"`
}
