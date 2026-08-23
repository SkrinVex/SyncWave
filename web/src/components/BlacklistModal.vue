<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <!-- Backdrop -->
    <div
      class="absolute inset-0 bg-black/80 backdrop-blur-sm"
      @click="$emit('close')"
    ></div>

    <!-- Modal Content -->
    <div
      class="relative w-full max-w-2xl bg-studio-surface border border-studio-border rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]"
    >
      <div class="px-6 py-4 border-b border-studio-border flex items-center justify-between">
        <h2 class="text-lg font-semibold text-zinc-100">Чёрный список треков</h2>
        <button
          @click="$emit('close')"
          class="text-zinc-400 hover:text-zinc-100 p-1 rounded transition-colors"
        >
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="p-4 border-b border-studio-border">
        <div class="relative">
          <svg class="absolute left-3 top-2.5 w-4 h-4 text-zinc-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchQuery"
            @input="debounceSearch"
            type="text"
            placeholder="Поиск по названию или исполнителю..."
            class="w-full pl-9 pr-4 py-2 bg-studio-elevated border border-studio-border focus:border-indigo-500/50 rounded-lg text-sm text-zinc-100 outline-none transition-colors placeholder:text-zinc-600"
          />
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-2 min-h-[300px]">
        <div v-if="blacklist.isLoading" class="flex justify-center py-10">
          <svg class="w-6 h-6 text-indigo-500 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
        </div>
        <div v-else-if="blacklist.items.length === 0" class="flex flex-col items-center justify-center py-12 text-zinc-500">
          <svg class="w-12 h-12 mb-3 text-zinc-700" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
          </svg>
          <p>В чёрном списке нет треков.</p>
          <p class="text-xs mt-1">Треки попадают сюда при удалении.</p>
        </div>
        <div v-else class="space-y-1">
          <div
            v-for="item in blacklist.items"
            :key="item.youtube_id"
            class="flex items-center justify-between p-3 bg-studio-elevated/50 hover:bg-studio-elevated rounded-lg group transition-colors"
          >
            <div class="min-w-0 pr-4">
              <h4 class="text-sm font-medium text-zinc-200 truncate">{{ item.title }}</h4>
              <p class="text-xs text-zinc-500 truncate">{{ item.artist }}</p>
            </div>
            <button
              @click="remove(item.youtube_id)"
              class="shrink-0 px-3 py-1.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 rounded text-xs font-medium opacity-0 group-hover:opacity-100 transition-all"
              title="Убрать из Чёрного списка (Трек снова скачается)"
            >
              Восстановить
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useBlacklistStore } from '../stores/blacklist'

const props = defineProps({
  open: Boolean
})

const emit = defineEmits(['close'])

const blacklist = useBlacklistStore()
const searchQuery = ref('')
let debounceTimer = null

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    searchQuery.value = ''
    blacklist.fetchBlacklist()
  }
})

function debounceSearch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    blacklist.fetchBlacklist(searchQuery.value)
  }, 300)
}

function remove(id) {
  blacklist.removeFromBlacklist(id)
}
</script>
