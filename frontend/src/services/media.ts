import { api } from './api'
import type { MediaUploadResponse } from '../types/media'

export async function uploadMediaFile(file: File, schoolId: string, ownerType = 'submission') {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('schoolId', schoolId)
  formData.append('ownerType', ownerType)

  const { data } = await api.post<MediaUploadResponse>('/medias/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

  return data
}

export async function deleteMedia(mediaId: string) {
  await api.delete(`/medias/${mediaId}`)
}
