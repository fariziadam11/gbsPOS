<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Tag from 'primevue/tag'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import DatePicker from 'primevue/datepicker'
import InputNumber from 'primevue/inputnumber'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'
import { useAd, useUpdateAd } from '../composables/useAds'
import { getErrorMessage } from '../api/client'
import VideoPlayer from '../components/VideoPlayer.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const adId = computed(() => Number(route.params.id))
const isEditing = ref(false)

const { data: ad, isLoading } = useAd(adId.value)
const { mutate: updateAd, isPending: isSaving } = useUpdateAd()

const storeTypeOptions = ['RETAIL', 'FNB', 'OUTFIT']

interface EditForm {
  name: string
  storeTypes: string[]
  playlistOrder: number
  isActive: boolean
  startDate: Date | null
  endDate: Date | null
  startTime: Date | null
  endTime: Date | null
}

const editForm = ref<EditForm>({
  name: '',
  storeTypes: [],
  playlistOrder: 0,
  isActive: true,
  startDate: null,
  endDate: null,
  startTime: null,
  endTime: null,
})

function parseApiDate(dateStr: string | null): Date | null {
  if (!dateStr) return null
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? null : d
}

function timeStrToDate(timeStr: string | null): Date | null {
  if (!timeStr) return null
  const [h, m, s] = timeStr.split(':').map(Number)
  const d = new Date()
  d.setHours(h, m, s || 0, 0)
  return d
}

function dateToTimeStr(date: Date | null): string | null {
  if (!date) return null
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

function dateToDateStr(date: Date | null): string | null {
  if (!date) return null
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function startEdit() {
  if (!ad.value) return
  editForm.value = {
    name: ad.value.name,
    storeTypes: [...ad.value.storeTypes],
    playlistOrder: ad.value.playlistOrder,
    isActive: ad.value.isActive,
    startDate: parseApiDate(ad.value.startDate),
    endDate: parseApiDate(ad.value.endDate),
    startTime: timeStrToDate(ad.value.startTime),
    endTime: timeStrToDate(ad.value.endTime),
  }
  isEditing.value = true
}

function saveEdit() {
  if (!ad.value) return
  const payload = {
    name: editForm.value.name,
    storeTypes: editForm.value.storeTypes,
    playlistOrder: editForm.value.playlistOrder,
    isActive: editForm.value.isActive,
    startDate: dateToDateStr(editForm.value.startDate),
    endDate: dateToDateStr(editForm.value.endDate),
    startTime: dateToTimeStr(editForm.value.startTime),
    endTime: dateToTimeStr(editForm.value.endTime),
  }
  updateAd(
    { id: ad.value.id, data: payload },
    {
      onSuccess: () => {
        toast.add({
          severity: 'success',
          summary: 'Saved',
          detail: 'Ad updated successfully.',
          life: 3000,
        })
        isEditing.value = false
      },
      onError: (err) => {
        toast.add({
          severity: 'error',
          summary: 'Error',
          detail: getErrorMessage(err),
          life: 5000,
        })
      },
    }
  )
}

function cancelEdit() {
  isEditing.value = false
}

function goBack() {
  router.push('/ads')
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not set'
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? 'Invalid date' : d.toLocaleDateString()
}

function formatTime(timeStr: string | null): string {
  if (!timeStr) return 'Not set'
  return timeStr
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
  <div class="detail-page">
    <div class="page-header">
      <Button icon="pi pi-arrow-left" text label="Back" @click="goBack" />
      <div v-if="!isLoading && ad && !isEditing" class="header-actions">
        <Button icon="pi pi-pencil" label="Edit" @click="startEdit" />
      </div>
      <div v-if="isEditing" class="header-actions">
        <Button
          icon="pi pi-check"
          label="Save"
          :loading="isSaving"
          @click="saveEdit"
        />
        <Button
          icon="pi pi-times"
          label="Cancel"
          severity="secondary"
          outlined
          @click="cancelEdit"
        />
      </div>
    </div>

    <div v-if="isLoading" class="loading-state">
      <i class="pi pi-spinner pi-spin" style="font-size: 2rem"></i>
      <p>Loading ad details...</p>
    </div>

    <div v-else-if="ad" class="detail-layout">
      <div class="detail-main">
        <Card>
          <template #title>
            <div class="detail-title">
              <span v-if="!isEditing">{{ ad.name }}</span>
              <InputText
                v-else
                v-model="editForm.name"
                placeholder="Ad name"
                style="width: 100%"
              />
            </div>
          </template>
          <template #content>
            <div class="detail-grid">
              <div class="detail-field">
                <label>Status</label>
                <div v-if="!isEditing" class="field-value">
                  <Tag
                    :value="ad.isActive ? 'Active' : 'Inactive'"
                    :severity="ad.isActive ? 'success' : 'danger'"
                  />
                </div>
                <ToggleSwitch v-else v-model="editForm.isActive" />
              </div>

              <div class="detail-field">
                <label>Store Types</label>
                <div v-if="!isEditing" class="field-value">
                  <Tag
                    v-for="st in ad.storeTypes"
                    :key="st"
                    :value="st"
                    severity="info"
                    class="store-tag"
                  />
                </div>
                <MultiSelect
                  v-else
                  v-model="editForm.storeTypes"
                  :options="storeTypeOptions"
                  placeholder="Select store types"
                  display="chip"
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>Playlist Order</label>
                <div v-if="!isEditing" class="field-value">
                  {{ ad.playlistOrder }}
                </div>
                <InputNumber
                  v-else
                  v-model="editForm.playlistOrder"
                  :min="0"
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>File</label>
                <div class="field-value">
                  {{ ad.filename }} ({{ formatFileSize(ad.fileSize) }})
                </div>
              </div>

              <div class="detail-field">
                <label>Schedule Start</label>
                <div v-if="!isEditing" class="field-value">
                  {{ formatDate(ad.startDate) }}
                </div>
                <DatePicker
                  v-else
                  v-model="editForm.startDate"
                  dateFormat="yy-mm-dd"
                  placeholder="Start date"
                  showIcon
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>Schedule End</label>
                <div v-if="!isEditing" class="field-value">
                  {{ formatDate(ad.endDate) }}
                </div>
                <DatePicker
                  v-else
                  v-model="editForm.endDate"
                  dateFormat="yy-mm-dd"
                  placeholder="End date"
                  showIcon
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>Time Start</label>
                <div v-if="!isEditing" class="field-value">
                  {{ formatTime(ad.startTime) }}
                </div>
                <DatePicker
                  v-else
                  v-model="editForm.startTime"
                  timeOnly
                  hourFormat="24"
                  placeholder="Start time"
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>Time End</label>
                <div v-if="!isEditing" class="field-value">
                  {{ formatTime(ad.endTime) }}
                </div>
                <DatePicker
                  v-else
                  v-model="editForm.endTime"
                  timeOnly
                  hourFormat="24"
                  placeholder="End time"
                  style="width: 100%"
                />
              </div>

              <div class="detail-field">
                <label>Created</label>
                <div class="field-value">
                  {{ new Date(ad.createdAt).toLocaleString() }}
                </div>
              </div>

              <div class="detail-field">
                <label>Updated</label>
                <div class="field-value">
                  {{ new Date(ad.updatedAt).toLocaleString() }}
                </div>
              </div>
            </div>
          </template>
        </Card>
      </div>

      <div class="detail-side">
        <Card>
          <template #title>Preview</template>
          <template #content>
            <VideoPlayer :adId="ad.id" :mimeType="ad.mimeType" />
          </template>
        </Card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 80px 0;
  color: var(--p-text-secondary-color);
}

.detail-layout {
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 20px;
}

@media (max-width: 1024px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
}

.detail-title {
  font-size: 20px;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.detail-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-field label {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--p-text-secondary-color);
}

.field-value {
  font-size: 14px;
  color: var(--p-text-color);
  min-height: 20px;
}

.store-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}
</style>
