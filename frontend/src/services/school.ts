import { api } from './api'

export interface CreateSchoolPayload {
  schoolName: string
}

export interface CreateSchoolResultSchool {
  schoolId: string
  schoolName: string
  schoolCode: string
  schoolLogo?: string
  schoolAddress: string
  schoolEmail: string
  schoolPhone: string
  schoolWebsite?: string
  isDeleted: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateSchoolResult {
  school: CreateSchoolResultSchool
  schoolUserId: string
  role: string
}

export async function createSchool(payload: CreateSchoolPayload) {
  const { data } = await api.post<CreateSchoolResult>('/schools', payload)
  return data
}
