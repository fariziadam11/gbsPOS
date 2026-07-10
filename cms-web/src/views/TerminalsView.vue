<script setup lang="ts">
import { ref, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { useTerminals } from '../composables/useTerminals'
import type { Terminal } from '../api/cartDisplay'

const { data: terminals, isLoading } = useTerminals()

const showDetailDialog = ref(false)
const selectedTerminal = ref<Terminal | null>(null)

const parsedPayload = computed(() => {
  if (!selectedTerminal.value?.payload) return null
  try {
    return JSON.parse(selectedTerminal.value.payload)
  } catch {
    return null
  }
})

function openDetail(terminal: Terminal) {
  selectedTerminal.value = terminal
  showDetailDialog.value = true
}

function formatDistanceToNow(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffHour < 24) return `${diffHour}h ago`
  if (diffDay < 7) return `${diffDay}d ago`

  return date.toLocaleDateString('id-ID')
}

function getStatusSeverity(status: string | undefined): string {
  switch (status) {
    case 'Transaksi':
      return 'info'
    case 'Selesai':
      return 'success'
    default:
      return 'secondary'
  }
}

function formatRupiah(value: string | number | undefined): string {
  if (!value) return 'Rp 0'
  const num = typeof value === 'string' ? parseInt(value.replace(/\./g, ''), 10) : value
  return `Rp ${num.toLocaleString('id-ID')}`
}

function getDeviceDisplayName(terminal: Terminal): string {
  const parts = [
    terminal.deviceManufacturer || 'Unknown',
    terminal.deviceModel || '',
  ].filter(Boolean)
  return parts.length > 1 ? parts.join(' ') : 'Unknown Device'
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Terminals</h1>

    <DataTable
      :value="terminals || []"
      :loading="isLoading"
      class="p-datatable-sm"
      stripedRows
      paginator
      :rows="20"
      :rowsPerPageOptions="[10, 20, 50]"
    >
      <Column header="Terminal ID" field="terminalId" sortable style="min-width: 200px">
        <template #body="{ data }">
          <span class="font-mono text-sm">{{ data.terminalId }}</span>
        </template>
      </Column>

      <Column header="Device" sortable style="min-width: 180px">
        <template #body="{ data }">
          <div>{{ getDeviceDisplayName(data) }}</div>
          <div class="text-xs text-color-secondary">{{ data.deviceBrand || 'N/A' }}</div>
        </template>
      </Column>

      <Column header="Android" sortable style="min-width: 120px">
        <template #body="{ data }">
          <div>{{ data.androidVersion || 'N/A' }}</div>
          <div class="text-xs text-color-secondary">SDK {{ data.sdkInt || '?' }}</div>
        </template>
      </Column>

      <Column header="App Version" sortable style="min-width: 120px">
        <template #body="{ data }">
          {{ data.appVersion || 'N/A' }}
        </template>
      </Column>

      <Column header="Status" sortable style="min-width: 120px">
        <template #body="{ data }">
          <Tag
            :value="parsedPayload?.TeksSelesai || 'None'"
            :severity="getStatusSeverity(parsedPayload?.TeksSelesai)"
          />
        </template>
      </Column>

      <Column header="Last Updated" sortable style="min-width: 120px">
        <template #body="{ data }">
          {{ formatDistanceToNow(data.updatedAt) }}
        </template>
      </Column>

      <Column header="" style="width: 80px">
        <template #body="{ data }">
          <Button icon="pi pi-eye" severity="secondary" text @click="openDetail(data)" />
        </template>
      </Column>

      <template #empty>
        <div class="text-center p-4 text-color-secondary">
          No terminals found. Terminals will appear here once they sync with the CMS.
        </div>
      </template>
    </DataTable>

    <Dialog
      v-model:visible="showDetailDialog"
      :header="`Terminal: ${selectedTerminal?.terminalId?.substring(0, 12)}...`"
      modal
      class="w-full max-w-2xl"
    >
      <div v-if="selectedTerminal" class="grid grid-cols-2 gap-4">
        <!-- Device Info -->
        <div class="col-span-2">
          <h3 class="text-lg font-semibold mb-3">Device Information</h3>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Device</div>
          <div>{{ getDeviceDisplayName(selectedTerminal) }}</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Brand</div>
          <div>{{ selectedTerminal.deviceBrand || 'N/A' }}</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Android Version</div>
          <div>{{ selectedTerminal.androidVersion || 'N/A' }} (SDK {{ selectedTerminal.sdkInt || '?' }})</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">App Version</div>
          <div>{{ selectedTerminal.appVersion || 'N/A' }}</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Last Updated</div>
          <div>{{ new Date(selectedTerminal.updatedAt).toLocaleString('id-ID') }}</div>
        </div>

        <!-- Cart Info -->
        <div class="col-span-2 mt-4">
          <h3 class="text-lg font-semibold mb-3">Current Cart Display</h3>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Status</div>
          <Tag
            :value="parsedPayload?.TeksSelesai || 'None'"
            :severity="getStatusSeverity(parsedPayload?.TeksSelesai)"
          />
        </div>
        <div>
          <div class="text-sm text-color-secondary">Cashier</div>
          <div>{{ parsedPayload?.Initial?.NamaKasir || 'N/A' }}</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Store</div>
          <div>{{ parsedPayload?.Initial?.NamaToko || 'N/A' }}</div>
        </div>
        <div>
          <div class="text-sm text-color-secondary">Items</div>
          <div>{{ parsedPayload?.DaftarBelanja?.length || 0 }}</div>
        </div>

        <!-- Cart Items -->
        <div class="col-span-2 mt-4" v-if="parsedPayload?.DaftarBelanja?.length">
          <h4 class="text-md font-medium mb-2">Items</h4>
          <DataTable :value="parsedPayload.DaftarBelanja" size="small" class="text-sm">
            <Column field="Deskripsi" header="Item" />
            <Column field="Harga" header="Price" style="width: 120px" />
            <Column field="Qty" header="Qty" style="width: 60px" />
            <Column field="Total" header="Total" style="width: 120px" />
          </DataTable>
        </div>

        <!-- Summary -->
        <div class="col-span-2 mt-4" v-if="parsedPayload?.Summary">
          <h4 class="text-md font-medium mb-2">Summary</h4>
          <div class="grid grid-cols-4 gap-4">
            <div>
              <div class="text-sm text-color-secondary">Subtotal (Hemat)</div>
              <div>{{ parsedPayload.Summary?.Hemat || '0' }}</div>
            </div>
            <div>
              <div class="text-sm text-color-secondary">Total</div>
              <div class="font-semibold">{{ parsedPayload.Summary?.Total || '0' }}</div>
            </div>
            <div>
              <div class="text-sm text-color-secondary">Bayar</div>
              <div>{{ parsedPayload.Summary?.Bayar || '0' }}</div>
            </div>
            <div>
              <div class="text-sm text-color-secondary">Kembali</div>
              <div>{{ parsedPayload.Summary?.Kembali || '0' }}</div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetailDialog = false" />
      </template>
    </Dialog>
  </div>
</template>
