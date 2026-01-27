<template>
  <div class="panel">
    <div class="row" style="justify-content: space-between; align-items: flex-end;">
      <div>
        <h1>Машины</h1>
        <p class="muted">Фильтры и сортировка</p>
      </div>
      <button class="secondary" @click="resetFilters">Сбросить фильтры</button>
    </div>

    <div class="spacer"></div>

    <div class="row">
      <label class="field">
        Марка
        <input v-model.trim="filters.mark" placeholder="Toyota" />
      </label>
      <label class="field">
        Категория
        <select v-model="filters.category">
          <option value="">Любая</option>
          <option value="economy">Economy</option>
          <option value="business">Business</option>
          <option value="luxury">Luxury</option>
        </select>
      </label>
      <label class="field">
        Статус
        <select v-model="filters.status">
          <option value="">Любой</option>
          <option value="available">Available</option>
          <option value="booked">Booked</option>
          <option value="maintenance">Maintenance</option>
        </select>
      </label>
      <label class="field">
        Мин. цена/час
        <input v-model.number="filters.min_price" type="number" min="0" />
      </label>
      <label class="field">
        Макс. цена/час
        <input v-model.number="filters.max_price" type="number" min="0" />
      </label>
      <label class="field">
        Сортировать
        <select v-model="filters.sort">
          <option value="">Без сортировки</option>
          <option value="price_per_hour">Цена</option>
          <option value="rating">Рейтинг</option>
          <option value="created_at">Дата</option>
        </select>
      </label>
      <label class="field">
        Порядок
        <select v-model="filters.order">
          <option value="asc">ASC</option>
          <option value="desc">DESC</option>
        </select>
      </label>
      <label class="field">
        Лимит
        <input v-model.number="filters.limit" type="number" min="1" max="200" />
      </label>
      <label class="field">
        Оффсет
        <input v-model.number="filters.offset" type="number" min="0" />
      </label>
    </div>
  </div>

  <div class="spacer"></div>

  <div v-if="pending" class="panel">Загрузка...</div>
  <div v-else-if="error" class="panel">Ошибка: {{ errorMessage }}</div>
  <div v-else class="grid">
    <div v-for="car in cars" :key="getCarId(car)" class="card">
      <div class="card__title">{{ car.mark }} {{ car.model }}</div>
      <div class="card__meta">Категория: {{ car.category }}</div>
      <div class="card__meta">Статус: <span class="badge">{{ car.status }}</span></div>
      <div class="card__meta">Цена: {{ car.price_per_hour }} / час</div>
      <div class="row">
        <NuxtLink :to="`/cars/${getCarId(car)}`">
          <button class="secondary">Открыть</button>
        </NuxtLink>
      </div>
    </div>
    <div v-if="cars.length === 0" class="panel">Ничего не найдено</div>
  </div>
</template>

<script setup lang="ts">
const { fetcher } = useApi()

type Car = {
  id?: number
  ID?: number
  mark: string
  model: string
  category: string
  status: string
  price_per_hour: number
  rating: number
}

const getCarId = (car: Car) => car.id ?? car.ID ?? 0

const filters = reactive({
  mark: '',
  category: '',
  status: '',
  min_price: undefined as number | undefined,
  max_price: undefined as number | undefined,
  sort: '',
  order: 'asc',
  limit: 20,
  offset: 0
})

const query = computed(() => {
  const q: Record<string, string> = {}
  if (filters.mark) q.mark = filters.mark
  if (filters.category) q.category = filters.category
  if (filters.status) q.status = filters.status
  if (filters.min_price !== undefined && filters.min_price !== null) q.min_price = String(filters.min_price)
  if (filters.max_price !== undefined && filters.max_price !== null) q.max_price = String(filters.max_price)
  if (filters.sort) q.sort = filters.sort
  if (filters.order) q.order = filters.order
  if (filters.limit) q.limit = String(filters.limit)
  if (filters.offset) q.offset = String(filters.offset)
  return q
})

const { data, pending, error, refresh } = await useAsyncData<Car[]>(
  'cars-list',
  () => fetcher(`/api/v1/cars`, { query: query.value })
)

watch(query, () => refresh(), { deep: true })

const cars = computed(() => data.value ?? [])

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const resetFilters = () => {
  filters.mark = ''
  filters.category = ''
  filters.status = ''
  filters.min_price = undefined
  filters.max_price = undefined
  filters.sort = ''
  filters.order = 'asc'
  filters.limit = 20
  filters.offset = 0
}
</script>
