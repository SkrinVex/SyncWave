<template>
  <div
    v-if="playerStore.currentTrack"
    class="fixed bottom-0 left-0 right-0 h-20 bg-studio-surface/95 backdrop-blur-xl border-t border-studio-border px-4 md:px-6 flex items-center justify-between gap-4 z-40 select-none shadow-2xl"
  >
    <!-- Left: Track Metadata & Cover -->
    <div class="flex items-center gap-3.5 min-w-0 w-1/4 max-w-[280px]">
      <div class="relative w-12 h-12 rounded-lg overflow-hidden bg-studio-elevated border border-studio-border shrink-0 group">
        <img
          :src="tracksStore.getTrackCoverUrl(playerStore.currentTrack)"
          :alt="playerStore.currentTrack.title"
          class="w-full h-full object-cover"
          @error="onImageError"
        />
        <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
          <a
            :href="`https://www.youtube.com/watch?v=${playerStore.currentTrack.youtube_id}`"
            target="_blank"
            rel="noopener noreferrer"
            title="Open on YouTube"
            class="text-zinc-200 hover:text-white"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>
            </svg>
          </a>
        </div>
      </div>

      <div class="min-w-0">
        <h4 class="text-sm font-semibold text-zinc-100 truncate leading-tight hover:text-indigo-400 transition-colors">
          {{ playerStore.currentTrack.title }}
        </h4>
        <p class="text-xs text-zinc-400 truncate mt-0.5">
          {{ playerStore.currentTrack.artist || 'Unknown Artist' }}
        </p>
      </div>
    </div>

    <!-- Center: Playback Controls & Progress Scrubber -->
    <div class="flex flex-col items-center gap-1.5 flex-1 max-w-2xl px-2">
      <!-- Buttons Row -->
      <div class="flex items-center gap-4 md:gap-6">
        <!-- Shuffle -->
        <button
          @click="playerStore.toggleShuffle"
          :class="[
            'transition-colors p-1 relative',
            playerStore.isShuffle ? 'text-indigo-400' : 'text-zinc-500 hover:text-zinc-300'
          ]"
          title="Shuffle"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M2 18h1.4c1.3 0 2.5-.6 3.3-1.7l6.6-8.6c.8-1.1 2-1.7 3.3-1.7H22M2 6h1.4c1.3 0 2.5.6 3.3 1.7l2 2.6M22 18h-5.4c-1.3 0-2.5-.6-3.3-1.7l-2-2.6M18 14l4 4-4 4M18 2l4 4-4 4"/>
          </svg>
          <span v-if="playerStore.isShuffle" class="absolute -bottom-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-indigo-400"></span>
        </button>

        <!-- Previous -->
        <button
          @click="playerStore.prev"
          class="text-zinc-400 hover:text-zinc-100 transition-colors p-1"
          title="Previous Track"
        >
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M6 6h2v12H6zm3.5 6 8.5 6V6z"/>
          </svg>
        </button>

        <!-- Play / Pause -->
        <button
          @click="playerStore.togglePlay"
          class="w-10 h-10 rounded-full bg-zinc-100 text-zinc-950 flex items-center justify-center hover:scale-105 active:scale-95 transition-all shadow-md"
          title="Play / Pause (Space)"
        >
          <svg v-if="playerStore.isPlaying" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M6 4h4v16H6zm8 0h4v16h-4z"/>
          </svg>
          <svg v-else class="w-5 h-5 ml-0.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </button>

        <!-- Next -->
        <button
          @click="playerStore.next"
          class="text-zinc-400 hover:text-zinc-100 transition-colors p-1"
          title="Next Track"
        >
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="m6 18 8.5-6L6 6v12zM16 6v12h2V6h-2z"/>
          </svg>
        </button>

        <!-- Loop -->
        <button
          @click="playerStore.toggleLoop"
          :class="[
            'transition-colors p-1 relative',
            playerStore.loopMode !== 'off' ? 'text-indigo-400' : 'text-zinc-500 hover:text-zinc-300'
          ]"
          title="Repeat"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m17 2 4 4-4 4M3 11v-1a4 4 0 0 1 4-4h14M7 22l-4-4 4-4M21 13v1a4 4 0 0 1-4 4H3"/>
          </svg>
          <span
            v-if="playerStore.loopMode === 'one'"
            class="absolute -top-1 -right-1 text-[9px] font-bold font-mono bg-indigo-500 text-zinc-950 px-1 rounded-full leading-none"
          >
            1
          </span>
          <span
            v-else-if="playerStore.loopMode === 'all'"
            class="absolute -bottom-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-indigo-400"
          ></span>
        </button>
      </div>

      <!-- Scrubber Row -->
      <div class="w-full flex items-center gap-3">
        <span class="text-[11px] font-mono text-zinc-400 tabular-nums w-9 text-right">
          {{ formatTime(playerStore.currentTime) }}
        </span>

        <div class="relative flex-1 flex items-center h-4 group">
          <!-- Background Bar -->
          <div class="absolute left-0 right-0 h-1 bg-zinc-800 rounded-full overflow-hidden">
            <!-- Buffer Progress -->
            <div
              class="h-full bg-zinc-700/60 transition-all duration-300"
              :style="{ width: `${playerStore.buffered}%` }"
            ></div>
          </div>
          <!-- Played Progress -->
          <div
            class="absolute left-0 h-1 bg-indigo-500 group-hover:bg-indigo-400 rounded-full pointer-events-none transition-colors"
            :style="{ width: `${progressPercentage}%` }"
          ></div>
          <!-- Range Slider -->
          <input
            type="range"
            min="0"
            :max="playerStore.duration || 100"
            :value="playerStore.currentTime"
            @input="onSeek"
            class="w-full relative z-10 opacity-0 group-hover:opacity-100 transition-opacity"
          />
        </div>

        <span class="text-[11px] font-mono text-zinc-400 tabular-nums w-9">
          {{ formatTime(playerStore.duration) }}
        </span>
      </div>
    </div>

    <!-- Right: Volume, Format Badge & Queue Toggle -->
    <div class="flex items-center justify-end gap-3 w-1/4 max-w-[280px]">
      <!-- Audio Format Pill -->
      <div class="hidden xl:flex items-center gap-1 px-2 py-0.5 rounded bg-zinc-800/80 border border-zinc-700/50 text-[10px] font-mono text-zinc-400 uppercase">
        <span>{{ playerStore.currentTrack.format || 'opus' }}</span>
        <span>•</span>
        <span>{{ playerStore.currentTrack.bitrate || 160 }}k</span>
      </div>

      <!-- Download Button -->
      <a
        :href="tracksStore.getTrackDownloadUrl(playerStore.currentTrack)"
        download
        title="Download Audio File"
        class="text-zinc-400 hover:text-zinc-100 transition-colors p-1.5 rounded-lg hover:bg-studio-hover hidden sm:block"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/>
        </svg>
      </a>

      <!-- Queue Button -->
      <button
        @click="playerStore.isQueueOpen = !playerStore.isQueueOpen"
        :class="[
          'p-1.5 rounded-lg transition-colors relative',
          playerStore.isQueueOpen ? 'bg-indigo-600/20 text-indigo-400 border border-indigo-500/30' : 'text-zinc-400 hover:text-zinc-100 hover:bg-studio-hover'
        ]"
        title="Queue"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/>
          <line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>
        </svg>
        <span
          v-if="playerStore.queue.length > 0"
          class="absolute -top-1 -right-1 text-[9px] font-mono font-bold bg-zinc-800 border border-zinc-700 text-zinc-300 px-1 rounded-full"
        >
          {{ playerStore.queue.length }}
        </span>
      </button>

      <!-- Volume Control -->
      <div class="hidden sm:flex items-center gap-2 group">
        <button
          @click="playerStore.toggleMute"
          class="text-zinc-400 hover:text-zinc-100 transition-colors p-1"
        >
          <svg v-if="playerStore.isMuted || playerStore.volume === 0" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="1" y1="1" x2="23" y2="23"/><path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
            <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0a7 7 0 0 1-.11 1.23"/>
          </svg>
          <svg v-else class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
            <path d="M15.54 8.46a5 5 0 0 1 0 7.07M19.07 4.93a10 10 0 0 1 0 14.14"/>
          </svg>
        </button>

        <input
          type="range"
          min="0"
          max="1"
          step="0.01"
          :value="playerStore.isMuted ? 0 : playerStore.volume"
          @input="onVolumeChange"
          class="w-16 md:w-20"
        />
      </div>

      <!-- Close Player Button -->
      <button
        @click="closePlayer"
        class="text-zinc-500 hover:text-zinc-200 transition-colors p-1.5 rounded-lg hover:bg-studio-hover ml-1"
        title="Close Player"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>

    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePlayerStore } from '../stores/player'
import { useTracksStore } from '../stores/tracks'

const playerStore = usePlayerStore()
const tracksStore = useTracksStore()

const progressPercentage = computed(() => {
  if (!playerStore.duration || playerStore.duration === 0) return 0
  return (playerStore.currentTime / playerStore.duration) * 100
})

function onSeek(e) {
  playerStore.seek(parseFloat(e.target.value))
}

function onVolumeChange(e) {
  playerStore.setVolume(parseFloat(e.target.value))
}

function formatTime(secs) {
  if (!secs || isNaN(secs)) return '0:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s < 10 ? '0' : ''}${s}`
}

function onImageError(e) {
  e.target.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 24 24" fill="none" stroke="%236366f1" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>'
}

function closePlayer() {
  playerStore.pause()
  playerStore.currentTrack = null
  playerStore.queue = []
}
</script>

