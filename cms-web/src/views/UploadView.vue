<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import DatePicker from 'primevue/datepicker'
import InputNumber from 'primevue/inputnumber'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'
import { useCreateAd } from '../composables/useAds'
import { getErrorMessage } from '../api/client'
import type { CreateAdRequest } from '../types/api'

const router = useRouter()
const toast = useToast()

const { mutate: createAd, isPending } = useCreateAd()

const selectedFile = ref<File | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

const storeTypeOptions = ['RETAIL', 'FNB', 'OUTFIT']

const form = ref({
  name: '',
  storeTypes: [] as string[],
  playlistOrder: 0,
  isActive: true,
  startDate: null as Date | null,
  endDate: null as Date | null,
  startTime: null as Date | null,
  endTime: null as Date | null,
})

function triggerFileSelect() {
  fileInputRef.value?.click()
}

function onFileSelected(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
  }
}

function dateToDateStr(date: Date | null): string | undefined {
  if (!date) return undefined
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function dateToTimeStr(date: Date | null): string | undefined {
  if (!date) return undefined
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

function handleSubmit() {
  if (!selectedFile.value) {
    toast.add({
      severity: 'warn',
      summary: 'Missing File',
      detail: 'Please select a video file to upload.',
      life: 4000,
    })
    return
  }
  if (!form.value.name) {
    toast.add({
      severity: 'warn',
      summary: 'Missing Name',
      detail: 'Please enter an ad name.',
      life: 4000,
    })
    return
  }
  if (form.value.storeTypes.length === 0) {
    toast.add({
      severity: 'warn',
      summary: 'Missing Store Types',
      detail: 'Please select at least one store type.',
      life: 4000,
    })
    return
  }

  const payload: CreateAdRequest = {
    file: selectedFile.value,
    name: form.value.name,
    storeTypes: form.value.storeTypes,
    playlistOrder: form.value.playlistOrder,
    startDate: dateToDateStr(form.value.startDate),
    endDate: dateToDateStr(form.value.endDate),
    startTime: dateToTimeStr(form.value.startTime),
    endTime: dateToTimeStr(form.value.endTime),
  }

  createAd(payload, {
    onSuccess: () => {
      toast.add({
        severity: 'success',
        summary: 'Uploaded',
        detail: 'Ad uploaded successfully.',
        life: 3000,
      })
      router.push('/ads')
    },
    onError: (err) => {
      toast.add({
        severity: 'error',
        summary: 'Upload Failed',
        detail: getErrorMessage(err),
        life: 5000,
      })
    },
  })
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="flex flex-column gap-3 lg:gap-4">
    <div class="flex flex-column md:flex-row md:align-items-start justify-content-between gap-3">
      <div>
        <h1 class="text-2xl lg:text-3xl font-semibold text-color m-0">Upload New Ad</h1>
        <p class="text-sm text-color-secondary mt-1 mb-0">Upload a video advertisement and configure its settings</p>
      </div>
      <div class="flex flex-wrap align-items-center gap-2"></div>
    </div>

    <Card>
      <template #content>
        <form @submit.prevent="handleSubmit" class="flex flex-column gap-4">
          <div class="flex flex-column gap-2">
            <label class="text-xs font-semibold uppercase text-color-secondary">Video File</label>
            <input
              ref="fileInputRef"
              type="file"
              accept="video/mp4,video/webm,video/quicktime"
              style="display: none"
              @change="onFileSelected"
            />
            <div class="file-area">
              <Button
                type="button"
                icon="pi pi-video"
                label="Select Video"
                outlined
                @click="triggerFileSelect"
              />
              <div v-if="selectedFile" class="file-info">
                <i class="pi pi-check-circle" style="color: var(--p-green-500)"></i>
                <span class="file-name">{{ selectedFile.name }}</span>
                <span class="file-size">({{ formatFileSize(selectedFile.size) }})</span>
              </div>
              <p v-else class="file-hint">
                Accepted formats: .mp4, .webm, .mov | Max size: 50MB
              </p>
            </div>
          </div>

          <div class="flex flex-column gap-2">
            <label class="text-xs font-semibold uppercase text-color-secondary">Ad Name</label>
            <InputText
              v-model="form.name"
              placeholder="Enter ad name"
              class="w-full"
            />
          </div>

          <div class="flex flex-wrap gap-3">
            <div class="flex-1" style="min-width: 200px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Store Types</label>
                <MultiSelect
                  v-model="form.storeTypes"
                  :options="storeTypeOptions"
                  placeholder="Select store types"
                  display="chip"
                  class="w-full"
                />
              </div>
            </div>
            <div class="flex-1" style="min-width: 140px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Playlist Order</label>
                <InputNumber
                  v-model="form.playlistOrder"
                  :min="0"
                  class="w-full"
                />
              </div>
            </div>
            <div class="flex-none" style="min-width: 120px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Active</label>
                <div class="flex align-items-center gap-2" style="min-height: 40px">
                  <ToggleSwitch v-model="form.isActive" />
                  <span class="text-sm font-medium text-color">{{ form.isActive ? 'Yes' : 'No' }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <div class="flex-1" style="min-width: 200px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Schedule Start Date</label>
                <DatePicker
                  v-model="form.startDate"
                  dateFormat="yy-mm-dd"
                  placeholder="Start date"
                  showIcon
                  class="w-full"
                />
              </div>
            </div>
            <div class="flex-1" style="min-width: 200px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Schedule End Date</label>
                <DatePicker
                  v-model="form.endDate"
                  dateFormat="yy-mm-dd"
                  placeholder="End date"
                  showIcon
                  class="w-full"
                />
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <div class="flex-1" style="min-width: 200px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">Start Time</label>
                <DatePicker
                  v-model="form.startTime"
                  timeOnly
                  hourFormat="24"
                  placeholder="Start time"
                  class="w-full"
                />
              </div>
            </div>
            <div class="flex-1" style="min-width: 200px">
              <div class="flex flex-column gap-2">
                <label class="text-xs font-semibold uppercase text-color-secondary">End Time</label>
                <DatePicker
                  v-model="form.endTime"
                  timeOnly
                  hourFormat="24"
                  placeholder="End time"
                  class="w-full"
                />
              </div>
            </div>
          </div>

          <div class="flex justify-content-end gap-2 pt-3 border-top-1 surface-border">
            <Button
              type="button"
              label="Cancel"
              severity="secondary"
              outlined
              @click="router.push('/ads')"
            />
            <Button
              type="submit"
              label="Upload Ad"
              icon="pi pi-upload"
              :loading="isPending"
            />
          </div>
        </form>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.file-area {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  padding: 24px;
  border: 2px dashed light-dark(var(--p-surface-300), var(--p-surface-600));
  border-radius: 10px;
  background: light-dark(var(--p-surface-50), var(--p-surface-900));
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.file-name {
  font-weight: 500;
  color: var(--p-text-color);
}

.file-size {
  color: var(--p-text-secondary-color);
}

.file-hint {
  margin: 0;
  font-size: 13px;
  color: var(--p-text-secondary-color);
}
</style>
