<template>
  <div v-if="event" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-gray-900">Event Detail</h2>
      <button
        class="text-gray-400 hover:text-gray-600"
        @click="emit('close')"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <div class="space-y-3">
      <div class="flex items-center gap-3">
        <img
          v-if="event.sender_avatar_url"
          :src="event.sender_avatar_url"
          :alt="event.sender_login"
          class="w-12 h-12 rounded-full"
        />
        <div>
          <p class="font-medium text-gray-900">{{ event.sender_login }}</p>
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
        </div>
      </div>
      <div>
        <h3 class="text-base font-medium text-gray-900">{{ event.title || 'No title' }}</h3>
        <p v-if="event.body" class="text-sm text-gray-600 mt-1 whitespace-pre-wrap">{{ event.body }}</p>
      </div>
      <dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
        <dt class="text-gray-500">Repository</dt>
        <dd class="text-gray-900">{{ event.repo_name }}</dd>
        <dt class="text-gray-500">Event Type</dt>
        <dd class="text-gray-900">{{ event.event_type }} / {{ event.action }}</dd>
        <dt class="text-gray-500">Occurred At</dt>
        <dd class="text-gray-900">{{ new Date(event.occurred_at).toLocaleString() }}</dd>
        <dt class="text-gray-500">Received At</dt>
        <dd class="text-gray-900">{{ new Date(event.received_at).toLocaleString() }}</dd>
        <dt class="text-gray-500">Delivery ID</dt>
        <dd class="text-gray-900 font-mono text-xs">{{ event.delivery_id }}</dd>
      </dl>
      <a
        :href="event.html_url"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 mt-2"
      >
        View on GitHub
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
        </svg>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Event } from '~/types/event'

interface Props {
  event: Event | null
}

defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()
</script>
