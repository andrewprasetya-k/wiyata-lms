import { api } from './api'
import type {
  AcademicYearsBySchoolResponse,
  AssessmentWeightsResponse,
  CreateAcademicYearPayload,
  CreateAssignmentCategoryPayload,
  CreateSubjectPayload,
  CreateTermPayload,
  SaveAssessmentWeightsPayload,
  SchoolAssignmentCategoriesResponse,
  SchoolSubjectsResponse,
  TermItem,
} from '../types/adminAcademic'

export async function getAcademicYearsBySchool(schoolCode: string) {
  const { data } = await api.get<AcademicYearsBySchoolResponse>(`/academic-years/school/${schoolCode}`)
  return data
}

export async function createAcademicYear(payload: CreateAcademicYearPayload) {
  const { data } = await api.post('/academic-years', payload)
  return data
}

export async function activateAcademicYear(academicYearId: string) {
  const { data } = await api.patch(`/academic-years/activate/${academicYearId}`)
  return data
}

export async function deactivateAcademicYear(academicYearId: string) {
  const { data } = await api.patch(`/academic-years/deactivate/${academicYearId}`)
  return data
}

export async function getTermsByAcademicYear(academicYearId: string) {
  const { data } = await api.get<TermItem[]>(`/terms/academic-year/${academicYearId}`)
  return data
}

export async function createTerm(payload: CreateTermPayload) {
  const { data } = await api.post('/terms', payload)
  return data
}

export async function activateTerm(termId: string) {
  const { data } = await api.patch(`/terms/activate/${termId}`)
  return data
}

export async function deactivateTerm(termId: string) {
  const { data } = await api.patch(`/terms/deactivate/${termId}`)
  return data
}

export async function getSubjectsBySchool(schoolCode: string) {
  const { data } = await api.get<SchoolSubjectsResponse>(`/subjects/school/${schoolCode}`)
  return data
}

export async function createSubject(payload: CreateSubjectPayload) {
  const { data } = await api.post('/subjects', payload)
  return data
}

export async function getAssignmentCategoriesBySchool(schoolCode: string) {
  const { data } = await api.get<SchoolAssignmentCategoriesResponse>(
    `/assignments/categories/school/${schoolCode}`,
  )
  return data
}

export async function createAssignmentCategory(payload: CreateAssignmentCategoryPayload) {
  const { data } = await api.post('/assignments/categories', payload)
  return data
}

export async function getAssessmentWeightsBySubject(subjectId: string) {
  const { data } = await api.get<AssessmentWeightsResponse>(
    `/grades/weights/subject/${subjectId}`,
  )
  return data
}

export async function saveAssessmentWeights(payload: SaveAssessmentWeightsPayload) {
  const { data } = await api.post('/grades/weights', payload)
  return data
}
