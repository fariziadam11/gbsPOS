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
  <div class="upload-page">
    <div class="page-header">
      <h1 class="page-title">Upload New Ad</h1>
      <p class="page-subtitle">Upload a video advertisement and configure its settings</p>
    </div>

    <Card class="upload-card">
      <template #content>
        <form @submit.prevent="handleSubmit" class="upload-form">
          <div class="form-section">
            <label class="section-label">Video File</label>
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

          <div class="form-section">
            <label class="section-label">Ad Name</label>
            <InputText
              v-model="form.name"
              placeholder="Enter ad name"
              style="width: 100%"
            />
          </div>

          <div class="form-row">
            <div class="form-section flex-1">
              <label class="section-label">Store Types</label>
              <MultiSelect
                v-model="form.storeTypes"
                :options="storeTypeOptions"
                placeholder="Select store types"
                display="chip"
                style="width: 100%"
              />
            </div>
            <div class="form-section flex-1" style="min-width: 140px">
              <label class="section-label">Playlist Order</label>
              <InputNumber
                v-model="form.playlistOrder"
                :min="0"
                style="width: 100%"
              />
            </div>
            <div class="form-section flex-none">
              <label class="section-label">Active</label>
              <div class="switch-wrapper">
                <ToggleSwitch v-model="form.isActive" />
                <span class="switch-label">{{ form.isActive ? 'Yes' : 'No' }}</span>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-section flex-1">
              <label class="section-label">Schedule Start Date</label>
              <DatePicker
                v-model="form.startDate"
                dateFormat="yy-mm-dd"
                placeholder="Start date"
                showIcon
                style="width: 100%"
              />
            </div>
            <div class="form-section flex-1">
              <label class="section-label">Schedule End Date</label>
              <DatePicker
                v-model="form.endDate"
                dateFormat="yy-mm-dd"
                placeholder="End date"
                showIcon
                style="width: 100%"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-section flex-1">
              <label class="section-label">Start Time</label>
              <DatePicker
                v-model="form.startTime"
                timeOnly
                hourFormat="24"
                placeholder="Start time"
                style="width: 100%"
              />
            </div>
            <div class="form-section flex-1">
              <label class="section-label">End Time</label>
              <DatePicker
                v-model="form.endTime"
                timeOnly
                hourFormat="24"
                placeholder="End time"
                style="width: 100%"
              />
            </div>
          </div>

          <div class="form-actions">
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
.upload-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 800px;
}

.page-header {
  margin-bottom: 4px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.page-subtitle {
  margin: 4px 0 0;
  color: var(--p-text-secondary-color);
  font-size: 14px;
}

.upload-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--p-text-secondary-color);
}

.form-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.flex-1 {
  flex: 1;
  min-width: 200px;
}

.flex-none {
  flex: 0 0 auto;
  min-width: 120px;
}

.switch-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 40px;
}

.switch-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--p-text-color);
  user-select: none;
}

.file-area {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  padding: 24px;
  border: 2px dashed var(--p-surface-300);
  border-radius: 10px;
  background: var(--p-surface-50);
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--p-surface-200);
}
</style>
