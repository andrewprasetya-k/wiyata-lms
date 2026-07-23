import { getActiveSchoolId } from './session'
import { fetchWsTicket } from './wsTicket'
import type { ChatSocketEvent } from '../types/chat'

type ChatSocketOptions = {
  onEvent: (event: ChatSocketEvent) => void
  onStatusChange?: (status: ChatSocketStatus) => void
  onOpen?: () => void
  onClose?: () => void
}

type ChatSocketConnection = {
  close: () => void
}

export type ChatSocketStatus = 'connecting' | 'connected' | 'disconnected'

const reconnectDelaysMs = [1000, 2000, 5000, 10000]

export function connectChatSocket(options: ChatSocketOptions): ChatSocketConnection {
  let socket: WebSocket | null = null
  let reconnectTimer: number | undefined
  let closedByClient = false
  let retryIndex = 0
  let hasOpenedCurrentSocket = false
  let failedBeforeOpenCount = 0

  async function connect() {
    setStatus('connecting')
    hasOpenedCurrentSocket = false

    // Fetched fresh on every (re)connect attempt — a WS ticket is
    // single-use, so reusing one from a prior attempt would never work.
    const url = await buildChatSocketUrl()
    if (closedByClient) return // caller closed while the ticket fetch was in flight
    if (!url) {
      setStatus('disconnected')
      return
    }

    socket = new WebSocket(url)
    socket.onopen = () => {
      retryIndex = 0
      failedBeforeOpenCount = 0
      hasOpenedCurrentSocket = true
      setStatus('connected')
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
      setStatus('disconnected')
      options.onClose?.()
      socket = null
      if (!hasOpenedCurrentSocket) {
        failedBeforeOpenCount += 1
      }
      if (failedBeforeOpenCount >= 5) return
      if (!closedByClient && getActiveSchoolId()) {
        const delay = reconnectDelaysMs[Math.min(retryIndex, reconnectDelaysMs.length - 1)]
        retryIndex += 1
        reconnectTimer = window.setTimeout(connect, delay)
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
      setStatus('disconnected')
    },
  }

  function setStatus(status: ChatSocketStatus) {
    options.onStatusChange?.(status)
  }
}

async function buildChatSocketUrl(): Promise<string> {
  const schoolId = getActiveSchoolId()
  if (!schoolId) return ''

  const ticket = await fetchWsTicket()
  if (!ticket) return ''

  const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
  const url = new URL(`${apiBase.replace(/\/$/, '')}/ws/chat`, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('ticket', ticket)
  url.searchParams.set('schoolId', schoolId)
  return url.toString()
}
