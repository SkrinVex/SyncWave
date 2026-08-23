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
              v-if="settingsStore.settings.cookies_status === 'valid'"
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-950 text-emerald-400 border border-emerald-800"
            >
              {{ i18n.t('settings.cookiesActive') }}
            </span>
            <span
              v-else-if="settingsStore.settings.cookies_status === 'expiring_soon'"
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-amber-950 text-amber-300 border border-amber-800"
            >
              СКОРО ИСТЕКУТ
            </span>
            <span
              v-else-if="settingsStore.settings.cookies_status === 'expired' || settingsStore.settings.cookies_status === 'invalid'"
              class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-rose-950 text-rose-300 border border-rose-800 animate-pulse"
            >
              ИСТЕКЛИ / ТРЕБУЕТСЯ ОБНОВЛЕНИЕ
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
          <div
            v-if="settingsStore.settings.cookies_error"
            class="mt-2 p-2.5 rounded-lg bg-rose-950/40 border border-rose-900/60 text-rose-300 text-xs font-mono"
          >
            ⚠️ {{ settingsStore.settings.cookies_error }}
          </div>
          <div
            v-if="settingsStore.settings.cookies_expires_at"
            class="mt-1 text-[11px] font-mono text-zinc-400"
          >
            Срок действия: {{ formatDateTime(settingsStore.settings.cookies_expires_at) }}
          </div>
        </div>
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

        <div class="w-10 h-10 rounded-full bg-studio-elevated border border-studio-border flex items-center justify-center text-zinc-400">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
          </svg>
        </div>

        <div>
          <span class="text-xs font-semibold text-indigo-400 hover:text-indigo-300">
            {{ i18n.t('settings.dropzoneTitle') }}
          </span>
          <span class="text-xs text-zinc-400"> {{ i18n.t('settings.dropzoneOr') }}</span>
          <p class="text-[10px] text-zinc-500 mt-1">
            {{ i18n.t('settings.dropzoneSub') }}
          </p>
        </div>
      </div>

      <div v-if="settingsStore.settings.cookies_updated_at" class="text-[10px] font-mono text-zinc-500">
        {{ i18n.t('settings.lastUpdated') }} {{ formatDateTime(settingsStore.settings.cookies_updated_at) }}
      </div>
    </div>

    <!-- Section 2: Network & Proxy -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-4 shadow-sm">
      <div class="flex items-start justify-between">
        <div>
          <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.proxyTitle') }}</h3>
          <p class="text-xs text-zinc-400 mt-0.5">{{ i18n.t('settings.proxyDesc') }}</p>
        </div>
      </div>

      <div class="space-y-2">
        <label class="block text-xs font-medium text-zinc-300">{{ i18n.t('settings.proxyLabel') }}</label>
        <div class="flex gap-2">
          <input
            type="text"
            v-model="proxyInput"
            :placeholder="i18n.t('settings.proxyPlaceholder')"
            class="flex-1 bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-indigo-500/60 font-mono"
          />
          <button
            type="button"
            @click="testProxy"
            :disabled="!proxyInput || testingProxy"
            class="px-4 py-2 text-xs font-medium bg-studio-elevated hover:bg-studio-hover border border-studio-border text-zinc-200 rounded-lg transition-colors disabled:opacity-50 flex items-center gap-1.5 shrink-0"
          >
            <span v-if="testingProxy" class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
            <span>{{ i18n.t('settings.testProxy') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Section 3: Audio Codecs & Concurrency -->
    <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 space-y-5 shadow-sm">
      <div>
        <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.audioTitle') }}</h3>
        <p class="text-xs text-zinc-400 mt-0.5">{{ i18n.t('settings.audioDesc') }}</p>
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

    <!-- Section 4: Admin Panel (Only for Administrator) -->
    <div v-if="authStore.user?.is_admin" class="bg-studio-surface border border-indigo-900/40 rounded-2xl p-6 space-y-6 shadow-sm">
      <div class="border-b border-studio-borderSubtle pb-4">
        <div class="flex items-center gap-2">
          <span class="px-2 py-0.5 bg-indigo-900/60 text-indigo-300 border border-indigo-700/60 rounded text-[10px] font-mono font-bold uppercase">
            Admin Panel
          </span>
          <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.adminPanelTitle') }}</h3>
        </div>
        <p class="text-xs text-zinc-400 mt-1">{{ i18n.t('settings.adminPanelSubtitle') }}</p>
      </div>

      <!-- Controls: Public Registration Toggle & Global Storage Limit -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Registration Toggle Card -->
        <div class="p-4 rounded-xl bg-studio-elevated border border-studio-border flex items-center justify-between gap-4">
          <div>
            <h4 class="text-xs font-semibold text-zinc-200">{{ i18n.t('settings.allowRegistrationLabel') }}</h4>
            <p class="text-[11px] text-zinc-400 mt-0.5">{{ i18n.t('settings.allowRegistrationDesc') }}</p>
          </div>
          <button
            type="button"
            @click="toggleRegistration"
            :class="[
              'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
              allowRegState ? 'bg-indigo-600' : 'bg-zinc-700'
            ]"
          >
            <span
              :class="[
                'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                allowRegState ? 'translate-x-5' : 'translate-x-0'
              ]"
            />
          </button>
        </div>

        <!-- Global Storage Limit Card -->
        <div class="p-4 rounded-xl bg-studio-elevated border border-studio-border space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-xs font-semibold text-zinc-200">{{ i18n.t('settings.globalLimitLabel') }}</h4>
            <span class="text-[10px] font-mono text-indigo-400">
              {{ globalLimitGb > 0 ? `${globalLimitGb} GB` : i18n.t('settings.quotaUnlimited') }}
            </span>
          </div>
          <div class="flex gap-2">
            <input
              type="number"
              v-model.number="globalLimitGb"
              min="0"
              placeholder="ГБ (0 = безлимит)"
              class="flex-1 bg-zinc-900 border border-studio-border rounded-lg px-3 py-1.5 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
            />
            <button
              @click="saveGlobalLimit"
              class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium transition-colors shrink-0"
            >
              Ок
            </button>
          </div>
        </div>
      </div>

      <!-- Users Table -->
      <div class="space-y-3 pt-2">
        <h4 class="text-xs font-semibold text-zinc-300 uppercase tracking-wider">Пользователи системы</h4>
        
        <div class="overflow-x-auto border border-studio-border rounded-xl">
          <table class="w-full text-left text-xs">
            <thead class="bg-studio-elevated/80 border-b border-studio-border text-zinc-400 font-mono">
              <tr>
                <th class="py-2.5 px-3.5">{{ i18n.t('settings.userTableUser') }}</th>
                <th class="py-2.5 px-3.5">{{ i18n.t('settings.userTableRole') }}</th>
                <th class="py-2.5 px-3.5">{{ i18n.t('settings.userTableUsage') }}</th>
                <th class="py-2.5 px-3.5">{{ i18n.t('settings.userTableTracks') }}</th>
                <th class="py-2.5 px-3.5 text-right">{{ i18n.t('settings.userTableActions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-studio-borderSubtle">
              <tr v-for="u in adminStore.users" :key="u.id" class="hover:bg-studio-elevated/30 transition-colors">
                <td class="py-3 px-3.5 font-medium text-zinc-100 flex items-center gap-2">
                  <div class="w-6 h-6 rounded-full bg-indigo-950 text-indigo-300 border border-indigo-800 flex items-center justify-center font-bold text-[10px]">
                    {{ u.username.charAt(0).toUpperCase() }}
                  </div>
                  <span>{{ u.username }}</span>
                </td>
                <td class="py-3 px-3.5">
                  <span
                    :class="[
                      'px-2 py-0.5 rounded text-[10px] font-mono font-semibold',
                      u.is_admin ? 'bg-indigo-950 text-indigo-300 border border-indigo-800' : 'bg-zinc-800 text-zinc-400'
                    ]"
                  >
                    {{ u.is_admin ? i18n.t('settings.roleAdmin') : i18n.t('settings.roleUser') }}
                  </span>
                </td>
                <td class="py-3 px-3.5">
                  <div class="space-y-1 max-w-[140px]">
                    <div class="flex justify-between text-[10px] font-mono text-zinc-400">
                      <span>{{ formatBytes(u.storage_used_bytes) }}</span>
                      <span>{{ u.storage_quota_bytes > 0 ? formatBytes(u.storage_quota_bytes) : '∞' }}</span>
                    </div>
                    <div class="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                      <div
                        class="h-full bg-indigo-500 rounded-full"
                        :style="{ width: getQuotaPercent(u.storage_used_bytes, u.storage_quota_bytes) + '%' }"
                      ></div>
                    </div>
                  </div>
                </td>
                <td class="py-3 px-3.5 font-mono text-zinc-300">
                  {{ u.tracks_count }}
                </td>
                <td class="py-3 px-3.5 text-right space-x-2">
                  <button
                    @click="openQuotaModal(u)"
                    class="px-2.5 py-1 bg-studio-elevated hover:bg-studio-hover border border-studio-border text-zinc-200 rounded text-[11px] font-medium transition-colors"
                  >
                    Квота
                  </button>
                  <button
                    v-if="!u.is_admin"
                    @click="promptDeleteUser(u)"
                    class="px-2.5 py-1 bg-rose-500/10 hover:bg-rose-500/20 border border-rose-500/20 text-rose-400 rounded text-[11px] font-medium transition-colors"
                  >
                    Удалить
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Section 5: System Information & Diagnostics -->
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

        <!-- Physical Host Disk Total / Free (Admin Only) -->
        <div v-if="authStore.user?.is_admin" class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.hostDiskTitle') }}</span>
          <span class="text-zinc-100 font-semibold">
            {{ formatBytes(settingsStore.settings.host_disk_used_bytes) }} / {{ formatBytes(settingsStore.settings.host_disk_total_bytes) }}
          </span>
          <span class="text-[10px] text-emerald-400 block mt-0.5">
            {{ formatBytes(settingsStore.settings.host_disk_free_bytes) }} свободно
          </span>
        </div>

        <!-- User Quota or SQLite DB Size -->
        <div class="p-3.5 rounded-xl bg-studio-elevated border border-studio-border">
          <span class="text-zinc-500 text-[10px] block">{{ i18n.t('settings.userQuotaTitle') }}</span>
          <span class="text-zinc-100 font-semibold">
            {{ formatBytes(settingsStore.settings.user_storage_usage_bytes) }}
            <template v-if="settingsStore.settings.user_storage_quota_bytes > 0">
              / {{ formatBytes(settingsStore.settings.user_storage_quota_bytes) }}
            </template>
          </span>
          <span class="text-[10px] text-zinc-400 block mt-0.5">
            БД: {{ formatBytes(settingsStore.settings.database_size_bytes) }}
          </span>
        </div>
      </div>
    </div>

    <!-- Section 6: Additional Actions (Blacklist, Danger Zone Cookies) -->
    <div class="space-y-4">
      <h3 class="text-base font-semibold text-zinc-100">{{ i18n.t('settings.dangerZone') }}</h3>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Blacklist Section -->
        <section class="bg-studio-surface border border-studio-border rounded-xl overflow-hidden shadow-sm flex flex-col justify-between">
          <div class="p-6">
            <h3 class="font-medium text-zinc-100 flex items-center gap-2">
              <svg class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
              </svg>
              {{ i18n.t('settings.blacklistTitle') }}
            </h3>
            <p class="text-sm text-zinc-400 mt-2 leading-relaxed">
              {{ i18n.t('settings.blacklistDesc') }}
            </p>
          </div>
          <div class="px-6 py-4 bg-studio-elevated/30 border-t border-studio-border">
            <button
              @click="showBlacklist = true"
              class="px-4 py-2 bg-studio-elevated hover:bg-studio-hover border border-studio-border rounded-lg text-sm font-medium transition-colors w-full"
            >
              {{ i18n.t('settings.manageBlacklist') }}
            </button>
          </div>
        </section>

        <!-- Danger Zone Cookies -->
        <section class="bg-studio-surface border border-rose-900/50 rounded-xl overflow-hidden shadow-sm flex flex-col justify-between">
          <div class="p-6">
            <h3 class="font-medium text-rose-400 flex items-center gap-2">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4m0 4h.01"/>
              </svg>
              {{ i18n.t('settings.cookiesDeleteTitle') }}
            </h3>
            <p class="text-sm text-zinc-400 mt-2 leading-relaxed">
              {{ i18n.t('settings.cookiesDeleteDesc') }}
            </p>
          </div>
          <div class="px-6 py-4 bg-rose-950/20 border-t border-rose-900/30">
            <button
              @click="showDeleteCookiesConfirm = true"
              class="px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 rounded-lg text-sm font-medium transition-colors w-full"
            >
              {{ i18n.t('settings.removeCookies') }}
            </button>
          </div>
        </section>
      </div>
    </div>

    <!-- Modals -->
    <BlacklistModal :open="showBlacklist" @close="showBlacklist = false" />
    <UserQuotaModal :open="showQuotaModal" :user="selectedUser" @close="showQuotaModal = false" @save="onSaveQuota" />

    <!-- Confirmation Modal for Delete Cookies -->
    <ConfirmModal
      :open="showDeleteCookiesConfirm"
      :title="i18n.t('confirm.deleteCookiesTitle') || 'Удаление Cookies'"
      :description="i18n.t('confirm.deleteCookiesDesc') || 'Вы уверены, что хотите удалить сохраненные cookies?'"
      :confirm-text="i18n.t('confirm.delete') || 'Удалить'"
      :cancel-text="i18n.t('confirm.cancel') || 'Отмена'"
      :danger="true"
      @confirm="handleDeleteCookies"
      @cancel="showDeleteCookiesConfirm = false"
    />

    <!-- Confirmation Modal for Delete User -->
    <ConfirmModal
      :open="showDeleteUserConfirm"
      title="Удаление пользователя"
      :description="`Вы действительно хотите навсегда удалить пользователя ${userToDelete?.username}? Все его треки и плейлисты будут удалены.`"
      confirm-text="Удалить аккаунт"
      cancel-text="Отмена"
      :danger="true"
      @confirm="handleDeleteUser"
      @cancel="showDeleteUserConfirm = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import { useAdminStore } from '../stores/admin'
import { useToastStore } from '../stores/toast'
import { useI18nStore } from '../stores/i18n'
import ConfirmModal from '../components/ConfirmModal.vue'
import BlacklistModal from '../components/BlacklistModal.vue'
import UserQuotaModal from '../components/UserQuotaModal.vue'

const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const adminStore = useAdminStore()
const toast = useToastStore()
const i18n = useI18nStore()

const isDragging = ref(false)
const proxyInput = ref('')
const audioFormat = ref('opus')
const maxConcurrent = ref(2)
const testingProxy = ref(false)
const allowRegState = ref(false)
const globalLimitGb = ref(0)

const showDeleteCookiesConfirm = ref(false)
const showBlacklist = ref(false)
const showQuotaModal = ref(false)
const selectedUser = ref(null)
const showDeleteUserConfirm = ref(false)
const userToDelete = ref(null)

onMounted(async () => {
  await settingsStore.fetchSettings()
  proxyInput.value = settingsStore.settings.http_proxy || ''
  audioFormat.value = settingsStore.settings.audio_format || 'opus'
  maxConcurrent.value = settingsStore.settings.max_concurrent || 2
  allowRegState.value = settingsStore.settings.allow_registration || false
  if (settingsStore.settings.global_storage_limit_bytes > 0) {
    globalLimitGb.value = Math.round(settingsStore.settings.global_storage_limit_bytes / (1024 * 1024 * 1024))
  } else {
    globalLimitGb.value = 0
  }

  if (authStore.user?.is_admin) {
    adminStore.fetchUsers()
  }
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

async function toggleRegistration() {
  const next = !allowRegState.value
  allowRegState.value = next
  await adminStore.setAllowRegistration(next)
}

async function saveGlobalLimit() {
  const bytes = globalLimitGb.value > 0 ? globalLimitGb.value * 1024 * 1024 * 1024 : 0
  await adminStore.setGlobalLimit(bytes)
}

function openQuotaModal(user) {
  selectedUser.value = user
  showQuotaModal.value = true
}

async function onSaveQuota({ userId, quotaBytes }) {
  showQuotaModal.value = false
  await adminStore.updateUserQuota(userId, quotaBytes)
}

function promptDeleteUser(user) {
  userToDelete.value = user
  showDeleteUserConfirm.value = true
}

async function handleDeleteUser() {
  if (!userToDelete.value) return
  showDeleteUserConfirm.value = false
  await adminStore.deleteUser(userToDelete.value.id)
  userToDelete.value = null
}

function getQuotaPercent(used, quota) {
  if (!quota || quota <= 0) return 0
  return Math.min(100, Math.round((used / quota) * 100))
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
