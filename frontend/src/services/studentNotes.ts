import { api } from './api'
import type {
  SaveStudentMaterialNotePayload,
  StudentMaterialNoteResponse,
} from '../types/studentNotes'

export async function getStudentMaterialNote(materialId: string) {
  const { data } = await api.get<StudentMaterialNoteResponse>(
    `/notes/material/${materialId}`,
  )
  return data
}

export async function saveStudentMaterialNote(
  materialId: string,
  payload: SaveStudentMaterialNotePayload,
) {
  const { data } = await api.put<StudentMaterialNoteResponse>(
    `/notes/material/${materialId}`,
    payload,
  )
  return data
}

export async function deleteStudentMaterialNote(materialId: string) {
  const { data } = await api.delete<{ message: string }>(
    `/notes/material/${materialId}`,
  )
  return data
}
