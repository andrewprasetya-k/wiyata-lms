import type { MediaAttachment, PaginatedResponse } from './classWorkspace'

export interface SubjectClassHeader {
  subjectClassId: string
  subjectCode: string
  subjectName?: string
  subjectColor?: string
  teacherId: string
  teacherName?: string
}

export interface AssignmentItem {
  assignmentId: string
  subjectClassId?: string
  subjectName?: string
  subjectCode?: string
  subjectColor?: string
  assignmentTitle: string
  assignmentDescription?: string
  deadline?: string | null
  categoryName?: string
  createdAt?: string
  updatedAt?: string
  allowLateSubmission?: boolean
  attachments?: MediaAttachment[]
}

export interface AssignmentListResponse {
  subjectClass: SubjectClassHeader
  data: PaginatedResponse<AssignmentItem>
}

export interface SubmitAssignmentPayload {
  schoolId: string
  mediaIds: string[]
}

export interface SubmitAssignmentResponse {
  message: string
}

export type MySubmissionStatus = 'not_submitted' | 'submitted' | 'graded'

export interface SubmissionAttachment extends MediaAttachment {}

export interface SubmissionAssessment {
  assessmentId: string
  score: number
  feedback: string
  assessedAt: string
  assessorName: string
}

export interface MySubmission {
  submissionId: string
  assignmentId: string
  submittedAt: string
  attachments?: SubmissionAttachment[]
  assessment: SubmissionAssessment | null
}

export interface MySubmissionResponse {
  status: MySubmissionStatus
  submission: MySubmission | null
}

export interface StudentAssignmentInboxSummary {
  totalAssignments: number
  notSubmittedCount: number
  submittedCount: number
  gradedCount: number
  overdueCount: number
}

export interface StudentAssignmentInboxItem {
  assignmentId: string
  subjectClassId: string
  assignmentTitle: string
  subjectName: string
  subjectCode?: string
  subjectColor?: string
  className: string
  classCode?: string
  categoryName?: string
  deadline?: string | null
  submissionId?: string | null
  submittedAt?: string | null
  score?: number | null
  isSubmitted: boolean
  isGraded: boolean
  isOverdue: boolean
  isSubmittedLate: boolean
}

export interface StudentAssignmentInboxResponse {
  summary: StudentAssignmentInboxSummary
  items: StudentAssignmentInboxItem[]
}
