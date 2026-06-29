import { getActiveSchoolId, getStoredToken } from './session'
import type { ChatSocketEvent } from '../types/chat'

type ChatSocketOptions = {
  onEvent: (event: ChatSocketEvent) => void
  onOpen?: () => void
  onClose?: () => void
}

type ChatSocketConnection = {
  close: () => void
}

const reconnectDelayMs = 3000

export function connectChatSocket(options: ChatSocketOptions): ChatSocketConnection {
  let socket: WebSocket | null = null
  let reconnectTimer: number | undefined
  let closedByClient = false

  function connect() {
    const url = buildChatSocketUrl()
    if (!url) return

    socket = new WebSocket(url)
    socket.onopen = () => {
      options.onOpen?.()
    }
    socket.onmessage = (message) => {
      try {
        options.onEvent(JSON.parse(message.data) as ChatSocketEvent)
      } catch {
        // Ignore malformed realtime events and keep REST/polling as source of truth.
      }
    }
    socket.onclose = () => {
      options.onClose?.()
      socket = null
      if (!closedByClient) {
        reconnectTimer = window.setTimeout(connect, reconnectDelayMs)
      }
    }
    socket.onerror = () => {
      socket?.close()
    }
  }

  connect()

  return {
    close() {
      closedByClient = true
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer)
      }
      socket?.close()
      socket = null
    },
  }
}

function buildChatSocketUrl() {
  const token = getStoredToken()
  const schoolId = getActiveSchoolId()
  if (!token || !schoolId) return ''

  const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
  const url = new URL(`${apiBase.replace(/\/$/, '')}/ws/chat`, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('token', token)
  url.searchParams.set('schoolId', schoolId)
  return url.toString()
}
