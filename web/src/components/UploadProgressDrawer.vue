<template>
  <transition
    enter-active-class="transition ease-out duration-300 transform"
    enter-from-class="translate-y-12 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition ease-in duration-200 transform"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-12 opacity-0"
  >
    <div
      v-if="uploadStore.tasks.length > 0"
      :class="[
        'fixed right-4 md:right-8 z-40 max-w-md w-full sm:w-96 select-none shadow-2xl transition-all duration-300',
        playerStore.currentTrack ? 'bottom-24' : 'bottom-6'
      ]"
    >
      <!-- Minimized Floating Widget -->
      <div
        v-if="uploadStore.isMinimized"
        @click="uploadStore.isMinimized = false"
        class="bg-studio-surface/95 backdrop-blur-xl border border-zinc-700/80 hover:border-indigo-500/60 rounded-2xl p-3.5 flex items-center justify-between gap-3 shadow-xl cursor-pointer group"
      >
        <div class="flex items-center gap-3 min-w-0">
          <div class="relative w-8 h-8 rounded-lg bg-indigo-950/80 border border-indigo-700/60 flex items-center justify-center text-indigo-400 shrink-0">
            <svg
              class="w-4 h-4"
              :class="{ 'animate-bounce': uploadStore.isUploading }"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
            </svg>
            <span
              v-if="uploadStore.isUploading"
              class="absolute -top-1 -right-1 w-2.5 h-2.5 rounded-full bg-indigo-500 animate-ping"
            ></span>
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between text-xs font-medium text-zinc-200 mb-1">
              <span class="truncate">{{ uploadStore.isUploading ? i18n.t('uploadTasks.title') : i18n.t('uploadTasks.completedTitle') }}</span>
              <span class="font-mono text-indigo-300 text-[11px] ml-2">{{ uploadStore.completedTasks }}/{{ uploadStore.totalTasks }}</span>
            </div>
            <!-- Progress Bar -->
            <div class="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
              <div
                class="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300 rounded-full"
                :style="{ width: `${uploadStore.overallProgress}%` }"
              ></div>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-1 shrink-0">
          <button
            @click.stop="uploadStore.isMinimized = false"
            class="p-1 rounded text-zinc-400 hover:text-zinc-100 transition-colors"
            :title="i18n.t('uploadTasks.show')"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="18 15 12 9 6 15"></polyline>
            </svg>
          </button>
          <button
            v-if="!uploadStore.isUploading"
            @click.stop="uploadStore.clearCompleted"
            class="p-1 rounded text-zinc-400 hover:text-rose-400 transition-colors text-xs"
            :title="i18n.t('uploadTasks.clear')"
          >
            ✕
          </button>
        </div>
      </div>

      <!-- Expanded Detailed Widget Card -->
      <div
        v-else
        class="bg-studio-surface/98 backdrop-blur-2xl border border-zinc-700/80 rounded-2xl overflow-hidden shadow-2xl flex flex-col max-h-[380px]"
      >
        <!-- Header -->
        <div class="p-3.5 bg-zinc-900/80 border-b border-zinc-800 flex items-center justify-between select-none">
          <div class="flex items-center gap-2.5">
            <div class="w-6 h-6 rounded-md bg-indigo-950 border border-indigo-700/60 flex items-center justify-center text-indigo-400">
              <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
              </svg>
            </div>
            <div>
              <h4 class="text-xs font-semibold text-zinc-100 leading-none">{{ i18n.t('uploadTasks.title') }}</h4>
              <p class="text-[10px] font-mono text-zinc-400 mt-0.5">
                {{ uploadStore.completedTasks }} / {{ uploadStore.totalTasks }} ({{ uploadStore.overallProgress }}%)
              </p>
            </div>
          </div>

          <div class="flex items-center gap-1.5">
            <button
              v-if="uploadStore.completedTasks > 0 || uploadStore.failedTasks > 0"
              @click="uploadStore.clearCompleted"
              class="text-[11px] text-zinc-400 hover:text-zinc-200 px-2 py-0.5 rounded transition-colors"
            >
              {{ i18n.t('uploadTasks.clear') }}
            </button>
            <button
              @click="uploadStore.isMinimized = true"
              class="p-1 text-zinc-400 hover:text-zinc-100 rounded transition-colors"
              :title="i18n.t('uploadTasks.hide')"
            >
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="6 9 12 15 18 9"></polyline>
              </svg>
            </button>
          </div>
        </div>

        <!-- Overall Progress Bar -->
        <div class="h-1 bg-zinc-800 w-full overflow-hidden">
          <div
            class="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300"
            :style="{ width: `${uploadStore.overallProgress}%` }"
          ></div>
        </div>

        <!-- Task List -->
        <div class="overflow-y-auto p-2 space-y-1.5 flex-1 divide-y divide-zinc-800/40">
          <div
            v-for="task in uploadStore.tasks"
            :key="task.id"
            class="pt-1.5 first:pt-0 flex items-center justify-between gap-3 text-xs"
          >
            <!-- Left: Filename & Status -->
            <div class="min-w-0 flex-1">
              <p class="text-xs text-zinc-200 truncate font-medium" :title="task.name">{{ task.name }}</p>
              
              <div class="flex items-center gap-2 mt-0.5">
                <span class="text-[10px] font-mono text-zinc-500">{{ formatBytes(task.size) }}</span>
                
                <!-- Status Badge -->
                <span
                  v-if="task.status === 'uploading'"
                  class="text-[10px] font-mono text-indigo-400 flex items-center gap-1"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse"></span>
                  {{ task.progress }}%
                </span>
                <span
                  v-else-if="task.status === 'processing'"
                  class="text-[10px] font-mono text-amber-400 flex items-center gap-1 animate-pulse"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
                  {{ i18n.t('uploadTasks.processing') }}
                </span>
                <span
                  v-else-if="task.status === 'done'"
                  class="text-[10px] font-mono text-emerald-400 flex items-center gap-1"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                  {{ i18n.t('uploadTasks.done') }}
                </span>
                <span
                  v-else-if="task.status === 'error'"
                  class="text-[10px] font-mono text-rose-400 truncate max-w-[180px]"
                  :title="task.error"
                >
                  ✕ {{ task.error || i18n.t('uploadTasks.error') }}
                </span>
                <span
                  v-else
                  class="text-[10px] font-mono text-zinc-500"
                >
                  {{ i18n.t('uploadTasks.pending') }}
                </span>
              </div>
            </div>

            <!-- Right: Dismiss/Action -->
            <div class="shrink-0 flex items-center">
              <svg
                v-if="task.status === 'done'"
                class="w-4 h-4 text-emerald-400"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
              <svg
                v-else-if="task.status === 'error'"
                @click="uploadStore.removeTask(task.id)"
                class="w-4 h-4 text-rose-400 cursor-pointer hover:text-rose-300"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
              </svg>
              <svg
                v-else-if="task.status === 'uploading' || task.status === 'processing'"
                class="w-4 h-4 text-indigo-400 animate-spin"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { useUploadStore } from '../stores/upload'
import { usePlayerStore } from '../stores/player'
import { useI18nStore } from '../stores/i18n'

const uploadStore = useUploadStore()
const playerStore = usePlayerStore()
const i18n = useI18nStore()

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

