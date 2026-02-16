import { ref } from 'vue'

interface User {
  id: number
  login: string
  display_name: string
  avatar_url: string | null
}

interface UseAuthReturn {
  user: ReturnType<typeof ref<User | null>>
  loading: ReturnType<typeof ref<boolean>>
  isAuthenticated: ReturnType<typeof ref<boolean>>
  fetchUser: () => Promise<void>
  login: () => void
  logout: () => Promise<void>
}

/**
 * Composable for managing authentication state.
 */
export const useAuth = (): UseAuthReturn => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase as string

  const user = ref<User | null>(null)
  const loading = ref<boolean>(true)
  const isAuthenticated = ref<boolean>(false)

  const fetchUser = async (): Promise<void> => {
    loading.value = true
    try {
      const response = await $fetch<User>(`${apiBase}/api/auth/me`, {
        credentials: 'include',
      })
      user.value = response
      isAuthenticated.value = true
    } catch {
      user.value = null
      isAuthenticated.value = false
    } finally {
      loading.value = false
    }
  }

  const login = (): void => {
    window.location.href = `${apiBase}/api/auth/login`
  }

  const logout = async (): Promise<void> => {
    try {
      await $fetch(`${apiBase}/api/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      })
    } catch {
      // ignore
    }
    user.value = null
    isAuthenticated.value = false
    navigateTo('/login')
  }

  return {
    user,
    loading,
    isAuthenticated,
    fetchUser,
    login,
    logout,
  }
}
