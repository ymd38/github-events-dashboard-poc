import { useAuth } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/login') {
    return
  }
  const { fetchUser, isAuthenticated } = useAuth()
  await fetchUser()
  if (!isAuthenticated.value) {
    return navigateTo('/login')
  }
})
