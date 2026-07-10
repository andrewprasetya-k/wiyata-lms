import type { ChatRoom } from '../types/chat'

export function getInitials(value?: string | null) {
  return (
    value
      ?.split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((word) => word.charAt(0).toUpperCase())
      .join('') || 'RS'
  )
}

export function isDirectMessageRoom(room: ChatRoom) {
  return room.roomType === 'dm'
}

export function isCustomGroupRoom(room: ChatRoom) {
  return room.roomType === 'group' && room.roomRefType == null
}

export function roomDisplayName(room?: ChatRoom | null) {
  if (!room) return 'Ruang Sekolah'
  if (isDirectMessageRoom(room)) {
    return room.dmTargetName || room.dmTargetEmail || 'Direct Message'
  }
  return room.roomName || 'Ruang Grup'
}

export function resolveChatError(error: unknown) {
  if (error instanceof Error && error.message) {
    return error.message
  }
  const maybeError = error as {
    response?: { status?: number; data?: { error?: string } }
  }
  if (
    maybeError.response?.status === 401 ||
    maybeError.response?.status === 403
  ) {
    return 'Kamu tidak memiliki akses ke chat sekolah ini.'
  }
  return maybeError.response?.data?.error || 'Gagal memuat chat. Coba lagi.'
}
