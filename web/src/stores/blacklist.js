import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'
import { useToastStore } from './toast'

export const useBlacklistStore = defineStore('blacklist', () => {
  const items = ref([])
  const isLoading = ref(false)
  const authStore = useAuthStore()
  const toast = useToastStore()

  async function fetchBlacklist(query = '') {
    isLoading.value = true
    try {
      const res = await fetch(`/api/v1/blacklist?q=${encodeURIComponent(query)}`, {
        headers: authStore.authHeaders()
      })
      if (res.ok) {
        items.value = await res.json() || []
      } else {
        throw new Error('Failed to fetch')
      }
    } catch (e) {
      console.error('Failed to fetch blacklist:', e)
      toast.error('Не удалось загрузить чёрный список')
    } finally {
      isLoading.value = false
    }
  }

  async function removeFromBlacklist(youtubeId) {
    try {
      const res = await fetch(`/api/v1/blacklist/${encodeURIComponent(youtubeId)}`, {
        method: 'DELETE',
        headers: authStore.authHeaders()
      })
      if (!res.ok) throw new Error('Failed to delete')
      items.value = items.value.filter(i => i.youtube_id !== youtubeId)
      toast.success('Трек удален из чёрного списка')
    } catch (e) {
      console.error('Failed to remove from blacklist:', e)
      toast.error('Ошибка при удалении')
    }
  }

  return {
    items,
    isLoading,
    fetchBlacklist,
    removeFromBlacklist
  }
})
