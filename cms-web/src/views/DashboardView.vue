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
  <div class="dashboard-page">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Dashboard</h1>
        <p class="page-subtitle">Sales overview and key metrics</p>
      </div>
      <div class="header-filters">
        <Select
          v-model="storeType"
          :options="storeTypes"
          showClear
          placeholder="All Stores"
          class="filter-select"
        />
        <Select
          v-model="selectedRange"
          :options="dateRangeOptions"
          optionLabel="label"
          placeholder="Date Range"
          class="filter-select"
        />
        <DatePicker
          v-if="isCustom"
          v-model="customDateRange"
          selectionMode="range"
          showClear
          placeholder="Custom Range"
          class="filter-select date-picker"
        />
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="stat-grid">
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Revenue</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="140px" height="36px" /></div>
          <div v-else class="stat-value revenue">{{ formatCurrency(summary?.totalRevenue || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Total Orders</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="80px" height="36px" /></div>
          <div v-else class="stat-value">{{ formatNumber(summary?.totalOrders || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Avg Order Value</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="120px" height="36px" /></div>
          <div v-else class="stat-value">{{ formatCurrency(summary?.avgOrderValue || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Voided</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="60px" height="36px" /></div>
          <div v-else class="stat-value void">{{ formatNumber(summary?.voidedCount || 0) }}</div>
        </template>
      </Card>
    </div>

    <!-- Charts Row -->
    <div class="charts-grid">
      <!-- Revenue Trend -->
      <Card class="chart-card trend-card">
        <template #title>
          <div class="card-title-row">
            <span>Revenue Trend</span>
          </div>
        </template>
        <template #content>
          <div v-if="showSkeletonCharts || revenueLoading" class="chart-skeleton">
            <Skeleton width="100%" height="100%" />
          </div>
          <div v-else-if="revenueError" class="chart-empty error">
            Failed to load revenue trend.
          </div>
          <div v-else-if="!hasRevenueData" class="chart-empty">
            <div>
              <div class="empty-title">No revenue data</div>
              <div class="empty-subtitle">{{ formattedDateRange }}</div>
              <div class="empty-hint">Try selecting a wider date range.</div>
            </div>
          </div>
          <div v-else class="chart-wrapper">
            <Chart type="bar" :data="revenueChartData" :options="revenueChartOptions" class="revenue-chart" />
          </div>
        </template>
      </Card>

      <!-- Payment Mix -->
      <Card class="chart-card payment-card">
        <template #title>
          <div class="card-title-row">
            <span>Payment Mix</span>
          </div>
        </template>
        <template #content>
          <div v-if="showSkeletonCharts || summaryLoading" class="chart-skeleton">
            <Skeleton width="100%" height="100%" />
          </div>
          <div v-else-if="summaryError" class="chart-empty error">
            Failed to load payment data.
          </div>
          <div v-else-if="!hasAnyPayment" class="chart-empty">
            <div>
              <div class="empty-title">No payment data</div>
              <div class="empty-subtitle">{{ formattedDateRange }}</div>
              <div class="empty-hint">Try selecting a wider date range.</div>
            </div>
          </div>
          <div v-else class="chart-wrapper payment-wrapper">
            <Chart type="doughnut" :data="paymentChartData" :options="paymentChartOptions" class="payment-chart" />
            <div class="payment-totals">
              <div class="payment-total-item">
                <span class="dot cash" />
                <span class="payment-total-label">Cash</span>
                <span class="payment-total-value">{{ formatCurrency(summary?.cashTotal || 0) }}</span>
              </div>
              <div class="payment-total-item">
                <span class="dot card" />
                <span class="payment-total-label">Card</span>
                <span class="payment-total-value">{{ formatCurrency(summary?.cardTotal || 0) }}</span>
              </div>
              <div class="payment-total-item">
                <span class="dot qris" />
                <span class="payment-total-label">QRIS</span>
                <span class="payment-total-value">{{ formatCurrency(summary?.qrisTotal || 0) }}</span>
              </div>
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Bottom Row -->
    <div class="bottom-grid">
      <!-- Top Products -->
      <Card class="table-card">
        <template #title>
          <div class="card-title-row">
            <span>Top Products</span>
          </div>
        </template>
        <template #content>
          <DataTable
            :value="topProducts || []"
            :loading="topLoading"
            tableStyle="min-width: 24rem"
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
            <Column field="revenue" header="Revenue" sortable style="width: 140px">
              <template #body="{ data }: { data: TopProduct }">{{ formatCurrency(data.revenue) }}</template>
            </Column>
            <template #empty>
              <div class="empty-state">No top products yet.</div>
            </template>
          </DataTable>
        </template>
      </Card>

      <!-- Recent Transactions -->
      <Card class="table-card">
        <template #title>
          <div class="card-title-row">
            <span>Recent Transactions</span>
            <Button as="router-link" to="/orders" label="View All" link size="small" />
          </div>
        </template>
        <template #content>
          <DataTable
            :value="recentOrders"
            :loading="ordersLoading"
            tableStyle="min-width: 24rem"
            size="small"
            class="dashboard-table"
          >
            <Column header="Order" style="width: 80px">
              <template #body="{ data }: { data: Order }">
                <span class="mono">{{ data.id.length > 8 ? data.id.substring(0, 8) : data.id }}</span>
              </template>
            </Column>
            <Column header="Date" style="width: 130px">
              <template #body="{ data }: { data: Order }">{{ formatOrderDate(data.timestamp) }}</template>
            </Column>
            <Column header="Total" style="width: 110px">
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
              <div class="empty-state">No recent transactions.</div>
            </template>
          </DataTable>
        </template>
      </Card>
    </div>

    <!-- Debug / Error banner -->
    <div v-if="summaryError || revenueError" class="error-banner">
      <i class="pi pi-exclamation-circle" />
      <span>Some dashboard data could not be loaded. {{ summaryErrorObj?.message || '' }}</span>
    </div>
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
  align-items: flex-start;
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
  color: var(--p-text-muted-color);
  font-size: 14px;
}
.header-filters {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.filter-select {
  width: 160px;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-card {
  border-radius: 12px;
}
.stat-card :deep(.p-card-content) {
  padding: 20px;
}
.stat-label {
  font-size: 13px;
  color: var(--p-text-muted-color);
  margin-bottom: 8px;
  font-weight: 500;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--p-text-color);
}
.stat-value.revenue {
  color: var(--p-primary-500);
}
.stat-value.void {
  color: var(--p-red-500);
}

.charts-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
}
.chart-card {
  border-radius: 12px;
}
.chart-card :deep(.p-card-title) {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
.card-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.chart-wrapper {
  height: 320px;
  position: relative;
}
.revenue-chart {
  height: 100%;
}
.payment-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}
.payment-chart {
  height: 200px;
  max-width: 240px;
}
.payment-totals {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.payment-total-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
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
.payment-total-label {
  color: var(--p-text-muted-color);
  width: 40px;
}
.payment-total-value {
  margin-left: auto;
  font-weight: 600;
  color: var(--p-text-color);
}
.chart-skeleton {
  height: 320px;
}
.chart-empty {
  height: 320px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--p-text-muted-color);
  text-align: center;
  padding: 20px;
}
.chart-empty.error {
  color: var(--p-red-500);
}
.empty-title {
  font-weight: 600;
  color: var(--p-text-color);
  margin-bottom: 4px;
}
.empty-subtitle {
  color: var(--p-text-muted-color);
  font-size: 13px;
  margin-bottom: 8px;
}
.empty-hint {
  color: var(--p-text-muted-color);
  font-size: 12px;
}

.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.table-card {
  border-radius: 12px;
}
.table-card :deep(.p-card-title) {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
.table-card :deep(.p-card-content) {
  padding-top: 8px;
}
.dashboard-table :deep(.p-datatable-header) {
  display: none;
}
.empty-state {
  text-align: center;
  padding: 32px;
  color: var(--p-text-muted-color);
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
.error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: 8px;
  background: var(--p-red-50);
  color: var(--p-red-600);
  font-size: 14px;
}

@media (max-width: 1024px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .charts-grid {
    grid-template-columns: 1fr;
  }
  .bottom-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }
  .filter-select {
    width: 100%;
  }
  .header-filters {
    width: 100%;
  }
  .chart-wrapper,
  .chart-skeleton,
  .chart-empty {
    height: 260px;
  }
}
</style>
