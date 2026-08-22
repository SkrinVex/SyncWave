<template>
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none max-w-sm w-full">
    <transition-group
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
      enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-for="t in toastStore.toasts"
        :key="t.id"
        class="pointer-events-auto flex items-center justify-between p-3.5 rounded-lg border shadow-xl backdrop-blur-md text-sm font-medium"
        :class="{
          'bg-zinc-900/95 border-zinc-800 text-zinc-200': t.type === 'info',
          'bg-emerald-950/90 border-emerald-800/60 text-emerald-200': t.type === 'success',
          'bg-rose-950/90 border-rose-800/60 text-rose-200': t.type === 'error'
        }"
      >
        <div class="flex items-center gap-2.5">
          <span v-if="t.type === 'success'" class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span v-else-if="t.type === 'error'" class="w-2 h-2 rounded-full bg-rose-400"></span>
          <span v-else class="w-2 h-2 rounded-full bg-indigo-400"></span>
          <span>{{ t.message }}</span>
        </div>
        <button
          @click="toastStore.remove(t.id)"
          class="ml-3 text-zinc-400 hover:text-zinc-100 text-xs px-1 py-0.5 rounded transition-colors"
        >
          ✕
        </button>
      </div>
    </transition-group>
  </div>
</template>

<script setup>
import { useToastStore } from '../stores/toast'
const toastStore = useToastStore()
</script>

