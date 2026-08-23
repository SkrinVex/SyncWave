import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'
import { useTracksStore } from './tracks'
import { usePlaylistsStore } from './playlists'
import { useSettingsStore } from './settings'

export const useSyncStore = defineStore('sync', () => {
  const authStore = useAuthStore()
  const tracksStore = useTracksStore()
  const playlistsStore = usePlaylistsStore()
  const settingsStore = useSettingsStore()

  const progress = ref({
    active: false,
    playlist_id: '',
    playlist_title: '',
    current_track_index: 0,
    total_tracks: 0,
    current_track_title: '',
    current_track_id: '',
    track_percentage: 0,
    percentage: 0,
    speed: '',
    eta: '',
    status_text: 'Idle',
  })

  const logs = ref([])
  let eventSource = null

  function connectSSE() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }

    const t = authStore.token
    if (!t) return

    eventSource = new EventSource(`/api/v1/sync/events?token=${encodeURIComponent(t)}`)

    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data)
        if (payload.type === 'progress') {
          progress.value = { ...progress.value, ...payload.data }
          if (!payload.data.active) {
            tracksStore.fetchTracks()
            tracksStore.fetchStats()
            playlistsStore.fetchPlaylists()
          }
        } else if (payload.type === 'log') {
          logs.value.unshift(payload.data)
          if (logs.value.length > 300) {
            logs.value.pop()
          }
        } else if (payload.type === 'track_updated') {
          tracksStore.updateTrack(payload.data)
          tracksStore.fetchStats()
        } else if (payload.type === 'playlist_updated') {
          playlistsStore.fetchPlaylists()
        } else if (payload.type === 'cookie_status') {
          if (payload.data) {
            settingsStore.settings.cookies_status = payload.data.status || 'expired'
            settingsStore.settings.cookies_valid = payload.data.is_valid || false
            settingsStore.settings.cookies_error = payload.data.error_reason || ''
            if (!payload.data.is_valid || payload.data.status === 'expired') {
              settingsStore.showCookieModal = true
            }
          }
        }
      } catch (e) {
        console.error('Failed to parse SSE event:', e)
      }
    }

    eventSource.onerror = () => {
      setTimeout(() => {
        if (authStore.isAuthenticated) {
          connectSSE()
        }
      }, 3000)
    }
  }

  function disconnectSSE() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  async function fetchLogs() {
    try {
      const res = await fetch('/api/v1/sync/logs?limit=150', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        logs.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch sync logs:', e)
    }
  }

  async function fetchInitialProgress() {
    try {
      const res = await fetch('/api/v1/sync/progress', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        progress.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch sync progress:', e)
    }
  }

  async function triggerSyncAll() {
    progress.value.active = true
    progress.value.status_text = 'Подготовка к синхронизации...'
    const res = await fetch('/api/v1/sync/trigger', {
      method: 'POST',
      headers: authStore.authHeaders(),
    })
    return res.ok
  }

  async function clearLogs() {
    const res = await fetch('/api/v1/sync/logs', {
      method: 'DELETE',
      headers: authStore.authHeaders(),
    })
    if (res.ok) {
      logs.value = []
    }
  }

  async function cancelSync() {
    try {
      const res = await fetch('/api/v1/sync/cancel', {
        method: 'POST',
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        progress.value.active = false
        progress.value.status_text = 'Отменено'
      }
      return res.ok
    } catch (e) {
      console.error('Failed to cancel sync:', e)
    }
  }

  return {
    progress,
    logs,
    connectSSE,
    disconnectSSE,
    fetchLogs,
    fetchInitialProgress,
    triggerSyncAll,
    cancelSync,
    clearLogs,
  }
})
