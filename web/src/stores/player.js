import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useTracksStore } from './tracks'

export const usePlayerStore = defineStore('player', () => {
  const tracksStore = useTracksStore()

  const audio = new Audio()
  audio.preload = 'auto'

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

  audio.addEventListener('waiting', () => isLoading.value = true)
  audio.addEventListener('playing', () => isLoading.value = false)
  audio.addEventListener('canplay', () => isLoading.value = false)

  audio.addEventListener('timeupdate', () => {
    currentTime.value = audio.currentTime
    if (audio.buffered.length > 0) {
      const end = audio.buffered.end(audio.buffered.length - 1)
      buffered.value = duration.value > 0 ? (end / duration.value) * 100 : 0
    }
  })

  audio.addEventListener('loadedmetadata', () => {
    duration.value = audio.duration || (currentTrack.value ? currentTrack.value.duration : 0)
  })

  audio.addEventListener('ended', () => {
    if (loopMode.value === 'one') {
      audio.currentTime = 0
      audio.play()
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
    console.error('Audio playback error:', e)
    isPlaying.value = false
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

  function playTrack(track, newQueue = null) {
    if (!track) return

    if (newQueue) {
      queue.value = [...newQueue]
      queueIndex.value = queue.value.findIndex(t => t.id === track.id)
      if (queueIndex.value === -1) {
        queue.value.unshift(track)
        queueIndex.value = 0
      }
    } else if (queue.value.length === 0) {
      queue.value = [track]
      queueIndex.value = 0
    }

    currentTrack.value = track
    isLoading.value = true
    const streamUrl = tracksStore.getTrackStreamUrl(track)
    audio.src = streamUrl
    audio.play().catch(e => {
      console.error('Play failed:', e)
      isLoading.value = false
    })
    isPlaying.value = true
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
      audio.play().catch(e => console.error('Resume failed:', e))
    }
  }

  function play() {
    if (currentTrack.value) audio.play()
  }

  function pause() {
    audio.pause()
  }

  function seek(seconds) {
    if (isNaN(seconds)) return
    audio.currentTime = seconds
    currentTime.value = seconds
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

    if (isShuffle.value) {
      const nextIdx = Math.floor(Math.random() * queue.value.length)
      queueIndex.value = nextIdx
      playTrack(queue.value[nextIdx])
      return
    }

    if (queueIndex.value < queue.value.length - 1) {
      queueIndex.value++
      playTrack(queue.value[queueIndex.value])
    } else if (loopMode.value === 'all') {
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
    seek,
    setVolume,
    toggleMute,
    next,
    prev,
    toggleLoop,
    toggleShuffle,
    addToQueue,
    playNext,
    removeFromQueue,
    clearQueue,
    clearMediaSession,
  }
})

