<template>
  <nav v-if="totalPages > 1" class="flex items-center justify-center gap-1.5 mt-5">
    <button
      :disabled="page <= 1"
      class="px-3 py-1.5 rounded-lg text-sm font-medium border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:border-gray-300 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-150"
      @click="emit('update:page', page - 1)"
    >
      ← Prev
    </button>
    <template v-for="p in visiblePages" :key="p">
      <button
        :class="[
          'w-9 h-9 rounded-lg text-sm font-medium border transition-all duration-150',
          p === page
            ? 'bg-gray-900 text-white border-gray-900 shadow-sm'
            : 'border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:border-gray-300',
        ]"
        @click="emit('update:page', p)"
      >
        {{ p }}
      </button>
    </template>
    <button
      :disabled="page >= totalPages"
      class="px-3 py-1.5 rounded-lg text-sm font-medium border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:border-gray-300 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-150"
      @click="emit('update:page', page + 1)"
    >
      Next →
    </button>
    <span class="text-xs text-gray-400 ml-1">
      {{ page }} / {{ totalPages }} ({{ total }} events)
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
