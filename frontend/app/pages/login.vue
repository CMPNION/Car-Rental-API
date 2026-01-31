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
const token = useCookie('token')
const { fetcher } = useApi()

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
    const response = await fetcher<{ token: string }>(`/auth/login`, {
      method: 'POST',
      body: loginForm
    })
    token.value = response.token
    navigateTo('/')
  } catch (err: any) {
    error.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Login failed'
  }
}

const register = async () => {
  error.value = ''
  try {
    const response = await fetcher<{ token: string }>(`/auth/register`, {
      method: 'POST',
      body: registerForm
    })
    token.value = response.token
    navigateTo('/')
  } catch (err: any) {
    error.value = err?.data?.message ?? err?.data?.error ?? err?.message ?? 'Registration failed'
  }
}
</script>
