<template>
  <div class="panel">
    <h1>Добро пожаловать</h1>
    <p class="muted">
      Здесь можно посмотреть машины, открыть страницу машины или управлять ими в админ-панели.
    </p>
    <div class="row">
      <NuxtLink to="/cars">
        <button>Список машин</button>
      </NuxtLink>
      <NuxtLink to="/admin/cars">
        <button class="secondary">Админ: машины</button>
      </NuxtLink>
      <button class="secondary" @click="logout">Выйти</button>
    </div>
  </div>
</template>

<script setup lang="ts">
const token = useCookie('token')

const logout = () => {
  token.value = null
  navigateTo('/login')
}
</script>
