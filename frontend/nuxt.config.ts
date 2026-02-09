// https://nuxt.com/docs/api/configuration/nuxt-config
const apiBase = process.env.NUXT_PUBLIC_API_BASE ?? 'http://localhost:4000'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: false },
  runtimeConfig: {
    public: {
      apiBase
    }
  },
  devServer: {
    host: '0.0.0.0'
  }
})
