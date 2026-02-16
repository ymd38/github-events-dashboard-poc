<template>
  <div class="flex items-center gap-2">
    <label class="text-sm font-medium text-gray-700">Filter:</label>
    <select
      :value="modelValue"
      class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      @change="onChange"
    >
      <option value="">All Events</option>
      <option value="issues">Issues</option>
      <option value="pull_request">Pull Requests</option>
    </select>
  </div>
</template>

<script setup lang="ts">
interface Props {
  modelValue: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const onChange = (e: globalThis.Event): void => {
  const target = e.target as HTMLSelectElement
  emit('update:modelValue', target.value)
}
</script>
