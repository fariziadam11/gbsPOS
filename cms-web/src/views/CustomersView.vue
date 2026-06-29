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
  <div class="flex flex-column gap-3 lg:gap-4">
    <div class="flex flex-column md:flex-row md:align-items-start justify-content-between gap-3">
      <div>
        <h1 class="text-2xl lg:text-3xl font-semibold text-color m-0">Customers</h1>
        <p class="text-sm text-color-secondary mt-1 mb-0">View customer profiles and order history</p>
      </div>
      <div class="flex flex-wrap align-items-center gap-2">
        <InputText v-model="searchQuery" placeholder="Search name or phone..." class="w-full sm:w-18rem" />
      </div>
    </div>

    <div class="surface-0 border-round-xl border-1 surface-border p-3">
      <DataTable :value="customers || []" :loading="isLoading" tableStyle="min-width: 40rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
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
          <div class="text-center p-5 text-color-secondary">No customers found.</div>
        </template>
      </DataTable>
    </div>

    <!-- Customer Detail Dialog -->
    <Dialog v-model:visible="showDetail" :modal="true" :style="{ width: '95vw', maxWidth: '650px' }" header="Customer Detail">
      <div v-if="detailLoading" class="text-center p-5 text-color-secondary">Loading...</div>
      <div v-else-if="selectedCustomer" class="detail-content">
        <div class="grid">
          <div class="col-12 md:col-6"><strong>Name:</strong> {{ selectedCustomer.customer.name }}</div>
          <div class="col-12 md:col-6"><strong>Phone:</strong> {{ selectedCustomer.customer.phone }}</div>
          <div class="col-12 md:col-6"><strong>Email:</strong> {{ selectedCustomer.customer.email || '-' }}</div>
          <div class="col-12 md:col-6"><strong>Address:</strong> {{ selectedCustomer.customer.address || '-' }}</div>
        </div>
        <div class="stats-row">
          <div class="stat-box"><strong>{{ selectedCustomer.customer.loyaltyPoints }}</strong><span>Points</span></div>
          <div class="stat-box"><strong>{{ formatCurrency(selectedCustomer.totalSpent) }}</strong><span>Total Spent</span></div>
          <div class="stat-box"><strong>{{ selectedCustomer.totalOrders }}</strong><span>Orders</span></div>
        </div>
        <div class="order-history mt-3" v-if="selectedCustomer.orderHistory?.length">
          <h3 class="m-0 mb-2 text-base">Order History</h3>
          <div class="overflow-x-auto">
            <table class="items-table w-full">
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
      </div>
      <div v-else class="text-center p-5 text-color-secondary">No data available.</div>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetail = false" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.stats-row { display: flex; gap: 12px; margin: 16px 0; }
.stat-box { flex: 1; text-align: center; padding: 16px; background: light-dark(var(--p-surface-50), var(--p-surface-800)); border-radius: 8px; }
.stat-box strong { display: block; font-size: 22px; color: var(--p-primary-color); }
.stat-box span { font-size: 12px; color: var(--p-text-secondary-color); }
.order-history { margin-top: 16px; }
.order-history h3 { margin: 0 0 8px; font-size: 16px; }
.items-table { width: 100%; border-collapse: collapse; }
.items-table th, .items-table td { padding: 8px; text-align: left; border-bottom: 1px solid light-dark(var(--p-surface-200), var(--p-surface-700)); font-size: 13px; }
.items-table th { font-weight: 600; color: var(--p-text-secondary-color); }
</style>
