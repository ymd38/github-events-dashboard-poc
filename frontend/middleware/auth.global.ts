import { useAuth } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware(async (to) => {
  const { fetchUser, isAuthenticated } = useAuth()
  await fetchUser()
  if (to.path === '/login' && isAuthenticated.value) {
    return navigateTo('/')
  }
  if (to.path === '/login') {
    return
  }
  if (!isAuthenticated.value) {
    return navigateTo('/login')
  }
})
