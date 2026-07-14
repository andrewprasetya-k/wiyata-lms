import { api } from './api'
import type { ClassGradeReportResponse } from '../types/teacherGrades'

export async function getClassGradeReport(classId: string, subjectId: string) {
  const { data } = await api.get<ClassGradeReportResponse>(
    `/grades/class/${classId}/subject/${subjectId}`,
  )
  return data
}
