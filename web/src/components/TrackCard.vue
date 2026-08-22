<template>
  <div
    @click="$emit('play', track)"
    :class="[
      'group relative bg-studio-surface hover:bg-studio-elevated border rounded-xl p-3 cursor-pointer transition-all duration-200 flex flex-col justify-between select-none shadow-sm hover:shadow-md',
      selected ? 'border-indigo-500/80 bg-indigo-950/30' : 'border-studio-border hover:border-zinc-700/80'
    ]"
  >
    <!-- Cover Image with Overlay Play Button -->
    <div class="relative w-full aspect-square rounded-lg overflow-hidden bg-studio-elevated border border-studio-borderSubtle mb-3">
      <img
        :src="tracksStore.getTrackCoverUrl(track)"
        :alt="track.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
        loading="lazy"
        @error="onImageError"
      />

      <!-- Checkbox overlay in top-left corner -->
      <div
        class="absolute top-2 left-2 z-10"
        @click.stop="$emit('toggle-select', track)"
      >
        <input
          type="checkbox"
          :checked="selected"
          class="w-4 h-4 rounded border-zinc-700 bg-zinc-900/90 text-indigo-600 focus:ring-0 focus:ring-offset-0 cursor-pointer shadow-md"
        />
      </div>

      <!-- Hover Play Overlay -->
      <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
        <div class="w-11 h-11 rounded-full bg-indigo-500 text-white flex items-center justify-center shadow-lg transform translate-y-2 group-hover:translate-y-0 transition-transform">
          <svg v-if="isCurrentAndPlaying" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M6 4h4v16H6zm8 0h4v16h-4z"/>
          </svg>
          <svg v-else class="w-5 h-5 ml-0.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </div>
      </div>

      <!-- Format / Status Badge -->
      <div class="absolute bottom-2 left-2 flex items-center gap-1.5">
        <span
          v-if="track.status === 'ready'"
          class="px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-black/70 backdrop-blur-md text-zinc-300 border border-white/10 uppercase"
        >
          {{ track.format || 'opus' }}
        </span>
        <span
          v-else-if="track.status === 'downloading'"
          class="px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-amber-950/80 text-amber-300 border border-amber-800 animate-pulse"
        >
          Downloading
        </span>
        <span
          v-else-if="track.status === 'failed'"
          class="px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-rose-950/80 text-rose-300 border border-rose-800"
        >
          Failed
        </span>
      </div>

      <!-- Duration Badge -->
      <div class="absolute bottom-2 right-2 px-1.5 py-0.5 rounded text-[10px] font-mono text-zinc-300 bg-black/70 backdrop-blur-md border border-white/10">
        {{ formatDuration(track.duration) }}
      </div>
    </div>

    <!-- Metadata -->
    <div class="min-w-0">
      <h3 class="text-sm font-semibold text-zinc-100 truncate group-hover:text-indigo-400 transition-colors" :title="track.title">
        {{ track.title }}
      </h3>
      <p class="text-xs text-zinc-400 truncate mt-0.5" :title="track.artist">
        {{ track.artist || 'Unknown Artist' }}
      </p>
    </div>

    <!-- Action Bar (Play Next, Add Queue, Download, Delete) -->
    <div class="flex items-center justify-between mt-3 pt-2 border-t border-studio-borderSubtle text-zinc-400 opacity-60 group-hover:opacity-100 transition-opacity">
      <div class="flex items-center gap-1">
        <button
          @click.stop="playerStore.playNext(track)"
          title="Play Next"
          class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 4 15 12 5 20 5 4"/><line x1="19" y1="5" x2="19" y2="19"/>
          </svg>
        </button>
        <button
          @click.stop="playerStore.addToQueue(track)"
          title="Add to Queue"
          class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
        </button>
      </div>

      <div class="flex items-center gap-1">
        <a
          :href="tracksStore.getTrackDownloadUrl(track)"
          download
          @click.stop
          title="Download Audio File"
          class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/>
          </svg>
        </a>
        <button
          @click.stop="$emit('delete', track)"
          title="Delete Track"
          class="p-1 hover:text-rose-400 hover:bg-studio-hover rounded transition-colors"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useTracksStore } from '../stores/tracks'
import { usePlayerStore } from '../stores/player'

const props = defineProps({
  track: {
    type: Object,
    required: true,
  },
  selected: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['play', 'delete', 'toggle-select'])

const tracksStore = useTracksStore()
const playerStore = usePlayerStore()

const isCurrentAndPlaying = computed(() => {
  return playerStore.currentTrack?.id === props.track.id && playerStore.isPlaying
})

function formatDuration(secs) {
  if (!secs) return '0:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s < 10 ? '0' : ''}${s}`
}

function onImageError(e) {
  e.target.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 24 24" fill="none" stroke="%236366f1" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>'
}
</script>
