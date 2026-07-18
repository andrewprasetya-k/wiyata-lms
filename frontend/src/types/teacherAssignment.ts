import type { SchoolHeader } from './adminAcademic'
import type { SubjectClassHeader } from './assignment'

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
  categoryId: string
  schoolId: string
  categoryName: string
  createdAt: string
}

export interface SchoolCategoriesResponse {
  school: SchoolHeader
  categories: AssignmentCategory[]
}

export interface TeacherSubmissionAttachment {
  mediaId: string
  mediaName?: string
  fileUrl?: string
  mimeType?: string
  fileSize?: number
  thumbnailUrl?: string
}

export interface TeacherSubmissionAssessment {
  score: number
  feedback: string
  assessor?: string
  assessorName?: string
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

export interface TeacherAssignmentHeader {
  assignmentId: string
  assignmentTitle: string
  subjectName?: string
  categoryName?: string
  deadline?: string
}

export interface TeacherSubmissionGroup {
  assignment: TeacherAssignmentHeader
  submissionCount: number
  gradedCount: number
  pendingCount: number
  submissions: TeacherSubmission[]
}

export interface TeacherSubmissionSummary {
  assignmentCount: number
  submissionCount: number
  gradedCount: number
  pendingCount: number
  lateCount: number
}

export interface TeacherSubjectClassSubmissionsResponse {
  subjectClass: SubjectClassHeader
  assignments: TeacherSubmissionGroup[]
  summary: TeacherSubmissionSummary
}

export interface TeacherSubmissionInboxSummary {
  totalSubmissions: number
  pendingCount: number
  gradedCount: number
  lateCount: number
}

export interface TeacherSubmissionInboxItem {
  assignmentId: string
  subjectClassId: string
  assignmentTitle: string
  subjectName: string
  subjectCode?: string
  subjectColor?: string
  className: string
  classCode?: string
  deadline?: string | null
  submissionCount: number
  pendingCount: number
  gradedCount: number
  lateCount: number
}

export interface TeacherSubmissionInboxResponse {
  summary: TeacherSubmissionInboxSummary
  items: TeacherSubmissionInboxItem[]
}

export interface TeacherAssignmentInboxSummary {
  totalAssignments: number
  activeAssignments: number
  overdueAssignments: number
  pendingReviewCount: number
  totalSubmissions: number
}

export interface TeacherAssignmentInboxItem {
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
  submissionCount: number
  pendingCount: number
  gradedCount: number
  lateCount: number
}

export interface TeacherAssignmentInboxResponse {
  summary: TeacherAssignmentInboxSummary
  items: TeacherAssignmentInboxItem[]
}
