export interface ChatLastMessage {
  messageId: string
  senderId: string
  senderName: string
  content: string
  messageType?: string
  attachmentCount?: number
  attachmentMimeType?: string
  createdAt: string
}

export interface ChatRoom {
  roomId: string
  roomName: string
  roomType: string
  roomRefType?: string | null
  roomRefId?: string | null
  schoolId: string
  schoolName?: string
  dmTargetUserId?: string | null
  dmTargetName?: string | null
  dmTargetEmail?: string | null
  lastMessage?: ChatLastMessage | null
  lastMessageAt?: string | null
  unreadCount: number
  canSend: boolean
}

export interface ChatMessage {
  messageId: string
  roomId: string
  senderId: string
  senderName: string
  senderRole: string
  content: string
  messageType: 'text' | string
  attachments: ChatAttachment[]
  createdAt: string
  isMine: boolean
  deliveryStatus?: 'pending' | 'sent' | 'failed'
  pendingFiles?: File[]
}

export interface ChatAttachment {
  attachmentId: string
  mediaId: string
  fileName: string
  mimeType: string
  sizeBytes: number
  url: string
  localUrl?: string
}

export interface ChatRoomsResponse {
  rooms: ChatRoom[]
}

export interface ChatRoomResponse {
  room: ChatRoom
}

export interface ChatMember {
  userId: string
  fullName: string
  email: string
  roles: string[]
}

export interface ChatGroupMember {
  userId: string
  fullName: string
  email: string
  role: 'admin' | 'member' | string
  joinedAt: string
  leftAt?: string | null
}

export interface ChatGroupInfo {
  roomId: string
  roomName: string
  roomType: string
  schoolId: string
  schoolName: string
  creator?: ChatMember | null
  admins: ChatGroupMember[]
  members: ChatGroupMember[]
  createdAt: string
  memberCount: number
}

export interface ChatMembersResponse {
  members: ChatMember[]
}

export interface ChatGroupInfoResponse {
  group: ChatGroupInfo
}

export interface ChatReadMember {
  userId: string
  fullName: string
  email: string
  lastReadMessageId?: string | null
  lastReadAt?: string | null
}

export interface ChatReadSummary {
  roomId: string
  lastReadMessageId?: string | null
  lastReadAt?: string | null
  members: ChatReadMember[]
}

export interface ChatMessagesResponse {
  messages: ChatMessage[]
  nextBefore?: string | null
  hasMore: boolean
}

export interface ChatMessagePayload {
  content?: string
  mediaIds?: string[]
}

export interface CreateChatGroupPayload {
  roomName: string
  memberUserIds: string[]
}

export interface OpenDirectMessagePayload {
  targetUserId: string
}

export interface UpdateChatGroupPayload {
  roomName: string
}

export interface AddChatGroupMembersPayload {
  memberUserIds: string[]
}

export interface MarkRoomReadPayload {
  lastReadMessageId?: string
}

export interface ChatReadEventPayload {
  roomId: string
  userId: string
  lastReadMessageId?: string | null
  lastReadAt: string
}

export interface RoomUpdatedEventPayload {
  reason: 'new_message' | 'message_read' | string
}

export interface BaseChatSocketEvent<TType extends string, TPayload> {
  type: TType
  roomId: string
  schoolId: string
  payload: TPayload
}

export type NewMessageEvent = BaseChatSocketEvent<'new_message', ChatMessage>

export type MessageReadEvent = BaseChatSocketEvent<'message_read', ChatReadEventPayload>

export type RoomUpdatedEvent = BaseChatSocketEvent<'room_updated', RoomUpdatedEventPayload>

export type UnknownChatSocketEvent = BaseChatSocketEvent<string, unknown>

export type ChatSocketEvent =
  | NewMessageEvent
  | MessageReadEvent
  | RoomUpdatedEvent
  | UnknownChatSocketEvent
