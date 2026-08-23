<template>
  <div class="h-screen w-screen flex flex-col bg-studio-bg text-zinc-100 overflow-hidden font-sans select-none antialiased">
    <!-- Auth Screen when unauthenticated -->
    <LoginView v-if="!authStore.isAuthenticated" />

    <!-- Main Dashboard Application -->
    <template v-else>
      <div
        class="flex-1 flex overflow-hidden transition-all duration-300"
        :class="{ 'pb-20': playerStore.currentTrack }"
      >
        <!-- Desktop Sidebar -->
        <Sidebar
          class="hidden md:flex"
          :current-view="currentView"
          @change-view="currentView = $event"
          @request-sign-out="showSignOutConfirm = true"
        />

        <!-- Mobile Drawer Backdrop -->
        <div
          v-if="mobileMenuOpen"
          class="fixed inset-0 bg-black/80 z-50 md:hidden backdrop-blur-sm"
          @click="mobileMenuOpen = false"
        >
          <div class="w-64 h-full" @click.stop>
            <Sidebar
              :current-view="currentView"
              @change-view="onMobileChangeView"
              @request-sign-out="showSignOutConfirm = true; mobileMenuOpen = false"
            />
          </div>
        </div>

        <!-- Main Content Area -->
        <main class="flex-1 flex flex-col min-w-0 bg-studio-bg overflow-hidden relative">
          <!-- Top Navigation Bar -->
          <Navbar
            @toggle-mobile-menu="mobileMenuOpen = !mobileMenuOpen"
            @open-sync="currentView = 'sync'"
          />

          <!-- Real-Time Sync Banner -->
          <SyncProgressBar />

          <!-- Dynamic Views Container -->
          <div class="flex-1 overflow-y-auto px-4 md:px-8 py-6">
            <div class="max-w-7xl mx-auto">
              <LibraryView
                v-show="currentView === 'library'"
                @open-add-playlist="isAddPlaylistOpen = true"
              />
              <PlaylistsView
                v-show="currentView === 'playlists'"
                @open-add-playlist="isAddPlaylistOpen = true"
              />
              <SyncView v-show="currentView === 'sync'" />
              <SettingsView v-show="currentView === 'settings'" />
            </div>
          </div>
        </main>
      </div>

      <!-- Persistent Audio Player Bar -->
      <Player />

      <!-- Slide-over Play Queue Drawer -->
      <QueueDrawer />

      <!-- Add YouTube Playlist Modal -->
      <AddPlaylistModal
        v-if="isAddPlaylistOpen"
        @close="isAddPlaylistOpen = false"
        @created="onPlaylistCreated"
      />

      <!-- Custom Confirm Modal for Sign Out -->
      <ConfirmModal
        :open="showSignOutConfirm"
        :title="i18n.t('confirm.signOutTitle')"
        :description="i18n.t('confirm.signOutDesc')"
        :confirm-text="i18n.t('nav.signOut')"
        :cancel-text="i18n.t('confirm.cancel')"
        :danger="true"
        @confirm="confirmSignOut"
        @cancel="showSignOutConfirm = false"
      />
    </template>

    <!-- Global Toast Notifications -->
    <Toast />
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { usePlayerStore } from './stores/player'
import { useTracksStore } from './stores/tracks'
import { usePlaylistsStore } from './stores/playlists'
import { useSettingsStore } from './stores/settings'
import { useSyncStore } from './stores/sync'
import { useI18nStore } from './stores/i18n'
import LoginView from './views/LoginView.vue'
import LibraryView from './views/LibraryView.vue'
import PlaylistsView from './views/PlaylistsView.vue'
import SyncView from './views/SyncView.vue'
import SettingsView from './views/SettingsView.vue'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import Player from './components/Player.vue'
import QueueDrawer from './components/QueueDrawer.vue'
import AddPlaylistModal from './components/AddPlaylistModal.vue'
import SyncProgressBar from './components/SyncProgressBar.vue'
import ConfirmModal from './components/ConfirmModal.vue'
import Toast from './components/Toast.vue'

const authStore = useAuthStore()
const playerStore = usePlayerStore()
const tracksStore = useTracksStore()
const playlistsStore = usePlaylistsStore()
const settingsStore = useSettingsStore()
const syncStore = useSyncStore()
const i18n = useI18nStore()

const currentView = ref('library')
const mobileMenuOpen = ref(false)
const isAddPlaylistOpen = ref(false)
const showSignOutConfirm = ref(false)

watch(currentView, (newView) => {
  if (newView === 'playlists') {
    playlistsStore.fetchPlaylists()
  } else if (newView === 'library') {
    tracksStore.fetchTracks()
    tracksStore.fetchStats()
  } else if (newView === 'sync') {
    syncStore.fetchLogs()
    syncStore.fetchInitialProgress()
  } else if (newView === 'settings') {
    settingsStore.fetchSettings()
  }
})

watch(() => authStore.isAuthenticated, (isAuth) => {
  if (isAuth) {
    syncStore.connectSSE()
    playlistsStore.fetchPlaylists()
    tracksStore.fetchTracks()
    tracksStore.fetchStats()
  } else {
    syncStore.disconnectSSE()
  }
}, { immediate: true })

onMounted(async () => {
  await authStore.checkStatus()
  window.addEventListener('keydown', handleGlobalHotkeys)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalHotkeys)
})

function onMobileChangeView(v) {
  currentView.value = v
  mobileMenuOpen.value = false
  if (v === 'playlists') playlistsStore.fetchPlaylists()
  if (v === 'library') tracksStore.fetchTracks()
}

function onPlaylistCreated() {
  currentView.value = 'playlists'
  playlistsStore.fetchPlaylists()
  tracksStore.fetchTracks(true)
}

function confirmSignOut() {
  showSignOutConfirm.value = false
  authStore.logout()
}

function handleGlobalHotkeys(e) {
  // Ignore hotkeys when typing in form inputs
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName)) {
    return
  }

  switch (e.code) {
    case 'Space':
      e.preventDefault()
      playerStore.togglePlay()
      break
    case 'ArrowLeft':
      e.preventDefault()
      playerStore.seek(Math.max(0, playerStore.currentTime - 5))
      break
    case 'ArrowRight':
      e.preventDefault()
      playerStore.seek(Math.min(playerStore.duration, playerStore.currentTime + 5))
      break
    case 'ArrowUp':
      e.preventDefault()
      playerStore.setVolume(Math.min(1, playerStore.volume + 0.05))
      break
    case 'ArrowDown':
      e.preventDefault()
      playerStore.setVolume(Math.max(0, playerStore.volume - 0.05))
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
