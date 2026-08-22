<template>
  <div class="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="bg-studio-surface border border-studio-border rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5">
      <!-- Modal Header -->
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('addModal.title') }}</h3>
          <p class="text-xs text-zinc-400 mt-0.5">{{ i18n.t('addModal.subtitle') }}</p>
        </div>
        <button
          @click="$emit('close')"
          class="text-zinc-400 hover:text-zinc-100 p-1.5 rounded-lg hover:bg-studio-hover transition-colors"
        >
          ✕
        </button>
      </div>

      <!-- Quick Preset: Liked Songs -->
      <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border flex items-center justify-between">
        <div class="pr-3">
          <h4 class="text-xs font-semibold text-zinc-200">{{ i18n.t('addModal.likedPresetTitle') }}</h4>
          <p class="text-[11px] text-zinc-400 mt-0.5">{{ i18n.t('addModal.likedPresetDesc') }}</p>
        </div>
        <button
          type="button"
          @click="fillLikedMusic"
          class="px-2.5 py-1.5 text-xs font-medium bg-indigo-600/20 hover:bg-indigo-600/40 border border-indigo-500/30 text-indigo-300 rounded-lg transition-colors shrink-0"
        >
          {{ i18n.t('addModal.usePreset') }}
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">
            {{ i18n.t('addModal.urlLabel') }} <span class="text-indigo-400">*</span>
          </label>
          <input
            type="text"
            v-model="urlOrId"
            required
            :placeholder="i18n.t('addModal.urlPlaceholder')"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-xs font-mono text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">
            {{ i18n.t('addModal.titleLabel') }} <span class="text-zinc-500 text-[11px]">{{ i18n.t('addModal.titleOptional') }}</span>
          </label>
          <input
            type="text"
            v-model="title"
            :placeholder="i18n.t('addModal.titlePlaceholder')"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
          />
        </div>

        <div class="grid grid-cols-2 gap-3 pt-1">
          <!-- Interval selector: disabled when autoSync is false -->
          <div :class="{ 'opacity-35 pointer-events-none select-none': !autoSync }">
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('addModal.syncIntervalLabel') }}</label>
            <select
              v-model="syncInterval"
              :disabled="!autoSync"
              class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
            >
              <option :value="15">{{ i18n.t('addModal.every15m') }}</option>
              <option :value="30">{{ i18n.t('addModal.every30m') }}</option>
              <option :value="60">{{ i18n.t('addModal.every1h') }}</option>
              <option :value="360">{{ i18n.t('addModal.every6h') }}</option>
              <option :value="720">{{ i18n.t('addModal.every12h') }}</option>
              <option :value="1440">{{ i18n.t('addModal.every24h') }}</option>
            </select>
          </div>

          <div class="flex items-center justify-between p-2.5 rounded-lg bg-studio-elevated border border-studio-border mt-auto">
            <div>
              <span class="text-xs font-medium text-zinc-200 block leading-tight">{{ i18n.t('addModal.autoSyncLabel') }}</span>
              <span class="text-[10px] text-zinc-400">{{ i18n.t('addModal.autoSyncDesc') }}</span>
            </div>
            <input
              type="checkbox"
              v-model="autoSync"
              class="w-4 h-4 rounded text-indigo-600 bg-zinc-800 border-zinc-700 focus:ring-indigo-500 focus:ring-offset-0 cursor-pointer"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-3">
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 rounded-lg hover:bg-studio-hover transition-colors"
          >
            {{ i18n.t('addModal.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="loading"
            class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors shadow-md disabled:opacity-50 flex items-center gap-2"
          >
            <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>{{ loading ? i18n.t('addModal.adding') : i18n.t('addModal.submit') }}</span>
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
import { useI18nStore } from '../stores/i18n'

const emit = defineEmits(['close', 'created'])
const playlistsStore = usePlaylistsStore()
const toast = useToastStore()
const i18n = useI18nStore()

const urlOrId = ref('')
const title = ref('')
const autoSync = ref(true)
const syncInterval = ref(60)
const loading = ref(false)

function fillLikedMusic() {
  urlOrId.value = 'LM'
  title.value = i18n.currentLang === 'ru' ? 'Понравившиеся' : 'Liked Music'
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
    toast.success(`${i18n.currentLang === 'ru' ? 'Добавлен плейлист' : 'Added playlist'}: ${pl.title}`)
    emit('created', pl)
    emit('close')
  } catch (e) {
    toast.error(e.message || 'Ошибка добавления плейлиста')
  } finally {
    loading.value = false
  }
}
</script>
