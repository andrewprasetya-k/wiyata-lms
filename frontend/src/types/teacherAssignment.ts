export interface CreateAssignmentPayload {
  schoolId: string
  subjectClassId: string
  categoryId: string
  assignmentTitle: string
  assignmentDescription: string
  deadline?: string
  allowLateSubmission: boolean
  mediaIds: string[]
}

export interface AssignmentCategory {
  asc_id: string
  asc_name: string
}

export interface SchoolCategoriesResponse {
  schoolId: string
  schoolName: string
  data: AssignmentCategory[]
}

export interface TeacherSubmissionAttachment {
  mediaId: string
  mediaName: string
  fileUrl?: string
  mimeType?: string
  fileSize?: number
}

export interface TeacherSubmissionAssessment {
  score: number
  feedback: string
  assessor?: string
  assessedAt?: string
}

export interface TeacherSubmission {
  submissionId: string
  studentName: string
  submittedAt: string
  isLate: boolean
  attachments?: TeacherSubmissionAttachment[]
  assessment?: TeacherSubmissionAssessment
}

export interface AssignmentWithSubmissionsResponse {
  assignment: {
    assignmentId: string
    assignmentTitle: string
    subjectName?: string
    categoryName?: string
    deadline?: string
  }
  submissions: TeacherSubmission[]
}
