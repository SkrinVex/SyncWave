<template>
  <div class="min-h-screen bg-[#070709] flex items-center justify-center p-4 select-none relative overflow-hidden">
    <!-- Ambient subtle background glow -->
    <div class="absolute -top-40 left-1/2 -translate-x-1/2 w-96 h-96 bg-indigo-600/10 rounded-full blur-3xl pointer-events-none"></div>

    <div class="w-full max-w-sm relative z-10">
      <!-- Brand Logo & Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 mb-3 shadow-inner">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M12 2v20M17 5v14M22 8v8M7 8v8M2 11v2" stroke-linecap="round"/>
          </svg>
        </div>
        <h1 class="text-xl font-bold tracking-tight text-zinc-100">SyncWave</h1>
        <p class="text-xs text-zinc-400 mt-1">Self-Hosted Music Archive & Streamer</p>
      </div>

      <!-- Auth Card -->
      <div class="bg-studio-surface border border-studio-border rounded-2xl p-6 sm:p-8 shadow-2xl space-y-6">
        <div>
          <h2 class="text-base font-semibold text-zinc-100">
            {{ authStore.needsSetup ? 'Initialize Admin' : 'Welcome Back' }}
          </h2>
          <p class="text-xs text-zinc-400 mt-0.5">
            {{ authStore.needsSetup ? 'Create your primary administrator credentials' : 'Sign in to access your music library' }}
          </p>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">Username</label>
            <input
              type="text"
              v-model="username"
              required
              autocomplete="username"
              placeholder="e.g. admin"
              class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-300 mb-1.5">Password</label>
            <input
              type="password"
              v-model="password"
              required
              autocomplete="current-password"
              placeholder="••••••••"
              class="w-full bg-studio-elevated border border-studio-border rounded-lg px-3.5 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/60"
            />
          </div>

          <button
            type="submit"
            :disabled="authStore.loading"
            class="w-full py-2.5 px-4 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-lg shadow-indigo-600/20 active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <span v-if="authStore.loading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>{{ authStore.needsSetup ? 'Create Admin & Start' : 'Sign In' }}</span>
          </button>
        </form>
      </div>

      <!-- Footer Info -->
      <div class="text-center mt-6 text-[11px] font-mono text-zinc-600">
        SyncWave Daemon • Zero Cloud Lock-in
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'

const authStore = useAuthStore()
const toast = useToastStore()

const username = ref('')
const password = ref('')

async function handleSubmit() {
  if (!username.value || !password.value) return

  try {
    if (authStore.needsSetup) {
      await authStore.setupAdmin(username.value, password.value)
      toast.success('Admin account created successfully!')
    } else {
      await authStore.login(username.value, password.value)
      toast.success('Welcome back!')
    }
  } catch (e) {
    toast.error(e.message || 'Authentication failed')
  }
}
</script>

