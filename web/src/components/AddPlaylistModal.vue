<template>
  <div class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="bg-studio-surface border border-studio-border rounded-xl max-w-md w-full p-6 shadow-2xl space-y-5">
      <!-- Modal Header -->
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-zinc-100">Add YouTube Playlist</h3>
          <p class="text-xs text-zinc-400 mt-0.5">Subscribe to automatic syncing & archiving</p>
        </div>
        <button
          @click="$emit('close')"
          class="text-zinc-400 hover:text-zinc-100 p-1.5 rounded-lg hover:bg-studio-hover transition-colors"
        >
          ✕
        </button>
      </div>

      <!-- Quick Preset: Liked Songs -->
      <div class="p-3 rounded-lg bg-studio-elevated border border-studio-borderSubtle flex items-center justify-between">
        <div>
          <h4 class="text-xs font-semibold text-zinc-200">Liked Music (Auto)</h4>
          <p class="text-[11px] text-zinc-400">Sync all your liked songs (requires cookies.txt)</p>
        </div>
        <button
          type="button"
          @click="fillLikedMusic"
          class="px-2.5 py-1 text-xs font-medium bg-indigo-600/30 hover:bg-indigo-600/50 border border-indigo-500/40 text-indigo-200 rounded-md transition-colors"
        >
          Use Preset
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">
            Playlist URL or ID <span class="text-indigo-400">*</span>
          </label>
          <input
            type="text"
            v-model="urlOrId"
            required
            placeholder="e.g. https://music.youtube.com/playlist?list=PL... or LM"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60 font-mono text-xs"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">
            Custom Title <span class="text-zinc-500">(optional, auto-extracted if empty)</span>
          </label>
          <input
            type="text"
            v-model="title"
            placeholder="e.g. Synthwave Favorites"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
          />
        </div>

        <div class="grid grid-cols-2 gap-3 pt-1">
          <div>
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">Sync Interval</label>
            <select
              v-model="syncInterval"
              class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
            >
              <option :value="15">Every 15 min</option>
              <option :value="30">Every 30 min</option>
              <option :value="60">Every 1 hour</option>
              <option :value="360">Every 6 hours</option>
              <option :value="720">Every 12 hours</option>
              <option :value="1440">Every 24 hours</option>
            </select>
          </div>

          <div class="flex items-center justify-between p-2.5 rounded-lg bg-studio-elevated border border-studio-border mt-auto">
            <div>
              <span class="text-xs font-medium text-zinc-200 block leading-tight">Auto-Sync</span>
              <span class="text-[10px] text-zinc-400">Background cron</span>
            </div>
            <input
              type="checkbox"
              v-model="autoSync"
              class="w-4 h-4 rounded text-indigo-600 bg-zinc-800 border-zinc-700 focus:ring-indigo-500 focus:ring-offset-0"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-3">
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 rounded-lg hover:bg-studio-hover transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="loading"
            class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors shadow-md disabled:opacity-50 flex items-center gap-2"
          >
            <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>{{ loading ? 'Adding...' : 'Add Playlist' }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { usePlaylistsStore } from '../stores/playlists'
import { useToastStore } from '../stores/toast'

const emit = defineEmits(['close', 'created'])
const playlistsStore = usePlaylistsStore()
const toast = useToastStore()

const urlOrId = ref('')
const title = ref('')
const autoSync = ref(true)
const syncInterval = ref(60)
const loading = ref(false)

function fillLikedMusic() {
  urlOrId.value = 'LM'
  title.value = 'Liked Music'
}

async function handleSubmit() {
  if (!urlOrId.value) return
  loading.value = true
  try {
    const pl = await playlistsStore.createPlaylist(
      title.value,
      urlOrId.value,
      autoSync.value,
      syncInterval.value
    )
    toast.success(`Added playlist: ${pl.title}`)
    emit('created', pl)
    emit('close')
  } catch (e) {
    toast.error(e.message || 'Failed to create playlist')
  } finally {
    loading.value = false
  }
}
</script>

