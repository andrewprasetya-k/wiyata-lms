import { api } from './api'
import type {
  NotificationListResponse,
  UnreadCountResponse,
} from '../types/dashboard'

export interface GetNotificationsParams {
  page?: number
  limit?: number
  unreadOnly?: boolean
}

export async function getNotifications(params: GetNotificationsParams = {}) {
  const { data } = await api.get<NotificationListResponse>('/notifications', {
    params,
  })
  return data
}

export async function getNotificationUnreadCount() {
  const { data } = await api.get<UnreadCountResponse>(
    '/notifications/unread-count',
  )
  return data
}

export async function markNotificationAsRead(notificationId: string) {
  const { data } = await api.patch<{ message: string }>(
    `/notifications/read/${notificationId}`,
  )
  return data
}

export async function markAllNotificationsAsRead() {
  const { data } = await api.patch<{ message: string }>(
    '/notifications/read-all',
  )
  return data
}
