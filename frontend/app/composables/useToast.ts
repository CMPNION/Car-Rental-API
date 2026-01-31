type ToastItem = {
  id: number
  message: string
  type?: 'success' | 'error' | 'info'
}

export const useToast = () => {
  const toasts = useState<ToastItem[]>('toasts', () => [])

  const push = (message: string, type: ToastItem['type'] = 'info') => {
    const id = Date.now() + Math.floor(Math.random() * 1000)
    toasts.value = [...toasts.value, { id, message, type }]
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, 3500)
  }

  const remove = (id: number) => {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  return { toasts, push, remove }
}
