<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in select-none text-left"
    @click.self="emit('close')"
  >
    <div
      class="bg-studio-surface border border-studio-border rounded-2xl max-w-lg w-full overflow-hidden shadow-2xl animate-scale-up"
    >
      <!-- Header -->
      <div class="p-6 border-b border-studio-borderSubtle flex items-start justify-between gap-4">
        <div class="flex items-center gap-3">
          <div
            :class="[
              'w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border',
              isExpired
                ? 'bg-rose-950/60 border-rose-800/80 text-rose-400'
                : (isExpiringSoon ? 'bg-amber-950/60 border-amber-800/80 text-amber-400' : 'bg-indigo-950/60 border-indigo-800/80 text-indigo-400')
            ]"
          >
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5"/>
              <path d="M8.5 8.5v.01"/><path d="M11.5 11.5v.01"/><path d="M8.5 15.5v.01"/><path d="M15.5 15.5v.01"/>
            </svg>
          </div>
          <div>
            <h3 class="text-base font-semibold text-zinc-100">
              {{ modalTitle }}
            </h3>
            <p class="text-xs text-zinc-400 mt-0.5">
              {{ i18n.t('cookiesModal.subtitle') }}
            </p>
          </div>
        </div>

        <button
          @click="emit('close')"
          class="text-zinc-500 hover:text-zinc-300 p-1.5 rounded-lg hover:bg-studio-hover transition-colors"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <!-- Body Content -->
      <div class="p-6 space-y-4 text-xs text-zinc-300">
        <!-- Status Banner -->
        <div
          :class="[
            'p-3.5 rounded-xl border flex items-start gap-3',
            isExpired
              ? 'bg-rose-950/30 border-rose-900/50 text-rose-200'
              : (isExpiringSoon ? 'bg-amber-950/30 border-amber-900/50 text-amber-200' : 'bg-studio-elevated border-studio-border text-zinc-300')
          ]"
        >
          <svg class="w-4 h-4 shrink-0 mt-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <div class="space-y-1">
            <p class="font-medium leading-relaxed">
              {{ statusMessage }}
            </p>
            <p v-if="settingsStore.settings.cookies_error" class="text-[11px] opacity-80 font-mono">
              {{ settingsStore.settings.cookies_error }}
            </p>
            <p v-if="settingsStore.settings.cookies_expires_at" class="text-[11px] opacity-75 font-mono">
              Срок действия: {{ formatDateTime(settingsStore.settings.cookies_expires_at) }}
            </p>
          </div>
        </div>

        <!-- Quick Export Steps Guide -->
        <div class="p-3.5 rounded-xl bg-studio-elevated/70 border border-studio-border space-y-2">
          <h4 class="font-semibold text-zinc-200 text-xs flex items-center gap-2">
            <span class="w-4 h-4 rounded-full bg-indigo-600/30 text-indigo-400 border border-indigo-500/30 flex items-center justify-center text-[10px] font-bold">1</span>
            Как получить свежий файл cookies.txt:
          </h4>
          <ol class="list-decimal list-inside space-y-1 text-zinc-400 text-[11px] leading-relaxed pl-1">
            <li>
              Установите расширение 
              <strong class="text-zinc-200">Get cookies.txt LOCALLY</strong> для Chrome / Firefox.
            </li>
            <li>
              Откройте 
              <a href="https://music.youtube.com" target="_blank" rel="noopener noreferrer" class="text-indigo-400 hover:underline">music.youtube.com</a>
              и убедитесь, что вошли в аккаунт.
            </li>
            <li>Нажмите на иконку расширения и выберите <strong class="text-zinc-200">Export as cookies.txt</strong>.</li>
            <li>Загрузите полученный файл ниже.</li>
          </ol>
        </div>

        <!-- Dropzone / File Upload -->
        <div
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onDrop"
          :class="[
            'border-2 border-dashed rounded-xl p-5 text-center transition-all cursor-pointer flex flex-col items-center justify-center gap-2',
            isDragging
              ? 'border-indigo-500 bg-indigo-950/20'
              : 'border-studio-border hover:border-zinc-600 bg-studio-elevated/40 hover:bg-studio-elevated/70'
          ]"
          @click="$refs.fileInput.click()"
        >
          <input
            ref="fileInput"
            type="file"
            accept=".txt"
            class="hidden"
            @change="onFileSelected"
          />

          <div v-if="isUploading" class="w-8 h-8 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
          <div v-else class="w-9 h-9 rounded-full bg-studio-elevated border border-studio-border flex items-center justify-center text-zinc-400">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
            </svg>
          </div>

          <div>
            <span class="text-xs font-semibold text-indigo-400 hover:text-indigo-300">
              {{ isUploading ? 'Проверка и загрузка...' : 'Нажмите для выбора файла cookies.txt' }}
            </span>
            <span class="text-xs text-zinc-400" v-if="!isUploading"> или перетащите файл</span>
            <p class="text-[10px] text-zinc-500 mt-0.5">
              Поддерживается стандартный формат Netscape HTTP Cookies
            </p>
          </div>
        </div>
      </div>

      <!-- Footer Buttons -->
      <div class="p-4 bg-studio-elevated/40 border-t border-studio-border flex items-center justify-end gap-2.5">
        <button
          type="button"
          @click="emit('close')"
          class="px-4 py-2 text-xs font-medium text-zinc-300 hover:text-white bg-studio-elevated hover:bg-studio-hover border border-studio-border rounded-lg transition-colors"
        >
          {{ i18n.t('confirm.cancel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useSettingsStore } from '../stores/settings'
import { useToastStore } from '../stores/toast'
import { useI18nStore } from '../stores/i18n'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close'])

const settingsStore = useSettingsStore()
const toast = useToastStore()
const i18n = useI18nStore()

const fileInput = ref(null)
const isDragging = ref(false)
const isUploading = ref(false)

const isExpired = computed(() => {
  return settingsStore.settings.cookies_status === 'expired' || settingsStore.settings.cookies_status === 'invalid'
})

const isExpiringSoon = computed(() => {
  return settingsStore.settings.cookies_status === 'expiring_soon'
})

const modalTitle = computed(() => {
  if (isExpired.value) return 'Срок действия YouTube Cookies истёк'
  if (isExpiringSoon.value) return 'Срок действия YouTube Cookies истекает'
  if (!settingsStore.settings.has_cookies) return 'Требуется настройка YouTube Cookies'
  return 'Обновление YouTube Cookies'
})

const statusMessage = computed(() => {
  if (isExpired.value) {
    return 'Сессия авторизации YouTube устарела. YouTube Music заблокировал доступ к трекам и плейлистам. Загрузите свежий файл cookies.txt для восстановления синхронизации.'
  }
  if (isExpiringSoon.value) {
    return 'Срок действия авторизационных cookies YouTube подходит к концу. Рекомендуется обновить их, чтобы синхронизация не прерывалась.'
  }
  if (!settingsStore.settings.has_cookies) {
    return 'Для скачивания треков из закрытых плейлистов, контента 18+ и синхронизации списка «Понравившиеся» необходима авторизация через cookies.txt.'
  }
  return 'Cookies настроены и активны. Вы можете обновить их в любой момент.'
})

async function onFileSelected(e) {
  const file = e.target.files[0]
  if (file) await handleUpload(file)
}

async function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files[0]
  if (file) await handleUpload(file)
}

async function handleUpload(file) {
  isUploading.value = true
  try {
    await settingsStore.uploadCookies(file)
    toast.success('cookies.txt успешно обновлен и активирован!')
    emit('close')
  } catch (e) {
    toast.error(e.message || 'Ошибка загрузки cookies')
  } finally {
    isUploading.value = false
  }
}

function formatDateTime(str) {
  if (!str) return ''
  return new Date(str).toLocaleDateString() + ' ' + new Date(str).toLocaleTimeString()
}
</script>

