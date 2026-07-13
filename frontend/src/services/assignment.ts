import { api } from './api'
import type {
  AssignmentItem,
  AssignmentListResponse,
  MySubmissionResponse,
  StudentAssignmentInboxResponse,
  SubmitAssignmentPayload,
  SubmitAssignmentResponse,
} from '../types/assignment'

export async function getSubjectAssignments(subjectClassId: string, page = 1, limit = 20) {
  const { data } = await api.get<AssignmentListResponse>(
    `/assignments/subject-class/${subjectClassId}`,
    {
      params: { page, limit },
    }
  )
  // Backend returns "data": null (not []) for an empty collection — see
  // internal/handler/assignment_handler.go's `var assignments []dto...`.
  // Normalize here so every caller can safely treat this as an array.
  data.data.data = data.data.data || []
  return data
}

export async function getSubjectAssignmentDetail(subjectClassId: string, assignmentId: string) {
  const response = await getSubjectAssignments(subjectClassId, 1, 100)
  const assignment =
    (response.data.data as AssignmentItem[]).find((item) => item.assignmentId === assignmentId) ??
    null

  return {
    subjectClass: response.subjectClass,
    assignment,
  }
}

export async function getStudentAssignmentDetail(assignmentId: string) {
  const { data } = await api.get<AssignmentItem>(
    `/assignments/student/${assignmentId}`,
  )
  return data
}

export async function submitAssignment(assignmentId: string, payload: SubmitAssignmentPayload) {
  const { data } = await api.post<SubmitAssignmentResponse>(
    `/assignments/submit/${assignmentId}`,
    payload,
  )
  return data
}

export async function deleteSubmission(submissionId: string) {
  const { data } = await api.delete<SubmitAssignmentResponse>(
    `/assignments/submit/${submissionId}`,
  )
  return data
}

export async function getMySubmissionByAssignment(assignmentId: string) {
  const { data } = await api.get<MySubmissionResponse>(
    `/assignments/my-submission/${assignmentId}`,
  )
  return data
}

export async function getStudentAssignmentInbox() {
  const { data } = await api.get<StudentAssignmentInboxResponse>(
    '/assignments/student-assignments',
  )
  return data
}
