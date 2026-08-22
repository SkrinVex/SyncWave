<template>
  <div class="min-h-screen bg-studio-bg text-zinc-100 flex flex-col font-sans antialiased selection:bg-indigo-500/30">
    <!-- Toast Notifications -->
    <Toast />

    <!-- Unauthenticated Flow -->
    <LoginView v-if="!authStore.isAuthenticated" />

    <!-- Authenticated Studio Dashboard -->
    <div v-else class="flex h-screen overflow-hidden">
      <!-- Desktop Sidebar -->
      <Sidebar
        :current-view="currentView"
        @change-view="currentView = $event"
        class="hidden md:flex"
      />

      <!-- Mobile Sidebar Drawer -->
      <div
        v-if="mobileMenuOpen"
        class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 md:hidden flex"
        @click="mobileMenuOpen = false"
      >
        <Sidebar
          :current-view="currentView"
          @change-view="currentView = $event; mobileMenuOpen = false"
          @click.stop
        />
      </div>

      <!-- Main Content Area -->
      <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
        <!-- Top Navbar -->
        <Navbar
          @toggle-mobile-menu="mobileMenuOpen = !mobileMenuOpen"
          @open-sync="currentView = 'sync'"
        />

        <!-- Active Sync Progress Banner -->
        <SyncProgressBar />

        <!-- View Body -->
        <main class="flex-1 overflow-y-auto px-4 md:px-8 py-6">
          <LibraryView
            v-if="currentView === 'library'"
            @open-add-playlist="showAddPlaylist = true"
          />
          <PlaylistsView
            v-else-if="currentView === 'playlists'"
            @open-add-playlist="showAddPlaylist = true"
          />
          <SyncView
            v-else-if="currentView === 'sync'"
          />
          <SettingsView
            v-else-if="currentView === 'settings'"
          />
        </main>
      </div>

      <!-- Persistent Audio Player -->
      <Player />

      <!-- Queue Drawer -->
      <QueueDrawer />

      <!-- Add Playlist Modal -->
      <AddPlaylistModal
        v-if="showAddPlaylist"
        @close="showAddPlaylist = false"
        @created="currentView = 'playlists'"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useAuthStore } from './stores/auth'
import { useSyncStore } from './stores/sync'
import { usePlayerStore } from './stores/player'

import Toast from './components/Toast.vue'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import Player from './components/Player.vue'
import QueueDrawer from './components/QueueDrawer.vue'
import SyncProgressBar from './components/SyncProgressBar.vue'
import AddPlaylistModal from './components/AddPlaylistModal.vue'

import LibraryView from './views/LibraryView.vue'
import PlaylistsView from './views/PlaylistsView.vue'
import SyncView from './views/SyncView.vue'
import SettingsView from './views/SettingsView.vue'
import LoginView from './views/LoginView.vue'

const authStore = useAuthStore()
const syncStore = useSyncStore()
const playerStore = usePlayerStore()

const currentView = ref('library')
const mobileMenuOpen = ref(false)
const showAddPlaylist = ref(false)

onMounted(async () => {
  await authStore.checkStatus()
  if (authStore.isAuthenticated) {
    syncStore.connectSSE()
  }
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  syncStore.disconnectSSE()
  window.removeEventListener('keydown', handleGlobalKeydown)
})

watch(() => authStore.isAuthenticated, (isAuthed) => {
  if (isAuthed) {
    syncStore.connectSSE()
  } else {
    syncStore.disconnectSSE()
  }
})

function handleGlobalKeydown(e) {
  const tag = e.target.tagName.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return

  switch (e.code) {
    case 'Space':
      e.preventDefault()
      playerStore.togglePlay()
      break
    case 'ArrowLeft':
      e.preventDefault()
      playerStore.seek(playerStore.currentTime - 5)
      break
    case 'ArrowRight':
      e.preventDefault()
      playerStore.seek(playerStore.currentTime + 5)
      break
    case 'ArrowUp':
      e.preventDefault()
      playerStore.setVolume(playerStore.volume + 0.05)
      break
    case 'ArrowDown':
      e.preventDefault()
      playerStore.setVolume(playerStore.volume - 0.05)
      break
    case 'KeyM':
      playerStore.toggleMute()
      break
    case 'KeyL':
      playerStore.toggleLoop()
      break
    case 'KeyS':
      playerStore.toggleShuffle()
      break
  }
}
</script>

