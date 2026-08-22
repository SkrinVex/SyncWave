import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('syncwave_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('syncwave_user') || 'null'))
  const needsSetup = ref(false)
  const loading = ref(false)

  const isAuthenticated = computed(() => !!token.value)

  function authHeaders() {
    return token.value ? { 'Authorization': `Bearer ${token.value}` } : {}
  }

  async function checkStatus() {
    try {
      const res = await fetch('/api/v1/auth/status')
      const data = await res.json()
      needsSetup.value = data.needs_setup
      if (token.value && !needsSetup.value) {
        await fetchMe()
      }
    } catch (e) {
      console.error('Failed to check auth status:', e)
    }
  }

  async function setupAdmin(username, password) {
    loading.value = true
    try {
      const res = await fetch('/api/v1/auth/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.error || 'Failed to setup admin')
      }
      const data = await res.json()
      setSession(data.token, data.user)
      needsSetup.value = false
      return true
    } finally {
      loading.value = false
    }
  }

  async function login(username, password) {
    loading.value = true
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.error || 'Invalid credentials')
      }
      const data = await res.json()
      setSession(data.token, data.user)
      return true
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const res = await fetch('/api/v1/auth/me', {
        headers: authHeaders(),
      })
      if (res.status === 401) {
        logout()
        return
      }
      if (res.ok) {
        user.value = await res.json()
        localStorage.setItem('syncwave_user', JSON.stringify(user.value))
      }
    } catch (e) {
      console.error('Failed to fetch user:', e)
    }
  }

  function setSession(newToken, newUser) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem('syncwave_token', newToken)
    localStorage.setItem('syncwave_user', JSON.stringify(newUser))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('syncwave_token')
    localStorage.removeItem('syncwave_user')
  }

  return {
    token,
    user,
    needsSetup,
    loading,
    isAuthenticated,
    authHeaders,
    checkStatus,
    setupAdmin,
    login,
    logout,
    fetchMe,
  }
})

