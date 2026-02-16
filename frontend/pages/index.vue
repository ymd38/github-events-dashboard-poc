<template>
  <div class="max-w-7xl mx-auto px-4 py-6">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-bold text-gray-900">Events</h2>
      <EventFilter v-model="filterType" @update:model-value="onFilterChange" />
    </div>
    <div v-if="loading && events.length === 0" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2">
        <EventList
          ref="eventListRef"
          :events="events"
          @select="onSelectEvent"
        />
        <Pagination
          :page="pagination.page"
          :total-pages="pagination.total_pages"
          :total="pagination.total"
          @update:page="onPageChange"
        />
      </div>
      <div class="lg:col-span-1">
        <EventDetail
          :event="selectedEvent"
          @close="selectedEvent = null"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Event } from '~/types/event'
import { useEvents } from '~/composables/useEvents'
import { useSSE } from '~/composables/useSSE'

const {
  events,
  pagination,
  loading,
  selectedEvent,
  filterType,
  fetchEvents,
  setPage,
  setFilter,
  prependEvent,
} = useEvents()

const eventListRef = ref<InstanceType<typeof import('~/components/EventList.vue').default> | null>(null)

const { connect } = useSSE({
  onNewEvent: (event) => {
    prependEvent(event)
    if (eventListRef.value) {
      eventListRef.value.markAsNew(event.id)
    }
  },
  onSessionExpired: () => {
    navigateTo('/login')
  },
  onReconnect: () => {
    fetchEvents()
  },
})

const onFilterChange = (eventType: string): void => {
  setFilter(eventType)
}

const onPageChange = (page: number): void => {
  setPage(page)
}

const onSelectEvent = (event: Event): void => {
  selectedEvent.value = event
}

onMounted(() => {
  fetchEvents()
  connect()
})
</script>
