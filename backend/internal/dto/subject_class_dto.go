package dto

type CreateSubjectClassDTO struct {
	ClassID      string `json:"classId" binding:"required,uuid"`
	SubjectID    string `json:"subjectId" binding:"required,uuid"`
	SchoolUserID string `json:"teacherId" binding:"required,uuid"`
}

type UpdateSubjectClassDTO struct {
	SubjectID    *string `json:"subjectId" binding:"omitempty,uuid"`
	SchoolUserID *string `json:"teacherId" binding:"omitempty,uuid"`
}

type SubjectClassResponseDTO struct {
	ID          string `json:"subjectClassId"`
	SubjectID   string `json:"subjectId"`
	SubjectName string `json:"subjectName,omitempty"`
	SubjectCode string `json:"subjectCode,omitempty"`
	TeacherID   string `json:"teacherId"`
	TeacherName string `json:"teacherName,omitempty"`
}

type SubjectClassHeaderDTO struct {
	ID          string `json:"subjectClassId"`
	SubjectCode string `json:"subjectCode"`
	SubjectName string `json:"subjectName,omitempty"`
	TeacherID   string `json:"teacherId"`
	TeacherName string `json:"teacherName,omitempty"`
}

type SubjectPerClassDTO struct {
	Class    ClassHeaderDTO            `json:"class"`
	Subjects []SubjectClassResponseDTO `json:"subjects"`
}
