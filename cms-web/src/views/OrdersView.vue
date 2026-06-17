<script setup lang="ts">
import { ref, watch } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import DatePicker from 'primevue/datepicker'
import { useOrders } from '../composables/useOrders'
import type { Order } from '../types/api'
import type { OrderFilters } from '../api/orders'

const filters = ref<OrderFilters>({
  storeType: undefined,
  isVoided: false,
  isSettled: undefined,
  paymentMethod: undefined,
})
const { data: orders, isLoading } = useOrders(filters)

const storeTypes = ['RETAIL', 'FNB', 'OUTFIT']
const paymentMethods = ['CASH', 'CARD', 'QRIS']
const showDetail = ref(false)
const selectedOrder = ref<Order | null>(null)

function viewOrder(order: Order) {
  selectedOrder.value = order
  showDetail.value = true
}

function formatCurrency(value: number): string {
  return `Rp ${value.toLocaleString('id-ID')}`
}

function formatTimestamp(ts: number): string {
  return new Date(ts).toLocaleString('id-ID')
}

function formatShortDate(ts: number): string {
  return new Date(ts).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function getPaymentSeverity(method: string): string {
  switch (method) {
    case 'CASH': return 'success'
    case 'CARD': return 'info'
    case 'QRIS': return 'warn'
    default: return 'secondary'
  }
}

const dateRange = ref<Date[] | null>(null)
watch(dateRange, (val) => {
  if (val && val.length === 2) {
    filters.value.startDate = val[0]!.getTime()
    filters.value.endDate = val[1]!.getTime()
  } else {
    filters.value.startDate = undefined
    filters.value.endDate = undefined
  }
})
</script>

<template>
  <div class="orders-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Transactions</h1>
        <p class="page-subtitle">View and audit all orders</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <Select v-model="filters.storeType" :options="storeTypes" showClear placeholder="Store" style="width: 130px" />
      <Select v-model="filters.paymentMethod" :options="paymentMethods" showClear placeholder="Payment" style="width: 130px" />
      <Select v-model="filters.isVoided" :options="[{ label: 'Active', value: false }, { label: 'Voided', value: true }]" optionLabel="label" optionValue="value" showClear placeholder="Status" style="width: 130px" />
      <DatePicker v-model="dateRange" selectionMode="range" showClear placeholder="Date Range" style="width: 240px" />
    </div>

    <div class="card">
      <DataTable :value="orders || []" :loading="isLoading" tableStyle="min-width: 50rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
        <Column header="Order ID" style="width: 80px">
          <template #body="{ data }: { data: Order }">
            <span class="order-id">{{ data.id.length > 8 ? data.id.substring(0, 8) : data.id }}</span>
          </template>
        </Column>
        <Column header="Date" sortable style="width: 140px">
          <template #body="{ data }: { data: Order }">{{ formatShortDate(data.timestamp) }}</template>
        </Column>
        <Column header="Total" sortable style="width: 130px">
          <template #body="{ data }: { data: Order }">{{ formatCurrency(data.total) }}</template>
        </Column>
        <Column header="Payment" sortable style="width: 90px">
          <template #body="{ data }: { data: Order }">
            <Tag :value="data.paymentMethod" :severity="getPaymentSeverity(data.paymentMethod)" />
          </template>
        </Column>
        <Column header="Store" sortable style="width: 80px">
          <template #body="{ data }: { data: Order }">{{ data.storeType }}</template>
        </Column>
        <Column header="Status" style="width: 90px">
          <template #body="{ data }: { data: Order }">
            <Tag :value="data.isVoided ? 'Voided' : data.isSettled ? 'Settled' : 'Active'" :severity="data.isVoided ? 'danger' : data.isSettled ? 'success' : 'info'" />
          </template>
        </Column>
        <Column header="Customer" style="width: 140px">
          <template #body="{ data }: { data: Order }">{{ data.customerName || data.customerPhone || '-' }}</template>
        </Column>
        <Column header="Items" style="width: 50px">
          <template #body="{ data }: { data: Order }">{{ data.items?.length || 0 }}</template>
        </Column>
        <Column header="Actions" style="width: 80px">
          <template #body="{ data }">
            <Button icon="pi pi-eye" text rounded size="small" title="View" @click="viewOrder(data)" />
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No transactions found.</div>
        </template>
      </DataTable>
    </div>

    <!-- Order Detail Dialog -->
    <Dialog v-model:visible="showDetail" :modal="true" :style="{ width: '600px' }" :header="`Order #${selectedOrder?.id?.substring(0, 8)}`">
      <div v-if="selectedOrder" class="order-detail">
        <div class="detail-grid">
          <div class="detail-item"><strong>Date:</strong> {{ formatTimestamp(selectedOrder.timestamp) }}</div>
          <div class="detail-item"><strong>Payment:</strong> {{ selectedOrder.paymentMethod }}</div>
          <div class="detail-item"><strong>Store:</strong> {{ selectedOrder.storeType }}</div>
          <div class="detail-item"><strong>Status:</strong> {{ selectedOrder.isVoided ? 'Voided' : selectedOrder.isSettled ? 'Settled' : 'Active' }}</div>
          <div class="detail-item" v-if="selectedOrder.cashReceived"><strong>Cash:</strong> {{ formatCurrency(selectedOrder.cashReceived) }}</div>
          <div class="detail-item" v-if="selectedOrder.changeAmount !== null"><strong>Change:</strong> {{ formatCurrency(selectedOrder.changeAmount) }}</div>
          <div class="detail-item" v-if="selectedOrder.customerName"><strong>Customer:</strong> {{ selectedOrder.customerName }} ({{ selectedOrder.customerPhone }})</div>
          <div class="detail-item" v-if="selectedOrder.bankName"><strong>Bank:</strong> {{ selectedOrder.bankName }}</div>
          <div class="detail-item" v-if="selectedOrder.transactionId"><strong>Tx ID:</strong> {{ selectedOrder.transactionId }}</div>
          <div class="detail-item" v-if="selectedOrder.voidReason"><strong>Void Reason:</strong> {{ selectedOrder.voidReason }}</div>
        </div>
        <div class="order-items">
          <h3>Items</h3>
          <table class="items-table">
            <thead>
              <tr><th>Product</th><th>Qty</th><th>Price</th><th>Subtotal</th></tr>
            </thead>
            <tbody>
              <tr v-for="item in selectedOrder.items" :key="item.id">
                <td>{{ item.productName }}</td>
                <td>{{ item.qty }}</td>
                <td>{{ formatCurrency(item.productPrice) }}</td>
                <td>{{ formatCurrency(item.subtotal) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="order-totals">
          <div class="total-row"><span>Subtotal</span><span>{{ formatCurrency(selectedOrder.subtotal) }}</span></div>
          <div class="total-row"><span>Tax (10%)</span><span>{{ formatCurrency(selectedOrder.tax) }}</span></div>
          <div v-if="selectedOrder.discountAmount" class="total-row"><span>Discount</span><span>-{{ formatCurrency(selectedOrder.discountAmount) }}</span></div>
          <div class="total-row grand"><span>Total</span><span>{{ formatCurrency(selectedOrder.total) }}</span></div>
        </div>
      </div>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetail = false" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.orders-page { display: flex; flex-direction: column; gap: 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; }
.page-title { margin: 0; font-size: 28px; font-weight: 600; color: var(--p-text-color); }
.page-subtitle { margin: 4px 0 0; color: var(--p-text-secondary-color); font-size: 14px; }
.filter-bar { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
.card { background: var(--p-surface-0); border-radius: 12px; border: 1px solid var(--p-surface-200); padding: 16px; }
.order-id { font-family: monospace; font-size: 13px; }
.empty-state { text-align: center; padding: 40px; color: var(--p-text-secondary-color); }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 16px; }
.detail-item { font-size: 14px; }
.order-items { margin: 16px 0; }
.order-items h3 { margin: 0 0 8px; font-size: 16px; }
.items-table { width: 100%; border-collapse: collapse; }
.items-table th, .items-table td { padding: 8px; text-align: left; border-bottom: 1px solid var(--p-surface-200); font-size: 13px; }
.items-table th { font-weight: 600; color: var(--p-text-secondary-color); }
.order-totals { border-top: 1px solid var(--p-surface-300); padding-top: 12px; }
.total-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 14px; }
.total-row.grand { font-weight: 700; font-size: 16px; padding-top: 8px; }
</style>
