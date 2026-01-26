<template>
  <div>
    <h1>Auth</h1>

    <h2>Login</h2>
    <form @submit.prevent="login">
      <div>
        <label>Email</label>
        <input v-model="loginForm.email" type="email" />
      </div>
      <div>
        <label>Password</label>
        <input v-model="loginForm.password" type="password" />
      </div>
      <button type="submit">Login</button>
    </form>

    <h2>Register</h2>
    <form @submit.prevent="register">
      <div>
        <label>First name</label>
        <input v-model="registerForm.first_name" type="text" />
      </div>
      <div>
        <label>Last name</label>
        <input v-model="registerForm.last_name" type="text" />
      </div>
      <div>
        <label>Email</label>
        <input v-model="registerForm.email" type="email" />
      </div>
      <div>
        <label>Password</label>
        <input v-model="registerForm.password" type="password" />
      </div>
      <button type="submit">Register</button>
    </form>

    <div v-if="error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'

const token = useCookie('token')
const config = useRuntimeConfig()
const apiBase = config.public.apiBase as string

const error = ref('')

const loginForm = reactive({
  email: '',
  password: ''
})

const registerForm = reactive({
  email: '',
  password: '',
  first_name: '',
  last_name: ''
})

const login = async () => {
  error.value = ''
  try {
    const response = await axios.post<{ token: string }>(`${apiBase}/auth/login`, loginForm)
    token.value = response.data.token
    navigateTo('/')
  } catch (err: any) {
    error.value = err?.response?.data?.error ?? 'Login failed'
  }
}

const register = async () => {
  error.value = ''
  try {
    const response = await axios.post<{ token: string }>(`${apiBase}/auth/register`, registerForm)
    token.value = response.data.token
    navigateTo('/')
  } catch (err: any) {
    error.value = err?.response?.data?.error ?? 'Registration failed'
  }
}
</script>
