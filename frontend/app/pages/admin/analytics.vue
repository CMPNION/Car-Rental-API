<template>
  <div class="panel">
    <h1>Аналитика</h1>
    <p class="muted">Ключевые показатели автопарка.</p>
  </div>

  <div class="spacer"></div>

  <div v-if="pending" class="grid">
    <div v-for="i in 3" :key="i" class="skeleton"></div>
  </div>
  <div v-else-if="error" class="panel">Ошибка: {{ errorMessage }}</div>
  <div v-else class="grid">
    <div class="card">
      <div class="card__title">Total Revenue</div>
      <div class="card__meta">{{ metrics.total_revenue }} ₽</div>
    </div>
    <div class="card">
      <div class="card__title">Active Rentals</div>
      <div class="card__meta">{{ metrics.active_rentals }}</div>
    </div>
    <div class="card">
      <div class="card__title">Fleet Load</div>
      <div class="card__meta">{{ metrics.fleet_load.toFixed(1) }}%</div>
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
  active_rentals: number
  fleet_load: number
  total_cars: number
}

const { data, pending, error } = await useAsyncData<Metrics>(
  'admin-metrics',
  () => authFetch(`/api/v1/admin/metrics`)
)

const metrics = computed(() => data.value ?? {
  total_revenue: 0,
  active_rentals: 0,
  fleet_load: 0,
  total_cars: 0
})

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})
</script>
