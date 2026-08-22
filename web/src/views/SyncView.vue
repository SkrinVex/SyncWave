<template>
  <div class="space-y-6 pb-28">
    <!-- Header -->
    <div class="flex items-center justify-between select-none">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-zinc-100">Sync & Daemon Logs</h2>
        <p class="text-xs text-zinc-400 mt-1">Real-time background worker telemetry and synchronization events</p>
      </div>

      <div class="flex items-center gap-2.5">
        <button
          @click="syncStore.clearLogs()"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-studio-elevated hover:bg-studio-hover text-zinc-300 border border-studio-border transition-colors"
        >
          Clear Logs
        </button>
        <button
          @click="triggerAll"
          :disabled="syncStore.progress.active"
          class="flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md active:scale-95 disabled:opacity-50"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': syncStore.progress.active }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
          </svg>
          <span>{{ syncStore.progress.active ? 'Syncing...' : 'Sync All' }}</span>
        </button>
      </div>
    </div>

    <!-- Active Task Card (if active) -->
    <div
      v-if="syncStore.progress.active"
      class="bg-studio-surface border border-emerald-900/60 rounded-xl p-5 shadow-lg space-y-4 relative overflow-hidden"
    >
      <div class="absolute top-0 left-0 right-0 h-1 bg-zinc-800">
        <div
          class="h-full bg-emerald-500 transition-all duration-300"
          :style="{ width: `${syncStore.progress.percentage}%` }"
        ></div>
      </div>

      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-1">
        <div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping"></span>
            <span class="text-xs font-mono uppercase font-bold text-emerald-400 tracking-wider">Active Download</span>
          </div>
          <h3 class="text-base font-semibold text-zinc-100 mt-1 truncate">
            {{ syncStore.progress.current_track_title || 'Preparing...' }}
          </h3>
          <p class="text-xs text-zinc-400 font-mono mt-0.5">
            Playlist: <span class="text-zinc-200">{{ syncStore.progress.playlist_title || 'YouTube Music' }}</span>
            <span v-if="syncStore.progress.total_tracks > 0"> • Track {{ syncStore.progress.current_track_index }} of {{ syncStore.progress.total_tracks }}</span>
          </p>
        </div>

        <div class="flex items-center gap-6 font-mono text-xs text-zinc-300">
          <div v-if="syncStore.progress.speed">
            <span class="text-zinc-500 block text-[10px]">SPEED</span>
            <strong>{{ syncStore.progress.speed }}</strong>
          </div>
          <div v-if="syncStore.progress.eta">
            <span class="text-zinc-500 block text-[10px]">ETA</span>
            <strong>{{ syncStore.progress.eta }}</strong>
          </div>
          <div class="text-right">
            <span class="text-zinc-500 block text-[10px]">PROGRESS</span>
            <strong class="text-emerald-400 text-sm">{{ Math.round(syncStore.progress.percentage) }}%</strong>
          </div>
        </div>
      </div>
    </div>

    <!-- Live Terminal Console -->
    <div class="bg-studio-surface border border-studio-border rounded-xl overflow-hidden shadow-lg flex flex-col h-[520px]">
      <!-- Terminal Header -->
      <div class="h-12 bg-studio-elevated border-b border-studio-border px-4 flex items-center justify-between select-none">
        <div class="flex items-center gap-2 text-xs font-mono text-zinc-400">
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-full bg-zinc-700"></span>
            <span class="w-2.5 h-2.5 rounded-full bg-zinc-700"></span>
            <span class="w-2.5 h-2.5 rounded-full bg-zinc-700"></span>
          </div>
          <span class="ml-2 text-zinc-300">syncwave.daemon.log</span>
        </div>

        <!-- Filter Pills -->
        <div class="flex items-center gap-1 text-[11px] font-mono">
          <button
            v-for="lvl in ['all', 'info', 'success', 'warn', 'error']"
            :key="lvl"
            @click="selectedLevel = lvl"
            :class="[
              'px-2 py-0.5 rounded uppercase font-semibold transition-colors',
              selectedLevel === lvl ? 'bg-zinc-700 text-white' : 'text-zinc-500 hover:text-zinc-300'
            ]"
          >
            {{ lvl }}
          </button>
        </div>
      </div>

      <!-- Log Lines Stream -->
      <div
        ref="logContainer"
        class="flex-1 p-4 overflow-y-auto font-mono text-xs space-y-1.5 bg-[#0a0a0d] selection:bg-zinc-800"
      >
        <div
          v-for="log in filteredLogs"
          :key="log.id || log.created_at"
          class="flex items-start gap-3 leading-relaxed"
        >
          <span class="text-zinc-600 text-[11px] shrink-0 tabular-nums">
            {{ formatLogTime(log.created_at) }}
          </span>

          <span
            :class="[
              'px-1.5 py-0.2 rounded text-[10px] uppercase font-bold shrink-0',
              log.level === 'info' ? 'bg-zinc-800 text-zinc-400' :
              log.level === 'success' ? 'bg-emerald-950 text-emerald-400 border border-emerald-800/60' :
              log.level === 'warn' ? 'bg-amber-950 text-amber-400 border border-amber-800/60' :
              'bg-rose-950 text-rose-400 border border-rose-800/60'
            ]"
          >
            {{ log.level }}
          </span>

          <span
            :class="[
              'flex-1 break-all',
              log.level === 'error' ? 'text-rose-300' :
              log.level === 'warn' ? 'text-amber-300' :
              log.level === 'success' ? 'text-emerald-300' : 'text-zinc-300'
            ]"
          >
            {{ log.message }}
          </span>
        </div>

        <div v-if="filteredLogs.length === 0" class="text-zinc-600 py-10 text-center">
          No logs matching current filter.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useSyncStore } from '../stores/sync'
import { useToastStore } from '../stores/toast'

const syncStore = useSyncStore()
const toast = useToastStore()

const selectedLevel = ref('all')
const logContainer = ref(null)

onMounted(() => {
  syncStore.fetchLogs()
  syncStore.fetchInitialProgress()
})

const filteredLogs = computed(() => {
  if (selectedLevel.value === 'all') return syncStore.logs
  return syncStore.logs.filter(l => l.level === selectedLevel.value)
})

async function triggerAll() {
  const ok = await syncStore.triggerSyncAll()
  if (ok) {
    toast.success('Sync started for all playlists')
  } else {
    toast.error('Failed to trigger sync')
  }
}

function formatLogTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString([], { hour12: false })
}
</script>

