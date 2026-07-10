import { api } from './api'
import type {
  ClassFeedResponse,
  CreateFeedPayload,
  FeedPost,
} from '../types/feed'
import type { UnreadCountResponse } from '../types/dashboard'

export async function getClassFeed(classId: string) {
  const { data } = await api.get<ClassFeedResponse>(`/feeds/class/${classId}`, {
    params: { page: 1, limit: 10 },
  })
  return data
}

export async function createClassFeed(payload: CreateFeedPayload) {
  const { data } = await api.post<{ message: string; feed?: FeedPost }>('/feeds', payload)
  return data
}

export async function getFeedUnreadCount() {
  const { data } = await api.get<UnreadCountResponse>('/feeds/unread-count')
  return data
}

export async function markFeedNotificationsRead() {
  const { data } = await api.patch<{ message: string }>('/feeds/read')
  return data
}
