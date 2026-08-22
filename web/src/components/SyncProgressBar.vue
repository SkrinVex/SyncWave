<template>
  <transition
    enter-active-class="transition-all ease-out duration-300"
    enter-from-class="-translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition-all ease-in duration-200"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="-translate-y-full opacity-0"
  >
    <div
      v-if="syncStore.progress.active"
      class="bg-studio-surface border-b border-emerald-900/50 px-6 py-2.5 flex items-center justify-between gap-4 text-xs font-mono select-none"
    >
      <div class="flex items-center gap-3 min-w-0">
        <div class="w-2 h-2 rounded-full bg-emerald-400 animate-ping shrink-0"></div>
        <span class="text-zinc-400 hidden sm:inline">SYNCING:</span>
        <span class="font-semibold text-emerald-300 truncate max-w-sm">
          {{ syncStore.progress.current_track_title || syncStore.progress.status_text }}
        </span>
        <span v-if="syncStore.progress.total_tracks > 0" class="text-zinc-500 text-[11px]">
          ({{ syncStore.progress.current_track_index }} of {{ syncStore.progress.total_tracks }})
        </span>
      </div>

      <!-- Right: Progress Bar & Telemetry -->
      <div class="flex items-center gap-4 shrink-0">
        <div class="hidden md:flex items-center gap-3 text-zinc-400 text-[11px]">
          <span v-if="syncStore.progress.speed">Speed: <strong class="text-zinc-200">{{ syncStore.progress.speed }}</strong></span>
          <span v-if="syncStore.progress.eta">ETA: <strong class="text-zinc-200">{{ syncStore.progress.eta }}</strong></span>
        </div>

        <div class="w-24 sm:w-36 h-2 bg-zinc-800 rounded-full overflow-hidden border border-zinc-700/50">
          <div
            class="h-full bg-emerald-500 rounded-full transition-all duration-300"
            :style="{ width: `${syncStore.progress.percentage}%` }"
          ></div>
        </div>

        <span class="text-emerald-400 font-bold w-10 text-right">
          {{ Math.round(syncStore.progress.percentage) }}%
        </span>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { useSyncStore } from '../stores/sync'
const syncStore = useSyncStore()
</script>

