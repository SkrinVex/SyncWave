<template>
  <transition
    enter-active-class="transform transition ease-in-out duration-300"
    enter-from-class="translate-x-full"
    enter-to-class="translate-x-0"
    leave-active-class="transform transition ease-in-out duration-300"
    leave-from-class="translate-x-0"
    leave-to-class="translate-x-full"
  >
    <div
      v-if="playerStore.isQueueOpen"
      class="fixed right-0 top-0 bottom-20 w-full max-w-sm bg-studio-surface/98 backdrop-blur-xl border-l border-studio-border z-30 flex flex-col shadow-2xl select-none"
    >
      <!-- Header -->
      <div class="h-16 px-5 border-b border-studio-border flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h3 class="font-semibold text-zinc-100 text-sm">Play Queue</h3>
          <span class="text-xs font-mono text-zinc-400 bg-zinc-800 px-2 py-0.5 rounded">
            {{ playerStore.queue.length }} tracks
          </span>
        </div>
        <div class="flex items-center gap-2">
          <button
            v-if="playerStore.queue.length > 1"
            @click="playerStore.clearQueue"
            class="text-xs text-zinc-400 hover:text-rose-400 transition-colors px-2 py-1 rounded"
          >
            Clear
          </button>
          <button
            @click="playerStore.isQueueOpen = false"
            class="text-zinc-400 hover:text-zinc-100 p-1.5 rounded-lg hover:bg-studio-hover transition-colors"
          >
            ✕
          </button>
        </div>
      </div>

      <!-- Now Playing Section -->
      <div v-if="playerStore.currentTrack" class="p-4 border-b border-studio-borderSubtle bg-zinc-900/40">
        <span class="text-[10px] font-mono text-indigo-400 uppercase font-semibold tracking-wider">Now Playing</span>
        <div class="flex items-center gap-3 mt-2">
          <img
            :src="tracksStore.getTrackCoverUrl(playerStore.currentTrack)"
            class="w-10 h-10 rounded object-cover bg-zinc-800 shrink-0"
          />
          <div class="min-w-0 flex-1">
            <p class="text-xs font-semibold text-zinc-100 truncate">{{ playerStore.currentTrack.title }}</p>
            <p class="text-[11px] text-zinc-400 truncate">{{ playerStore.currentTrack.artist }}</p>
          </div>
          <span class="text-[10px] font-mono text-zinc-400">{{ formatDuration(playerStore.currentTrack.duration) }}</span>
        </div>
      </div>

      <!-- Queue List -->
      <div class="flex-1 overflow-y-auto p-3 space-y-1">
        <div
          v-for="(track, index) in playerStore.queue"
          :key="`${track.id}-${index}`"
          @click="playerStore.playTrack(track)"
          :class="[
            'flex items-center gap-3 p-2 rounded-lg cursor-pointer transition-all group text-left',
            index === playerStore.queueIndex
              ? 'bg-indigo-950/40 border border-indigo-800/60 text-indigo-200'
              : 'hover:bg-studio-hover text-zinc-300'
          ]"
        >
          <div class="w-5 text-center font-mono text-xs text-zinc-400 group-hover:hidden">
            {{ index + 1 }}
          </div>
          <button
            class="w-5 text-center text-indigo-400 hidden group-hover:block"
            title="Play"
          >
            ▶
          </button>

          <img
            :src="tracksStore.getTrackCoverUrl(track)"
            class="w-8 h-8 rounded object-cover bg-zinc-800 shrink-0"
          />

          <div class="min-w-0 flex-1">
            <p class="text-xs font-medium truncate" :class="{ 'font-semibold text-indigo-300': index === playerStore.queueIndex }">
              {{ track.title }}
            </p>
            <p class="text-[10px] text-zinc-400 truncate">{{ track.artist }}</p>
          </div>

          <span class="text-[10px] font-mono text-zinc-400 shrink-0">{{ formatDuration(track.duration) }}</span>

          <button
            @click.stop="playerStore.removeFromQueue(index)"
            class="text-zinc-500 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity p-1 text-xs"
            title="Remove from queue"
          >
            ✕
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { usePlayerStore } from '../stores/player'
import { useTracksStore } from '../stores/tracks'

const playerStore = usePlayerStore()
const tracksStore = useTracksStore()

function formatDuration(secs) {
  if (!secs) return '0:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s < 10 ? '0' : ''}${s}`
}
</script>

