import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'

export const useTracksStore = defineStore('tracks', () => {
  const authStore = useAuthStore()

  const tracks = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(50)
  const totalPages = ref(1)
  const loading = ref(false)

  const searchQuery = ref('')
  const selectedPlaylist = ref('')
  const selectedStatus = ref('')
  const sortBy = ref('created_at')
  const sortOrder = ref('desc')

  const stats = ref({
    total_tracks: 0,
    ready_tracks: 0,
    failed_tracks: 0,
    total_storage_size: 0,
    total_duration: 0,
  })

  function getTrackStreamUrl(track) {
    if (!track) return ''
    const t = authStore.token
    return `/api/v1/tracks/${track.id}/stream${t ? `?token=${encodeURIComponent(t)}` : ''}`
  }

  function getTrackCoverUrl(track) {
    if (!track) return ''
    const t = authStore.token
    const tokenParam = t ? `token=${encodeURIComponent(t)}` : ''
    const timestampParam = track.updated_at ? `t=${new Date(track.updated_at).getTime()}` : ''
    const qs = [tokenParam, timestampParam].filter(Boolean).join('&')
    return `/api/v1/tracks/${track.id}/cover${qs ? '?' + qs : ''}`
  }

  function getTrackDownloadUrl(track) {
    if (!track) return ''
    const t = authStore.token
    return `/api/v1/tracks/${track.id}/download${t ? `?token=${encodeURIComponent(t)}` : ''}`
  }

  function updateTrack(updatedTrack) {
    const idx = tracks.value.findIndex(t => t.id === updatedTrack.id)
    if (idx !== -1) {
      tracks.value[idx] = updatedTrack
    } else {
      tracks.value.unshift(updatedTrack)
    }
  }

  async function fetchTracks(resetPage = false) {
    if (resetPage) page.value = 1
    loading.value = true
    try {
      const params = new URLSearchParams({
        page: page.value,
        page_size: pageSize.value,
        sort_by: sortBy.value,
        order: sortOrder.value,
      })

      if (searchQuery.value) params.set('q', searchQuery.value)
      if (selectedPlaylist.value) params.set('playlist_id', selectedPlaylist.value)
      if (selectedStatus.value) params.set('status', selectedStatus.value)

      const res = await fetch(`/api/v1/tracks?${params.toString()}`, {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        tracks.value = data.tracks || []
        total.value = data.total || 0
        page.value = data.page || 1
        totalPages.value = data.total_pages || 1
      }
    } catch (e) {
      console.error('Failed to fetch tracks:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const res = await fetch('/api/v1/tracks/stats', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        stats.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch stats:', e)
    }
  }

  async function deleteTrack(id) {
    try {
      const res = await fetch(`/api/v1/tracks/${id}`, {
        method: 'DELETE',
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        tracks.value = tracks.value.filter(t => t.id !== id)
        total.value--
        fetchStats()
        return true
      }
    } catch (e) {
      console.error('Failed to delete track:', e)
    }
    return false
  }

  async function batchDelete(ids) {
    if (!ids || ids.length === 0) return false
    try {
      const res = await fetch('/api/v1/tracks/batch-delete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authStore.authHeaders(),
        },
        body: JSON.stringify({ ids }),
      })
      if (res.ok) {
        const idSet = new Set(ids)
        tracks.value = tracks.value.filter(t => !idSet.has(t.id))
        total.value = Math.max(0, total.value - ids.length)
        fetchStats()
        return true
      }
    } catch (e) {
      console.error('Failed to batch delete tracks:', e)
    }
    return false
  }

  async function fetchAllReadyTracks(playlistId = '') {
    try {
      const url = `/api/v1/tracks/ready${playlistId ? `?playlist_id=${encodeURIComponent(playlistId)}` : ''}`
      const res = await fetch(url, {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        return Array.isArray(data) ? data : []
      }
    } catch (e) {
      console.error('Failed to fetch all ready tracks:', e)
    }
    return []
  }

  return {
    tracks,
    total,
    page,
    pageSize,
    totalPages,
    loading,
    searchQuery,
    selectedPlaylist,
    selectedStatus,
    sortBy,
    sortOrder,
    stats,
    getTrackStreamUrl,
    getTrackCoverUrl,
    getTrackDownloadUrl,
    updateTrack,
    fetchTracks,
    fetchAllReadyTracks,
    fetchStats,
    deleteTrack,
    batchDelete,
  }
})

