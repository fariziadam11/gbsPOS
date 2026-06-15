<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'
import { useAds, useToggleAd, useDeleteAd } from '../composables/useAds'
import { useAuthStore } from '../stores/auth'
import { getErrorMessage } from '../api/client'
import type { Ad } from '../types/api'

const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()

const page = ref(1)
const limit = ref(100)
const { data: adsData, isLoading } = useAds(page, limit)
const { mutate: toggleAd, isPending: isToggling } = useToggleAd()
const { mutate: deleteAd, isPending: isDeleting } = useDeleteAd()

function onToggle(ad: Ad) {
  toggleAd(ad.id, {
    onSuccess: () => {
      toast.add({
        severity: 'success',
        summary: 'Updated',
        detail: `${ad.name} is now ${ad.isActive ? 'inactive' : 'active'}.`,
        life: 3000,
      })
    },
    onError: (err) => {
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: getErrorMessage(err),
        life: 5000,
      })
    },
  })
}

function onDelete(ad: Ad) {
  if (!confirm(`Delete "${ad.name}"? This cannot be undone.`)) return
  deleteAd(ad.id, {
    onSuccess: () => {
      toast.add({
        severity: 'success',
        summary: 'Deleted',
        detail: `${ad.name} deleted.`,
        life: 3000,
      })
    },
    onError: (err) => {
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: getErrorMessage(err),
        life: 5000,
      })
    },
  })
}

function viewAd(ad: Ad) {
  router.push(`/ads/${ad.id}`)
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not set'
  return new Date(dateStr).toLocaleDateString()
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
  <div class="ads-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Ads</h1>
        <p class="page-subtitle">Manage advertisement videos and playlist</p>
      </div>
      <Button
        v-if="authStore.isAdmin"
        label="Upload Ad"
        icon="pi pi-upload"
        @click="router.push('/upload')"
      />
    </div>

    <div class="card">
      <DataTable
        :value="adsData?.ads || []"
        :loading="isLoading"
        stripedRows
        size="small"
        tableStyle="min-width: 60rem"
      >
        <Column field="id" header="ID" style="width: 60px" />
        <Column field="name" header="Name" sortable />
        <Column header="Status" style="width: 120px">
          <template #body="{ data }: { data: Ad }">
            <Tag
              :value="data.isActive ? 'Active' : 'Inactive'"
              :severity="data.isActive ? 'success' : 'danger'"
            />
          </template>
        </Column>
        <Column header="Store Types">
          <template #body="{ data }: { data: Ad }">
            <Tag
              v-for="st in data.storeTypes"
              :key="st"
              :value="st"
              severity="info"
              class="store-tag"
            />
          </template>
        </Column>
        <Column header="Schedule" style="width: 180px">
          <template #body="{ data }: { data: Ad }">
            <span v-if="data.startDate || data.endDate">
              {{ formatDate(data.startDate) }} - {{ formatDate(data.endDate) }}
            </span>
            <span v-else class="text-secondary">Always</span>
          </template>
        </Column>
        <Column field="playlistOrder" header="Order" sortable style="width: 90px" />
        <Column header="File" style="width: 150px">
          <template #body="{ data }: { data: Ad }">
            {{ formatFileSize(data.fileSize) }}
          </template>
        </Column>
        <Column header="Uploaded" style="width: 160px">
          <template #body="{ data }: { data: Ad }">
            {{ new Date(data.createdAt).toLocaleString() }}
          </template>
        </Column>
        <Column header="Actions" style="width: 160px">
          <template #body="{ data }: { data: Ad }">
            <div class="action-buttons">
              <Button
                icon="pi pi-eye"
                text
                rounded
                title="View"
                @click="viewAd(data)"
              />
              <ToggleSwitch
                :modelValue="data.isActive"
                :disabled="isToggling"
                @update:modelValue="onToggle(data)"
              />
              <Button
                icon="pi pi-trash"
                text
                rounded
                severity="danger"
                title="Delete"
                :disabled="isDeleting"
                @click="onDelete(data)"
              />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No ads found.</div>
        </template>
      </DataTable>
    </div>
  </div>
</template>

<style scoped>
.ads-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--p-text-color);
}

.page-subtitle {
  margin: 4px 0 0;
  color: var(--p-text-secondary-color);
  font-size: 14px;
}

.card {
  background: var(--p-surface-0);
  border-radius: 12px;
  border: 1px solid var(--p-surface-200);
  padding: 20px;
}

.store-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
  color: var(--p-text-secondary-color);
}

.text-secondary {
  color: var(--p-text-secondary-color);
}
</style>
