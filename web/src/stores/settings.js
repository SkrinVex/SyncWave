import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'

export const useSettingsStore = defineStore('settings', () => {
  const authStore = useAuthStore()

  const settings = ref({
    http_proxy: '',
    audio_format: 'opus',
    audio_quality: '0 (Best)',
    max_concurrent: 2,
    has_cookies: false,
    cookies_valid: false,
    cookies_status: 'missing',
    cookies_expires_at: '',
    cookies_error: '',
    cookies_updated_at: '',
    ytdlp_version: '',
    ffmpeg_version: '',
    storage_usage_bytes: 0,
    database_size_bytes: 0,
    total_tracks_count: 0,
    total_playlists_count: 0,
    user_storage_usage_bytes: 0,
    user_storage_quota_bytes: 0,
    host_disk_total_bytes: 0,
    host_disk_used_bytes: 0,
    host_disk_free_bytes: 0,
    is_admin: false,
  })
  const loading = ref(false)
  const showCookieModal = ref(false)

  async function fetchSettings() {
    loading.value = true
    try {
      const res = await fetch('/api/v1/settings', {
        headers: authStore.authHeaders(),
      })
      if (res.ok) {
        settings.value = await res.json()
      }
    } catch (e) {
      console.error('Failed to fetch settings:', e)
    } finally {
      loading.value = false
    }
  }

  async function updateSettings(payload) {
    const res = await fetch('/api/v1/settings', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...authStore.authHeaders(),
      },
      body: JSON.stringify(payload),
    })
    if (!res.ok) {
      const errText = await res.text()
      let errMsg = errText
      try {
        const json = JSON.parse(errText)
        if (json.error) errMsg = json.error
      } catch {}
      throw new Error(errMsg || 'Не удалось сохранить настройки')
    }
    await fetchSettings()
  }

  async function uploadCookies(file) {
    const formData = new FormData()
    formData.append('cookies', file)

    const res = await fetch('/api/v1/settings/cookies', {
      method: 'POST',
      headers: authStore.authHeaders(),
      body: formData,
    })

    if (!res.ok) {
      const errText = await res.text()
      let errMsg = errText
      try {
        const json = JSON.parse(errText)
        if (json.error) errMsg = json.error
      } catch {}
      throw new Error(errMsg || 'Не удалось загрузить cookies.txt')
    }
    await fetchSettings()
  }

  async function deleteCookies() {
    const res = await fetch('/api/v1/settings/cookies', {
      method: 'DELETE',
      headers: authStore.authHeaders(),
    })
    if (res.ok) {
      await fetchSettings()
    }
  }

  async function testProxy(proxyUrl) {
    const res = await fetch('/api/v1/settings/test-proxy', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...authStore.authHeaders(),
      },
      body: JSON.stringify({ proxy_url: proxyUrl }),
    })
    if (!res.ok) {
      const errText = await res.text()
      let errMsg = errText
      try {
        const json = JSON.parse(errText)
        if (json.error) errMsg = json.error
      } catch {}
      throw new Error(errMsg || 'Ошибка проверки прокси')
    }
    return true
  }

  return {
    settings,
    loading,
    showCookieModal,
    fetchSettings,
    updateSettings,
    uploadCookies,
    deleteCookies,
    testProxy,
  }
})
