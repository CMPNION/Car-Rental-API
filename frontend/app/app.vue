<template>
  <div class="app">
    <NuxtRouteAnnouncer />
    <header class="topbar">
      <div class="container topbar__inner">
        <div class="brand">Car Rental</div>
        <nav class="nav">
          <NuxtLink to="/">Главная</NuxtLink>
          <NuxtLink to="/cars">Машины</NuxtLink>
          <NuxtLink v-if="isAuthed" to="/rentals">Мои аренды</NuxtLink>
          <NuxtLink v-if="isAuthed" to="/profile">Профиль</NuxtLink>
          <NuxtLink v-if="isAdmin" to="/admin/cars">Админ</NuxtLink>
          <NuxtLink v-if="isAdmin" to="/admin/rentals">Аренды</NuxtLink>
          <NuxtLink v-if="isAdmin" to="/admin/analytics">Аналитика</NuxtLink>
        </nav>
        <form class="nav__search" @submit.prevent="submitSearch">
          <input v-model.trim="globalSearch" placeholder="Поиск авто..." />
        </form>
        <div class="nav__actions">
          <button class="secondary" @click="toggleTheme">
            {{ theme === 'dark' ? 'Светлая' : 'Тёмная' }}
          </button>
          <NuxtLink v-if="!isAuthed" to="/login" class="link-btn">Войти</NuxtLink>
          <button v-else class="secondary" @click="logout">Выйти</button>
        </div>
      </div>
    </header>
    <main class="container">
      <NuxtPage />
    </main>

    <div class="toast-container">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="`toast--${toast.type || 'info'}`"
        @click="removeToast(toast.id)"
      >
        {{ toast.message }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const token = useCookie('token')
const { authFetch } = useApi()
const { toasts, remove: removeToast } = useToast()

const globalSearch = ref('')
const themeCookie = useCookie('theme', { default: () => 'light' })
const theme = ref(themeCookie.value)

const isAuthed = computed(() => Boolean(token.value))
const isAdmin = ref(false)

const refreshAdmin = async () => {
  if (!token.value) {
    isAdmin.value = false
    return
  }

  try {
    const data = await authFetch<{ role: string; is_admin: boolean }>(`/auth/me`)
    isAdmin.value = Boolean(data?.is_admin)
  } catch {
    isAdmin.value = false
  }
}

onMounted(() => {
  refreshAdmin()
  applyTheme()
})

watch(
  () => token.value,
  () => refreshAdmin()
)

const logout = () => {
  token.value = null
  isAdmin.value = false
  navigateTo('/login')
}

const submitSearch = () => {
  const query = globalSearch.value
  if (!query) {
    navigateTo('/cars')
    return
  }
  navigateTo(`/cars?search=${encodeURIComponent(query)}`)
}

const applyTheme = () => {
  if (process.client) {
    document.body.classList.toggle('theme-dark', theme.value === 'dark')
  }
}

const toggleTheme = () => {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  themeCookie.value = theme.value
  applyTheme()
}
</script>

<style>
:root {
  color-scheme: light;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial,
    "Apple Color Emoji", "Segoe UI Emoji";
  background: #f6f7fb;
  color: #1f2937;
}

body.theme-dark {
  background: #0f172a;
  color: #e2e8f0;
}

a {
  color: inherit;
  text-decoration: none;
}

.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 24px;
}

.topbar {
  background: #111827;
  color: #f9fafb;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.08);
}

body.theme-dark .topbar {
  background: #0b1120;
}

.topbar__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
}

.brand {
  font-weight: 700;
  letter-spacing: 0.3px;
}

.nav {
  display: flex;
  gap: 16px;
  font-size: 14px;
}

.nav a {
  opacity: 0.9;
}

.nav a.router-link-active {
  opacity: 1;
  font-weight: 600;
}

.nav__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav__search input {
  border-radius: 999px;
  padding: 8px 12px;
  min-width: 220px;
}

.link-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 10px;
  background: #ffffff;
  color: #111827;
  font-weight: 600;
}

.panel {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.04);
}

body.theme-dark .panel,
body.theme-dark .card {
  background: #111827;
  border-color: #1f2937;
  box-shadow: none;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.04);
}

.card__title {
  font-weight: 600;
  font-size: 16px;
}

.card__meta {
  font-size: 13px;
  color: #6b7280;
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
}

input,
select,
textarea {
  border: 1px solid #d1d5db;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 14px;
  background: #fff;
}

body.theme-dark input,
body.theme-dark select,
body.theme-dark textarea {
  background: #0f172a;
  color: #e2e8f0;
  border-color: #334155;
}

button {
  border: none;
  border-radius: 10px;
  padding: 10px 14px;
  font-weight: 600;
  background: #111827;
  color: #fff;
  cursor: pointer;
}

button.secondary {
  background: #e5e7eb;
  color: #111827;
}

button.danger {
  background: #dc2626;
}

.muted {
  color: #6b7280;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
  background: #e5e7eb;
  color: #111827;
}

.badge--pending {
  background: #fef9c3;
  color: #92400e;
}

.badge--active {
  background: #dcfce7;
  color: #166534;
}

.badge--cancelled {
  background: #fee2e2;
  color: #991b1b;
}

.badge--completed {
  background: #e0e7ff;
  color: #3730a3;
}

.badge--available {
  background: #ecfeff;
  color: #0e7490;
}

.badge--booked {
  background: #fef3c7;
  color: #92400e;
}

.badge--maintenance {
  background: #f1f5f9;
  color: #475569;
}

.skeleton {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton 1.4s ease infinite;
  border-radius: 10px;
  min-height: 140px;
}

.skeleton-line {
  height: 12px;
  width: 100%;
  margin-bottom: 8px;
}

@keyframes skeleton {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.toast-container {
  position: fixed;
  right: 20px;
  bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 1000;
}

.toast {
  padding: 12px 16px;
  border-radius: 12px;
  background: #111827;
  color: #f9fafb;
  font-size: 14px;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.2);
  cursor: pointer;
}

.toast--success { background: #16a34a; }
.toast--error { background: #dc2626; }
.toast--info { background: #2563eb; }

.spacer {
  height: 16px;
}
</style>
