import { api } from './api'
import type { ClassFeedResponse } from '../types/feed'

export async function getClassFeed(classId: string) {
  const { data } = await api.get<ClassFeedResponse>(`/feeds/class/${classId}`, {
    params: { page: 1, limit: 10 },
  })
  return data
}
