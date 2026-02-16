import { ref, type Ref } from 'vue'
import type { Event, Pagination, EventListResponse } from '~/types/event'

interface UseEventsReturn {
  events: Ref<Event[]>
  pagination: Ref<Pagination>
  loading: Ref<boolean>
  error: Ref<string | null>
  selectedEvent: Ref<Event | null>
  filterType: Ref<string>
  fetchEvents: () => Promise<void>
  fetchEventById: (id: number) => Promise<void>
  setPage: (page: number) => void
  setFilter: (eventType: string) => void
  prependEvent: (event: Event) => void
}

/**
 * Composable for managing event data, pagination, and filtering.
 */
export const useEvents = (): UseEventsReturn => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase as string

  const events = ref<Event[]>([])
  const pagination = ref<Pagination>({
    page: 1,
    per_page: 20,
    total: 0,
    total_pages: 0,
  })
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)
  const selectedEvent = ref<Event | null>(null)
  const filterType = ref<string>('')

  const fetchEvents = async (): Promise<void> => {
    loading.value = true
    error.value = null
    try {
      const params = new URLSearchParams({
        page: String(pagination.value.page),
        per_page: String(pagination.value.per_page),
      })
      if (filterType.value) {
        params.set('event_type', filterType.value)
      }
      const response = await $fetch<EventListResponse>(
        `${apiBase}/api/events?${params.toString()}`,
        { credentials: 'include' }
      )
      events.value = response.events
      pagination.value = response.pagination
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'Failed to fetch events'
      error.value = message
    } finally {
      loading.value = false
    }
  }

  const fetchEventById = async (id: number): Promise<void> => {
    loading.value = true
    error.value = null
    try {
      const response = await $fetch<Event>(
        `${apiBase}/api/events/${id}`,
        { credentials: 'include' }
      )
      selectedEvent.value = response
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'Failed to fetch event'
      error.value = message
    } finally {
      loading.value = false
    }
  }

  const setPage = (page: number): void => {
    pagination.value.page = page
    fetchEvents()
  }

  const setFilter = (eventType: string): void => {
    filterType.value = eventType
    pagination.value.page = 1
    fetchEvents()
  }

  const prependEvent = (event: Event): void => {
    if (filterType.value && event.event_type !== filterType.value) {
      return
    }
    events.value = [event, ...events.value]
    pagination.value.total += 1
  }

  return {
    events,
    pagination,
    loading,
    error,
    selectedEvent,
    filterType,
    fetchEvents,
    fetchEventById,
    setPage,
    setFilter,
    prependEvent,
  }
}
