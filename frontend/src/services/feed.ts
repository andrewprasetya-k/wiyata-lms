import { api } from './api'
import type { ClassFeedResponse, CreateFeedPayload } from '../types/feed'

export async function getClassFeed(classId: string) {
  const { data } = await api.get<ClassFeedResponse>(`/feeds/class/${classId}`, {
    params: { page: 1, limit: 10 },
  })
  return data
}

export async function createClassFeed(payload: CreateFeedPayload) {
  const { data } = await api.post<{ message: string }>('/feeds', payload)
  return data
}
