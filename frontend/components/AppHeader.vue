<template>
  <header class="bg-gray-900 text-white shadow-md">
    <div class="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <svg class="w-7 h-7 text-white" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
        </svg>
        <h1 class="text-sm font-semibold tracking-tight">GitHub Events Dashboard</h1>
      </div>
      <div v-if="user" class="flex items-center gap-3">
        <img
          v-if="safeAvatarUrl(user.avatar_url)"
          :src="safeAvatarUrl(user.avatar_url) || ''"
          :alt="user.login"
          class="w-8 h-8 rounded-full ring-2 ring-gray-700"
        />
        <div
          v-else
          class="w-8 h-8 rounded-full bg-gray-700 ring-2 ring-gray-600 flex items-center justify-center text-sm font-bold"
        >
          {{ (user.display_name || user.login).charAt(0).toUpperCase() }}
        </div>
        <span class="text-sm text-gray-300 hidden sm:block">{{ user.display_name || user.login }}</span>
        <button
          class="text-xs text-gray-400 hover:text-white border border-gray-700 hover:border-gray-500 transition-all duration-150 px-3 py-1.5 rounded-lg"
          @click="logout"
        >
          Sign out
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { safeAvatarUrl } from '~/utils/url'

const { user, fetchUser, logout } = useAuth()

onMounted(() => {
  fetchUser()
})
</script>
