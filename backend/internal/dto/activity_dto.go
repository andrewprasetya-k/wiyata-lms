package dto

type AcademicActivityResponseDTO struct {
	Items []AcademicActivityItemDTO `json:"items"`
}

type AcademicActivityItemDTO struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Date        string                 `json:"date"`
	Time        string                 `json:"time"`
	Priority    string                 `json:"priority"`
	Subject     *ActivitySubjectDTO    `json:"subject,omitempty"`
	Class       *ActivityClassDTO      `json:"class,omitempty"`
	Link        string                 `json:"link"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ActivitySubjectDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Code  string `json:"code"`
	Color string `json:"color,omitempty"`
}

type ActivityClassDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
