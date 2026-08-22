<template>
  <div class="space-y-8 pb-28 max-w-4xl select-none">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold tracking-tight text-zinc-100">Settings & Configuration</h2>
      <p class="text-xs text-zinc-400 mt-1">Manage authentication tokens, yt-dlp parameters, network proxies, and storage</p>
    </div>

    <!-- Section 1: YouTube Authentication (cookies.txt) -->
    <div class="bg-studio-surface border border-studio-border rounded-xl p-6 space-y-5 shadow-sm">
      <div class="flex items-start justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <h3 class="text-base font-semibold text-zinc-100">YouTube Session & Cookies</h3>
            <span
              v-if="settingsStore.settings.has_cookies"
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-950 text-emerald-400 border border-emerald-800"
            >
              ACTIVE
            </span>
            <span
              v-else
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-zinc-800 text-zinc-400 border border-zinc-700"
            >
              NOT CONFIGURED
            </span>
          </div>
          <p class="text-xs text-zinc-400 mt-1">
            Required for syncing private playlists, age-restricted tracks, and the YouTube Music «Liked Songs» list.
          </p>
        </div>

        <button
          v-if="settingsStore.settings.has_cookies"
          @click="deleteCookies"
          class="text-xs text-rose-400 hover:text-rose-300 font-medium px-2.5 py-1 rounded bg-rose-950/40 border border-rose-900/60 transition-colors"
        >
          Remove Cookies
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
            <span class="text-indigo-400 font-semibold">Click to upload</span> or drag and drop <code class="font-mono text-zinc-300">cookies.txt</code>
          </p>
          <p class="text-[11px] text-zinc-400 mt-0.5">
            Exported from your browser with extensions like "Get cookies.txt locally" (Netscape HTTP format)
          </p>
        </div>
      </div>

      <div v-if="settingsStore.settings.cookies_updated_at" class="text-[11px] font-mono text-zinc-400">
        Last updated: <span class="text-zinc-200">{{ formatDateTime(settingsStore.settings.cookies_updated_at) }}</span>
      </div>
    </div>

    <!-- Section 2: Network & Residential Proxy -->
    <div class="bg-studio-surface border border-studio-border rounded-xl p-6 space-y-5 shadow-sm">
      <div>
        <h3 class="text-base font-semibold text-zinc-100">Network & Proxy</h3>
        <p class="text-xs text-zinc-400 mt-1">
          Route yt-dlp traffic through an optional residential or datacenter proxy to avoid datacenter IP bans.
        </p>
      </div>

      <div class="space-y-3">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">HTTP / SOCKS5 Proxy URL</label>
          <div class="flex items-center gap-3">
            <input
              type="text"
              v-model="proxyInput"
              placeholder="e.g. socks5://user:pass@proxy.example.com:1080 or http://127.0.0.1:8080"
              class="flex-1 bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-xs font-mono text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60"
            />
            <button
              @click="testProxy"
              :disabled="testingProxy || !proxyInput"
              class="px-3.5 py-2 rounded-lg text-xs font-medium bg-studio-elevated hover:bg-studio-hover text-zinc-200 border border-studio-border disabled:opacity-40 transition-colors flex items-center gap-2 shrink-0"
            >
              <span v-if="testingProxy" class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>Test Connection</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 3: Audio Preferences -->
    <div class="bg-studio-surface border border-studio-border rounded-xl p-6 space-y-5 shadow-sm">
      <div>
        <h3 class="text-base font-semibold text-zinc-100">Audio Codecs & Concurrency</h3>
        <p class="text-xs text-zinc-400 mt-1">Choose target extraction codec and worker concurrency limits</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">Audio Extraction Codec</label>
          <select
            v-model="audioFormat"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
          >
            <option value="opus">Opus (High Efficiency, Recommended)</option>
            <option value="m4a">AAC / M4A (Maximum Compatibility)</option>
            <option value="mp3">MP3 (Legacy)</option>
            <option value="flac">FLAC (Lossless Container)</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-300 mb-1.5">Max Concurrent Downloads</label>
          <select
            v-model="maxConcurrent"
            class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
          >
            <option :value="1">1 Worker (Safest against rate-limits)</option>
            <option :value="2">2 Workers (Default recommended)</option>
            <option :value="3">3 Workers (Fast)</option>
          </select>
        </div>
      </div>

      <div class="flex justify-end pt-2">
        <button
          @click="saveSettings"
          class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-md active:scale-95"
        >
          Save Preferences
        </button>
      </div>
    </div>

    <!-- Section 4: System Information & Diagnostics -->
    <div class="bg-studio-surface border border-studio-border rounded-xl p-6 space-y-4 shadow-sm font-mono text-xs text-zinc-300">
      <h3 class="text-base font-semibold text-zinc-100 font-sans">System Diagnostics</h3>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-2">
        <div class="p-3 rounded-lg bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">YT-DLP BINARY</span>
          <span class="text-zinc-100 font-semibold">{{ settingsStore.settings.ytdlp_version || 'Installed' }}</span>
        </div>

        <div class="p-3 rounded-lg bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">FFMPEG RUNTIME</span>
          <span class="text-zinc-100 font-semibold truncate block">{{ settingsStore.settings.ffmpeg_version || 'Ready' }}</span>
        </div>

        <div class="p-3 rounded-lg bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">STORAGE USAGE</span>
          <span class="text-zinc-100 font-semibold">{{ formatBytes(settingsStore.settings.storage_usage_bytes) }}</span>
        </div>

        <div class="p-3 rounded-lg bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">SQLITE DB SIZE</span>
          <span class="text-zinc-100 font-semibold">{{ formatBytes(settingsStore.settings.database_size_bytes) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '../stores/settings'
import { useToastStore } from '../stores/toast'

const settingsStore = useSettingsStore()
const toast = useToastStore()

const isDragging = ref(false)
const proxyInput = ref('')
const audioFormat = ref('opus')
const maxConcurrent = ref(2)
const testingProxy = ref(false)

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
    toast.success('cookies.txt uploaded successfully')
  } catch (e) {
    toast.error(e.message || 'Failed to upload cookies')
  }
}

async function deleteCookies() {
  if (confirm('Delete saved cookies.txt?')) {
    await settingsStore.deleteCookies()
    toast.success('Cookies removed')
  }
}

async function testProxy() {
  if (!proxyInput.value) return
  testingProxy.value = true
  try {
    await settingsStore.testProxy(proxyInput.value)
    toast.success('Proxy connection verified successfully!')
  } catch (e) {
    toast.error(e.message || 'Proxy connection test failed')
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
    toast.success('Settings saved successfully')
  } catch (e) {
    toast.error(e.message || 'Failed to save settings')
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

