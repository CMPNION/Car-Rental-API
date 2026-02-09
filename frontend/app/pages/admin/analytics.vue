<template>
  <div class="analytics">
    <section class="panel hero">
      <div>
        <h1>Аналитика</h1>
        <p class="muted">Сводка по выручке, загрузке автопарка и активности клиентов.</p>
      </div>
      <div class="hero__kpi">
        <div class="kpi">
          <span>Выручка всего</span>
          <strong>{{ formatCurrency(metrics.total_revenue) }}</strong>
          <small>30 дней: {{ formatCurrency(metrics.revenue_last_30_days) }}</small>
        </div>
        <div class="kpi">
          <span>Аренд всего</span>
          <strong>{{ formatNumber(metrics.total_rentals) }}</strong>
          <small>Активных: {{ formatNumber(metrics.rentals_by_status.active) }}</small>
        </div>
      </div>
    </section>

    <div class="spacer"></div>

    <div v-if="pending" class="grid grid--kpi">
      <div v-for="i in 6" :key="i" class="skeleton"></div>
    </div>
    <div v-else-if="error" class="panel">Ошибка: {{ errorMessage }}</div>
    <div v-else>
      <section class="grid grid--kpi">
        <div class="card">
          <div class="card__title">Выручка</div>
          <div class="card__meta">{{ formatCurrency(metrics.total_revenue) }}</div>
          <div class="card__hint">7 дней: {{ formatCurrency(revenueLast7Total) }}</div>
        </div>
        <div class="card">
          <div class="card__title">Загрузка флота</div>
          <div class="card__meta">{{ formatPercent(metrics.fleet_load) }}</div>
          <div class="card__hint">Машин в парке: {{ formatNumber(metrics.total_cars) }}</div>
        </div>
        <div class="card">
          <div class="card__title">Пользователи</div>
          <div class="card__meta">{{ formatNumber(metrics.total_users) }}</div>
          <div class="card__hint">Средний рейтинг: {{ metrics.average_user_rating.toFixed(2) }}</div>
        </div>
        <div class="card">
          <div class="card__title">Авто</div>
          <div class="card__meta">{{ formatNumber(metrics.total_cars) }}</div>
          <div class="card__hint">Средний рейтинг: {{ metrics.average_car_rating.toFixed(2) }}</div>
        </div>
        <div class="card">
          <div class="card__title">Завершенные</div>
          <div class="card__meta">{{ formatNumber(metrics.rentals_by_status.completed) }}</div>
          <div class="card__hint">Отмененные: {{ formatNumber(metrics.rentals_by_status.cancelled) }}</div>
        </div>
        <div class="card">
          <div class="card__title">Ожидают оплаты</div>
          <div class="card__meta">{{ formatNumber(metrics.rentals_by_status.pending) }}</div>
          <div class="card__hint">Активные: {{ formatNumber(metrics.rentals_by_status.active) }}</div>
        </div>
      </section>

      <div class="spacer"></div>

      <section class="panel">
        <div class="section-head">
          <h2>Выручка по дням</h2>
          <p class="muted">Последние 7 дней</p>
        </div>
        <div class="chart">
          <div
            v-for="point in revenueSeries"
            :key="point.day"
            class="chart__bar"
            :style="{ '--h': point.percent + '%' }"
          >
            <span class="chart__value">{{ formatCompact(point.revenue) }}</span>
            <span class="chart__label">{{ point.label }}</span>
          </div>
        </div>
      </section>

      <div class="spacer"></div>

      <section class="grid grid--split">
        <div class="panel">
          <div class="section-head">
            <h2>Лидеры по арендам</h2>
            <p class="muted">Топ-5 машин</p>
          </div>
          <div v-if="topCars.length === 0" class="muted">Нет данных.</div>
          <ul v-else class="list">
            <li v-for="car in topCars" :key="car.car_id">
              <span>{{ car.mark }} {{ car.model }}</span>
              <strong>{{ formatNumber(car.rentals) }}</strong>
            </li>
          </ul>
        </div>
        <div class="panel">
          <div class="section-head">
            <h2>Лидеры по выручке</h2>
            <p class="muted">Топ-5 клиентов</p>
          </div>
          <div v-if="topUsers.length === 0" class="muted">Нет данных.</div>
          <ul v-else class="list">
            <li v-for="user in topUsers" :key="user.user_id">
              <span>
                {{ user.name }}
                <small>{{ user.email }}</small>
              </span>
              <strong>{{ formatCurrency(user.spend) }}</strong>
            </li>
          </ul>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import admin from '~/middleware/admin'

definePageMeta({
  middleware: [admin]
})

const { authFetch } = useApi()

type Metrics = {
  total_revenue: number
  revenue_last_30_days: number
  revenue_last_7_days: { day: string; revenue: number }[]
  total_users: number
  total_cars: number
  total_rentals: number
  rentals_by_status: {
    pending: number
    active: number
    completed: number
    cancelled: number
  }
  fleet_load: number
  average_car_rating: number
  average_user_rating: number
  top_cars_by_rentals: { car_id: number; mark: string; model: string; rentals: number }[]
  top_users_by_spend: { user_id: number; name: string; email: string; spend: number }[]
}

const { data, pending, error } = await useAsyncData<Metrics>(
  'admin-metrics',
  () => authFetch(`/api/v1/admin/metrics`)
)

const metrics = computed(() => data.value ?? {
  total_revenue: 0,
  revenue_last_30_days: 0,
  revenue_last_7_days: [],
  total_users: 0,
  total_cars: 0,
  total_rentals: 0,
  rentals_by_status: {
    pending: 0,
    active: 0,
    completed: 0,
    cancelled: 0
  },
  fleet_load: 0,
  average_car_rating: 0,
  average_user_rating: 0,
  top_cars_by_rentals: [],
  top_users_by_spend: []
})

const revenueDays = computed(() => Array.isArray(metrics.value.revenue_last_7_days)
  ? metrics.value.revenue_last_7_days
  : [])

const topCars = computed(() => Array.isArray(metrics.value.top_cars_by_rentals)
  ? metrics.value.top_cars_by_rentals
  : [])

const topUsers = computed(() => Array.isArray(metrics.value.top_users_by_spend)
  ? metrics.value.top_users_by_spend
  : [])

const revenueLast7Total = computed(() => revenueDays.value
  .reduce((sum, item) => sum + (item.revenue || 0), 0))

const revenueSeries = computed(() => {
  const list = revenueDays.value
  const max = Math.max(1, ...list.map((item) => item.revenue || 0))
  return list.map((item) => ({
    day: item.day,
    revenue: item.revenue || 0,
    percent: Math.round(((item.revenue || 0) / max) * 100),
    label: formatDay(item.day)
  }))
})

const formatCurrency = (value: number) => new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  maximumFractionDigits: 0
}).format(value || 0)

const formatNumber = (value: number) => new Intl.NumberFormat('ru-RU').format(value || 0)

const formatCompact = (value: number) => new Intl.NumberFormat('ru-RU', {
  notation: 'compact',
  maximumFractionDigits: 1
}).format(value || 0)

const formatPercent = (value: number) => `${(value || 0).toFixed(1)}%`

const formatDay = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })
}

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;600;700&family=IBM+Plex+Mono:wght@400;500&display=swap');

.analytics {
  font-family: 'Space Grotesk', ui-sans-serif, system-ui;
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: relative;
}

.analytics::before {
  content: '';
  position: absolute;
  inset: -120px -20px auto;
  height: 240px;
  background: radial-gradient(circle at 20% 20%, rgba(15, 118, 110, 0.18), transparent 55%),
    radial-gradient(circle at 80% 0%, rgba(59, 130, 246, 0.18), transparent 50%);
  pointer-events: none;
}

.hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  border-radius: 18px;
  position: relative;
  overflow: hidden;
  animation: fadeIn 0.6s ease;
}

.hero__kpi {
  display: grid;
  gap: 12px;
}

.kpi {
  background: linear-gradient(135deg, rgba(15, 118, 110, 0.15), rgba(37, 99, 235, 0.12));
  border: 1px solid rgba(15, 118, 110, 0.25);
  border-radius: 14px;
  padding: 12px 16px;
  min-width: 220px;
}

.kpi span {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #0f172a;
  opacity: 0.7;
}

.kpi strong {
  display: block;
  font-size: 22px;
  margin: 6px 0 4px;
}

.kpi small {
  color: #0f172a;
  opacity: 0.65;
}

.grid--kpi {
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.grid--split {
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 20px;
}

.card__hint {
  font-size: 12px;
  opacity: 0.7;
}

.section-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.chart {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(72px, 1fr));
  gap: 14px;
  align-items: end;
  min-height: 180px;
}

.chart__bar {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 10px 8px;
  border-radius: 14px;
  background: rgba(15, 118, 110, 0.08);
  position: relative;
}

.chart__bar::before {
  content: '';
  position: absolute;
  bottom: 36px;
  width: 100%;
  height: var(--h, 0%);
  background: linear-gradient(180deg, rgba(15, 118, 110, 0.8), rgba(37, 99, 235, 0.75));
  border-radius: 12px;
  transition: height 0.6s ease;
}

.chart__value {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 12px;
  z-index: 1;
}

.chart__label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  opacity: 0.6;
  z-index: 1;
}

.list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
}

.list li span {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.list li small {
  opacity: 0.6;
  font-size: 12px;
}

.skeleton {
  height: 110px;
  background: linear-gradient(90deg, rgba(15, 118, 110, 0.1), rgba(37, 99, 235, 0.12), rgba(15, 118, 110, 0.08));
  border-radius: 14px;
  animation: shimmer 1.2s infinite linear;
  background-size: 200% 100%;
}

@keyframes shimmer {
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 900px) {
  .hero {
    flex-direction: column;
  }

  .hero__kpi {
    width: 100%;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }
}
</style>
