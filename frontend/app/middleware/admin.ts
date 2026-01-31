export default defineNuxtRouteMiddleware(async () => {
  const { authFetch } = useApi()

  try {
    const data = await authFetch<{ role: string; is_admin: boolean }>(`/auth/me`)
    if (!data?.is_admin) {
      return navigateTo('/')
    }
  } catch {
    return navigateTo('/')
  }
})
