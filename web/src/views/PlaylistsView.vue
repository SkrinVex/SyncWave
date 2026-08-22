<template>
  <div class="space-y-6 pb-28">
    <!-- Header -->
    <div class="flex items-center justify-between select-none">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-zinc-100">Playlists & Subscriptions</h2>
        <p class="text-xs text-zinc-400 mt-1">Manage tracked YouTube Music playlists for automated background sync</p>
      </div>

      <button
        @click="$emit('open-add-playlist')"
        class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md active:scale-95"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        <span>Add Playlist</span>
      </button>
    </div>

    <!-- Playlists Grid -->
    <div v-if="playlistsStore.loading" class="py-20 flex flex-col items-center justify-center gap-3 text-zinc-400">
      <div class="w-8 h-8 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
      <p class="text-xs font-mono">Loading playlists...</p>
    </div>

    <div
      v-else-if="playlistsStore.playlists.length === 0"
      class="py-20 border border-dashed border-zinc-800 rounded-2xl flex flex-col items-center justify-center text-center p-8 select-none"
    >
      <div class="w-14 h-14 rounded-2xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mb-4">
        <svg class="w-7 h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15V6M18.5 18a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5ZM12 12H3M16 6H3M12 18H3"/>
        </svg>
      </div>
      <h3 class="text-base font-semibold text-zinc-200">No playlists added</h3>
      <p class="text-xs text-zinc-400 max-w-sm mt-1 mb-6">
        Add YouTube Music playlists or your Liked Songs to begin archiving and streaming.
      </p>
      <button
        @click="$emit('open-add-playlist')"
        class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-lg active:scale-95"
      >
        + Add Your First Playlist
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="pl in playlistsStore.playlists"
        :key="pl.id"
        class="bg-studio-surface border border-studio-border hover:border-zinc-700/80 rounded-xl p-5 flex flex-col justify-between space-y-4 shadow-sm transition-all select-none"
      >
        <!-- Header: Title & Track Count -->
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-semibold text-sm text-zinc-100 truncate" :title="pl.title">
                {{ pl.title }}
              </h3>
              <span
                v-if="pl.youtube_id === 'LM' || pl.youtube_id === 'liked'"
                class="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-indigo-950 border border-indigo-800 text-indigo-300"
              >
                LIKES
              </span>
            </div>
            <p class="text-[11px] font-mono text-zinc-500 truncate mt-0.5" :title="pl.youtube_id">
              {{ pl.youtube_id }}
            </p>
          </div>

          <span class="px-2 py-0.5 rounded-full text-xs font-mono bg-studio-elevated border border-studio-border text-zinc-300 shrink-0">
            {{ pl.track_count }} tracks
          </span>
        </div>

        <!-- Badges & Status -->
        <div class="flex items-center justify-between text-xs pt-1 border-t border-studio-borderSubtle">
          <div class="flex items-center gap-2">
            <span
              v-if="pl.status === 'syncing'"
              class="flex items-center gap-1.5 text-emerald-400 text-xs font-mono font-medium animate-pulse"
            >
              <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
              Syncing
            </span>
            <span
              v-else-if="pl.status === 'error'"
              class="flex items-center gap-1.5 text-rose-400 text-xs font-mono font-medium"
              :title="pl.error_message"
            >
              <span class="w-2 h-2 rounded-full bg-rose-400"></span>
              Error
            </span>
            <span v-else class="flex items-center gap-1.5 text-zinc-400 text-xs font-mono">
              <span class="w-1.5 h-1.5 rounded-full bg-zinc-600"></span>
              Idle
            </span>
          </div>

          <div class="text-[11px] text-zinc-500 font-mono">
            <span v-if="pl.last_synced_at">Synced {{ formatTimeAgo(pl.last_synced_at) }}</span>
            <span v-else>Never synced</span>
          </div>
        </div>

        <!-- Action Footer -->
        <div class="flex items-center justify-between gap-2 pt-2 border-t border-studio-borderSubtle">
          <span class="text-[11px] text-zinc-400 font-mono">
            {{ pl.auto_sync ? `Every ${pl.sync_interval_minutes}m` : 'Manual only' }}
          </span>

          <div class="flex items-center gap-2">
            <button
              @click="triggerSync(pl)"
              :disabled="syncStore.progress.active"
              class="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700/60 transition-colors disabled:opacity-50 flex items-center gap-1.5"
            >
              <svg class="w-3 h-3 text-indigo-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
              </svg>
              <span>Sync Now</span>
            </button>

            <button
              @click="deletePlaylist(pl)"
              class="p-1.5 rounded-lg text-zinc-500 hover:text-rose-400 hover:bg-studio-hover transition-colors"
              title="Delete Playlist"
            >
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { usePlaylistsStore } from '../stores/playlists'
import { useSyncStore } from '../stores/sync'
import { useToastStore } from '../stores/toast'

defineEmits(['open-add-playlist'])

const playlistsStore = usePlaylistsStore()
const syncStore = useSyncStore()
const toast = useToastStore()

onMounted(() => {
  playlistsStore.fetchPlaylists()
})

async function triggerSync(pl) {
  const ok = await playlistsStore.triggerSync(pl.id)
  if (ok) {
    toast.success(`Sync scheduled for ${pl.title}`)
  } else {
    toast.error('Failed to trigger sync')
  }
}

async function deletePlaylist(pl) {
  if (confirm(`Remove "${pl.title}" playlist subscription? (Existing downloaded tracks will remain intact in your library)`)) {
    const ok = await playlistsStore.deletePlaylist(pl.id)
    if (ok) {
      toast.success(`Removed ${pl.title}`)
    } else {
      toast.error('Failed to remove playlist')
    }
  }
}

function formatTimeAgo(dateStr) {
  if (!dateStr) return 'never'
  const diffSec = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diffSec < 60) return 'just now'
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`
  return `${Math.floor(diffSec / 86400)}d ago`
}
</script>

