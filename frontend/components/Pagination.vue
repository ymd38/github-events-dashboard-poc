<template>
  <nav v-if="totalPages > 1" class="flex items-center justify-center gap-2 mt-4">
    <button
      :disabled="page <= 1"
      class="px-3 py-1 rounded-md text-sm border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
      @click="emit('update:page', page - 1)"
    >
      Previous
    </button>
    <template v-for="p in visiblePages" :key="p">
      <button
        :class="[
          'px-3 py-1 rounded-md text-sm border',
          p === page
            ? 'bg-blue-600 text-white border-blue-600'
            : 'border-gray-300 bg-white hover:bg-gray-50'
        ]"
        @click="emit('update:page', p)"
      >
        {{ p }}
      </button>
    </template>
    <button
      :disabled="page >= totalPages"
      class="px-3 py-1 rounded-md text-sm border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
      @click="emit('update:page', page + 1)"
    >
      Next
    </button>
    <span class="text-sm text-gray-500 ml-2">
      Page {{ page }} of {{ totalPages }} ({{ total }} events)
    </span>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  page: number
  totalPages: number
  total: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:page': [value: number]
}>()

const maxVisiblePages = 5

const visiblePages = computed<number[]>(() => {
  const pages: number[] = []
  let start = Math.max(1, props.page - Math.floor(maxVisiblePages / 2))
  const end = Math.min(props.totalPages, start + maxVisiblePages - 1)
  start = Math.max(1, end - maxVisiblePages + 1)
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})
</script>
