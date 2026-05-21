import type { PaginatedResponse } from './classWorkspace'

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
}

export interface AssignmentListResponse {
  subjectClass: SubjectClassHeader
  data: PaginatedResponse<AssignmentItem>
}
