<template>
  <div class="panel" v-if="!hasValidId">Некорректный ID машины.</div>
  <div class="panel" v-else-if="pending">Загрузка...</div>
  <div class="panel" v-else-if="error">Ошибка: {{ errorMessage }}</div>
  <div class="panel" v-else-if="car">
    <div class="row" style="justify-content: space-between; align-items: flex-start;">
      <div>
        <h1>{{ car.mark }} {{ car.model }}</h1>
        <p class="muted">ID: {{ car.id ?? car.ID }}</p>
      </div>
      <span class="badge" :class="statusClass(car.status)">{{ car.status }}</span>
    </div>

    <div class="spacer"></div>

    <div class="panel" style="background: #f8fafc;">
      <div class="row" style="justify-content: space-between; align-items: center;">
        <div>
          <div class="card__title">Бронирование</div>
          <div class="card__meta">Выберите даты и подтвердите бронь.</div>
        </div>
        <button
          class="secondary"
          :disabled="car.status !== 'available'"
          @click="showBooking = !showBooking"
        >
          {{ showBooking ? 'Скрыть' : 'Забронировать' }}
        </button>
      </div>

      <div v-if="car.status !== 'available'" class="muted" style="margin-top: 8px;">
        Бронирование недоступно: статус машины {{ car.status }}.
      </div>

      <div v-if="showBooking" class="spacer"></div>

      <form v-if="showBooking" class="row" @submit.prevent="submitBooking">
        <label class="field">
          Дата начала
          <input v-model="bookingForm.start" type="datetime-local" required />
        </label>
        <label class="field">
          Дата окончания
          <input v-model="bookingForm.end" type="datetime-local" required />
        </label>
        <div class="field" style="min-width: 180px;">
          Итоговая цена
          <div class="card__title">{{ formattedTotal }}</div>
          <div class="card__meta">Предварительный расчёт</div>
          <div class="card__meta">Формула: (Rate × Duration) × (1 − Discount)</div>
        </div>
        <button type="submit" :disabled="bookingDisabled">Подтвердить</button>
      </form>

      <div v-if="bookings.length > 0" class="card__meta">
        Ближайшие занятые даты:
        <div v-for="(b, index) in bookings.slice(0, 3)" :key="index">
          {{ formatDate(b.start_date) }} — {{ formatDate(b.end_date) }}
        </div>
      </div>

      <p v-if="hasOverlap" class="muted">Выбранные даты пересекаются с текущими бронированиями.</p>
      <p v-if="bookingError" class="muted">{{ bookingError }}</p>
    </div>

    <div class="spacer"></div>

    <div class="row">
      <div class="card" style="flex: 1;">
        <div class="card__title">Галерея</div>
        <div class="gallery">
          <div v-for="(img, index) in gallery" :key="index" class="gallery__item">
            <img :src="img" alt="car image" />
          </div>
        </div>
      </div>
      <div class="card" style="flex: 1;">
        <div class="card__title">Характеристики</div>
        <div v-if="specs.length === 0" class="card__meta">{{ car.metadata || 'Нет описания' }}</div>
        <div v-else class="card__meta" v-for="spec in specs" :key="spec.key">
          {{ spec.key }}: {{ spec.value }}
        </div>
        <div class="card__meta">Категория: {{ car.category }}</div>
        <div class="card__meta">Цена/час: {{ car.price_per_hour }}</div>
        <div class="card__meta">Рейтинг: {{ car.rating }}</div>
      </div>
    </div>

    <div class="spacer"></div>

    <NuxtLink to="/cars">
      <button class="secondary">Назад к списку</button>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
const { fetcher, authFetch } = useApi()
const route = useRoute()
const token = useCookie('token')

type Car = {
  id?: number
  ID?: number
  mark: string
  model: string
  category: string
  status: string
  price_per_hour: number
  rating: number
  metadata: string
}

const carId = computed(() => {
  const raw = Array.isArray(route.params.id) ? route.params.id[0] : route.params.id
  return Number(raw)
})
const hasValidId = computed(() => Number.isInteger(carId.value) && carId.value > 0)

const { data, pending, error } = await useAsyncData<Car>(
  'car-detail',
  () => {
    if (!hasValidId.value) {
      return Promise.resolve(null as unknown as Car)
    }
    return fetcher(`/api/v1/cars/${carId.value}`)
  },
  { watch: [carId] }
)

const car = computed(() => data.value ?? null)

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const statusClass = (status: string) => {
  const normalized = status?.toLowerCase()
  return normalized ? `badge--${normalized}` : ''
}

const showBooking = ref(false)
const bookingError = ref('')
const isBooking = ref(false)

const bookingForm = reactive({
  start: '',
  end: ''
})

const profileRating = ref(0)

const totalPrice = computed(() => {
  if (!car.value) return 0
  const start = new Date(bookingForm.start)
  const end = new Date(bookingForm.end)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return 0
  const hours = (end.getTime() - start.getTime()) / 36e5
  if (hours <= 0) return 0
  const base = car.value.price_per_hour * hours
  const discount = profileRating.value > 4.5 ? 0.1 : 0
  const surcharge = profileRating.value > 0 && profileRating.value < 2 ? 0.2 : 0
  const final = base * (1 - discount + surcharge)
  return Math.round(final * 100) / 100
})

const formattedTotal = computed(() => (totalPrice.value > 0 ? `${totalPrice.value} ₽` : '—'))

const { data: bookingsData } = await useAsyncData<any[]>(
  'car-bookings',
  () => {
    if (!hasValidId.value) {
      return Promise.resolve([])
    }
    return fetcher(`/api/v1/cars/${carId.value}/bookings`)
  },
  { watch: [carId] }
)

const bookings = computed(() => bookingsData.value ?? [])

const hasOverlap = computed(() => {
  const start = new Date(bookingForm.start)
  const end = new Date(bookingForm.end)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return false
  return bookings.value.some((b) => {
    const bStart = new Date(b.start_date)
    const bEnd = new Date(b.end_date)
    return start < bEnd && end > bStart
  })
})

const bookingDisabled = computed(() => {
  if (!car.value || car.value.status !== 'available') return true
  if (!bookingForm.start || !bookingForm.end) return true
  if (totalPrice.value <= 0) return true
  if (hasOverlap.value) return true
  return isBooking.value
})

const loadProfile = async () => {
  if (!token.value) return
  try {
    const profile = await authFetch<{ rating: number }>(`/api/v1/users/me`)
    profileRating.value = profile?.rating ?? 0
  } catch {
    profileRating.value = 0
  }
}

const submitBooking = async () => {
  bookingError.value = ''
  if (!car.value) return
  isBooking.value = true

  try {
    await authFetch(`/api/v1/rentals`, {
      method: 'POST',
      body: {
        car_id: car.value.id ?? car.value.ID,
        start_date: new Date(bookingForm.start).toISOString(),
        end_date: new Date(bookingForm.end).toISOString()
      }
    })
    showBooking.value = false
    bookingForm.start = ''
    bookingForm.end = ''
    await refreshNuxtData('car-detail')
  } catch (err: any) {
    bookingError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось создать бронь'
  } finally {
    isBooking.value = false
  }
}

const parseMetadata = () => {
  if (!car.value?.metadata) return {}
  try {
    const data = JSON.parse(car.value.metadata)
    return typeof data === 'object' && data ? data : {}
  } catch {
    return {}
  }
}

const specs = computed(() => {
  const metadata = parseMetadata() as Record<string, any>
  const entries = Object.entries(metadata).filter(([key]) => key !== 'images')
  return entries.map(([key, value]) => ({ key, value }))
})

const gallery = computed(() => {
  const metadata = parseMetadata() as Record<string, any>
  if (Array.isArray(metadata.images) && metadata.images.length > 0) {
    return metadata.images
  }
  return [
    'https://images.unsplash.com/photo-1503376780353-7e6692767b70?q=80&w=800&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1493238792000-8113da705763?q=80&w=800&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1503736334956-4c8f8e92946d?q=80&w=800&auto=format&fit=crop'
  ]
})

const formatDate = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('ru-RU')
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.gallery {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
}

.gallery__item img {
  width: 100%;
  height: 110px;
  object-fit: cover;
  border-radius: 10px;
}
</style>
