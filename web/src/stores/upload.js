import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { useTracksStore } from './tracks'
import { usePlaylistsStore } from './playlists'
import { useToastStore } from './toast'
import { useI18nStore } from './i18n'

export const useUploadStore = defineStore('upload', () => {
  const authStore = useAuthStore()
  const tracksStore = useTracksStore()
  const playlistsStore = usePlaylistsStore()
  const toast = useToastStore()
  const i18n = useI18nStore()

  const tasks = ref([])
  const isDrawerOpen = ref(true)
  const isMinimized = ref(false)

  const isUploading = computed(() => {
    return tasks.value.some(t => t.status === 'uploading' || t.status === 'processing' || t.status === 'pending')
  })

  const totalTasks = computed(() => tasks.value.length)
  const completedTasks = computed(() => tasks.value.filter(t => t.status === 'done').length)
  const failedTasks = computed(() => tasks.value.filter(t => t.status === 'error').length)

  const overallProgress = computed(() => {
    if (tasks.value.length === 0) return 0
    const sum = tasks.value.reduce((acc, t) => acc + (t.progress || 0), 0)
    return Math.round(sum / tasks.value.length)
  })

  async function uploadFiles(fileList, playlistId = '') {
    if (!fileList || fileList.length === 0) return

    const newFiles = Array.from(fileList)
    const newTasks = newFiles.map((file, idx) => ({
      id: `${Date.now()}-${idx}-${Math.random().toString(36).substring(2, 7)}`,
      file,
      name: file.name,
      size: file.size,
      progress: 0,
      status: 'pending', // 'pending' | 'uploading' | 'processing' | 'done' | 'error'
      error: '',
    }))

    tasks.value.push(...newTasks)
    isMinimized.value = false

    // Process uploads in batches of 3
    const batchSize = 3
    for (let i = 0; i < newTasks.length; i += batchSize) {
      const batch = newTasks.slice(i, i + batchSize)
      await Promise.all(batch.map(task => uploadSingleFile(task, playlistId)))
    }

    // Refresh store after all finish
    await Promise.all([
      tracksStore.fetchTracks(),
      tracksStore.fetchStats(),
      playlistsStore.fetchPlaylists(),
    ])

    const successCount = newTasks.filter(t => t.status === 'done').length
    const errorCount = newTasks.filter(t => t.status === 'error').length

    if (successCount > 0 && errorCount === 0) {
      toast.success(
        i18n.currentLang === 'ru'
          ? `Успешно загружено треков: ${successCount}`
          : `Successfully uploaded ${successCount} tracks`
      )
    } else if (successCount > 0 && errorCount > 0) {
      toast.warn(
        i18n.currentLang === 'ru'
          ? `Загружено: ${successCount}, с ошибками: ${errorCount}`
          : `Uploaded ${successCount} tracks, ${errorCount} failed`
      )
    } else if (errorCount > 0) {
      toast.error(
        i18n.currentLang === 'ru'
          ? 'Ошибка при загрузке аудиофайлов'
          : 'Failed to upload audio files'
      )
    }
  }

  function uploadSingleFile(task, playlistId) {
    return new Promise((resolve) => {
      task.status = 'uploading'
      task.progress = 0

      const xhr = new XMLHttpRequest()
      const formData = new FormData()
      formData.append('files', task.file, task.name)
      if (playlistId) {
        formData.append('playlist_id', playlistId)
      }

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const pct = Math.round((e.loaded / e.total) * 90) // 0-90% upload
          task.progress = pct
          if (pct >= 90) {
            task.status = 'processing'
          }
        }
      })

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText)
            if (data.errors && data.errors.length > 0 && (!data.uploaded || data.uploaded.length === 0)) {
              task.status = 'error'
              task.error = data.errors.join(', ')
              task.progress = 100
            } else {
              task.status = 'done'
              task.progress = 100
            }
          } catch (e) {
            task.status = 'done'
            task.progress = 100
          }
        } else {
          task.status = 'error'
          try {
            const err = JSON.parse(xhr.responseText)
            task.error = err.error || err.message || `HTTP ${xhr.status}`
          } catch (e) {
            task.error = `HTTP ${xhr.status}`
          }
          task.progress = 100
        }
        resolve()
      })

      xhr.addEventListener('error', () => {
        task.status = 'error'
        task.error = 'Сетевая ошибка при передаче файла'
        task.progress = 100
        resolve()
      })

      xhr.addEventListener('abort', () => {
        task.status = 'error'
        task.error = 'Загрузка отменена'
        resolve()
      })

      const headers = authStore.authHeaders()
      xhr.open('POST', '/api/v1/tracks/upload')
      for (const [k, v] of Object.entries(headers)) {
        xhr.setRequestHeader(k, v)
      }

      xhr.send(formData)
    })
  }

  function clearCompleted() {
    tasks.value = tasks.value.filter(t => t.status === 'uploading' || t.status === 'processing' || t.status === 'pending')
  }

  function removeTask(id) {
    tasks.value = tasks.value.filter(t => t.id !== id)
  }

  return {
    tasks,
    isDrawerOpen,
    isMinimized,
    isUploading,
    totalTasks,
    completedTasks,
    failedTasks,
    overallProgress,
    uploadFiles,
    clearCompleted,
    removeTask,
  }
})

