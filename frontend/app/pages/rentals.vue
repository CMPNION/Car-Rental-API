<template>
  <div class="panel">
    <h1>Мои аренды</h1>
    <p class="muted">История бронирований и управление оплатой.</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <div class="row" style="justify-content: space-between; align-items: center;">
      <h2>Активные и предстоящие</h2>
      <button class="secondary" @click="refresh">Обновить</button>
    </div>
    <div v-if="pending">Загрузка...</div>
    <div v-else-if="error">Ошибка: {{ errorMessage }}</div>
    <div v-else-if="activeRentals.length === 0" class="muted">Нет активных аренд.</div>

    <div v-else class="grid">
      <div v-for="rental in activeRentals" :key="getRentalId(rental)" class="card">
        <div class="card__title">Аренда #{{ getRentalId(rental) }}</div>
        <div class="card__meta">Машина ID: {{ rental.car_id }}</div>
        <div class="card__meta">С {{ formatDate(rental.start_date) }}</div>
        <div class="card__meta">По {{ formatDate(rental.end_date) }}</div>
        <div class="card__meta">Сумма: {{ rental.total_price }}</div>
        <div class="card__meta" v-if="transactionMap[getRentalId(rental)]">
          Transaction ID: {{ transactionMap[getRentalId(rental)] }}
        </div>
        <div class="card__meta">Осталось: {{ countdown(rental) }}</div>
        <div class="card__meta">
          Статус:
          <span class="badge" :class="statusClass(rental.status)">{{ rental.status }}</span>
        </div>
        <div class="row">
          <button
            v-if="rental.status === 'pending'"
            :disabled="isActionLoading[getRentalId(rental)]"
            @click="payRental(rental)"
          >
            Оплатить
          </button>
          <button
            v-if="rental.status === 'pending'"
            class="secondary"
            :disabled="isActionLoading[getRentalId(rental)] || isCloseToStart(rental.start_date)"
            @click="cancelRental(rental)"
          >
            Отменить
          </button>
        </div>
        <p v-if="actionErrors[getRentalId(rental)]" class="muted">
          {{ actionErrors[getRentalId(rental)] }}
        </p>
      </div>
    </div>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>История поездок</h2>
    <div v-if="completedRentals.length === 0" class="muted">Завершённых аренд нет.</div>
    <div v-else class="grid">
      <div v-for="rental in completedRentals" :key="getRentalId(rental)" class="card">
        <div class="card__title">Аренда #{{ getRentalId(rental) }}</div>
        <div class="card__meta">Машина ID: {{ rental.car_id }}</div>
        <div class="card__meta">Период: {{ formatDate(rental.start_date) }} — {{ formatDate(rental.end_date) }}</div>
        <div class="card__meta">Сумма: {{ rental.total_price }}</div>
        <div class="card__meta" v-if="transactionMap[getRentalId(rental)]">
          Transaction ID: {{ transactionMap[getRentalId(rental)] }}
        </div>
        <div class="card__meta">
          Статус:
          <span class="badge" :class="statusClass(rental.status)">{{ rental.status }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { authFetch } = useApi()
const { push } = useToast()

type Rental = {
  id?: number
  ID?: number
  user_id: number
  car_id: number
  start_date: string
  end_date: string
  total_price: number
  status: string
}

const getRentalId = (rental: Rental) => rental.id ?? rental.ID ?? 0

const { data, pending, error, refresh } = await useAsyncData<Rental[]>(
  'my-rentals',
  () => authFetch(`/api/v1/rentals`)
)

const rentals = computed(() => data.value ?? [])
const activeRentals = computed(() =>
  rentals.value.filter((r) => r.status === 'pending' || r.status === 'active')
)
const completedRentals = computed(() => rentals.value.filter((r) => r.status === 'completed'))

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const statusClass = (status: string) => {
  const normalized = status?.toLowerCase()
  return normalized ? `badge--${normalized}` : ''
}

const formatDate = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('ru-RU')
}

const isCloseToStart = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return false
  return date.getTime() - Date.now() <= 24 * 60 * 60 * 1000
}

const countdown = (rental: Rental) => {
  const now = Date.now()
  const start = new Date(rental.start_date).getTime()
  const end = new Date(rental.end_date).getTime()
  const target = rental.status === 'active' ? end : start
  const diff = Math.max(target - now, 0)
  const hours = Math.floor(diff / 36e5)
  const minutes = Math.floor((diff % 36e5) / 60000)
  return `${hours}ч ${minutes}м`
}

const isActionLoading = reactive<Record<number, boolean>>({})
const actionErrors = reactive<Record<number, string>>({})

const payRental = async (rental: Rental) => {
  const id = getRentalId(rental)
  actionErrors[id] = ''
  isActionLoading[id] = true
  try {
    await authFetch(`/api/v1/rentals/${id}/pay`, { method: 'POST' })
    await refresh()
    push('Оплата прошла успешно', 'success')
  } catch (err: any) {
    actionErrors[id] = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось оплатить'
  } finally {
    isActionLoading[id] = false
  }
}

const cancelRental = async (rental: Rental) => {
  const id = getRentalId(rental)
  actionErrors[id] = ''
  isActionLoading[id] = true
  try {
    await authFetch(`/api/v1/rentals/${id}/cancel`, { method: 'POST' })
    await refresh()
    push('Бронь отменена', 'info')
  } catch (err: any) {
    actionErrors[id] = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось отменить'
  } finally {
    isActionLoading[id] = false
  }
}

const { data: txData, refresh: refreshTx } = await useAsyncData<any[]>(
  'rentals-transactions',
  () => authFetch(`/api/v1/transactions`)
)

const transactionMap = computed<Record<number, number>>(() => {
  const map: Record<number, number> = {}
  ;(txData.value ?? []).forEach((tx: any) => {
    if (tx.rental_id && tx.type === 'payment') {
      map[tx.rental_id] = tx.id
    }
  })
  return map
})

watch(rentals, () => refreshTx())
</script>
