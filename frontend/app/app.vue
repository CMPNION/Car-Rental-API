<template>
  <div class="app">
    <NuxtRouteAnnouncer />
    <header class="topbar">
      <div class="container topbar__inner">
        <div class="brand">Car Rental</div>
        <nav class="nav">
          <NuxtLink to="/">Главная</NuxtLink>
          <NuxtLink to="/cars">Машины</NuxtLink>
          <NuxtLink v-if="isAdmin" to="/admin/cars">Админ</NuxtLink>
        </nav>
        <div class="nav__actions">
          <NuxtLink v-if="!isAuthed" to="/login" class="link-btn">Войти</NuxtLink>
          <button v-else class="secondary" @click="logout">Выйти</button>
        </div>
      </div>
    </header>
    <main class="container">
      <NuxtPage />
    </main>
  </div>
</template>

<script setup lang="ts">
const token = useCookie('token')
const { authFetch } = useApi()

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
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  background: #e5e7eb;
  color: #111827;
}

.spacer {
  height: 16px;
}
</style>
