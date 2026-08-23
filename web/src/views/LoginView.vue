<template>
  <div class="min-h-screen bg-studio-bg flex items-center justify-center p-4 select-none relative overflow-hidden">
    <div class="w-full max-w-sm relative z-10">
      <!-- Brand Logo & Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-studio-surface border border-studio-border text-indigo-400 mb-3.5 shadow-md">
          <!-- Bespoke Modern Studio Soundwave Logo -->
          <svg class="w-8 h-8" viewBox="0 0 32 32" fill="none">
            <rect x="4" y="12" width="2.5" height="8" rx="1.25" fill="currentColor" opacity="0.6"/>
            <rect x="9.5" y="8" width="2.5" height="16" rx="1.25" fill="currentColor" opacity="0.8"/>
            <rect x="15" y="4" width="2.5" height="24" rx="1.25" fill="#818cf8"/>
            <rect x="20.5" y="8" width="2.5" height="16" rx="1.25" fill="currentColor" opacity="0.8"/>
            <rect x="26" y="12" width="2.5" height="8" rx="1.25" fill="currentColor" opacity="0.6"/>
          </svg>
        </div>
        <h1 class="text-xl font-bold tracking-tight text-zinc-100">SyncWave</h1>
        <p class="text-xs text-zinc-400 mt-1">{{ i18n.t('auth.subtitle') }}</p>
      </div>

      <!-- Auth Card -->
      <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 sm:p-8 shadow-xl space-y-6">
        <!-- Tab Switcher (When not in first-time setup and registration is allowed) -->
        <div v-if="!authStore.needsSetup && authStore.allowRegistration" class="flex p-1 bg-studio-elevated border border-studio-border rounded-xl">
          <button
            type="button"
            @click="authMode = 'login'"
            :class="[
              'flex-1 py-1.5 text-xs font-semibold rounded-lg transition-all',
              authMode === 'login' ? 'bg-indigo-600 text-white shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
            ]"
          >
            {{ i18n.t('auth.tabLogin') }}
          </button>
          <button
            type="button"
            @click="authMode = 'register'"
            :class="[
              'flex-1 py-1.5 text-xs font-semibold rounded-lg transition-all',
              authMode === 'register' ? 'bg-indigo-600 text-white shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
            ]"
          >
            {{ i18n.t('auth.tabRegister') }}
          </button>
        </div>

        <div>
          <h2 class="text-base font-semibold text-zinc-100">
            <template v-if="authStore.needsSetup">
              {{ i18n.t('auth.initAdmin') }}
            </template>
            <template v-else-if="authMode === 'register'">
              {{ i18n.t('auth.registerTitle') }}
            </template>
            <template v-else>
              {{ i18n.t('auth.welcomeBack') }}
            </template>
          </h2>
          <p class="text-xs text-zinc-400 mt-0.5">
            <template v-if="authStore.needsSetup">
              {{ i18n.t('auth.initAdminPrompt') }}
            </template>
            <template v-else-if="authMode === 'register'">
              {{ i18n.t('auth.registerPrompt') }}
            </template>
            <template v-else>
              {{ i18n.t('auth.signInPrompt') }}
            </template>
          </p>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('auth.username') }}</label>
            <input
              type="text"
              v-model="username"
              required
              autocomplete="username"
              :placeholder="i18n.t('auth.usernamePlaceholder')"
              class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">{{ i18n.t('auth.password') }}</label>
            <div class="relative">
              <input
                :type="showPassword ? 'text' : 'password'"
                v-model="password"
                required
                autocomplete="current-password"
                placeholder="••••••••"
                class="w-full bg-studio-elevated border border-studio-border rounded-lg pl-3.5 pr-10 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 p-1 transition-colors"
                :title="showPassword ? i18n.t('auth.hidePassword') : i18n.t('auth.showPassword')"
              >
                <!-- Eye open -->
                <svg v-if="!showPassword" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
                <!-- Eye closed -->
                <svg v-else class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/>
                  <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/>
                  <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/>
                  <line x1="2" y1="2" x2="22" y2="22"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Solid, Clean Button (Zero Gradient) -->
          <button
            type="submit"
            :disabled="authStore.loading"
            class="w-full py-2.5 px-4 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-colors shadow-sm active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2 mt-2"
          >
            <span v-if="authStore.loading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>
              <template v-if="authStore.needsSetup">
                {{ i18n.t('auth.createAdminButton') }}
              </template>
              <template v-else-if="authMode === 'register'">
                {{ i18n.t('auth.registerButton') }}
              </template>
              <template v-else>
                {{ i18n.t('auth.signInButton') }}
              </template>
            </span>
          </button>
        </form>
      </div>

      <!-- Professional Slogan / Footer -->
      <div class="text-center mt-6 text-xs text-zinc-400 select-none">
        {{ i18n.t('auth.slogan') }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useI18nStore } from '../stores/i18n'
import { useToastStore } from '../stores/toast'

const authStore = useAuthStore()
const i18n = useI18nStore()
const toast = useToastStore()

const authMode = ref('login') // 'login' | 'register'
const username = ref('')
const password = ref('')
const showPassword = ref(false)

async function handleSubmit() {
  if (!username.value || !password.value) return

  try {
    if (authStore.needsSetup) {
      await authStore.setupAdmin(username.value, password.value)
      toast.success('Администратор успешно создан!')
    } else if (authMode.value === 'register') {
      await authStore.register(username.value, password.value)
      toast.success('Аккаунт успешно зарегистрирован!')
    } else {
      await authStore.login(username.value, password.value)
      toast.success('Добро пожаловать в SyncWave!')
    }
  } catch (e) {
    toast.error(e.message || 'Ошибка авторизации')
  }
}
</script>
