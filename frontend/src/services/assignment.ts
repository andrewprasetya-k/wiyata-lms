import { api } from './api'
import type { AssignmentListResponse } from '../types/assignment'

export async function getSubjectAssignments(subjectClassId: string, page = 1, limit = 20) {
  const { data } = await api.get<AssignmentListResponse>(
    `/assignments/subject-class/${subjectClassId}`,
    {
      params: { page, limit },
    }
  )
  return data
}
