import type { MediaAttachment, PaginatedResponse } from './classWorkspace'

export interface SubjectClassHeader {
  subjectClassId: string
  subjectCode: string
  subjectName?: string
  teacherId: string
  teacherName?: string
}

export interface AssignmentItem {
  assignmentId: string
  assignmentTitle: string
  assignmentDescription?: string
  deadline: string
  categoryName?: string
  createdAt?: string
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

export interface SubmissionAttachment {
  mediaId: string
  mediaName: string
  fileUrl: string
  mimeType: string
  fileSize: number
}

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
