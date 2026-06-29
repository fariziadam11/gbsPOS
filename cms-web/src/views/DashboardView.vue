<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Select from 'primevue/select'
import DatePicker from 'primevue/datepicker'
import Card from 'primevue/card'
import Skeleton from 'primevue/skeleton'
import Chart from 'primevue/chart'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import { useDashboardSummary, useRevenueTrend, useTopProducts } from '../composables/useDashboard'
import { useOrders } from '../composables/useOrders'
import type { TopProduct, Order } from '../types/api'
import type { DashboardDateRange } from '../api/dashboard'

interface DateRangeOption {
  label: string
  value: string
  days: number
}

const storeTypes = ['RETAIL', 'FNB', 'OUTFIT']
const dateRangeOptions: DateRangeOption[] = [
  { label: 'Today', value: 'today', days: 1 },
  { label: 'Last 7 Days', value: 'last7', days: 7 },
  { label: 'Last 30 Days', value: 'last30', days: 30 },
  { label: 'Last 90 Days', value: 'last90', days: 90 },
  { label: 'This Month', value: 'thisMonth', days: 0 },
  { label: 'All Time', value: 'allTime', days: -1 },
  { label: 'Custom', value: 'custom', days: -1 },
]

const storeType = ref<string | undefined>(undefined)
const selectedRange = ref<DateRangeOption>(dateRangeOptions[2]) // default Last 30 Days
const customDateRange = ref<Date[] | null>(null)

const isCustom = computed(() => selectedRange.value.value === 'custom')

function formatDate(d: Date): string {
  return d.toISOString().split('T')[0]
}

const dateRange = computed<DashboardDateRange>(() => {
  const today = new Date()

  if (selectedRange.value.value === 'today') {
    const d = formatDate(today)
    return { startDate: d, endDate: d }
  }

  if (selectedRange.value.value === 'thisMonth') {
    const start = new Date(today.getFullYear(), today.getMonth(), 1)
    return { startDate: formatDate(start), endDate: formatDate(today) }
  }

  if (selectedRange.value.value === 'allTime') {
    return { startDate: '2000-01-01', endDate: formatDate(today) }
  }

  if (selectedRange.value.value === 'custom' && customDateRange.value && customDateRange.value.length === 2) {
    return {
      startDate: formatDate(customDateRange.value[0]!),
      endDate: formatDate(customDateRange.value[1]!),
    }
  }

  const start = new Date(today)
  start.setDate(today.getDate() - selectedRange.value.days + 1)
  return { startDate: formatDate(start), endDate: formatDate(today) }
})

const { data: summary, isLoading: summaryLoading, isError: summaryError, error: summaryErrorObj } = useDashboardSummary(storeType, dateRange)
const { data: revenueTrend, isLoading: revenueLoading, isError: revenueError } = useRevenueTrend(dateRange, storeType)
const { data: topProducts, isLoading: topLoading } = useTopProducts(10, storeType)
const { data: orders, isLoading: ordersLoading } = useOrders(computed(() => ({ storeType: storeType.value, isVoided: undefined, isSettled: undefined })))

const recentOrders = computed(() => (orders.value || []).slice(0, 5))

function formatCurrency(value: number): string {
  return `Rp ${value.toLocaleString('id-ID')}`
}

function formatNumber(value: number): string {
  return value.toLocaleString('id-ID')
}

function formatShortDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}

function formatOrderDate(ts: number): string {
  return new Date(ts).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

function getOrderStatus(order: Order): { label: string; severity: string } {
  if (order.isVoided) return { label: 'Voided', severity: 'danger' }
  if (order.isSettled) return { label: 'Settled', severity: 'success' }
  return { label: 'Active', severity: 'info' }
}

function getPaymentSeverity(method: string): string {
  switch (method) {
    case 'CASH': return 'success'
    case 'CARD': return 'info'
    case 'QRIS': return 'warn'
    default: return 'secondary'
  }
}

const documentStyle = computed(() => getComputedStyle(document.documentElement))

const chartColors = computed(() => ({
  primary: documentStyle.value.getPropertyValue('--p-primary-500').trim() || '#6366f1',
  primaryLight: documentStyle.value.getPropertyValue('--p-primary-200').trim() || '#c7d2fe',
  green: documentStyle.value.getPropertyValue('--p-green-500').trim() || '#22c55e',
  blue: documentStyle.value.getPropertyValue('--p-blue-500').trim() || '#3b82f6',
  purple: documentStyle.value.getPropertyValue('--p-purple-500').trim() || '#a855f7',
  text: documentStyle.value.getPropertyValue('--p-text-color').trim() || '#1f2937',
  textSecondary: documentStyle.value.getPropertyValue('--p-text-muted-color').trim() || '#6b7280',
  border: documentStyle.value.getPropertyValue('--p-content-border-color').trim() || '#e5e7eb',
}))

const revenueChartData = computed(() => {
  const labels = (revenueTrend.value || []).map((p) => formatShortDate(p.date))
  const revenue = (revenueTrend.value || []).map((p) => p.revenue)
  const orderCounts = (revenueTrend.value || []).map((p) => p.orders)

  return {
    labels,
    datasets: [
      {
        type: 'bar' as const,
        label: 'Revenue',
        data: revenue,
        backgroundColor: chartColors.value.primary,
        borderColor: chartColors.value.primary,
        borderWidth: 0,
        borderRadius: 4,
        order: 1,
        yAxisID: 'y',
      },
      {
        type: 'line' as const,
        label: 'Orders',
        data: orderCounts,
        borderColor: chartColors.value.purple,
        backgroundColor: chartColors.value.purple,
        borderWidth: 2,
        pointRadius: 3,
        tension: 0.3,
        order: 0,
        yAxisID: 'y1',
      },
    ],
  }
})

const revenueChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index', intersect: false },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: { color: chartColors.value.text, usePointStyle: true },
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const label = context.dataset.label || ''
          const value = context.parsed.y
          return label === 'Revenue' ? `${label}: ${formatCurrency(value)}` : `${label}: ${formatNumber(value)}`
        },
      },
    },
  },
  scales: {
    x: {
      ticks: { color: chartColors.value.textSecondary },
      grid: { display: false },
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      ticks: {
        color: chartColors.value.textSecondary,
        callback: (value: any) => formatCurrency(Number(value)).replace('Rp ', ''),
      },
      grid: { color: chartColors.value.border },
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      ticks: { color: chartColors.value.textSecondary },
      grid: { display: false },
    },
  },
}))

const paymentChartData = computed(() => ({
  labels: ['Cash', 'Card', 'QRIS'],
  datasets: [
    {
      data: [summary.value?.cashTotal || 0, summary.value?.cardTotal || 0, summary.value?.qrisTotal || 0],
      backgroundColor: [chartColors.value.green, chartColors.value.blue, chartColors.value.purple],
      hoverBackgroundColor: [chartColors.value.green, chartColors.value.blue, chartColors.value.purple],
    },
  ],
}))

const paymentChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '60%',
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: { color: chartColors.value.text, usePointStyle: true },
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const label = context.label || ''
          const value = context.parsed
          return `${label}: ${formatCurrency(value)}`
        },
      },
    },
  },
}))

const hasAnyPayment = computed(() => {
  const s = summary.value
  return !!s && (s.cashTotal > 0 || s.cardTotal > 0 || s.qrisTotal > 0)
})

const hasRevenueData = computed(() => (revenueTrend.value || []).some((p) => p.revenue > 0 || p.orders > 0))

const formattedDateRange = computed(() => {
  if (!dateRange.value.startDate || !dateRange.value.endDate) return ''
  const start = new Date(dateRange.value.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  const end = new Date(dateRange.value.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  return start === end ? start : `${start} - ${end}`
})

const showSkeletonCharts = ref(true)
watch([revenueLoading, summaryLoading], ([r, s]) => {
  if (!r && !s) {
    setTimeout(() => { showSkeletonCharts.value = false }, 50)
  } else {
    showSkeletonCharts.value = true
  }
}, { immediate: true })
</script>

<template>
  <div class="flex flex-column gap-3 lg:gap-4">
    <!-- Header -->
    <div class="flex flex-column md:flex-row md:align-items-start justify-content-between gap-3">
      <div>
        <h1 class="text-2xl lg:text-3xl font-semibold text-color m-0">Dashboard</h1>
        <p class="text-sm text-color-secondary mt-1 mb-0">Sales overview and key metrics</p>
      </div>
      <div class="flex flex-column sm:flex-row gap-2 w-full md:w-auto">
        <Select
          v-model="storeType"
          :options="storeTypes"
          showClear
          placeholder="All Stores"
          class="w-full sm:w-10rem"
        />
        <Select
          v-model="selectedRange"
          :options="dateRangeOptions"
          optionLabel="label"
          placeholder="Date Range"
          class="w-full sm:w-10rem"
        />
        <DatePicker
          v-if="isCustom"
          v-model="customDateRange"
          selectionMode="range"
          showClear
          placeholder="Custom Range"
          class="w-full sm:w-13rem"
        />
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid">
      <div class="col-12 md:col-6 lg:col-3">
        <Card class="h-full border-round-xl shadow-1">
          <template #content>
            <div class="text-xs lg:text-sm text-color-secondary font-medium mb-2">Revenue</div>
            <div v-if="summaryLoading"><Skeleton width="8rem" height="2rem" /></div>
            <div v-else class="text-xl lg:text-2xl font-bold text-primary">{{ formatCurrency(summary?.totalRevenue || 0) }}</div>
          </template>
        </Card>
      </div>
      <div class="col-12 md:col-6 lg:col-3">
        <Card class="h-full border-round-xl shadow-1">
          <template #content>
            <div class="text-xs lg:text-sm text-color-secondary font-medium mb-2">Total Orders</div>
            <div v-if="summaryLoading"><Skeleton width="5rem" height="2rem" /></div>
            <div v-else class="text-xl lg:text-2xl font-bold text-color">{{ formatNumber(summary?.totalOrders || 0) }}</div>
          </template>
        </Card>
      </div>
      <div class="col-12 md:col-6 lg:col-3">
        <Card class="h-full border-round-xl shadow-1">
          <template #content>
            <div class="text-xs lg:text-sm text-color-secondary font-medium mb-2">Avg Order Value</div>
            <div v-if="summaryLoading"><Skeleton width="7rem" height="2rem" /></div>
            <div v-else class="text-xl lg:text-2xl font-bold text-color">{{ formatCurrency(summary?.avgOrderValue || 0) }}</div>
          </template>
        </Card>
      </div>
      <div class="col-12 md:col-6 lg:col-3">
        <Card class="h-full border-round-xl shadow-1">
          <template #content>
            <div class="text-xs lg:text-sm text-color-secondary font-medium mb-2">Voided</div>
            <div v-if="summaryLoading"><Skeleton width="3rem" height="2rem" /></div>
            <div v-else class="text-xl lg:text-2xl font-bold text-red-500">{{ formatNumber(summary?.voidedCount || 0) }}</div>
          </template>
        </Card>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid">
      <!-- Revenue Trend -->
      <div class="col-12 lg:col-8">
        <Card class="h-full border-round-xl shadow-1">
          <template #title>
            <div class="flex align-items-center justify-content-between gap-2 text-base font-semibold">
              <span>Revenue Trend</span>
            </div>
          </template>
          <template #content>
            <div v-if="showSkeletonCharts || revenueLoading" class="h-16rem lg:h-20rem">
              <Skeleton width="100%" height="100%" />
            </div>
            <div v-else-if="revenueError" class="h-16rem lg:h-20rem flex align-items-center justify-content-center text-red-500 text-center p-3">
              Failed to load revenue trend.
            </div>
            <div v-else-if="!hasRevenueData" class="h-16rem lg:h-20rem flex align-items-center justify-content-center text-color-secondary text-center p-3">
              <div>
                <div class="font-semibold text-color mb-1">No revenue data</div>
                <div class="text-color-secondary text-sm mb-2">{{ formattedDateRange }}</div>
                <div class="text-color-secondary text-xs">Try selecting a wider date range.</div>
              </div>
            </div>
            <div v-else class="h-16rem lg:h-20rem relative">
              <Chart type="bar" :data="revenueChartData" :options="revenueChartOptions" class="h-full" />
            </div>
          </template>
        </Card>
      </div>

      <!-- Payment Mix -->
      <div class="col-12 lg:col-4">
        <Card class="h-full border-round-xl shadow-1">
          <template #title>
            <div class="flex align-items-center justify-content-between gap-2 text-base font-semibold">
              <span>Payment Mix</span>
            </div>
          </template>
          <template #content>
            <div v-if="showSkeletonCharts || summaryLoading" class="h-16rem lg:h-20rem">
              <Skeleton width="100%" height="100%" />
            </div>
            <div v-else-if="summaryError" class="h-16rem lg:h-20rem flex align-items-center justify-content-center text-red-500 text-center p-3">
              Failed to load payment data.
            </div>
            <div v-else-if="!hasAnyPayment" class="h-16rem lg:h-20rem flex align-items-center justify-content-center text-color-secondary text-center p-3">
              <div>
                <div class="font-semibold text-color mb-1">No payment data</div>
                <div class="text-color-secondary text-sm mb-2">{{ formattedDateRange }}</div>
                <div class="text-color-secondary text-xs">Try selecting a wider date range.</div>
              </div>
            </div>
            <div v-else class="h-16rem lg:h-20rem flex flex-column align-items-center justify-content-center gap-3">
              <Chart type="doughnut" :data="paymentChartData" :options="paymentChartOptions" class="h-10rem lg:h-12rem max-w-15rem" />
              <div class="w-full flex flex-column gap-2">
                <div class="flex align-items-center gap-2 text-sm">
                  <span class="dot cash" />
                  <span class="text-color-secondary w-3rem">Cash</span>
                  <span class="font-semibold text-color ml-auto">{{ formatCurrency(summary?.cashTotal || 0) }}</span>
                </div>
                <div class="flex align-items-center gap-2 text-sm">
                  <span class="dot card" />
                  <span class="text-color-secondary w-3rem">Card</span>
                  <span class="font-semibold text-color ml-auto">{{ formatCurrency(summary?.cardTotal || 0) }}</span>
                </div>
                <div class="flex align-items-center gap-2 text-sm">
                  <span class="dot qris" />
                  <span class="text-color-secondary w-3rem">QRIS</span>
                  <span class="font-semibold text-color ml-auto">{{ formatCurrency(summary?.qrisTotal || 0) }}</span>
                </div>
              </div>
            </div>
          </template>
        </Card>
      </div>
    </div>

    <!-- Bottom Row -->
    <div class="grid">
      <!-- Top Products -->
      <div class="col-12 lg:col-6">
        <Card class="h-full border-round-xl shadow-1">
          <template #title>
            <div class="flex align-items-center justify-content-between gap-2 text-base font-semibold">
              <span>Top Products</span>
            </div>
          </template>
          <template #content>
            <DataTable
              :value="topProducts || []"
              :loading="topLoading"
              tableStyle="min-width: 20rem"
              stripedRows
              size="small"
              class="dashboard-table"
            >
              <Column header="#" style="width: 40px">
                <template #body="{ index }">{{ index + 1 }}</template>
              </Column>
              <Column field="productName" header="Product" sortable />
              <Column field="totalSold" header="Sold" sortable style="width: 80px">
                <template #body="{ data }: { data: TopProduct }">{{ formatNumber(data.totalSold) }}</template>
              </Column>
              <Column field="revenue" header="Revenue" sortable style="width: 120px">
                <template #body="{ data }: { data: TopProduct }">{{ formatCurrency(data.revenue) }}</template>
              </Column>
              <template #empty>
                <div class="text-center p-4 text-color-secondary">No top products yet.</div>
              </template>
            </DataTable>
          </template>
        </Card>
      </div>

      <!-- Recent Transactions -->
      <div class="col-12 lg:col-6">
        <Card class="h-full border-round-xl shadow-1">
          <template #title>
            <div class="flex align-items-center justify-content-between gap-2 text-base font-semibold">
              <span>Recent Transactions</span>
              <Button as="router-link" to="/orders" label="View All" link size="small" />
            </div>
          </template>
          <template #content>
            <DataTable
              :value="recentOrders"
              :loading="ordersLoading"
              tableStyle="min-width: 20rem"
              size="small"
              class="dashboard-table"
            >
              <Column header="Order" style="width: 80px">
                <template #body="{ data }: { data: Order }">
                  <span class="mono">{{ data.id.length > 8 ? data.id.substring(0, 8) : data.id }}</span>
                </template>
              </Column>
              <Column header="Date" style="width: 120px">
                <template #body="{ data }: { data: Order }">{{ formatOrderDate(data.timestamp) }}</template>
              </Column>
              <Column header="Total" style="width: 100px">
                <template #body="{ data }: { data: Order }">{{ formatCurrency(data.total) }}</template>
              </Column>
              <Column header="Payment" style="width: 80px">
                <template #body="{ data }: { data: Order }">
                  <Tag :value="data.paymentMethod" :severity="getPaymentSeverity(data.paymentMethod)" />
                </template>
              </Column>
              <Column header="Status" style="width: 90px">
                <template #body="{ data }: { data: Order }">
                  <Tag :value="getOrderStatus(data).label" :severity="getOrderStatus(data).severity" />
                </template>
              </Column>
              <template #empty>
                <div class="text-center p-4 text-color-secondary">No recent transactions.</div>
              </template>
            </DataTable>
          </template>
        </Card>
      </div>
    </div>

    <!-- Debug / Error banner -->
    <div v-if="summaryError || revenueError" class="error-banner flex align-items-center gap-2 p-3 border-round-lg text-sm">
      <i class="pi pi-exclamation-circle" />
      <span>Some dashboard data could not be loaded. {{ summaryErrorObj?.message || '' }}</span>
    </div>
  </div>
</template>

<style scoped>
.error-banner {
  background-color: light-dark(var(--p-red-50), var(--p-red-900));
  color: light-dark(var(--p-red-600), var(--p-red-300));
}
.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot.cash { background: var(--p-green-500); }
.dot.card { background: var(--p-blue-500); }
.dot.qris { background: var(--p-purple-500); }
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
.dashboard-table :deep(.p-datatable-header) {
  display: none;
}
</style>
