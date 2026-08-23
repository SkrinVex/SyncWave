<template>
  <div
    @click="$emit('play', track)"
    :class="[
      'flex items-center gap-3 md:gap-4 px-3 md:px-4 py-2.5 rounded-lg cursor-pointer transition-all select-none group text-left border',
      selected ? 'bg-indigo-950/40 border-indigo-500/50 text-indigo-100' : (
        isCurrentTrack
          ? 'bg-studio-elevated border-indigo-500/40 text-indigo-200'
          : 'border-transparent hover:bg-studio-elevated hover:border-studio-border text-zinc-300'
      )
    ]"
  >
    <!-- Checkbox for multi-select -->
    <div class="shrink-0 flex items-center" @click.stop="$emit('toggle-select', track)">
      <input
        type="checkbox"
        :checked="selected"
        class="w-4 h-4 rounded border-zinc-700 bg-zinc-900 text-indigo-600 focus:ring-0 focus:ring-offset-0 cursor-pointer"
      />
    </div>

    <!-- Index / Play Button -->
    <div class="hidden md:block w-6 text-center shrink-0">
      <span class="font-mono text-xs text-zinc-400 group-hover:hidden" :class="{ 'text-indigo-400': isCurrentTrack }">
        {{ index }}
      </span>
      <span class="text-indigo-400 hidden group-hover:block">
        ▶
      </span>
    </div>

    <!-- Cover Thumbnail -->
    <div class="relative w-10 h-10 rounded-md overflow-hidden bg-studio-elevated border border-studio-borderSubtle shrink-0">
      <img
        :src="tracksStore.getTrackCoverUrl(track)"
        :alt="track.title"
        class="w-full h-full object-cover"
        loading="lazy"
        @error="onImageError"
      />
      <!-- Playing / Loading Indicator Overlay -->
      <div v-if="isCurrentTrack" class="absolute inset-0 bg-black/60 flex items-center justify-center">
        <template v-if="playerStore.isLoading">
          <svg class="w-4 h-4 text-indigo-400 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
        </template>
        <template v-else-if="playerStore.isPlaying">
          <div class="flex items-end gap-[1.5px] h-3">
            <div class="w-[2.5px] bg-indigo-400 rounded-sm animate-[eq_0.8s_ease-in-out_infinite_alternate]" style="animation-delay: -0.4s"></div>
            <div class="w-[2.5px] bg-indigo-400 rounded-sm animate-[eq_0.8s_ease-in-out_infinite_alternate]" style="animation-delay: -0.2s"></div>
            <div class="w-[2.5px] bg-indigo-400 rounded-sm animate-[eq_0.8s_ease-in-out_infinite_alternate]"></div>
          </div>
        </template>
        <template v-else>
          <svg class="w-4 h-4 text-indigo-400 ml-0.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </template>
      </div>
    </div>

    <!-- Title & Artist -->
    <div class="min-w-0 flex-1">
      <h4 class="text-sm font-medium truncate" :class="{ 'font-semibold text-indigo-300': isCurrentTrack }">
        {{ track.title }}
      </h4>
      <p class="text-xs text-zinc-400 truncate mt-0.5">{{ track.artist || 'Unknown Artist' }}</p>
    </div>

    <!-- Album (Hidden on mobile) -->
    <div class="hidden md:block w-1/4 min-w-0">
      <p class="text-xs text-zinc-400 truncate">{{ track.album || '-' }}</p>
    </div>

    <!-- Format & Bitrate (Hidden on small screens) -->
    <div class="hidden lg:flex items-center gap-1.5 shrink-0">
      <span class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-zinc-800 border border-zinc-700 text-zinc-400 uppercase">
        {{ track.format || 'opus' }}
      </span>
    </div>

    <!-- Duration -->
    <div class="w-12 text-right shrink-0">
      <span class="text-xs font-mono text-zinc-400">{{ formatDuration(track.duration) }}</span>
    </div>

    <!-- Actions Menu -->
    <div class="hidden md:flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
      <button
        @click.stop="playerStore.playNext(track)"
        title="Play Next"
        class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polygon points="5 4 15 12 5 20 5 4"/><line x1="19" y1="5" x2="19" y2="19"/>
        </svg>
      </button>
      <button
        @click.stop="playerStore.addToQueue(track)"
        title="Add to Queue"
        class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
      </button>
      <a
        :href="tracksStore.getTrackDownloadUrl(track)"
        download
        @click.stop
        title="Download Audio File"
        class="p-1 hover:text-zinc-100 hover:bg-studio-hover rounded transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/>
        </svg>
      </a>
      <button
        @click.stop="$emit('delete', track)"
        title="Delete"
        class="p-1 hover:text-rose-400 hover:bg-studio-hover rounded transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
        </svg>
      </button>
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
  index: {
    type: Number,
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

const isCurrentTrack = computed(() => playerStore.currentTrack?.id === props.track.id)
const isCurrentPlaying = computed(() => isCurrentTrack.value && playerStore.isPlaying)

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
