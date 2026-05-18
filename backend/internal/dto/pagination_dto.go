package dto

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalItems int64       `json:"totalItems"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"totalPages"`
}
