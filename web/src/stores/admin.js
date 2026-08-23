import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'
import { useToastStore } from './toast'

export const useAdminStore = defineStore('admin', () => {
  const users = ref([])
  const loading = ref(false)
  const authStore = useAuthStore()
  const toast = useToastStore()

  async function fetchUsers() {
    loading.value = true
    try {
      const res = await fetch('/api/v1/admin/users', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        users.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch users:', e)
    } finally {
      loading.value = false
    }
  }

  async function updateUserQuota(userId, quotaBytes) {
    try {
      const res = await fetch(`/api/v1/admin/users/${userId}/quota`, {
        method: 'PUT',
        headers: {
          ...authStore.authHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ quota_bytes: quotaBytes }),
      })
      if (!res.ok) throw new Error('Failed to update quota')
      toast.success('Квота пользователя успешно обновлена')
      await fetchUsers()
    } catch (e) {
      toast.error(e.message || 'Ошибка обновления квоты')
    }
  }

  async function deleteUser(userId) {
    try {
      const res = await fetch(`/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: authStore.authHeaders(),
      })
      if (!res.ok) {
        const errText = await res.text()
        let errMsg = errText
        try {
          const json = JSON.parse(errText)
          if (json.error) errMsg = json.error
        } catch {}
        throw new Error(errMsg || 'Failed to delete user')
      }
      toast.success('Пользователь успешно удален')
      await fetchUsers()
    } catch (e) {
      toast.error(e.message || 'Ошибка удаления пользователя')
    }
  }

  async function setAllowRegistration(allow) {
    try {
      const res = await fetch('/api/v1/admin/registration', {
        method: 'POST',
        headers: {
          ...authStore.authHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ allow }),
      })
      if (!res.ok) throw new Error('Failed to update registration setting')
      toast.success(allow ? 'Регистрация открыта' : 'Регистрация закрыта')
    } catch (e) {
      toast.error(e.message || 'Ошибка изменения настройки регистрации')
    }
  }

  async function setGlobalLimit(limitBytes) {
    try {
      const res = await fetch('/api/v1/admin/global-limit', {
        method: 'POST',
        headers: {
          ...authStore.authHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ limit_bytes: limitBytes }),
      })
      if (!res.ok) throw new Error('Failed to update global limit')
      toast.success('Глобальный лимит сохранен')
    } catch (e) {
      toast.error(e.message || 'Ошибка сохранения глобального лимита')
    }
  }

  return {
    users,
    loading,
    fetchUsers,
    updateUserQuota,
    deleteUser,
    setAllowRegistration,
    setGlobalLimit,
  }
})
