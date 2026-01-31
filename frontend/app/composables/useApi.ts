type ApiResponse<T> = {
  status?: 'ok' | 'error'
  message?: string
  data?: T
}

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

  const unwrap = <T>(response: T | ApiResponse<T>) => {
    if (response && typeof response === 'object' && 'status' in response) {
      const envelope = response as ApiResponse<T>
      if (envelope.status === 'ok') {
        return envelope.data as T
      }
      if (envelope.status === 'error') {
        const err: any = new Error(envelope.message || 'Request failed')
        err.data = envelope
        throw err
      }
    }
    return response as T
  }

  const fetcher = async <T>(path: string, options: any = {}) => {
    const response = await $fetch<T | ApiResponse<T>>(resolveUrl(path), options)
    return unwrap(response)
  }

  const authFetch = async <T>(path: string, options: any = {}) => {
    const headers = {
      ...(options.headers ?? {}),
      Authorization: `Bearer ${token.value ?? ''}`
    }
    const response = await $fetch<T | ApiResponse<T>>(resolveUrl(path), { ...options, headers })
    return unwrap(response)
  }

  return { fetcher, authFetch }
}
