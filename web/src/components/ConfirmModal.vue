<template>
  <transition
    enter-active-class="transition ease-out duration-200"
    enter-from-class="opacity-0 scale-95"
    enter-to-class="opacity-100 scale-100"
    leave-active-class="transition ease-in duration-150"
    leave-from-class="opacity-100 scale-100"
    leave-to-class="opacity-0 scale-95"
  >
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm select-none"
      @click.self="$emit('cancel')"
    >
      <div class="bg-studio-surface border border-studio-border rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5 text-left">
        <!-- Header & Icon -->
        <div class="flex items-start gap-4">
          <div
            :class="[
              'w-10 h-10 rounded-xl flex items-center justify-center shrink-0',
              danger ? 'bg-rose-950/60 border border-rose-800/60 text-rose-400' : 'bg-indigo-950/60 border border-indigo-800/60 text-indigo-400'
            ]"
          >
            <svg v-if="danger" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 9v4M12 17h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/>
            </svg>
            <svg v-else class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h3 class="text-base font-semibold text-zinc-100 leading-tight">
              {{ title }}
            </h3>
            <p class="text-xs text-zinc-400 mt-1.5 leading-relaxed">
              {{ description }}
            </p>
          </div>
        </div>

        <!-- Buttons Footer -->
        <div class="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            @click="$emit('cancel')"
            class="px-4 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 rounded-lg hover:bg-studio-hover transition-colors"
          >
            {{ cancelText || 'Отмена' }}
          </button>

          <button
            type="button"
            @click="$emit('confirm')"
            :class="[
              'px-4 py-2 text-xs font-semibold rounded-lg transition-all shadow-md active:scale-95 text-white',
              danger
                ? 'bg-rose-600 hover:bg-rose-500'
                : 'bg-indigo-600 hover:bg-indigo-500'
            ]"
          >
            {{ confirmText || 'Подтвердить' }}
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  confirmText: {
    type: String,
    default: '',
  },
  cancelText: {
    type: String,
    default: '',
  },
  danger: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['confirm', 'cancel'])

function handleKeydown(e) {
  if (!props.open) return
  if (e.key === 'Escape') {
    emit('cancel')
  } else if (e.key === 'Enter') {
    emit('confirm')
  }
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onUnmounted(() => window.removeEventListener('keydown', handleKeydown))
</script>

