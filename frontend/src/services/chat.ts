import { api } from './api'
import type {
  ChatMembersResponse,
  ChatMessage,
  ChatMessagePayload,
  ChatMessagesResponse,
  ChatRoomResponse,
  ChatRoomsResponse,
  AddChatGroupMembersPayload,
  CreateChatGroupPayload,
  ChatGroupInfoResponse,
  MarkRoomReadPayload,
  UpdateChatGroupPayload,
} from '../types/chat'

export async function openSchoolChatRoom() {
  const { data } = await api.post<ChatRoomResponse>('/chat/school/open')
  return data.room
}

export async function getChatRooms(search = '') {
  const { data } = await api.get<ChatRoomsResponse>('/chat/rooms', {
    params: {
      search: search || undefined,
    },
  })
  return data.rooms ?? []
}

export async function searchChatMembers(search = '', excludeRoomId?: string | null) {
  const { data } = await api.get<ChatMembersResponse>('/chat/members', {
    params: {
      search: search || undefined,
      excludeRoomId: excludeRoomId || undefined,
    },
  })
  return data.members ?? []
}

export async function createChatGroup(payload: CreateChatGroupPayload) {
  const { data } = await api.post<ChatRoomResponse>('/chat/groups', payload)
  return data.room
}

export async function getChatGroupInfo(roomId: string) {
  const { data } = await api.get<ChatGroupInfoResponse>(`/chat/groups/${roomId}`)
  return data.group
}

export async function renameChatGroup(roomId: string, payload: UpdateChatGroupPayload) {
  const { data } = await api.patch<ChatRoomResponse>(`/chat/groups/${roomId}`, payload)
  return data.room
}

export async function leaveChatGroup(roomId: string) {
  const { data } = await api.post<{ message: string }>(`/chat/groups/${roomId}/leave`)
  return data
}

export async function addChatGroupMembers(roomId: string, payload: AddChatGroupMembersPayload) {
  const { data } = await api.post<{ message: string }>(`/chat/groups/${roomId}/members`, payload)
  return data
}

export async function removeChatGroupMember(roomId: string, userId: string) {
  const { data } = await api.delete<{ message: string }>(`/chat/groups/${roomId}/members/${userId}`)
  return data
}

export async function getMessages(
  roomId: string,
  params: { limit?: number; before?: string | null } = {},
) {
  const { data } = await api.get<ChatMessagesResponse>(`/chat/rooms/${roomId}/messages`, {
    params: {
      limit: params.limit,
      before: params.before || undefined,
    },
  })
  return {
    ...data,
    messages: data.messages ?? [],
  }
}

export async function sendMessage(roomId: string, content: string) {
  const payload: ChatMessagePayload = { content }
  const { data } = await api.post<ChatMessage>(`/chat/rooms/${roomId}/messages`, payload)
  return data
}

export async function markRoomRead(roomId: string, payload: MarkRoomReadPayload = {}) {
  const { data } = await api.patch<{ message: string }>(`/chat/rooms/${roomId}/read`, payload)
  return data
}
