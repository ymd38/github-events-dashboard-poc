import { ref, onUnmounted } from 'vue'
import type { Event } from '~/types/event'

interface UseSSEOptions {
  onNewEvent: (event: Event) => void
  onSessionExpired?: () => void
  onReconnect?: () => void
}

interface UseSSEReturn {
  connected: ReturnType<typeof ref<boolean>>
  connect: () => void
  disconnect: () => void
}

const RECONNECT_DELAY_MS = 3000
const MAX_RECONNECT_DELAY_MS = 30000

/**
 * Composable for managing SSE connection with auto-reconnect.
 */
export const useSSE = (options: UseSSEOptions): UseSSEReturn => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase as string

  const connected = ref<boolean>(false)
  let eventSource: EventSource | null = null
  let reconnectDelay = RECONNECT_DELAY_MS
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  const connect = (): void => {
    if (eventSource) {
      eventSource.close()
    }
    const url = `${apiBase}/api/events/stream`
    eventSource = new EventSource(url, { withCredentials: true })
    eventSource.addEventListener('open', () => {
      connected.value = true
      reconnectDelay = RECONNECT_DELAY_MS
    })
    eventSource.addEventListener('new_event', (e: MessageEvent) => {
      try {
        const event: Event = JSON.parse(e.data)
        options.onNewEvent(event)
      } catch {
        console.error('Failed to parse SSE event data')
      }
    })
    eventSource.addEventListener('session_expired', () => {
      disconnect()
      if (options.onSessionExpired) {
        options.onSessionExpired()
      }
    })
    eventSource.addEventListener('error', () => {
      connected.value = false
      if (eventSource) {
        eventSource.close()
        eventSource = null
      }
      scheduleReconnect()
    })
  }

  const disconnect = (): void => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connected.value = false
  }

  const scheduleReconnect = (): void => {
    if (reconnectTimer) {
      return
    }
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      if (options.onReconnect) {
        options.onReconnect()
      }
      connect()
      reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY_MS)
    }, reconnectDelay)
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    connect,
    disconnect,
  }
}
