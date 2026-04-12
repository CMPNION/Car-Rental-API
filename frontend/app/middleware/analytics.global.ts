export default defineNuxtRouteMiddleware((to, from) => {
  if (import.meta.server) return

  // Track page view on client side
  const trackPageView = async (path: string) => {
    try {
      await $fetch(`/api/metrics?action=track&path=${encodeURIComponent(path)}`)
    } catch {
      // Ignore tracking errors
    }
  }

  trackPageView(to.path)
})
