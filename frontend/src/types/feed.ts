import type { MediaAttachment, PaginatedResponse } from './classWorkspace'

export interface FeedClassHeader {
  classId: string
  classTitle: string
  classCode: string
}

export interface FeedPost {
  feedId: string
  content: string
  creatorName?: string
  createdAt: string
  attachments?: MediaAttachment[]
  commentCount?: number
}

export interface ClassFeedResponse {
  class: FeedClassHeader
  data: PaginatedResponse<FeedPost>
}

export interface CreateFeedPayload {
  schoolId: string
  classId: string
  content: string
}
