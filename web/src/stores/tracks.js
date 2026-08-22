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
    return `/api/v1/tracks/${track.id}/cover${t ? `?token=${encodeURIComponent(t)}` : ''}`
  }

  function getTrackDownloadUrl(track) {
    if (!track) return ''
    const t = authStore.token
    return `/api/v1/tracks/${track.id}/download${t ? `?token=${encodeURIComponent(t)}` : ''}`
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
    fetchTracks,
    fetchStats,
    deleteTrack,
  }
})

