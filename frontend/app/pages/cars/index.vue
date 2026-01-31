<template>
  <div class="catalog">
    <aside class="panel catalog__filters">
      <div class="row" style="justify-content: space-between; align-items: center;">
        <h2>Фильтры</h2>
        <button class="secondary" @click="resetFilters">Сбросить</button>
      </div>

      <label class="field">
        Поиск по названию
        <input v-model.trim="filters.search" placeholder="Toyota Camry" />
      </label>

      <div class="field">
        Категории
        <label class="row"><input type="checkbox" value="economy" v-model="filters.categories" /> Economy</label>
        <label class="row"><input type="checkbox" value="business" v-model="filters.categories" /> Business</label>
        <label class="row"><input type="checkbox" value="luxury" v-model="filters.categories" /> Luxury</label>
      </div>

      <label class="field">
        Диапазон цены (₽/час)
        <div class="row">
          <input v-model.number="filters.min_price" type="number" min="0" style="max-width: 90px;" />
          <span>—</span>
          <input v-model.number="filters.max_price" type="number" min="0" style="max-width: 90px;" />
        </div>
        <input v-model.number="filters.max_price" type="range" min="0" max="500" />
      </label>

      <label class="field">
        Только свободные
        <input type="checkbox" v-model="filters.onlyAvailable" />
      </label>

      <label class="field">
        Сортировка
        <select v-model="filters.sorting">
          <option value="">Без сортировки</option>
          <option value="price_asc">Сначала дешевые</option>
          <option value="price_desc">Сначала дорогие</option>
          <option value="rating_desc">По рейтингу</option>
        </select>
      </label>
    </aside>

    <section class="catalog__list">
      <div class="row" style="justify-content: space-between; align-items: center;">
        <div>
          <h1>Каталог автомобилей</h1>
          <p class="muted">Найдено: {{ filteredCars.length }}</p>
        </div>
      </div>

      <div v-if="pending" class="grid">
        <div v-for="i in 6" :key="i" class="skeleton"></div>
      </div>
      <div v-else-if="error" class="panel">Ошибка: {{ errorMessage }}</div>
      <div v-else class="grid">
        <div v-for="car in filteredCars" :key="getCarId(car)" class="card">
          <div class="card__image">{{ car.mark.slice(0, 1) }}</div>
          <div class="card__title">{{ car.mark }} {{ car.model }}</div>
          <div class="card__meta">
            <span class="badge" :class="statusClass(car.category)">{{ car.category }}</span>
          </div>
          <div class="card__meta">Цена: {{ car.price_per_hour }} / час</div>
          <div class="row">
            <NuxtLink :to="`/cars/${getCarId(car)}`">
              <button class="secondary">Подробнее</button>
            </NuxtLink>
          </div>
        </div>
        <div v-if="filteredCars.length === 0" class="panel">Ничего не найдено</div>
      </div>
    </section>
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

const route = useRoute()

const filters = reactive({
  search: String(route.query.search ?? ''),
  categories: [] as string[],
  min_price: undefined as number | undefined,
  max_price: 200 as number | undefined,
  onlyAvailable: false,
  sorting: ''
})

const query = computed(() => {
  const q: Record<string, string> = {}
  if (filters.onlyAvailable) q.status = 'available'
  if (filters.min_price !== undefined && filters.min_price !== null) q.min_price = String(filters.min_price)
  if (filters.max_price !== undefined && filters.max_price !== null) q.max_price = String(filters.max_price)

  if (filters.sorting === 'price_asc') {
    q.sort = 'price_per_hour'
    q.order = 'asc'
  }
  if (filters.sorting === 'price_desc') {
    q.sort = 'price_per_hour'
    q.order = 'desc'
  }
  if (filters.sorting === 'rating_desc') {
    q.sort = 'rating'
    q.order = 'desc'
  }
  return q
})

const { data, pending, error, refresh } = await useAsyncData<Car[]>(
  'cars-list',
  () => fetcher(`/api/v1/cars`, { query: query.value })
)

watch(query, () => refresh(), { deep: true })

const cars = computed(() => data.value ?? [])

const filteredCars = computed(() => {
  let list = [...cars.value]
  if (filters.search) {
    const term = filters.search.toLowerCase()
    list = list.filter((car) => `${car.mark} ${car.model}`.toLowerCase().includes(term))
  }
  if (filters.categories.length > 0) {
    list = list.filter((car) => filters.categories.includes(car.category))
  }
  return list
})

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const statusClass = (status: string) => {
  const normalized = status?.toLowerCase()
  return normalized ? `badge--${normalized}` : ''
}

const resetFilters = () => {
  filters.search = ''
  filters.categories = []
  filters.min_price = undefined
  filters.max_price = 200
  filters.onlyAvailable = false
  filters.sorting = ''
}
</script>

<style scoped>
.catalog {
  display: grid;
  grid-template-columns: minmax(220px, 280px) 1fr;
  gap: 20px;
}

.catalog__filters {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.catalog__list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card__image {
  height: 120px;
  border-radius: 12px;
  background: linear-gradient(135deg, #94a3b8, #e2e8f0);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 700;
  color: #111827;
}

@media (max-width: 900px) {
  .catalog {
    grid-template-columns: 1fr;
  }
}
</style>
