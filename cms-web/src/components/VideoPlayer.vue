<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import apiClient from '../api/client'

const props = defineProps<{
  adId: number
  mimeType?: string
}>()

const videoUrl = ref<string | null>(null)
const isLoading = ref(false)
const error = ref<string | null>(null)

let currentObjectUrl: string | null = null

async function loadVideo() {
  if (currentObjectUrl) {
    URL.revokeObjectURL(currentObjectUrl)
    currentObjectUrl = null
  }
  videoUrl.value = null
  error.value = null
  isLoading.value = true

  try {
    const response = await apiClient.get(`/v1/ads/download/${props.adId}`, {
      responseType: 'blob',
    })
    currentObjectUrl = URL.createObjectURL(response.data as Blob)
    videoUrl.value = currentObjectUrl
  } catch (err: any) {
    if (err.response?.status === 401) {
      error.value = 'Authentication required to preview video.'
    } else {
      error.value = 'Failed to load video preview.'
    }
  } finally {
    isLoading.value = false
  }
}

watch(() => props.adId, loadVideo, { immediate: true })

onUnmounted(() => {
  if (currentObjectUrl) {
    URL.revokeObjectURL(currentObjectUrl)
  }
})
</script>

<template>
  <div class="video-player-wrapper">
    <div v-if="isLoading" class="video-state">
      <i class="pi pi-spinner pi-spin" style="font-size: 1.5rem"></i>
      <span>Loading preview...</span>
    </div>
    <div v-else-if="error" class="video-state error">
      <i class="pi pi-exclamation-circle" style="font-size: 1.5rem"></i>
      <span>{{ error }}</span>
    </div>
    <video
      v-else-if="videoUrl"
      controls
      class="video-player"
      preload="metadata"
    >
      <source :src="videoUrl" :type="mimeType || 'video/mp4'" />
      Your browser does not support the video tag.
    </video>
  </div>
</template>

<style scoped>
.video-player-wrapper {
  background: var(--p-surface-900);
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  aspect-ratio: 16 / 9;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.video-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--p-text-secondary-color);
  padding: 24px;
  text-align: center;
}

.video-state.error {
  color: var(--p-red-400);
}
</style>
