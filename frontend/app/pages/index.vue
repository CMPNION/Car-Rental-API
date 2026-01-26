<template>
  <div>
    <h1>Protected</h1>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else>{{ data }}</div>
    <button @click="logout">Logout</button>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'

const token = useCookie('token')
const config = useRuntimeConfig()
const apiBase = config.public.apiBase as string

const { data, error, pending } = await useAsyncData('hello', async () => {
  const response = await axios.get<string>(`${apiBase}/hello`, {
    headers: {
      Authorization: `Bearer ${token.value}`
    }
  })
  return response.data
})

const logout = () => {
  token.value = null
  navigateTo('/login')
}
</script>
