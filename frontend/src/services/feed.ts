import { api } from './api'
import type {
  ClassFeedResponse,
  CreateFeedCommentPayload,
  CreateFeedPayload,
  FeedComment,
} from '../types/feed'
import type { UnreadCountResponse } from '../types/dashboard'

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

export async function getFeedComments(feedId: string) {
  const { data } = await api.get<FeedComment[]>('/comments', {
    params: { type: 'feed', id: feedId },
  })
  return data ?? []
}

export async function createFeedComment(feedId: string, content: string) {
  const payload: CreateFeedCommentPayload = {
    sourceType: 'feed',
    sourceId: feedId,
    content,
  }
  const { data } = await api.post<{ message: string }>('/comments', payload)
  return data
}

export async function deleteFeedComment(commentId: string) {
  const { data } = await api.delete<{ message: string }>(`/comments/${commentId}`)
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
