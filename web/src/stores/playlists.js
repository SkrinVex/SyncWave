import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'
import { useSyncStore } from './sync'

export const usePlaylistsStore = defineStore('playlists', () => {
  const authStore = useAuthStore()

  const playlists = ref([])
  const loading = ref(false)

  async function fetchPlaylists() {
    loading.value = true
    try {
      const res = await fetch('/api/v1/playlists', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        playlists.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch playlists:', e)
    } finally {
      loading.value = false
    }
  }

  async function createPlaylist(title, urlOrId, autoSync = true, syncInterval = 60) {
    const res = await fetch('/api/v1/playlists', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...authStore.authHeaders(),
      },
      body: JSON.stringify({
        title,
        url_or_id: urlOrId,
        auto_sync: autoSync,
        sync_interval_minutes: syncInterval,
      }),
    })

    if (!res.ok) {
      const errText = await res.text()
      let errMsg = errText
      try {
        const json = JSON.parse(errText)
        if (json.error) errMsg = json.error
      } catch {}
      throw new Error(errMsg || 'Не удалось добавить плейлист')
    }

    const newPl = await res.json()
    playlists.value.unshift(newPl)
    return newPl
  }

  async function deletePlaylist(id) {
    const res = await fetch(`/api/v1/playlists/${id}`, {
      method: 'DELETE',
      headers: authStore.authHeaders(),
    })
    if (res.ok) {
      playlists.value = playlists.value.filter(p => p.id !== id)
      return true
    }
    return false
  }

  async function triggerSync(id) {
    const syncStore = useSyncStore()
    syncStore.progress.active = true
    syncStore.progress.status_text = 'Запуск синхронизации...'
    const res = await fetch(`/api/v1/playlists/${id}/sync`, {
      method: 'POST',
      headers: authStore.authHeaders(),
    })
    return res.ok
  }

  return {
    playlists,
    loading,
    fetchPlaylists,
    createPlaylist,
    deletePlaylist,
    triggerSync,
  }
})
