import { api } from './api'
import type { 
  AssignmentWithSubmissionsResponse,
  CreateAssignmentPayload, 
  SchoolCategoriesResponse,
  TeacherSubmissionInboxResponse,
  TeacherSubjectClassSubmissionsResponse,
} from '../types/teacherAssignment'

export async function createAssignment(payload: CreateAssignmentPayload) {
  const { data } = await api.post('/assignments', payload)
  return data
}

export async function getAssignmentCategories(schoolCode: string) {
  const { data } = await api.get<SchoolCategoriesResponse>(`/assignments/categories/school/${schoolCode}`)
  return data
}

export async function getAssignmentDetailWithSubmissions(assignmentId: string) {
  const { data } = await api.get<AssignmentWithSubmissionsResponse>(`/assignments/${assignmentId}`)
  return data
}

export async function getSubjectClassSubmissions(subjectClassId: string) {
  const { data } = await api.get<TeacherSubjectClassSubmissionsResponse>(
    `/assignments/subject-class/submissions/${subjectClassId}`,
  )
  return data
}

export async function getTeacherSubmissionInbox() {
  const { data } = await api.get<TeacherSubmissionInboxResponse>('/assignments/teacher-submissions')
  return data
}

export async function assessSubmission(submissionId: string, payload: { score: number; feedback: string }) {
  const { data } = await api.post(`/assignments/assess/${submissionId}`, payload)
  return data
}
