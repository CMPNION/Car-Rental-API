<template>
  <div class="panel">
    <h1>Профиль</h1>
    <p class="muted">Управляйте данными и балансом.</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>Данные пользователя</h2>
    <form class="row" @submit.prevent="saveProfile">
      <label class="field">
        Имя
        <input v-model.trim="profileForm.first_name" />
      </label>
      <label class="field">
        Фамилия
        <input v-model.trim="profileForm.last_name" />
      </label>
      <label class="field">
        Email
        <input :value="profile.email" disabled />
      </label>
      <button type="submit">Сохранить</button>
    </form>
    <p v-if="profileMessage" class="muted">{{ profileMessage }}</p>
    <div class="card__meta">Ваш рейтинг: {{ profile.rating }}</div>
    <div class="card__meta">Скидка: {{ ratingDiscount }}</div>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>Кошелёк</h2>
    <div class="row" style="justify-content: space-between;">
      <div>
        <div class="muted">Текущий баланс</div>
        <div class="card__title">{{ balanceDisplay }}</div>
      </div>
      <form class="row" @submit.prevent="topUp">
        <label class="field">
          Пополнить
          <input v-model.number="topUpAmount" type="number" min="1" />
        </label>
        <button type="submit" :disabled="isToppingUp">Пополнить</button>
      </form>
    </div>
    <p v-if="walletError" class="muted">{{ walletError }}</p>
  </div>

  <div class="spacer"></div>

  <div class="panel">
    <h2>Транзакции</h2>
    <div v-if="pending">Загрузка...</div>
    <div v-else-if="error">Ошибка: {{ errorMessage }}</div>
    <div v-else-if="transactions.length === 0" class="muted">Транзакций пока нет.</div>
    <div v-else class="grid">
      <div v-for="tx in transactions" :key="tx.id" class="card">
        <div class="card__title">{{ tx.type === 'topup' ? 'Пополнение' : 'Оплата аренды' }}</div>
        <div class="card__meta">Сумма: {{ tx.amount }}</div>
        <div class="card__meta">Статус: {{ tx.status }}</div>
        <div class="card__meta">Дата: {{ formatDate(tx.created_at) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { authFetch } = useApi()
const { push } = useToast()

type Profile = {
  first_name: string
  last_name: string
  email: string
  rating: number
  balance: number
}

type Transaction = {
  id: number
  type: string
  amount: number
  status: string
  created_at: string
}

const profile = reactive<Profile>({
  first_name: '',
  last_name: '',
  email: '',
  rating: 0,
  balance: 0
})

const profileForm = reactive({
  first_name: '',
  last_name: ''
})

const profileMessage = ref('')

const ratingDiscount = computed(() =>
  profile.rating > 4.5 ? '10%' : profile.rating > 0 && profile.rating < 2 ? '-20%' : '0%'
)

const loadProfile = async () => {
  const data = await authFetch<Profile>(`/api/v1/users/me`)
  Object.assign(profile, data)
  profileForm.first_name = data.first_name
  profileForm.last_name = data.last_name
}

const saveProfile = async () => {
  profileMessage.value = ''
  try {
    await authFetch(`/api/v1/users/me`, {
      method: 'PATCH',
      body: {
        first_name: profileForm.first_name,
        last_name: profileForm.last_name
      }
    })
    await loadProfile()
    push('Профиль обновлён', 'success')
  } catch (err: any) {
    profileMessage.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось обновить профиль'
  }
}

const topUpAmount = ref<number | null>(null)
const isToppingUp = ref(false)
const walletError = ref('')

const balanceDisplay = computed(() => `${profile.balance} ₽`)

const topUp = async () => {
  walletError.value = ''
  if (!topUpAmount.value || topUpAmount.value <= 0) {
    walletError.value = 'Введите сумму пополнения'
    return
  }

  isToppingUp.value = true
  try {
    const response = await authFetch<{ balance: number }>(`/api/v1/users/balance`, {
      method: 'PATCH',
      body: { amount: topUpAmount.value }
    })
    profile.balance = response.balance
    topUpAmount.value = null
    push('Баланс пополнен', 'success')
  } catch (err: any) {
    walletError.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Не удалось пополнить'
  } finally {
    isToppingUp.value = false
  }
}

const { data, pending, error, refresh } = await useAsyncData<Transaction[]>(
  'profile-transactions',
  () => authFetch(`/api/v1/transactions`)
)

const transactions = computed(() => data.value ?? [])

const errorMessage = computed(() => {
  if (!error.value) return ''
  return (error.value as any)?.data?.message ?? (error.value as any)?.data?.error ?? (error.value as any)?.message ?? 'Ошибка запроса'
})

const formatDate = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('ru-RU')
}

onMounted(async () => {
  await loadProfile()
  await refresh()
})
</script>
