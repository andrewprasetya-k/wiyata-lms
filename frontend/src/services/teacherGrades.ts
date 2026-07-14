import { api } from './api'
import type {
  ClassGradeReportResponse,
  StudentGradeDetailResponse,
  StudentReportResponse,
} from '../types/teacherGrades'

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

export async function getStudentReport(classId: string, studentId: string) {
  const { data } = await api.get<StudentReportResponse>(
    `/grades/class/${classId}/student/${studentId}/report`,
  )
  return data
}
