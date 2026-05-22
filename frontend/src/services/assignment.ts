import { api } from './api'
import type {
  AssignmentItem,
  AssignmentListResponse,
  MySubmissionResponse,
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

export async function submitAssignment(assignmentId: string, payload: SubmitAssignmentPayload) {
  const { data } = await api.post<SubmitAssignmentResponse>(
    `/assignments/submit/${assignmentId}`,
    payload,
  )
  return data
}

export async function getMySubmissionByAssignment(assignmentId: string) {
  const { data } = await api.get<MySubmissionResponse>(
    `/assignments/my-submission/${assignmentId}`,
  )
  return data
}
