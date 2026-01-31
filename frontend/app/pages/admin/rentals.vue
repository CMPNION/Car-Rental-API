<template>
  <div class="panel">
    <h1>Аренды (Админ)</h1>
    <p class="muted">Управление активными и pending арендами.</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <div class="row" style="justify-content: space-between; align-items: center;">
      <div class="row">
        <label class="field">
          User ID
          <input v-model.number="filters.user_id" type="number" min="1" />
        </label>
      </div>
      <div class="row">
        <button class="secondary" @click="resetFilters">Сбросить</button>
        <button @click="refresh">Обновить</button>
      </div>
    </div>

    <div v-if="pending">Загрузка...</div>
    <div v-else-if="error">Ошибка: {{ errorMessage }}</div>
    <div v-else-if="rentals.length === 0" class="muted">Нет аренд.</div>

    <div v-else class="grid">
      <div v-for="rental in rentals" :key="getRentalId(rental)" class="card">
        <div class="card__title">Аренда #{{ getRentalId(rental) }}</div>
        <div class="card__meta">User ID: {{ rental.user_id }}</div>
        <div class="card__meta">Car ID: {{ rental.car_id }}</div>
        <div class="card__meta">С {{ formatDate(rental.start_date) }}</div>
        <div class="card__meta">По {{ formatDate(rental.end_date) }}</div>
        <div class="card__meta">Сумма: {{ rental.total_price }}</div>
        <div class="card__meta">
          Статус:
          <span class="badge" :class="statusClass(rental.status)">{{ rental.status }}</span>
        </div>
        <div class="row">
          <button
            v-if="rental.status === 'active'"
            :disabled="isActionLoading[getRentalId(rental)]"
            @click="finishRental(rental)"
          >
            Завершить
          </button>
          <button
            v-if="rental.status === 'pending'"
            class="secondary"
            :disabled="isActionLoading[getRentalId(rental)]"
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
</template>

<script setup lang="ts">
import admin from '~/middleware/admin'

definePageMeta({
  middleware: [admin]
})

const { authFetch } = useApi()

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

const filters = reactive({
  user_id: undefined as number | undefined
})

const query = computed(() => {
  const q: Record<string, string> = {}
  if (filters.user_id) q.user_id = String(filters.user_id)
  return q
})

const { data, pending, error, refresh } = await useAsyncData<Rental[]>(
  'admin-rentals',
  () => authFetch(`/api/v1/rentals`, { query: query.value })
)

watch(query, () => refresh(), { deep: true })

const rentals = computed(() => data.value ?? [])

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

const isActionLoading = reactive<Record<number, boolean>>({})
const actionErrors = reactive<Record<number, string>>({})

const finishRental = async (rental: Rental) => {
  const id = getRentalId(rental)
  actionErrors[id] = ''
  isActionLoading[id] = true
  try {
    await authFetch(`/api/v1/rentals/${id}/finish`, { method: 'POST' })
    await refresh()
  } catch (err: any) {
    actionErrors[id] = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось завершить'
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
  } catch (err: any) {
    actionErrors[id] = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось отменить'
  } finally {
    isActionLoading[id] = false
  }
}

const resetFilters = () => {
  filters.user_id = undefined
}
</script>
