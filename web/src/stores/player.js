import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useTracksStore } from './tracks'
import { useToastStore } from './toast'

export const usePlayerStore = defineStore('player', () => {
  const tracksStore = useTracksStore()
  const toast = useToastStore()

  const audio = new Audio()
  audio.preload = 'metadata'

  const currentTrack = ref(null)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const buffered = ref(0)
  const volume = ref(parseFloat(localStorage.getItem('syncwave_volume') || '0.8'))
  const isMuted = ref(false)
  const loopMode = ref(localStorage.getItem('syncwave_loop') || 'off') // 'off' | 'all' | 'one'
  const isShuffle = ref(localStorage.getItem('syncwave_shuffle') === 'true')
  const queue = ref([])
  const queueIndex = ref(-1)
  const isQueueOpen = ref(false)
  const isLoading = ref(false)

  // Audio setup
  audio.volume = volume.value

  audio.addEventListener('loadstart', () => {
    isLoading.value = true
  })

  audio.addEventListener('waiting', () => {
    isLoading.value = true
  })

  audio.addEventListener('playing', () => {
    isLoading.value = false
    isPlaying.value = true
  })

  audio.addEventListener('canplay', () => {
    isLoading.value = false
  })

  audio.addEventListener('canplaythrough', () => {
    isLoading.value = false
  })

  audio.addEventListener('stalled', () => {
    if (!isPlaying.value) {
      isLoading.value = false
    }
  })

  audio.addEventListener('timeupdate', () => {
    currentTime.value = audio.currentTime
    if (audio.duration && !isNaN(audio.duration) && isFinite(audio.duration)) {
      duration.value = audio.duration
    }
    if (audio.buffered.length > 0) {
      const end = audio.buffered.end(audio.buffered.length - 1)
      buffered.value = duration.value > 0 ? (end / duration.value) * 100 : 0
    }
  })

  audio.addEventListener('loadedmetadata', () => {
    if (audio.duration && !isNaN(audio.duration) && isFinite(audio.duration)) {
      duration.value = audio.duration
    } else if (currentTrack.value && currentTrack.value.duration) {
      duration.value = currentTrack.value.duration
    }
    isLoading.value = false
  })

  audio.addEventListener('ended', () => {
    if (loopMode.value === 'one') {
      audio.currentTime = 0
      audio.play().catch(e => console.error('Loop restart failed:', e))
    } else {
      next()
    }
  })

  audio.addEventListener('play', () => {
    isPlaying.value = true
  })

  audio.addEventListener('pause', () => {
    isPlaying.value = false
  })

  audio.addEventListener('error', (e) => {
    console.error('Audio playback error:', e, audio.error)
    isLoading.value = false
    isPlaying.value = false
    if (currentTrack.value) {
      toast.error(`Не удалось воспроизвести: ${currentTrack.value.title || 'трек'}`)
    }
  })

  function setupMediaSession(track) {
    if (!('mediaSession' in navigator) || !track) return

    const coverUrl = tracksStore.getTrackCoverUrl(track)

    navigator.mediaSession.metadata = new MediaMetadata({
      title: track.title || 'Unknown Title',
      artist: track.artist || 'Unknown Artist',
      album: track.album || 'YouTube Music',
      artwork: [
        { src: coverUrl, sizes: '96x96', type: 'image/jpeg' },
        { src: coverUrl, sizes: '128x128', type: 'image/jpeg' },
        { src: coverUrl, sizes: '256x256', type: 'image/jpeg' },
        { src: coverUrl, sizes: '512x512', type: 'image/jpeg' },
      ],
    })

    navigator.mediaSession.setActionHandler('play', () => play())
    navigator.mediaSession.setActionHandler('pause', () => pause())
    navigator.mediaSession.setActionHandler('previoustrack', () => prev())
    navigator.mediaSession.setActionHandler('nexttrack', () => next())
    navigator.mediaSession.setActionHandler('seekto', (details) => {
      if (details.fastSeek && 'fastSeek' in audio) {
        audio.fastSeek(details.seekTime)
      } else {
        seek(details.seekTime)
      }
    })
  }

  function clearMediaSession() {
    if ('mediaSession' in navigator) {
      navigator.mediaSession.metadata = null
      navigator.mediaSession.setActionHandler('play', null)
      navigator.mediaSession.setActionHandler('pause', null)
      navigator.mediaSession.setActionHandler('previoustrack', null)
      navigator.mediaSession.setActionHandler('nexttrack', null)
      navigator.mediaSession.setActionHandler('seekto', null)
    }
  }

  function shuffleArray(array) {
    const arr = [...array]
    for (let i = arr.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[arr[i], arr[j]] = [arr[j], arr[i]]
    }
    return arr
  }

  function reshuffleQueue() {
    if (queue.value.length <= 1) return
    const cur = currentTrack.value
    if (!cur) {
      queue.value = shuffleArray(queue.value)
      queueIndex.value = 0
      return
    }

    const otherTracks = queue.value.filter(t => t.id !== cur.id)
    const shuffledOthers = shuffleArray(otherTracks)
    queue.value = [cur, ...shuffledOthers]
    queueIndex.value = 0
  }

  function playTrack(track, newQueue = null) {
    if (!track) return

    if (newQueue && Array.isArray(newQueue) && newQueue.length > 0) {
      if (isShuffle.value) {
        const otherTracks = newQueue.filter(t => t.id !== track.id)
        queue.value = [track, ...shuffleArray(otherTracks)]
        queueIndex.value = 0
      } else {
        queue.value = [...newQueue]
        queueIndex.value = queue.value.findIndex(t => t.id === track.id)
        if (queueIndex.value === -1) {
          queue.value.unshift(track)
          queueIndex.value = 0
        }
      }
    } else if (queue.value.length === 0) {
      queue.value = [track]
      queueIndex.value = 0
    } else {
      const idx = queue.value.findIndex(t => t.id === track.id)
      if (idx !== -1) {
        queueIndex.value = idx
      } else {
        queue.value.splice(queueIndex.value + 1, 0, track)
        queueIndex.value = queueIndex.value + 1
      }
    }

    // Immediately update reactive track metadata in UI
    currentTrack.value = { ...track }
    currentTime.value = 0
    duration.value = track.duration || 0
    buffered.value = 0
    isLoading.value = true
    isPlaying.value = true

    const streamUrl = tracksStore.getTrackStreamUrl(track)
    audio.src = streamUrl

    const playPromise = audio.play()
    if (playPromise !== undefined) {
      playPromise.then(() => {
        isLoading.value = false
        isPlaying.value = true
      }).catch(e => {
        if (e.name !== 'AbortError') {
          console.error('Play failed:', e)
          isPlaying.value = false
        }
        isLoading.value = false
      })
    }

    setupMediaSession(track)
  }

  function togglePlay() {
    if (!currentTrack.value) {
      if (queue.value.length > 0) {
        playTrack(queue.value[0])
      }
      return
    }

    if (isPlaying.value) {
      audio.pause()
    } else {
      const p = audio.play()
      if (p !== undefined) {
        p.catch(e => console.error('Resume failed:', e))
      }
    }
  }

  function play() {
    if (currentTrack.value) {
      const p = audio.play()
      if (p !== undefined) {
        p.catch(e => console.error('Play failed:', e))
      }
    }
  }

  function pause() {
    audio.pause()
  }

  function seek(seconds) {
    if (isNaN(seconds)) return
    const safeSec = Math.max(0, Math.min(seconds, duration.value || 999999))
    audio.currentTime = safeSec
    currentTime.value = safeSec
  }

  function setVolume(val) {
    const v = Math.max(0, Math.min(1, val))
    volume.value = v
    audio.volume = isMuted.value ? 0 : v
    localStorage.setItem('syncwave_volume', v.toString())
  }

  function toggleMute() {
    isMuted.value = !isMuted.value
    audio.volume = isMuted.value ? 0 : volume.value
  }

  function next() {
    if (queue.value.length === 0) return

    if (queueIndex.value < queue.value.length - 1) {
      queueIndex.value++
      playTrack(queue.value[queueIndex.value])
    } else if (loopMode.value === 'all') {
      if (isShuffle.value) {
        reshuffleQueue()
      }
      queueIndex.value = 0
      playTrack(queue.value[0])
    } else {
      isPlaying.value = false
    }
  }

  function prev() {
    if (currentTime.value > 3 || queueIndex.value <= 0) {
      seek(0)
    } else if (queueIndex.value > 0) {
      queueIndex.value--
      playTrack(queue.value[queueIndex.value])
    }
  }

  function toggleLoop() {
    if (loopMode.value === 'off') loopMode.value = 'all'
    else if (loopMode.value === 'all') loopMode.value = 'one'
    else loopMode.value = 'off'
    localStorage.setItem('syncwave_loop', loopMode.value)
  }

  function toggleShuffle() {
    isShuffle.value = !isShuffle.value
    localStorage.setItem('syncwave_shuffle', isShuffle.value.toString())
    if (isShuffle.value) {
      reshuffleQueue()
    }
  }

  function addToQueue(track) {
    queue.value.push(track)
    if (!currentTrack.value) {
      playTrack(track)
    }
  }

  function playNext(track) {
    if (queue.value.length === 0 || queueIndex.value === -1) {
      playTrack(track)
      return
    }
    queue.value.splice(queueIndex.value + 1, 0, track)
  }

  function removeFromQueue(index) {
    if (index === queueIndex.value) {
      next()
    }
    queue.value.splice(index, 1)
    if (index < queueIndex.value) {
      queueIndex.value--
    }
  }

  function clearQueue() {
    if (currentTrack.value) {
      queue.value = [currentTrack.value]
      queueIndex.value = 0
    } else {
      queue.value = []
      queueIndex.value = -1
    }
  }

  function stop() {
    audio.pause()
    audio.src = ''
    currentTrack.value = null
    isPlaying.value = false
    currentTime.value = 0
    duration.value = 0
    buffered.value = 0
    queue.value = []
    queueIndex.value = -1
    clearMediaSession()
  }

  return {
    currentTrack,
    isPlaying,
    currentTime,
    duration,
    buffered,
    volume,
    isMuted,
    loopMode,
    isShuffle,
    queue,
    queueIndex,
    isQueueOpen,
    isLoading,
    playTrack,
    togglePlay,
    play,
    pause,
    stop,
    seek,
    setVolume,
    toggleMute,
    next,
    prev,
    toggleLoop,
    toggleShuffle,
    reshuffleQueue,
    addToQueue,
    playNext,
    removeFromQueue,
    clearQueue,
    clearMediaSession,
  }
})
