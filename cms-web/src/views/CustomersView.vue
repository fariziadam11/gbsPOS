<script setup lang="ts">
import { ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { useCustomers } from '../composables/useCustomers'
import { getCustomer } from '../api/customers'
import type { CustomerDetail } from '../types/api'

const searchQuery = ref<string | undefined>(undefined)
const { data: customers, isLoading } = useCustomers(searchQuery)

const showDetail = ref(false)
const detailLoading = ref(false)
const selectedCustomer = ref<CustomerDetail | null>(null)

async function viewCustomer(id: number) {
  showDetail.value = true
  detailLoading.value = true
  try {
    const res = await getCustomer(id)
    if (res.success) selectedCustomer.value = res.data
  } catch {
    selectedCustomer.value = null
  } finally {
    detailLoading.value = false
  }
}

function formatCurrency(value: number): string {
  return `Rp ${value.toLocaleString('id-ID')}`
}

function formatShortDate(ts: number): string {
  return new Date(ts).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="customers-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Customers</h1>
        <p class="page-subtitle">View customer profiles and order history</p>
      </div>
      <InputText v-model="searchQuery" placeholder="Search name or phone..." style="width: 280px" />
    </div>

    <div class="card">
      <DataTable :value="customers || []" :loading="isLoading" tableStyle="min-width: 50rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
        <Column field="name" header="Name" sortable />
        <Column field="phone" header="Phone" sortable style="width: 150px" />
        <Column field="email" header="Email" style="width: 200px">
          <template #body="{ data }">{{ data.email || '-' }}</template>
        </Column>
        <Column header="Loyalty Points" sortable style="width: 130px">
          <template #body="{ data }">
            <Tag :value="`${data.loyaltyPoints} pts`" severity="info" />
          </template>
        </Column>
        <Column header="Actions" style="width: 80px">
          <template #body="{ data }">
            <Button icon="pi pi-eye" text rounded size="small" title="View" @click="viewCustomer(data.id)" />
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No customers found.</div>
        </template>
      </DataTable>
    </div>

    <!-- Customer Detail Dialog -->
    <Dialog v-model:visible="showDetail" :modal="true" :style="{ width: '650px' }" header="Customer Detail">
      <div v-if="detailLoading" class="detail-loading">Loading...</div>
      <div v-else-if="selectedCustomer" class="detail-content">
        <div class="customer-info">
          <div class="info-item"><strong>Name:</strong> {{ selectedCustomer.customer.name }}</div>
          <div class="info-item"><strong>Phone:</strong> {{ selectedCustomer.customer.phone }}</div>
          <div class="info-item"><strong>Email:</strong> {{ selectedCustomer.customer.email || '-' }}</div>
          <div class="info-item"><strong>Address:</strong> {{ selectedCustomer.customer.address || '-' }}</div>
        </div>
        <div class="stats-row">
          <div class="stat-box"><strong>{{ selectedCustomer.customer.loyaltyPoints }}</strong><span>Points</span></div>
          <div class="stat-box"><strong>{{ formatCurrency(selectedCustomer.totalSpent) }}</strong><span>Total Spent</span></div>
          <div class="stat-box"><strong>{{ selectedCustomer.totalOrders }}</strong><span>Orders</span></div>
        </div>
        <div class="order-history" v-if="selectedCustomer.orderHistory?.length">
          <h3>Order History</h3>
          <table class="items-table">
            <thead>
              <tr><th>Order ID</th><th>Date</th><th>Total</th><th>Payment</th></tr>
            </thead>
            <tbody>
              <tr v-for="order in selectedCustomer.orderHistory" :key="order.id">
                <td>{{ order.id.substring(0, 8) }}</td>
                <td>{{ formatShortDate(order.timestamp) }}</td>
                <td>{{ formatCurrency(order.total) }}</td>
                <td>{{ order.paymentMethod }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-else class="detail-loading">No data available.</div>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetail = false" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.customers-page { display: flex; flex-direction: column; gap: 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; }
.page-title { margin: 0; font-size: 28px; font-weight: 600; color: var(--p-text-color); }
.page-subtitle { margin: 4px 0 0; color: var(--p-text-secondary-color); font-size: 14px; }
.card { background: var(--p-surface-0); border-radius: 12px; border: 1px solid var(--p-surface-200); padding: 16px; }
.empty-state { text-align: center; padding: 40px; color: var(--p-text-secondary-color); }
.customer-info { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 16px; }
.info-item { font-size: 14px; }
.stats-row { display: flex; gap: 12px; margin: 16px 0; }
.stat-box { flex: 1; text-align: center; padding: 16px; background: var(--p-surface-50); border-radius: 8px; }
.stat-box strong { display: block; font-size: 22px; color: var(--p-primary-color); }
.stat-box span { font-size: 12px; color: var(--p-text-secondary-color); }
.order-history { margin-top: 16px; }
.order-history h3 { margin: 0 0 8px; font-size: 16px; }
.items-table { width: 100%; border-collapse: collapse; }
.items-table th, .items-table td { padding: 8px; text-align: left; border-bottom: 1px solid var(--p-surface-200); font-size: 13px; }
.items-table th { font-weight: 600; color: var(--p-text-secondary-color); }
.detail-loading { text-align: center; padding: 40px; color: var(--p-text-secondary-color); }
</style>
