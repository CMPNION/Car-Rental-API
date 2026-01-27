export const useApi = () => {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase as string
  const token = useCookie('token')

  const resolveUrl = (path: string) => {
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path
    }
    if (path.startsWith('/')) {
      return `${baseURL}${path}`
    }
    return `${baseURL}/${path}`
  }

  const fetcher = async <T>(path: string, options: any = {}) => {
    return await $fetch<T>(resolveUrl(path), options)
  }

  const authFetch = async <T>(path: string, options: any = {}) => {
    const headers = {
      ...(options.headers ?? {}),
      Authorization: `Bearer ${token.value ?? ''}`
    }
    return await $fetch<T>(resolveUrl(path), { ...options, headers })
  }

  return { fetcher, authFetch }
}
