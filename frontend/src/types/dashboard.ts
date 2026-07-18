export interface NotificationItem {
  notificationId: string
  type: string
  title: string
  message: string
  link?: string
  isRead: boolean
  createdAt: string
}

export interface NotificationListResponse {
  data: NotificationItem[]
  unreadCount: number
  totalItems: number
  page: number
  limit: number
  totalPages: number
}

export interface UnreadCountResponse {
  unreadCount: number
}
