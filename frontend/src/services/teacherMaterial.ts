import { api } from './api'
import type { 
  CreateMaterialPayload, 
  MaterialListResponse 
} from '../types/teacherMaterial'

export async function createMaterial(payload: CreateMaterialPayload) {
  const { data } = await api.post('/materials', payload)
  return data
}

export async function updateMaterial(id: string, payload: Partial<CreateMaterialPayload>) {
  const { data } = await api.patch(`/materials/${id}`, payload)
  return data
}

export async function deleteMaterial(id: string) {
  const { data } = await api.delete(`/materials/${id}`)
  return data
}

export async function getSubjectMaterials(subjectClassId: string) {
  const { data } = await api.get<MaterialListResponse>(`/materials`, {
    params: { subjectClassId, page: 1, limit: 50 },
  })
  return data
}
