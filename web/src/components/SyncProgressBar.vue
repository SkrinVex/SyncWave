<template>
  <transition
    enter-active-class="transition-all ease-out duration-200"
    enter-from-class="-translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition-all ease-in duration-150"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="-translate-y-full opacity-0"
  >
    <div
      v-if="syncStore.progress.active"
      class="bg-studio-surface/95 backdrop-blur-md border-b border-indigo-500/20 px-4 md:px-6 py-1.5 flex items-center justify-between gap-3 text-xs font-mono select-none sticky top-16 z-10 shadow-sm"
    >
      <!-- Left: Indicator & Track Title -->
      <div class="flex items-center gap-2.5 min-w-0">
        <div class="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse shrink-0"></div>
        <span class="text-zinc-500 text-[10px] uppercase font-semibold hidden sm:inline tracking-wider">SYNC:</span>
        <span class="font-medium text-zinc-200 truncate max-w-xs md:max-w-md">
          {{ syncStore.progress.current_track_title || syncStore.progress.status_text || 'Синхронизация...' }}
        </span>
        <span
          v-if="syncStore.progress.total_tracks > 0"
          class="text-zinc-400 text-[10px] bg-zinc-800/80 px-1.5 py-0.5 rounded shrink-0 border border-zinc-700/50"
        >
          {{ syncStore.progress.current_track_index }} / {{ syncStore.progress.total_tracks }}
        </span>
      </div>

      <!-- Right: Track Progress Bar, Speed, ETA & Cancel Button -->
      <div class="flex items-center gap-3 shrink-0">
        <div v-if="syncStore.progress.speed" class="hidden md:flex items-center gap-1 text-[11px] text-zinc-400">
          <span>{{ syncStore.progress.speed }}</span>
        </div>

        <div v-if="syncStore.progress.eta" class="hidden sm:flex items-center gap-1 text-[11px] text-zinc-500">
          <span>ETA: {{ syncStore.progress.eta }}</span>
        </div>

        <!-- Track Download Progress Bar -->
        <div class="flex items-center gap-2">
          <div class="w-20 sm:w-32 h-1.5 bg-zinc-800 rounded-full overflow-hidden border border-zinc-700/50">
            <div
              class="h-full bg-gradient-to-r from-indigo-500 to-emerald-400 rounded-full transition-all duration-150"
              :style="{ width: `${Math.max(syncStore.progress.track_percentage || 0, 2)}%` }"
            ></div>
          </div>

          <span class="text-emerald-400 font-bold text-[11px] tabular-nums w-9 text-right">
            {{ Math.round(syncStore.progress.track_percentage || 0) }}%
          </span>
        </div>

        <!-- Cancel Sync Button -->
        <button
          @click="syncStore.cancelSync()"
          title="Отменить синхронизацию"
          class="ml-1 text-zinc-400 hover:text-red-400 hover:bg-red-500/10 p-1 rounded transition-colors flex items-center gap-1 text-[11px]"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
          <span class="hidden sm:inline text-[10px]">Отмена</span>
        </button>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { useSyncStore } from '../stores/sync'
const syncStore = useSyncStore()
</script>
