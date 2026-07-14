import { api } from './api'
import type { ClassGradeReportResponse, StudentGradeDetailResponse } from '../types/teacherGrades'

export async function getClassGradeReport(classId: string, subjectId: string) {
  const { data } = await api.get<ClassGradeReportResponse>(
    `/grades/class/${classId}/subject/${subjectId}`,
  )
  return data
}

export async function getStudentGradeDetail(classId: string, subjectId: string, studentId: string) {
  const { data } = await api.get<StudentGradeDetailResponse>(
    `/grades/class/${classId}/subject/${subjectId}/student/${studentId}`,
  )
  return data
}
