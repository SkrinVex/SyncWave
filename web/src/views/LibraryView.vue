<template>
  <div
    class="space-y-6 pb-32 text-left relative"
    @dragover.prevent="onDragOver"
    @dragleave.prevent="onDragLeave"
    @drop.prevent="onDrop"
  >
    <!-- Drag and Drop Fullscreen Dropzone Overlay -->
    <div
      v-if="isDraggingOver"
      class="fixed inset-0 z-50 bg-indigo-950/80 backdrop-blur-md border-4 border-dashed border-indigo-500 flex flex-col items-center justify-center pointer-events-none animate-pulse select-none"
    >
      <div class="w-20 h-20 rounded-3xl bg-indigo-600/30 border border-indigo-400 flex items-center justify-center text-indigo-300 mb-4 shadow-2xl">
        <svg class="w-10 h-10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
        </svg>
      </div>
      <h3 class="text-xl font-bold text-white">{{ i18n.t('library.dragDropTitle') }}</h3>
      <p class="text-xs text-indigo-200 mt-2">MP3, FLAC, M4A, Opus, Ogg, WAV, AAC</p>
    </div>

    <!-- Hidden file input for file selection -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      accept="audio/*,.mp3,.m4a,.flac,.opus,.ogg,.wav,.aac,.webm,.wma"
      class="hidden"
      @change="onFilesSelected"
    />

    <!-- Header: Title & Action Controls -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 select-none">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-zinc-100">{{ i18n.t('library.title') }}</h2>
        <p class="text-xs text-zinc-400 mt-1">
          {{ i18n.t('library.tracksCount') }} <span class="font-mono text-zinc-200">{{ tracksStore.total }}</span>
          <span v-if="tracksStore.searchQuery"> {{ i18n.t('library.matching') }} "<strong class="text-indigo-400">{{ tracksStore.searchQuery }}</strong>"</span>
        </p>
      </div>

      <!-- Action Buttons & View Mode -->
      <div class="flex items-center gap-2.5 flex-wrap">
        <!-- Upload Music Button -->
        <button
          @click="triggerUploadDialog"
          class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white transition-all shadow-md active:scale-95"
          :title="i18n.t('library.uploadBtn')"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
          </svg>
          <span>{{ i18n.t('library.uploadBtn') }}</span>
        </button>

        <!-- Play All -->
        <button
          v-if="tracksStore.tracks.length > 0"
          @click="playAll(false)"
          class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md active:scale-95"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8 5v14l11-7z"/>
          </svg>
          <span>{{ i18n.t('library.playAll') }}</span>
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
          <span>{{ i18n.t('library.shuffle') }}</span>
        </button>

        <!-- Select All toggle button -->
        <button
          v-if="tracksStore.tracks.length > 0"
          @click="toggleSelectAll"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium bg-studio-elevated hover:bg-studio-hover text-zinc-300 border border-studio-border transition-all"
        >
          <input
            type="checkbox"
            :checked="isAllSelected"
            class="w-3.5 h-3.5 rounded border-zinc-700 bg-zinc-900 text-indigo-600 focus:ring-0 cursor-pointer pointer-events-none"
          />
          <span>{{ isAllSelected ? 'Снять выбор' : 'Выбрать все' }}</span>
        </button>

        <!-- Sort Dropdown -->
        <div class="flex items-center rounded-lg bg-studio-elevated border border-studio-border p-1 text-xs">
          <select
            v-model="tracksStore.sortBy"
            @change="tracksStore.fetchTracks(true)"
            class="bg-transparent text-zinc-300 text-xs px-2 py-1 focus:outline-none cursor-pointer"
          >
            <option value="created_at" class="bg-zinc-900">{{ i18n.t('library.newest') }}</option>
            <option value="title" class="bg-zinc-900">{{ i18n.t('library.sortTitle') }}</option>
            <option value="artist" class="bg-zinc-900">{{ i18n.t('library.sortArtist') }}</option>
            <option value="duration" class="bg-zinc-900">{{ i18n.t('library.sortDuration') }}</option>
          </select>
          <button
            @click="toggleSortOrder"
            class="px-1.5 py-1 text-zinc-400 hover:text-zinc-100 transition-colors"
            title="Порядок сортировки"
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
            title="Сетка"
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
            title="Список"
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
        {{ i18n.t('library.allTracks') }} ({{ tracksStore.stats.total_tracks }})
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
      <p class="text-xs font-mono">Загрузка треков...</p>
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
      <h3 class="text-base font-semibold text-zinc-200">{{ i18n.t('library.emptyTitle') }}</h3>
      <p class="text-xs text-zinc-400 max-w-sm mt-1 mb-6">
        {{ i18n.t('library.emptyDesc') }}
      </p>
      <div class="flex items-center gap-3 flex-wrap justify-center">
        <button
          @click="triggerUploadDialog"
          class="px-4 py-2 text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg transition-all shadow-lg active:scale-95 flex items-center gap-2"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
          </svg>
          <span>{{ i18n.t('library.uploadFirstBtn') }}</span>
        </button>
        <button
          @click="$emit('open-add-playlist')"
          class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-lg active:scale-95"
        >
          {{ i18n.t('library.addPlaylistBtn') }}
        </button>
      </div>
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
        :selected="selectedTrackIds.has(track.id)"
        @play="handlePlay"
        @delete="promptDeleteTrack"
        @toggle-select="toggleTrackSelect"
      />
    </div>

    <!-- Table View Layout -->
    <div v-else class="bg-studio-surface border border-studio-border rounded-xl overflow-hidden p-2 space-y-1">
      <div class="flex items-center gap-3 md:gap-4 px-3 md:px-4 py-2 text-[11px] font-mono text-zinc-500 uppercase tracking-wider select-none border-b border-studio-borderSubtle">
        <span class="w-4"></span>
        <span class="hidden md:block w-6 text-center">#</span>
        <span class="w-10"></span>
        <span class="flex-1">Название и Исполнитель</span>
        <span class="hidden md:block w-1/4">Альбом</span>
        <span class="hidden lg:block">Формат</span>
        <span class="w-12 text-right">Время</span>
        <span class="hidden md:block w-20"></span>
      </div>
      <TrackRow
        v-for="(track, idx) in tracksStore.tracks"
        :key="track.id"
        :track="track"
        :index="(tracksStore.page - 1) * tracksStore.pageSize + idx + 1"
        :selected="selectedTrackIds.has(track.id)"
        @play="handlePlay"
        @delete="promptDeleteTrack"
        @toggle-select="toggleTrackSelect"
      />
    </div>

    <!-- Floating Batch Actions Bar -->
    <transition
      enter-active-class="transition-all ease-out duration-200"
      enter-from-class="translate-y-16 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition-all ease-in duration-150"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-16 opacity-0"
    >
      <div
        v-if="selectedTrackIds.size > 0"
        class="fixed bottom-24 left-1/2 -translate-x-1/2 z-40 bg-studio-surface/95 backdrop-blur-md border border-zinc-700/80 rounded-2xl shadow-2xl px-5 py-3 flex items-center gap-4 text-xs select-none"
      >
        <div class="flex items-center gap-2 font-mono text-zinc-200">
          <span class="w-2 h-2 rounded-full bg-indigo-500"></span>
          <span>Выбрано: <strong>{{ selectedTrackIds.size }}</strong></span>
        </div>

        <button
          @click="clearSelection"
          class="px-2.5 py-1 text-zinc-400 hover:text-zinc-200 transition-colors"
        >
          Снять выбор
        </button>

        <button
          @click="showBatchDeleteConfirm = true"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-red-600 hover:bg-red-500 text-white font-medium shadow-md transition-colors"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
          <span>Удалить выбранные ({{ selectedTrackIds.size }})</span>
        </button>
      </div>
    </transition>

    <!-- Pagination Controls -->
    <div
      v-if="tracksStore.totalPages > 1"
      class="flex items-center justify-between pt-4 border-t border-studio-border text-xs font-mono select-none"
    >
      <span class="text-zinc-400">
        {{ i18n.t('library.page') }} {{ tracksStore.page }} {{ i18n.t('library.of') }} {{ tracksStore.totalPages }}
      </span>
      <div class="flex items-center gap-2">
        <button
          @click="changePage(tracksStore.page - 1)"
          :disabled="tracksStore.page <= 1"
          class="px-3 py-1.5 rounded bg-studio-elevated border border-studio-border text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          {{ i18n.t('library.previous') }}
        </button>
        <button
          @click="changePage(tracksStore.page + 1)"
          :disabled="tracksStore.page >= tracksStore.totalPages"
          class="px-3 py-1.5 rounded bg-studio-elevated border border-studio-border text-zinc-300 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          {{ i18n.t('library.next') }}
        </button>
      </div>
    </div>

    <!-- Custom Confirm Modal for Single Delete Track -->
    <ConfirmModal
      :open="!!trackToDelete"
      :title="i18n.t('confirm.deleteTrackTitle')"
      :description="i18n.t('confirm.deleteTrackDesc', { title: trackToDelete?.title || '' })"
      :confirm-text="i18n.t('confirm.delete')"
      :cancel-text="i18n.t('confirm.cancel')"
      :danger="true"
      @confirm="confirmDeleteTrack"
      @cancel="trackToDelete = null"
    />

    <!-- Custom Confirm Modal for Batch Delete -->
    <ConfirmModal
      :open="showBatchDeleteConfirm"
      title="Массовое удаление треков"
      :description="`Вы действительно хотите удалить ${selectedTrackIds.size} выбранных треков? Файлы будут удалены с сервера.`"
      confirm-text="Удалить выбранные"
      cancel-text="Отмена"
      :danger="true"
      @confirm="confirmBatchDelete"
      @cancel="showBatchDeleteConfirm = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useTracksStore } from '../stores/tracks'
import { usePlaylistsStore } from '../stores/playlists'
import { usePlayerStore } from '../stores/player'
import { useUploadStore } from '../stores/upload'
import { useToastStore } from '../stores/toast'
import { useI18nStore } from '../stores/i18n'
import TrackCard from '../components/TrackCard.vue'
import TrackRow from '../components/TrackRow.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

defineEmits(['open-add-playlist'])

const tracksStore = useTracksStore()
const playlistsStore = usePlaylistsStore()
const playerStore = usePlayerStore()
const uploadStore = useUploadStore()
const toast = useToastStore()
const i18n = useI18nStore()

const viewMode = ref(localStorage.getItem('syncwave_view_mode') || 'grid')
const trackToDelete = ref(null)
const selectedTrackIds = ref(new Set())
const showBatchDeleteConfirm = ref(false)
const fileInputRef = ref(null)
const isDraggingOver = ref(false)

const isAllSelected = computed(() => {
  return tracksStore.tracks.length > 0 && tracksStore.tracks.every(t => selectedTrackIds.value.has(t.id))
})

onMounted(() => {
  tracksStore.fetchTracks()
  tracksStore.fetchStats()
  playlistsStore.fetchPlaylists()
})

function triggerUploadDialog() {
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
    fileInputRef.value.click()
  }
}

function onFilesSelected(e) {
  const files = e.target.files
  if (files && files.length > 0) {
    uploadStore.uploadFiles(files, tracksStore.selectedPlaylist)
  }
}

function onDragOver(e) {
  isDraggingOver.value = true
}

function onDragLeave(e) {
  if (!e.currentTarget.contains(e.relatedTarget)) {
    isDraggingOver.value = false
  }
}

function onDrop(e) {
  isDraggingOver.value = false
  if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
    uploadStore.uploadFiles(e.dataTransfer.files, tracksStore.selectedPlaylist)
  }
}

function toggleTrackSelect(track) {
  const newSet = new Set(selectedTrackIds.value)
  if (newSet.has(track.id)) {
    newSet.delete(track.id)
  } else {
    newSet.add(track.id)
  }
  selectedTrackIds.value = newSet
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedTrackIds.value = new Set()
  } else {
    const newSet = new Set()
    tracksStore.tracks.forEach(t => newSet.add(t.id))
    selectedTrackIds.value = newSet
  }
}

function clearSelection() {
  selectedTrackIds.value = new Set()
}

async function confirmBatchDelete() {
  showBatchDeleteConfirm.value = false
  const ids = Array.from(selectedTrackIds.value)
  if (ids.length === 0) return

  const ok = await tracksStore.batchDelete(ids)
  if (ok) {
    toast.success(`Удалено треков: ${ids.length}`)
    selectedTrackIds.value = new Set()
  } else {
    toast.error('Не удалось удалить выбранные треки')
  }
}

function selectPlaylist(id) {
  tracksStore.selectedPlaylist = id
  selectedTrackIds.value = new Set()
  tracksStore.fetchTracks(true)
}

function toggleSortOrder() {
  tracksStore.sortOrder = tracksStore.sortOrder === 'asc' ? 'desc' : 'asc'
  tracksStore.fetchTracks(true)
}

function changePage(p) {
  tracksStore.page = p
  selectedTrackIds.value = new Set()
  tracksStore.fetchTracks()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function handlePlay(track) {
  // Fetch the full list of tracks in the current scope (selected playlist or entire library)
  const fullReadyTracks = await tracksStore.fetchAllReadyTracks(tracksStore.selectedPlaylist)
  const queueToUse = fullReadyTracks.length > 0 ? fullReadyTracks : tracksStore.tracks
  playerStore.playTrack(track, queueToUse)
}

async function playAll(shuffle = false) {
  const fullReadyTracks = await tracksStore.fetchAllReadyTracks(tracksStore.selectedPlaylist)
  const queueToUse = fullReadyTracks.length > 0 ? fullReadyTracks : tracksStore.tracks
  if (queueToUse.length === 0) return

  if (shuffle) {
    playerStore.isShuffle = true
    const rndIndex = Math.floor(Math.random() * queueToUse.length)
    playerStore.playTrack(queueToUse[rndIndex], queueToUse)
  } else {
    playerStore.isShuffle = false
    playerStore.playTrack(queueToUse[0], queueToUse)
  }
}

function promptDeleteTrack(track) {
  trackToDelete.value = track
}

async function confirmDeleteTrack() {
  if (!trackToDelete.value) return
  const track = trackToDelete.value
  trackToDelete.value = null
  const ok = await tracksStore.deleteTrack(track.id)
  if (ok) {
    selectedTrackIds.value.delete(track.id)
    toast.success(i18n.currentLang === 'ru' ? `Удален трек: ${track.title}` : `Deleted ${track.title}`)
  } else {
    toast.error('Не удалось удалить трек')
  }
}
</script>
