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
      <span class="badge">{{ car.status }}</span>
    </div>

    <div class="spacer"></div>

    <div class="row">
      <div class="card" style="flex: 1;">
        <div class="card__title">Детали</div>
        <div class="card__meta">Категория: {{ car.category }}</div>
        <div class="card__meta">Цена/час: {{ car.price_per_hour }}</div>
        <div class="card__meta">Рейтинг: {{ car.rating }}</div>
      </div>
      <div class="card" style="flex: 1;">
        <div class="card__title">Описание</div>
        <div class="card__meta">{{ car.metadata || 'Нет описания' }}</div>
      </div>
    </div>

    <div class="spacer"></div>

    <NuxtLink to="/cars">
      <button class="secondary">Назад к списку</button>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
const { fetcher } = useApi()
const route = useRoute()

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
  return (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})
</script>
