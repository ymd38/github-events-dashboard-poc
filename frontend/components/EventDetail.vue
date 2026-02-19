<template>
  <div v-if="event" class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
    <div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
      <h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">Event Detail</h2>
      <button
        class="w-7 h-7 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-700 hover:bg-gray-100 transition-all"
        @click="emit('close')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="px-5 py-4 space-y-4">
      <div class="flex items-center gap-3">
        <img
          v-if="event.sender_avatar_url"
          :src="event.sender_avatar_url"
          :alt="event.sender_login"
          class="w-11 h-11 rounded-full ring-2 ring-gray-100"
        />
        <div
          v-else
          class="w-11 h-11 rounded-full bg-gray-100 flex items-center justify-center text-gray-600 text-base font-bold"
        >
          {{ event.sender_login.charAt(0).toUpperCase() }}
        </div>
        <div>
          <p class="font-semibold text-gray-900 text-sm">{{ event.sender_login }}</p>
          <span
            :class="[
              'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold mt-0.5',
              event.event_type === 'issues'
                ? 'bg-green-100 text-green-700'
                : 'bg-violet-100 text-violet-700',
            ]"
          >
            {{ event.event_type === 'issues' ? 'Issue Opened' : 'PR Merged' }}
          </span>
        </div>
      </div>

      <div>
        <h3 class="text-base font-semibold text-gray-900 leading-snug">{{ event.title || 'No title' }}</h3>
        <div v-if="event.body" class="mt-2 bg-gray-50 rounded-lg p-3 max-h-40 overflow-y-auto">
          <p class="text-sm text-gray-600 whitespace-pre-wrap leading-relaxed">{{ event.body }}</p>
        </div>
      </div>

      <div class="border-t border-gray-100 pt-3">
        <dl class="space-y-2 text-sm">
          <div class="flex gap-3">
            <dt class="w-24 shrink-0 text-gray-400">Repository</dt>
            <dd class="text-gray-800 font-medium min-w-0 truncate">{{ event.repo_name }}</dd>
          </div>
          <div class="flex gap-3">
            <dt class="w-24 shrink-0 text-gray-400">Type</dt>
            <dd class="text-gray-800">{{ event.event_type }} / {{ event.action }}</dd>
          </div>
          <div class="flex gap-3">
            <dt class="w-24 shrink-0 text-gray-400">Occurred</dt>
            <dd class="text-gray-800">{{ new Date(event.occurred_at).toLocaleString() }}</dd>
          </div>
          <div class="flex gap-3">
            <dt class="w-24 shrink-0 text-gray-400">Received</dt>
            <dd class="text-gray-800">{{ new Date(event.received_at).toLocaleString() }}</dd>
          </div>
          <div class="flex gap-3">
            <dt class="w-24 shrink-0 text-gray-400">Delivery ID</dt>
            <dd class="text-gray-600 font-mono text-xs break-all">{{ event.delivery_id }}</dd>
          </div>
        </dl>
      </div>

      <a
        :href="event.html_url"
        target="_blank"
        rel="noopener noreferrer"
        class="flex items-center justify-center gap-2 w-full text-sm font-medium text-gray-700 bg-gray-50 hover:bg-gray-100 border border-gray-200 rounded-lg px-4 py-2.5 transition-all duration-150"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
        </svg>
        View on GitHub
      </a>
    </div>
  </div>

  <div
    v-else
    class="bg-white rounded-xl border border-dashed border-gray-200 flex flex-col items-center justify-center text-center p-8 min-h-[220px]"
  >
    <svg class="w-10 h-10 text-gray-200 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
    </svg>
    <p class="text-sm font-medium text-gray-400">Select an event to view details</p>
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
