<template>
  <div class="panel">
    <h1>Администрирование машин</h1>
    <p class="muted">Создание, обновление и удаление машин.</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>Создать машину</h2>
    <form class="row" @submit.prevent="createCar">
      <label class="field">
        Марка
        <input v-model.trim="createForm.mark" required />
      </label>
      <label class="field">
        Модель
        <input v-model.trim="createForm.model" required />
      </label>
      <label class="field">
        Категория
        <select v-model="createForm.category" required>
          <option value="economy">Economy</option>
          <option value="business">Business</option>
          <option value="luxury">Luxury</option>
        </select>
      </label>
      <label class="field">
        Статус
        <select v-model="createForm.status">
          <option value="available">Available</option>
          <option value="booked">Booked</option>
          <option value="maintenance">Maintenance</option>
        </select>
      </label>
      <label class="field">
        Цена/час
        <input v-model.number="createForm.price_per_hour" type="number" min="1" required />
      </label>
      <label class="field" style="flex: 1; min-width: 240px;">
        Описание
        <input v-model.trim="createForm.metadata" placeholder="Краткое описание" />
      </label>
      <button type="submit">Создать</button>
    </form>
    <p v-if="createError" class="muted">{{ createError }}</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>Список машин</h2>
    <div v-if="pending">Загрузка...</div>
    <div v-else-if="error">Ошибка: {{ errorMessage }}</div>
    <div v-else class="grid">
      <div v-for="car in cars" :key="getCarId(car)" class="card">
        <div class="card__title">{{ car.mark }} {{ car.model }}</div>
        <div class="card__meta">ID: {{ getCarId(car) }}</div>
        <div class="card__meta">Категория: {{ car.category }}</div>
        <div class="card__meta">
          Статус:
          <span class="badge" :class="statusClass(car.status)">{{ car.status }}</span>
        </div>
        <div class="card__meta">Цена/час: {{ car.price_per_hour }}</div>
        <div class="row">
          <label class="field" style="min-width: 160px;">
            Статус
            <select v-model="statusUpdates[getCarId(car)]">
              <option value="available">Available</option>
              <option value="booked">Booked</option>
              <option value="maintenance">Maintenance</option>
            </select>
          </label>
          <button
            class="secondary"
            @click="updateStatus(getCarId(car))"
            :disabled="statusUpdates[getCarId(car)] === car.status"
          >
            Обновить статус
          </button>
        </div>
        <div class="row">
          <button class="secondary" @click="startEdit(car)">Редактировать</button>
          <button class="danger" @click="deleteCar(getCarId(car))">Удалить</button>
        </div>
      </div>
    </div>
  </div>

  <div class="spacer"></div>

  <div v-if="editing" class="panel">
    <h2>Редактирование машины #{{ editForm.id }}</h2>
    <form class="row" @submit.prevent="updateCar">
      <label class="field">
        Марка
        <input v-model.trim="editForm.mark" />
      </label>
      <label class="field">
        Модель
        <input v-model.trim="editForm.model" />
      </label>
      <label class="field">
        Категория
        <select v-model="editForm.category">
          <option value="">Не менять</option>
          <option value="economy">Economy</option>
          <option value="business">Business</option>
          <option value="luxury">Luxury</option>
        </select>
      </label>
      <label class="field">
        Статус
        <select v-model="editForm.status">
          <option value="">Не менять</option>
          <option value="available">Available</option>
          <option value="booked">Booked</option>
          <option value="maintenance">Maintenance</option>
        </select>
      </label>
      <label class="field">
        Цена/час
        <input v-model.number="editForm.price_per_hour" type="number" min="1" />
      </label>
      <label class="field" style="flex: 1; min-width: 240px;">
        Описание
        <input v-model.trim="editForm.metadata" />
      </label>
      <button type="submit">Сохранить</button>
      <button type="button" class="secondary" @click="cancelEdit">Отмена</button>
    </form>
    <p v-if="editError" class="muted">{{ editError }}</p>
  </div>
</template>

<script setup lang="ts">
import admin from '~/middleware/admin'

definePageMeta({
  middleware: [admin]
})

const { fetcher, authFetch } = useApi()

type Car = {
  id?: number
  ID?: number
  mark: string
  model: string
  category: string
  status: string
  price_per_hour: number
  metadata: string
}

const getCarId = (car: Car) => car.id ?? car.ID ?? 0

const { data, pending, error, refresh } = await useAsyncData<Car[]>(
  'admin-cars',
  () => fetcher(`/api/v1/cars`)
)

const cars = computed(() => data.value ?? [])

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const statusClass = (status: string) => {
  const normalized = status?.toLowerCase()
  return normalized ? `badge--${normalized}` : ''
}

const createForm = reactive({
  mark: '',
  model: '',
  category: 'economy',
  status: 'available',
  price_per_hour: 1,
  metadata: ''
})

const createError = ref('')

const statusUpdates = reactive<Record<number, string>>({})

watch(
  cars,
  (list) => {
    list.forEach((car) => {
      const id = getCarId(car)
      if (!statusUpdates[id]) {
        statusUpdates[id] = car.status
      }
    })
  },
  { immediate: true }
)

const createCar = async () => {
  createError.value = ''
  try {
    await authFetch(`${'/api/v1/cars'}`, {
      method: 'POST',
      body: {
        mark: createForm.mark,
        model: createForm.model,
        category: createForm.category,
        status: createForm.status,
        price_per_hour: createForm.price_per_hour,
        metadata: createForm.metadata
      }
    })
    await refresh()
    createForm.mark = ''
    createForm.model = ''
    createForm.category = 'economy'
    createForm.status = 'available'
    createForm.price_per_hour = 1
    createForm.metadata = ''
  } catch (err: any) {
    createError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось создать машину'
  }
}

const editing = ref(false)
const editError = ref('')

const editForm = reactive({
  id: 0,
  mark: '',
  model: '',
  category: '',
  status: '',
  price_per_hour: undefined as number | undefined,
  metadata: ''
})

const startEdit = (car: Car) => {
  editing.value = true
  editError.value = ''
  editForm.id = getCarId(car)
  editForm.mark = car.mark
  editForm.model = car.model
  editForm.category = car.category
  editForm.status = car.status
  editForm.price_per_hour = car.price_per_hour
  editForm.metadata = car.metadata
}

const cancelEdit = () => {
  editing.value = false
}

const updateCar = async () => {
  editError.value = ''
  const payload: Record<string, any> = {}

  if (editForm.mark) payload.mark = editForm.mark
  if (editForm.model) payload.model = editForm.model
  if (editForm.category) payload.category = editForm.category
  if (editForm.status) payload.status = editForm.status
  if (editForm.price_per_hour !== undefined) payload.price_per_hour = editForm.price_per_hour
  if (editForm.metadata) payload.metadata = editForm.metadata

  try {
    await authFetch(`/api/v1/cars/${editForm.id}`, {
      method: 'PUT',
      body: payload
    })
    await refresh()
    editing.value = false
  } catch (err: any) {
    editError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось обновить машину'
  }
}

const deleteCar = async (id: number) => {
  try {
    await authFetch(`/api/v1/cars/${id}`, {
      method: 'DELETE',
    })
    await refresh()
  } catch (err: any) {
    editError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось удалить машину'
  }
}

const updateStatus = async (id: number) => {
  editError.value = ''
  try {
    await authFetch(`/api/v1/cars/${id}`, {
      method: 'PUT',
      body: { status: statusUpdates[id] }
    })
    await refresh()
  } catch (err: any) {
    editError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось обновить статус'
  }
}
</script>
