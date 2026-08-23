<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-black/80 backdrop-blur-sm" @click="$emit('close')"></div>

    <!-- Modal Card -->
    <div class="relative w-full max-w-md bg-studio-surface border border-studio-border rounded-2xl p-6 shadow-2xl space-y-5 text-left">
      <div class="flex items-center justify-between">
        <h3 class="text-base font-semibold text-zinc-100">Настройка квоты диска</h3>
        <button @click="$emit('close')" class="text-zinc-400 hover:text-zinc-200 p-1 transition-colors">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="space-y-1">
        <p class="text-xs text-zinc-400">Пользователь: <span class="font-semibold text-zinc-200">{{ user?.username }}</span></p>
        <p class="text-xs text-zinc-400">Текущее использование: <span class="font-mono text-indigo-400">{{ formatBytes(user?.storage_used_bytes) }}</span></p>
      </div>

      <div class="space-y-3">
        <label class="block text-xs font-medium text-zinc-300">Лимит дискового пространства</label>
        
        <div class="grid grid-cols-4 gap-2">
          <button
            type="button"
            @click="quotaGb = 5"
            :class="['py-2 text-xs font-medium rounded-lg border transition-all', quotaGb === 5 ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-studio-elevated border-studio-border text-zinc-300 hover:text-white']"
          >
            5 ГБ
          </button>
          <button
            type="button"
            @click="quotaGb = 10"
            :class="['py-2 text-xs font-medium rounded-lg border transition-all', quotaGb === 10 ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-studio-elevated border-studio-border text-zinc-300 hover:text-white']"
          >
            10 ГБ
          </button>
          <button
            type="button"
            @click="quotaGb = 25"
            :class="['py-2 text-xs font-medium rounded-lg border transition-all', quotaGb === 25 ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-studio-elevated border-studio-border text-zinc-300 hover:text-white']"
          >
            25 ГБ
          </button>
          <button
            type="button"
            @click="quotaGb = 0"
            :class="['py-2 text-xs font-medium rounded-lg border transition-all', quotaGb === 0 ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-studio-elevated border-studio-border text-zinc-300 hover:text-white']"
          >
            Безлимит
          </button>
        </div>

        <div class="flex items-center gap-2 pt-2">
          <input
            type="number"
            v-model.number="quotaGb"
            min="0"
            step="1"
            class="flex-1 bg-studio-elevated border border-studio-border rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500/60"
            placeholder="Свой размер (ГБ)"
          />
          <span class="text-xs text-zinc-400 font-medium">ГБ (0 = без ограничений)</span>
        </div>
      </div>

      <div class="flex items-center justify-end gap-2 pt-3 border-t border-studio-border">
        <button
          type="button"
          @click="$emit('close')"
          class="px-4 py-2 text-xs font-medium text-zinc-300 hover:text-white bg-studio-elevated hover:bg-studio-hover rounded-lg transition-colors"
        >
          Отмена
        </button>
        <button
          type="button"
          @click="save"
          class="px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-all shadow-md active:scale-95"
        >
          Сохранить квоту
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  open: Boolean,
  user: Object,
})

const emit = defineEmits(['close', 'save'])

const quotaGb = ref(10)

watch(() => props.user, (u) => {
  if (u) {
    if (u.storage_quota_bytes > 0) {
      quotaGb.value = Math.round(u.storage_quota_bytes / (1024 * 1024 * 1024))
    } else {
      quotaGb.value = 0
    }
  }
}, { immediate: true })

function save() {
  const bytes = quotaGb.value > 0 ? quotaGb.value * 1024 * 1024 * 1024 : 0
  emit('save', { userId: props.user.id, quotaBytes: bytes })
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>
