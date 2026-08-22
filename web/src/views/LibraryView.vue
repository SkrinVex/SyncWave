<template>
  <div class="space-y-6 pb-28">
    <!-- Header: Title & Action Controls -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 select-none">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-zinc-100">Music Library</h2>
        <p class="text-xs text-zinc-400 mt-1">
          Showing <span class="font-mono text-zinc-200">{{ tracksStore.total }}</span> archived tracks
          <span v-if="tracksStore.searchQuery"> matching "<strong class="text-indigo-400">{{ tracksStore.searchQuery }}</strong>"</span>
        </p>
      </div>

      <!-- Action Buttons & View Mode -->
      <div class="flex items-center gap-2.5 flex-wrap">
        <!-- Play All -->
        <button
          v-if="tracksStore.tracks.length > 0"
          @click="playAll(false)"
          class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md active:scale-95"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8 5v14l11-7z"/>
          </svg>
          <span>Play All</span>
        </button>

        <!-- Shuffle All -->
        <button
          v-if="tracksStore.tracks.length > 0"
          @click="playAll(true)"
          class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-studio-elevated hover:bg-studio-hover text-zinc-200 border border-studio-border transition-all active:scale-95"
        >
          <svg class="w-3.5 h-3.5 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M2 18h1.4c1.3 0 2.5-.6 3.3-1.7l6.6-8.6c.8-1.1 2-1.7 3.3-1.7H22M2 6h1.4c1.3 0 2.5.6 3.3 1.7l2 2.6M22 18h-5.4c-1.3 0-2.5-.6-3.3-1.7l-2-2.6M18 14l4 4-4 4M18 2l4 4-4 4"/>
          </svg>
          <span>Shuffle</span>
        </button>

        <!-- Sort Dropdown -->
        <div class="flex items-center rounded-lg bg-studio-elevated border border-studio-border p-1 text-xs">
          <select
            v-model="tracksStore.sortBy"
            @change="tracksStore.fetchTracks(true)"
            class="bg-transparent text-zinc-300 text-xs px-2 py-1 focus:outline-none cursor-pointer"
          >
            <option value="created_at" class="bg-zinc-900">Newest Added</option>
            <option value="title" class="bg-zinc-900">Title</option>
            <option value="artist" class="bg-zinc-900">Artist</option>
            <option value="duration" class="bg-zinc-900">Duration</option>
          </select>
          <button
            @click="toggleSortOrder"
            class="px-1.5 py-1 text-zinc-400 hover:text-zinc-100 transition-colors"
            title="Toggle Asc/Desc"
          >
            {{ tracksStore.sortOrder === 'asc' ? '↑' : '↓' }}
          </button>
        </div>

        <!-- Grid vs Table View Mode -->
        <div class="flex items-center rounded-lg bg-studio-elevated border border-studio-border p-1">
          <button
            @click="viewMode = 'grid'"
            :class="[
              'p-1.5 rounded transition-colors',
              viewMode === 'grid' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200'
            ]"
            title="Grid View"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>
            </svg>
          </button>
          <button
            @click="viewMode = 'table'"
            :class="[
              'p-1.5 rounded transition-colors',
              viewMode === 'table' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200'
            ]"
            title="Table View"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Playlist Filter Pills -->
    <div v-if="playlistsStore.playlists.length > 0" class="flex items-center gap-2 overflow-x-auto pb-1 text-xs select-none">
      <button
        @click="selectPlaylist('')"
        :class="[
          'px-3 py-1.5 rounded-full font-medium transition-all whitespace-nowrap',
          tracksStore.selectedPlaylist === ''
            ? 'bg-zinc-100 text-zinc-950 shadow-sm'
            : 'bg-studio-elevated text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover border border-studio-border'
        ]"
      >
        All Tracks ({{ tracksStore.stats.total_tracks }})
      </button>
      <button
        v-for="pl in playlistsStore.playlists"
        :key="pl.id"
        @click="selectPlaylist(pl.id)"
        :class="[
          'px-3 py-1.5 rounded-full font-medium transition-all whitespace-nowrap flex items-center gap-1.5',
          tracksStore.selectedPlaylist === pl.id
            ? 'bg-zinc-100 text-zinc-950 shadow-sm'
            : 'bg-studio-elevated text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover border border-studio-border'
        ]"
      >
        <span>{{ pl.title }}</span>
        <span class="text-[10px] opacity-70">({{ pl.track_count }})</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="tracksStore.loading" class="py-20 flex flex-col items-center justify-center gap-3 text-zinc-400">
      <div class="w-8 h-8 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
      <p class="text-xs font-mono">Loading library...</p>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="tracksStore.tracks.length === 0"
      class="py-20 border border-dashed border-zinc-800 rounded-2xl flex flex-col items-center justify-center text-center p-8 select-none"
    >
      <div class="w-14 h-14 rounded-2xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mb-4">
        <svg class="w-7 h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
        </svg>
      </div>
      <h3 class="text-base font-semibold text-zinc-200">No tracks found</h3>
      <p class="text-xs text-zinc-400 max-w-sm mt-1 mb-6">
        Add a YouTube Music playlist or Liked Songs to automatically sync tracks to your server.
      </p>
      <button
        @click="$emit('open-add-playlist')"
        class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-lg active:scale-95"
      >
        + Add Playlist
      </button>
    </div>

    <!-- Grid View Layout -->
    <div
      v-else-if="viewMode === 'grid'"
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4"
    >
      <TrackCard
        v-for="track in tracksStore.tracks"
        :key="track.id"
        :track="track"
        @play="handlePlay"
        @delete="handleDelete"
      />
    </div>

    <!-- Table View Layout -->
    <div v-else class="bg-studio-surface border border-studio-border rounded-xl overflow-hidden p-2 space-y-1">
      <div class="flex items-center gap-4 px-4 py-2 text-[11px] font-mono text-zinc-500 uppercase tracking-wider select-none border-b border-studio-borderSubtle">
        <span class="w-6 text-center">#</span>
        <span class="w-10"></span>
        <span class="flex-1">Title & Artist</span>
        <span class="hidden md:block w-1/4">Album</span>
        <span class="hidden lg:block">Format</span>
        <span class="w-12 text-right">Time</span>
        <span class="w-20"></span>
      </div>
      <TrackRow
        v-for="(track, idx) in tracksStore.tracks"
        :key="track.id"
        :track="track"
        :index="(tracksStore.page - 1) * tracksStore.pageSize + idx + 1"
        @play="handlePlay"
        @delete="handleDelete"
      />
    </div>

    <!-- Pagination Controls -->
    <div
      v-if="tracksStore.totalPages > 1"
      class="flex items-center justify-between pt-4 border-t border-studio-border text-xs font-mono select-none"
    >
      <span class="text-zinc-400">
        Page {{ tracksStore.page }} of {{ tracksStore.totalPages }}
      </span>
      <div class="flex items-center gap-2">
        <button
          @click="changePage(tracksStore.page - 1)"
          :disabled="tracksStore.page <= 1"
          class="px-3 py-1.5 rounded bg-studio-elevated border border-studio-border text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Previous
        </button>
        <button
          @click="changePage(tracksStore.page + 1)"
          :disabled="tracksStore.page >= tracksStore.totalPages"
          class="px-3 py-1.5 rounded bg-studio-elevated border border-studio-border text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTracksStore } from '../stores/tracks'
import { usePlaylistsStore } from '../stores/playlists'
import { usePlayerStore } from '../stores/player'
import { useToastStore } from '../stores/toast'
import TrackCard from '../components/TrackCard.vue'
import TrackRow from '../components/TrackRow.vue'

defineEmits(['open-add-playlist'])

const tracksStore = useTracksStore()
const playlistsStore = usePlaylistsStore()
const playerStore = usePlayerStore()
const toast = useToastStore()

const viewMode = ref(localStorage.getItem('syncwave_view_mode') || 'grid')

onMounted(() => {
  tracksStore.fetchTracks()
  tracksStore.fetchStats()
  playlistsStore.fetchPlaylists()
})

function selectPlaylist(id) {
  tracksStore.selectedPlaylist = id
  tracksStore.fetchTracks(true)
}

function toggleSortOrder() {
  tracksStore.sortOrder = tracksStore.sortOrder === 'asc' ? 'desc' : 'asc'
  tracksStore.fetchTracks(true)
}

function changePage(p) {
  tracksStore.page = p
  tracksStore.fetchTracks()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handlePlay(track) {
  playerStore.playTrack(track, tracksStore.tracks)
}

function playAll(shuffle = false) {
  if (tracksStore.tracks.length === 0) return
  if (shuffle) {
    playerStore.isShuffle = true
    const rndIndex = Math.floor(Math.random() * tracksStore.tracks.length)
    playerStore.playTrack(tracksStore.tracks[rndIndex], tracksStore.tracks)
  } else {
    playerStore.isShuffle = false
    playerStore.playTrack(tracksStore.tracks[0], tracksStore.tracks)
  }
}

async function handleDelete(track) {
  if (confirm(`Are you sure you want to delete "${track.title}"?`)) {
    const ok = await tracksStore.deleteTrack(track.id)
    if (ok) {
      toast.success(`Deleted ${track.title}`)
    } else {
      toast.error('Failed to delete track')
    }
  }
}
</script>

