package dto

type AdminSchoolMemberImportRowDTO struct {
	RowNumber int      `json:"rowNumber"`
	FullName  string   `json:"fullName"`
	Email     string   `json:"email"`
	Role      string   `json:"role"`
	ClassCode string   `json:"classCode,omitempty"`
	Status    string   `json:"status"`
	Errors    []string `json:"errors"`
}

type AdminSchoolMemberImportPreviewResponseDTO struct {
	Rows         []AdminSchoolMemberImportRowDTO `json:"rows"`
	ValidCount   int                             `json:"validCount"`
	InvalidCount int                             `json:"invalidCount"`
}

type AdminSchoolMemberImportCommitRequestDTO struct {
	DefaultPassword string                          `json:"defaultPassword" binding:"required,min=6"`
	Rows            []AdminSchoolMemberImportRowDTO `json:"rows" binding:"required"`
}

type AdminSchoolMemberImportResultDTO struct {
	RowNumber         int    `json:"rowNumber"`
	FullName          string `json:"fullName"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	ClassCode         string `json:"classCode,omitempty"`
	Status            string `json:"status"`
	Reason            string `json:"reason,omitempty"`
	UserCreated       bool   `json:"userCreated"`
	MembershipAction  string `json:"membershipAction,omitempty"`
	EmailNotification string `json:"emailNotification,omitempty"`
}

type AdminSchoolMemberImportCommitResponseDTO struct {
	ImportedCount int                                `json:"importedCount"`
	SkippedCount  int                                `json:"skippedCount"`
	FailedCount   int                                `json:"failedCount"`
	Results       []AdminSchoolMemberImportResultDTO `json:"results"`
}

type AdminSchoolMemberCreateDTO struct {
	FullName  string `json:"fullName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      string `json:"role" binding:"required"`
	ClassCode string `json:"classCode,omitempty"`
}

type AdminSchoolMemberResponseDTO struct {
	SchoolUserID      string   `json:"schoolUserId"`
	UserID            string   `json:"userId"`
	FullName          string   `json:"fullName"`
	Email             string   `json:"email"`
	Roles             []string `json:"roles"`
	ClassCodes        []string `json:"classCodes,omitempty"`
	CreatedAt         string   `json:"createdAt"`
	DeletedAt         *string  `json:"deletedAt,omitempty"`
	UserCreated       bool     `json:"userCreated"`
	MembershipAction  string   `json:"membershipAction,omitempty"`
	EmailNotification string   `json:"emailNotification,omitempty"`
}

type AdminSchoolMemberListResponseDTO struct {
	Data       []AdminSchoolMemberResponseDTO `json:"data"`
	TotalItems int64                          `json:"totalItems"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
	TotalPages int                            `json:"totalPages"`
}
