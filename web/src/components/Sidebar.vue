<template>
  <aside class="w-64 bg-studio-surface border-r border-studio-border flex flex-col justify-between h-full select-none shrink-0">
    <!-- Brand / Header -->
    <div>
      <div class="h-16 flex items-center px-6 gap-3 border-b border-studio-borderSubtle">
        <div class="w-8 h-8 rounded-lg bg-studio-elevated border border-studio-border flex items-center justify-center text-indigo-400">
          <!-- Soundwave Icon -->
          <svg class="w-5 h-5" viewBox="0 0 32 32" fill="none">
            <rect x="5" y="12" width="2.5" height="8" rx="1.25" fill="currentColor" opacity="0.6"/>
            <rect x="10" y="8" width="2.5" height="16" rx="1.25" fill="currentColor" opacity="0.85"/>
            <rect x="15" y="4" width="2.5" height="24" rx="1.25" fill="#818cf8"/>
            <rect x="20" y="8" width="2.5" height="16" rx="1.25" fill="currentColor" opacity="0.85"/>
            <rect x="25" y="12" width="2.5" height="8" rx="1.25" fill="currentColor" opacity="0.6"/>
          </svg>
        </div>
        <div>
          <h1 class="text-base font-semibold tracking-tight text-zinc-100 leading-none">SyncWave</h1>
          <span class="text-[10px] text-zinc-400 font-mono tracking-wider uppercase">Self-Hosted Audio</span>
        </div>
      </div>

      <!-- Nav Links -->
      <nav class="p-3 space-y-1">
        <button
          @click="$emit('change-view', 'library')"
          :class="[
            'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all text-left',
            currentView === 'library'
              ? 'bg-studio-elevated text-zinc-100 border border-zinc-800 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover'
          ]"
        >
          <svg class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 18V5l12-2v13M9 18a3 3 0 1 1-6 0 3 3 0 0 1 6 0zm12-2a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
          </svg>
          <span>{{ i18n.t('nav.library') }}</span>
          <span v-if="tracksStore.total > 0" class="ml-auto text-xs font-mono text-zinc-400 px-1.5 py-0.5 rounded bg-zinc-800/80">
            {{ tracksStore.total }}
          </span>
        </button>

        <button
          @click="$emit('change-view', 'playlists')"
          :class="[
            'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all text-left',
            currentView === 'playlists'
              ? 'bg-studio-elevated text-zinc-100 border border-zinc-800 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover'
          ]"
        >
          <svg class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15V6M18.5 18a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5ZM12 12H3M16 6H3M12 18H3"/>
          </svg>
          <span>{{ i18n.t('nav.playlists') }}</span>
          <span v-if="playlistsStore.playlists.length > 0" class="ml-auto text-xs font-mono text-zinc-400 px-1.5 py-0.5 rounded bg-zinc-800/80">
            {{ playlistsStore.playlists.length }}
          </span>
        </button>

        <button
          @click="$emit('change-view', 'sync')"
          :class="[
            'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all text-left',
            currentView === 'sync'
              ? 'bg-studio-elevated text-zinc-100 border border-zinc-800 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover'
          ]"
        >
          <div class="relative flex items-center">
            <svg class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
            </svg>
            <span
              v-if="syncStore.progress.active"
              class="absolute -top-1 -right-1 w-2 h-2 rounded-full bg-emerald-400 animate-ping"
            ></span>
          </div>
          <span>{{ i18n.t('nav.sync') }}</span>
          <span
            v-if="syncStore.progress.active"
            class="ml-auto text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-emerald-950 border border-emerald-800 text-emerald-400 animate-pulse"
          >
            Live
          </span>
        </button>

        <button
          @click="$emit('change-view', 'settings')"
          :class="[
            'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all text-left',
            currentView === 'settings'
              ? 'bg-studio-elevated text-zinc-100 border border-zinc-800 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-studio-hover'
          ]"
        >
          <svg class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
          <span>{{ i18n.t('nav.settings') }}</span>
        </button>
      </nav>
    </div>

    <!-- Bottom Actions & User Profile -->
    <div class="p-3 border-t border-studio-borderSubtle space-y-3">
      <!-- Install PWA Button -->
      <button
        v-if="!isStandalone"
        @click="installPwa"
        class="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-lg text-xs font-medium bg-indigo-600 hover:bg-indigo-500 text-white shadow-md transition-colors md:hidden"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/>
        </svg>
        <span>Установить приложение</span>
      </button>

      <!-- Quick Sync Trigger Button -->
      <button
        @click="triggerQuickSync"
        :disabled="syncStore.progress.active"
        class="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-lg text-xs font-medium bg-zinc-800/80 hover:bg-zinc-700/80 text-zinc-200 border border-zinc-700/60 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <svg
          class="w-3.5 h-3.5 text-indigo-400"
          :class="{ 'animate-spin': syncStore.progress.active }"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
        >
          <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
        </svg>
        <span>{{ syncStore.progress.active ? i18n.t('nav.syncing') : i18n.t('nav.syncAll') }}</span>
      </button>

      <!-- User Card -->
      <div class="flex items-center justify-between px-2 pt-1 text-xs">
        <div class="flex items-center gap-2 overflow-hidden">
          <div class="w-6 h-6 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 font-mono text-[10px]">
            {{ authStore.user?.username?.charAt(0).toUpperCase() || 'U' }}
          </div>
          <span class="font-medium text-zinc-300 truncate max-w-[100px]">{{ authStore.user?.username }}</span>
        </div>
        <button
          @click="$emit('request-sign-out')"
          :title="i18n.t('nav.signOut')"
          class="text-zinc-400 hover:text-rose-400 transition-colors p-1"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
          </svg>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useTracksStore } from '../stores/tracks'
import { usePlaylistsStore } from '../stores/playlists'
import { useSyncStore } from '../stores/sync'
import { useToastStore } from '../stores/toast'
import { useI18nStore } from '../stores/i18n'

defineProps({
  currentView: {
    type: String,
    required: true,
  }
})

defineEmits(['change-view', 'request-sign-out'])

const authStore = useAuthStore()
const tracksStore = useTracksStore()
const playlistsStore = usePlaylistsStore()
const syncStore = useSyncStore()
const toast = useToastStore()
const i18n = useI18nStore()

const deferredPrompt = ref(null)
const isStandalone = ref(false)

const handleInstallPrompt = (e) => {
  e.preventDefault()
  deferredPrompt.value = e
}

onMounted(() => {
  isStandalone.value = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone
  window.addEventListener('beforeinstallprompt', handleInstallPrompt)
})

onUnmounted(() => {
  window.removeEventListener('beforeinstallprompt', handleInstallPrompt)
})

async function installPwa() {
  if (deferredPrompt.value) {
    deferredPrompt.value.prompt()
    const { outcome } = await deferredPrompt.value.userChoice
    if (outcome === 'accepted') {
      deferredPrompt.value = null
    }
  } else {
    // Fallback for iOS / mobile browsers where beforeinstallprompt doesn't fire
    const msg = i18n.currentLang === 'ru' 
      ? 'Чтобы установить приложение, нажмите "Поделиться" или меню браузера и выберите "На экран домой".'
      : 'To install the app, tap "Share" or the browser menu and select "Add to Home Screen".'
    toast.info(msg, 5000)
  }
}

async function triggerQuickSync() {
  const ok = await syncStore.triggerSyncAll()
  if (ok) {
    toast.success(i18n.currentLang === 'ru' ? 'Синхронизация запущена для всех плейлистов' : 'Sync triggered for all playlists')
  } else {
    toast.error(i18n.currentLang === 'ru' ? 'Не удалось запустить синхронизацию' : 'Failed to trigger sync')
  }
}
</script>
