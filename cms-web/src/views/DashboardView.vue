<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Paginator from 'primevue/paginator'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '../stores/auth'
import { useAds, useToggleAd, useDeleteAd } from '../composables/useAds'
import { getErrorMessage } from '../api/client'
import type { Ad } from '../types/api'

const router = useRouter()
const authStore = useAuthStore()
const confirm = useConfirm()
const toast = useToast()

const page = ref(1)
const limit = ref(20)

const { data: adsData, isLoading } = useAds(page, limit)
const { mutate: toggleAd } = useToggleAd()
const { mutate: deleteAd } = useDeleteAd()

function onPageChange(event: { page: number; rows: number }) {
  page.value = event.page + 1
  limit.value = event.rows
}

function viewAd(id: number) {
  router.push(`/ads/${id}`)
}

function handleToggle(ad: Ad) {
  toggleAd(ad.id, {
    onSuccess: () => {
      toast.add({
        severity: 'success',
        summary: 'Updated',
        detail: `Ad "${ad.name}" status changed.`,
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

function confirmDelete(ad: Ad) {
  confirm.require({
    message: `Are you sure you want to delete "${ad.name}"?`,
    header: 'Confirm Delete',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel',
    rejectProps: { severity: 'secondary', outlined: true },
    acceptLabel: 'Delete',
    acceptProps: { severity: 'danger' },
    accept: () => {
      deleteAd(ad.id, {
        onSuccess: () => {
          toast.add({
            severity: 'success',
            summary: 'Deleted',
            detail: `Ad "${ad.name}" has been deleted.`,
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
    },
  })
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-'
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
  <div class="dashboard-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Ad Management</h1>
        <p class="page-subtitle">Manage video advertisements for all store types</p>
      </div>
      <Button
        v-if="authStore.isAdmin"
        label="Upload New Ad"
        icon="pi pi-plus"
        @click="router.push('/upload')"
      />
    </div>

    <div class="card">
      <DataTable
        :value="adsData?.ads || []"
        :loading="isLoading"
        tableStyle="min-width: 60rem"
        stripedRows
        paginatorTemplate=""
      >
        <Column field="id" header="ID" sortable style="width: 60px" />
        <Column field="name" header="Name" sortable />
        <Column field="filename" header="File" sortable>
          <template #body="{ data }">
            <span :title="data.filename" class="filename-text">{{ data.filename }}</span>
          </template>
        </Column>
        <Column field="fileSize" header="Size" sortable>
          <template #body="{ data }">
            {{ formatFileSize(data.fileSize) }}
          </template>
        </Column>
        <Column field="storeTypes" header="Store Types">
          <template #body="{ data }">
            <Tag
              v-for="st in data.storeTypes"
              :key="st"
              :value="st"
              severity="info"
              class="store-tag"
            />
          </template>
        </Column>
        <Column field="playlistOrder" header="Order" sortable style="width: 80px" />
        <Column field="isActive" header="Active" style="width: 100px">
          <template #body="{ data }">
            <ToggleSwitch
              :modelValue="data.isActive"
              @update:modelValue="() => handleToggle(data)"
            />
          </template>
        </Column>
        <Column field="startDate" header="Start" sortable>
          <template #body="{ data }">
            {{ formatDate(data.startDate) }}
          </template>
        </Column>
        <Column field="endDate" header="End" sortable>
          <template #body="{ data }">
            {{ formatDate(data.endDate) }}
          </template>
        </Column>
        <Column header="Actions" style="width: 160px">
          <template #body="{ data }">
            <div class="actions">
              <Button
                icon="pi pi-eye"
                text
                rounded
                title="View"
                @click="viewAd(data.id)"
              />
              <Button
                icon="pi pi-trash"
                text
                rounded
                severity="danger"
                title="Delete"
                @click="confirmDelete(data)"
              />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No ads found.</div>
        </template>
      </DataTable>

      <Paginator
        :rows="limit"
        :totalRecords="adsData?.pagination.total || 0"
        :rowsPerPageOptions="[10, 20, 50]"
        :first="(page - 1) * limit"
        @page="onPageChange"
        template="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink RowsPerPageDropdown"
        class="table-paginator"
      />
    </div>

    <ConfirmDialog />
  </div>
</template>

<style scoped>
.dashboard-page {
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
  padding: 16px;
}

.store-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.actions {
  display: flex;
  gap: 4px;
}

.filename-text {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--p-text-secondary-color);
}

.table-paginator {
  border-top: 1px solid var(--p-surface-200);
  padding-top: 12px;
  margin-top: 8px;
}
</style>
