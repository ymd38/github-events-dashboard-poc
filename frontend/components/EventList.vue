<template>
  <div class="space-y-3">
    <div
      v-for="event in events"
      :key="event.id"
      :class="[
        'bg-white rounded-lg shadow-sm border border-gray-200 p-4 cursor-pointer hover:shadow-md transition-shadow',
        isNew(event.id) ? 'ring-2 ring-blue-400 bg-blue-50 animate-pulse-once' : ''
      ]"
      @click="emit('select', event)"
    >
      <div class="flex items-start gap-3">
        <img
          v-if="event.sender_avatar_url"
          :src="event.sender_avatar_url"
          :alt="event.sender_login"
          class="w-10 h-10 rounded-full"
        />
        <div class="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center text-gray-600 text-sm font-bold" v-else>
          {{ event.sender_login.charAt(0).toUpperCase() }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1">
            <span
              :class="[
                'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                event.event_type === 'issues'
                  ? 'bg-green-100 text-green-800'
                  : 'bg-purple-100 text-purple-800'
              ]"
            >
              {{ event.event_type === 'issues' ? 'Issue Opened' : 'PR Merged' }}
            </span>
            <span class="text-sm text-gray-500">{{ event.repo_name }}</span>
          </div>
          <p class="text-sm font-medium text-gray-900 truncate">
            {{ event.title || 'No title' }}
          </p>
          <div class="flex items-center gap-2 mt-1 text-xs text-gray-500">
            <span>{{ event.sender_login }}</span>
            <span>&middot;</span>
            <span>{{ formatDate(event.occurred_at) }}</span>
          </div>
        </div>
      </div>
    </div>
    <div v-if="events.length === 0" class="text-center py-12 text-gray-500">
      <svg class="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
      <p class="text-lg font-medium">No events yet</p>
      <p class="text-sm mt-1">Events will appear here when GitHub webhooks are received.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Event } from '~/types/event'

interface Props {
  events: Event[]
}

defineProps<Props>()

const emit = defineEmits<{
  select: [event: Event]
}>()

const newEventIds = ref<Set<number>>(new Set())

const isNew = (id: number): boolean => newEventIds.value.has(id)

const markAsNew = (id: number): void => {
  newEventIds.value.add(id)
  setTimeout(() => {
    newEventIds.value.delete(id)
  }, 5000)
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

defineExpose({ markAsNew })
</script>

<style scoped>
@keyframes pulse-once {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
.animate-pulse-once {
  animation: pulse-once 1s ease-in-out 2;
}
</style>
