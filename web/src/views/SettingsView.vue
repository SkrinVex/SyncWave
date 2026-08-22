<template>
  <div class="space-y-8 pb-28 max-w-4xl select-none text-left">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold tracking-tight text-zinc-100">{{ i18n.t('settings.title') }}</h2>
      <p class="text-xs text-zinc-400 mt-1">{{ i18n.t('settings.subtitle') }}</p>
    </div>

    <!-- Section 0: Language Switcher -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-3 shadow-sm">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.language') }}</h3>
          <p class="text-xs text-zinc-400 mt-0.5">Выберите язык интерфейса / Select UI language</p>
        </div>
        <div class="flex items-center gap-1 bg-studio-elevated p-1 rounded-xl border border-studio-border text-xs">
          <button
            @click="i18n.setLanguage('ru')"
            :class="[
              'px-3 py-1.5 rounded-lg font-medium transition-all',
              i18n.currentLang === 'ru' ? 'bg-indigo-600 text-white shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
            ]"
          >
            🇷🇺 Русский
          </button>
          <button
            @click="i18n.setLanguage('en')"
            :class="[
              'px-3 py-1.5 rounded-lg font-medium transition-all',
              i18n.currentLang === 'en' ? 'bg-indigo-600 text-white shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
            ]"
          >
            🇬🇧 English
          </button>
        </div>
      </div>
    </div>

    <!-- Section 1: YouTube Authentication (cookies.txt) -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-5 shadow-sm">
      <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <div class="flex items-center gap-2.5">
            <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.cookiesTitle') }}</h3>
            <span
              v-if="settingsStore.settings.has_cookies"
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-950 text-emerald-400 border border-emerald-800"
            >
              {{ i18n.t('settings.cookiesActive') }}
            </span>
            <span
              v-else
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-zinc-800 text-zinc-400 border border-zinc-700"
            >
              {{ i18n.t('settings.cookiesNotConfigured') }}
            </span>
          </div>
          <p class="text-xs text-zinc-400 mt-1 leading-relaxed">
            {{ i18n.t('settings.cookiesDesc') }}
          </p>
        </div>

        <button
          v-if="settingsStore.settings.has_cookies"
          @click="showDeleteCookiesConfirm = true"
          class="text-xs text-rose-400 hover:text-rose-300 font-medium px-3 py-1.5 rounded-lg bg-rose-950/40 border border-rose-900/60 transition-colors shrink-0 self-start"
        >
          {{ i18n.t('settings.removeCookies') }}
        </button>
      </div>

      <!-- Upload Dropzone -->
      <div
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="onDrop"
        :class="[
          'border-2 border-dashed rounded-xl p-6 text-center transition-all cursor-pointer flex flex-col items-center justify-center gap-2',
          isDragging
            ? 'border-indigo-500 bg-indigo-950/20'
            : 'border-zinc-800 hover:border-zinc-700 bg-studio-elevated/50'
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
        <svg class="w-8 h-8 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
        </svg>
        <div>
          <p class="text-xs font-medium text-zinc-200">
            <span class="text-indigo-400 font-semibold">{{ i18n.t('settings.dropzoneTitle') }}</span> {{ i18n.t('settings.dropzoneOr') }} <code class="font-mono text-zinc-300">cookies.txt</code>
          </p>
          <p class="text-[11px] text-zinc-400 mt-0.5">
            {{ i18n.t('settings.dropzoneSub') }}
          </p>
        </div>
      </div>

      <div v-if="settingsStore.settings.cookies_updated_at" class="text-[11px] font-mono text-zinc-400">
        {{ i18n.t('settings.lastUpdated') }} <span class="text-zinc-200">{{ formatDateTime(settingsStore.settings.cookies_updated_at) }}</span>
      </div>
    </div>

    <!-- Section 2: Network & Residential Proxy -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-5 shadow-sm">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.proxyTitle') }}</h3>
          <p class="text-xs text-zinc-400 mt-1 leading-relaxed">
            {{ i18n.t('settings.proxyDesc') }}
          </p>
        </div>

        <button
          @click="showProxyHelp = !showProxyHelp"
          class="text-xs font-medium text-indigo-400 hover:text-indigo-300 flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-indigo-950/40 border border-indigo-800/50 transition-colors shrink-0"
        >
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3M12 17h.01"/>
          </svg>
          <span>{{ i18n.t('settings.proxyHelpBtn') }}</span>
        </button>
      </div>

      <!-- Expandable Proxy Guide Help Card -->
      <div v-if="showProxyHelp" class="p-4 rounded-xl bg-studio-elevated border border-indigo-900/40 text-xs text-zinc-300 space-y-2">
        <h4 class="font-semibold text-indigo-300 text-xs">Зачем нужен прокси и где его взять:</h4>
        <ul class="list-disc list-inside space-y-1 text-[11px] text-zinc-400 leading-relaxed">
          <li><strong>Датацентровые IP (VPS)</strong>: Хостинги типа Hetzner, DigitalOcean, OVH часто попадают во временный троттлинг YouTube (ошибка 429).</li>
          <li><strong>Резидентские прокси (Residential)</strong>: Прокси с IP домашних провайдеров полностью исключают блокировки.</li>
          <li><strong>Формат строки</strong>: <code class="font-mono text-zinc-200">http://username:password@ip:port</code> или <code class="font-mono text-zinc-200">socks5://username:password@ip:port</code>.</li>
          <li><strong>Популярные сервисы</strong>: Webshare, Proxy-Seller, BrightData, SmartProxy.</li>
        </ul>
      </div>

      <div class="space-y-3">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('settings.proxyLabel') }}</label>
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
            <input
              type="text"
              v-model="proxyInput"
              :placeholder="i18n.t('settings.proxyPlaceholder')"
              class="flex-1 bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-xs font-mono text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60"
            />
            <button
              @click="testProxy"
              :disabled="testingProxy || !proxyInput"
              class="px-3.5 py-2 rounded-lg text-xs font-medium bg-studio-elevated hover:bg-studio-hover text-zinc-200 border border-studio-border disabled:opacity-40 transition-colors flex items-center justify-center gap-2 shrink-0"
            >
              <span v-if="testingProxy" class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>{{ i18n.t('settings.testProxy') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 3: Audio Preferences -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-5 shadow-sm">
      <div>
        <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.audioTitle') }}</h3>
        <p class="text-xs text-zinc-400 mt-1 leading-relaxed">{{ i18n.t('settings.audioDesc') }}</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('settings.audioCodecLabel') }}</label>
          <select
            v-model="audioFormat"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60 cursor-pointer"
          >
            <option value="opus">{{ i18n.t('settings.opusDesc') }}</option>
            <option value="m4a">{{ i18n.t('settings.m4aDesc') }}</option>
            <option value="mp3">{{ i18n.t('settings.mp3Desc') }}</option>
            <option value="flac">{{ i18n.t('settings.flacDesc') }}</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('settings.maxConcurrentLabel') }}</label>
          <select
            v-model="maxConcurrent"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60 cursor-pointer"
          >
            <option :value="1">{{ i18n.t('settings.worker1') }}</option>
            <option :value="2">{{ i18n.t('settings.worker2') }}</option>
            <option :value="3">{{ i18n.t('settings.worker3') }}</option>
          </select>
        </div>
      </div>

      <div class="flex justify-end pt-2">
        <button
          @click="saveSettings"
          class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-md active:scale-95"
        >
          {{ i18n.t('settings.saveBtn') }}
        </button>
      </div>
    </div>

    <!-- Section 4: System Information & Diagnostics -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-4 shadow-sm font-mono text-xs text-zinc-300">
      <h3 class="text-base font-semibold text-zinc-100 font-sans">{{ i18n.t('settings.diagTitle') }}</h3>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-2">
        <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.ytdlpVer') }}</span>
          <span class="text-zinc-100 font-semibold">{{ settingsStore.settings.ytdlp_version || 'Ready' }}</span>
        </div>

        <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.ffmpegVer') }}</span>
          <span class="text-zinc-100 font-semibold truncate block">{{ settingsStore.settings.ffmpeg_version || 'Ready' }}</span>
        </div>

        <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.storageUsage') }}</span>
          <span class="text-zinc-100 font-semibold">{{ formatBytes(settingsStore.settings.storage_usage_bytes) }}</span>
        </div>

        <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.dbSize') }}</span>
          <span class="text-zinc-100 font-semibold">{{ formatBytes(settingsStore.settings.database_size_bytes) }}</span>
        </div>
      </div>
    </div>

    <!-- Confirmation Modal for Delete Cookies -->
    <ConfirmModal
      :open="showDeleteCookiesConfirm"
      :title="i18n.t('confirm.deleteCookiesTitle')"
      :description="i18n.t('confirm.deleteCookiesDesc')"
      :confirm-text="i18n.t('confirm.delete')"
      :cancel-text="i18n.t('confirm.cancel')"
      :danger="true"
      @confirm="handleDeleteCookies"
      @cancel="showDeleteCookiesConfirm = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '../stores/settings'
import { useToastStore } from '../stores/toast'
import { useI18nStore } from '../stores/i18n'
import ConfirmModal from '../components/ConfirmModal.vue'

const settingsStore = useSettingsStore()
const toast = useToastStore()
const i18n = useI18nStore()

const isDragging = ref(false)
const proxyInput = ref('')
const audioFormat = ref('opus')
const maxConcurrent = ref(2)
const testingProxy = ref(false)
const showProxyHelp = ref(false)
const showDeleteCookiesConfirm = ref(false)

onMounted(async () => {
  await settingsStore.fetchSettings()
  proxyInput.value = settingsStore.settings.http_proxy || ''
  audioFormat.value = settingsStore.settings.audio_format || 'opus'
  maxConcurrent.value = settingsStore.settings.max_concurrent || 2
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
  try {
    await settingsStore.uploadCookies(file)
    toast.success(i18n.currentLang === 'ru' ? 'cookies.txt успешно загружен' : 'cookies.txt uploaded successfully')
  } catch (e) {
    toast.error(e.message || 'Ошибка загрузки cookies')
  }
}

async function handleDeleteCookies() {
  showDeleteCookiesConfirm.value = false
  await settingsStore.deleteCookies()
  toast.success(i18n.currentLang === 'ru' ? 'Файл cookies.txt удален' : 'Cookies removed')
}

async function testProxy() {
  if (!proxyInput.value) return
  testingProxy.value = true
  try {
    await settingsStore.testProxy(proxyInput.value)
    toast.success(i18n.currentLang === 'ru' ? 'Соединение через прокси успешно проверено!' : 'Proxy connection verified successfully!')
  } catch (e) {
    toast.error(e.message || 'Ошибка проверки прокси')
  } finally {
    testingProxy.value = false
  }
}

async function saveSettings() {
  try {
    await settingsStore.updateSettings({
      http_proxy: proxyInput.value,
      audio_format: audioFormat.value,
      max_concurrent: maxConcurrent.value,
    })
    toast.success(i18n.currentLang === 'ru' ? 'Настройки успешно сохранены' : 'Settings saved successfully')
  } catch (e) {
    toast.error(e.message || 'Ошибка сохранения настроек')
  }
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDateTime(str) {
  if (!str) return ''
  return new Date(str).toLocaleString()
}
</script>
