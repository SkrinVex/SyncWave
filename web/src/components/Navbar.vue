<template>
  <header class="h-16 bg-studio-surface/90 backdrop-blur-md border-b border-studio-border px-6 flex items-center justify-between gap-4 select-none sticky top-0 z-20">
    <!-- Left: Mobile Menu Toggle & Title or Search -->
    <div class="flex items-center gap-4 flex-1 max-w-xl">
      <button
        @click="$emit('toggle-mobile-menu')"
        class="md:hidden text-zinc-400 hover:text-zinc-100 p-1 rounded-lg"
      >
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 6h16M4 12h16M4 18h16"/>
        </svg>
      </button>

      <!-- Global Search Bar -->
      <div class="relative w-full">
        <svg class="w-4 h-4 text-zinc-400 absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>
        </svg>
        <input
          type="text"
          v-model="tracksStore.searchQuery"
          @input="onSearchInput"
          placeholder="Search by title, artist, or album..."
          class="w-full bg-studio-elevated border border-studio-border rounded-lg pl-9 pr-8 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/50 transition-all font-sans"
        />
        <button
          v-if="tracksStore.searchQuery"
          @click="clearSearch"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 text-xs px-1"
        >
          ✕
        </button>
      </div>
    </div>

    <!-- Right: Telemetry, Sync Indicator & Stats -->
    <div class="flex items-center gap-3">
      <!-- Live Sync Pill (when active) -->
      <div
        v-if="syncStore.progress.active"
        @click="$emit('open-sync')"
        class="cursor-pointer hidden sm:flex items-center gap-2.5 px-3 py-1.5 rounded-full bg-emerald-950/70 border border-emerald-800/80 text-emerald-300 text-xs hover:bg-emerald-900/60 transition-colors"
      >
        <div class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></div>
        <span class="font-mono font-medium truncate max-w-[140px]">
          {{ syncStore.progress.current_track_title || 'Syncing...' }}
        </span>
        <span class="font-mono text-[10px] bg-emerald-900/90 text-emerald-200 px-1.5 py-0.5 rounded">
          {{ Math.round(syncStore.progress.percentage) }}%
        </span>
      </div>

      <!-- Storage & Stats Pill -->
      <div class="hidden lg:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-studio-elevated border border-studio-border text-xs font-mono text-zinc-400">
        <svg class="w-3.5 h-3.5 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
        </svg>
        <span>{{ formatBytes(tracksStore.stats.total_storage_size) }}</span>
        <span class="text-zinc-600">•</span>
        <span>{{ tracksStore.stats.total_tracks }} tracks</span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useTracksStore } from '../stores/tracks'
import { useSyncStore } from '../stores/sync'

defineEmits(['toggle-mobile-menu', 'open-sync'])

const tracksStore = useTracksStore()
const syncStore = useSyncStore()

let debounceTimeout = null
function onSearchInput() {
  clearTimeout(debounceTimeout)
  debounceTimeout = setTimeout(() => {
    tracksStore.fetchTracks(true)
  }, 250)
}

function clearSearch() {
  tracksStore.searchQuery = ''
  tracksStore.fetchTracks(true)
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

