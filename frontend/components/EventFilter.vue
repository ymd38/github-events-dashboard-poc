<template>
  <div class="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
    <button
      v-for="option in options"
      :key="option.value"
      :class="[
        'px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-150',
        modelValue === option.value
          ? 'bg-white text-gray-900 shadow-sm'
          : 'text-gray-500 hover:text-gray-700',
      ]"
      @click="emit('update:modelValue', option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
const options = [
  { value: '', label: 'All' },
  { value: 'issues', label: 'Issues' },
  { value: 'pull_request', label: 'Pull Requests' },
]

interface Props {
  modelValue: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
