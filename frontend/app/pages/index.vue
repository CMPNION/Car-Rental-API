<template>
  <section class="hero">
    <div class="hero__content">
      <h1>Аренда авто без лишних шагов</h1>
      <p class="muted">
        Подберите автомобиль за минуту, забронируйте онлайн и управляйте поездками в личном кабинете.
      </p>
      <form class="hero__form" @submit.prevent="submitHero">
        <label class="field">
          Дата начала
          <input v-model="heroForm.start" type="datetime-local" required />
        </label>
        <label class="field">
          Дата окончания
          <input v-model="heroForm.end" type="datetime-local" required />
        </label>
        <label class="field">
          Категория
          <select v-model="heroForm.category">
            <option value="">Любая</option>
            <option value="economy">Economy</option>
            <option value="business">Business</option>
            <option value="luxury">Luxury</option>
          </select>
        </label>
        <button type="submit">Найти авто</button>
      </form>
    </div>
    <div class="hero__image"></div>
  </section>

  <div class="spacer"></div>

  <section class="panel">
    <h2>Как это работает</h2>
    <div class="grid">
      <div class="card">
        <div class="card__title">Выбери авто</div>
        <div class="card__meta">Фильтры и каталог помогут быстро найти нужный вариант.</div>
      </div>
      <div class="card">
        <div class="card__title">Забронируй</div>
        <div class="card__meta">Онлайн бронирование с защитой от двойной аренды.</div>
      </div>
      <div class="card">
        <div class="card__title">Наслаждайся поездкой</div>
        <div class="card__meta">Оплата, статусы и завершение аренды — в личном кабинете.</div>
      </div>
    </div>
  </section>

  <div class="spacer"></div>

  <section class="panel">
    <div class="row" style="justify-content: space-between;">
      <div>
        <h2>Топ‑предложения</h2>
        <p class="muted">Самые высокие рейтинги за неделю</p>
      </div>
      <NuxtLink to="/cars">
        <button class="secondary">В каталог</button>
      </NuxtLink>
    </div>

    <div v-if="pending" class="row">
      <div v-for="i in 4" :key="i" class="skeleton" style="min-width: 240px; height: 160px;"></div>
    </div>
    <div v-else class="slider">
      <div v-for="car in topCars" :key="getCarId(car)" class="card slider__card">
        <div class="card__title">{{ car.mark }} {{ car.model }}</div>
        <div class="card__meta">Рейтинг: {{ car.rating }}</div>
        <div class="card__meta">Цена/час: {{ car.price_per_hour }}</div>
        <NuxtLink :to="`/cars/${getCarId(car)}`">
          <button class="secondary">Подробнее</button>
        </NuxtLink>
      </div>
    </div>
  </section>

  <div class="spacer"></div>

  <footer class="panel footer">
    <div>
      <div class="card__title">Car Rental</div>
      <div class="card__meta">Поддержка: support@carrental.kz</div>
    </div>
    <div>
      <div class="card__meta">Соцсети</div>
      <div class="row">
        <a href="#">Instagram</a>
        <a href="#">Telegram</a>
        <a href="#">TikTok</a>
      </div>
    </div>
    <div>
      <div class="card__meta">Документы</div>
      <div class="row">
        <a href="#">Условия аренды</a>
        <a href="#">Политика</a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
const { fetcher } = useApi()

type Car = {
  id?: number
  ID?: number
  mark: string
  model: string
  rating: number
  price_per_hour: number
}

const getCarId = (car: Car) => car.id ?? car.ID ?? 0

const heroForm = reactive({
  start: '',
  end: '',
  category: ''
})

const submitHero = () => {
  const params = new URLSearchParams()
  if (heroForm.category) params.set('category', heroForm.category)
  if (heroForm.start) params.set('start', heroForm.start)
  if (heroForm.end) params.set('end', heroForm.end)
  navigateTo(`/cars?${params.toString()}`)
}

const { data, pending } = await useAsyncData<Car[]>(
  'top-cars',
  () => fetcher(`/api/v1/cars`, { query: { sort: 'rating', order: 'desc', limit: 5 } })
)

const topCars = computed(() => data.value ?? [])
</script>

<style scoped>
.hero {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
  align-items: center;
  background: linear-gradient(135deg, #111827, #1f2937);
  color: #f9fafb;
  padding: 32px;
  border-radius: 18px;
}

.hero__image {
  background: url('https://images.unsplash.com/photo-1503376780353-7e6692767b70?q=80&w=1200&auto=format&fit=crop') center/cover;
  min-height: 240px;
  border-radius: 16px;
}

.hero__form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-top: 20px;
}

.slider {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 8px;
}

.slider__card {
  min-width: 240px;
}

.footer {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 20px;
}
</style>
