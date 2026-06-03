<script setup lang="ts">
import { ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Select from 'primevue/select'
import Card from 'primevue/card'
import Skeleton from 'primevue/skeleton'
import { useDashboardSummary, useRevenueTrend, useTopProducts } from '../composables/useDashboard'
import type { TopProduct } from '../types/api'

const storeType = ref<string | undefined>(undefined)
const days = ref(7)
const topLimit = ref(10)

const { data: summary, isLoading: summaryLoading } = useDashboardSummary(storeType)
const { data: revenueTrend, isLoading: revenueLoading } = useRevenueTrend(days, storeType)
const { data: topProducts, isLoading: topLoading } = useTopProducts(topLimit, storeType)

const storeTypes = ['RETAIL', 'FNB', 'OUTFIT']

function formatCurrency(value: number): string {
  return `Rp ${value.toLocaleString('id-ID')}`
}

function formatNumber(value: number): string {
  return value.toLocaleString('id-ID')
}

function maxRevenue(points: { revenue: number }[]): number {
  if (!points || points.length === 0) return 100
  return Math.max(...points.map((p) => p.revenue), 1)
}
</script>

<template>
  <div class="dashboard-page">
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
          style="width: 160px"
        />
      </div>
    </div>

    <!-- Stat Cards -->
    <div class="stat-grid">
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Today's Revenue</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="120px" height="36px" /></div>
          <div v-else class="stat-value revenue">{{ formatCurrency(summary?.totalRevenue || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Total Orders</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="60px" height="36px" /></div>
          <div v-else class="stat-value">{{ formatNumber(summary?.totalOrders || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Avg Order Value</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="100px" height="36px" /></div>
          <div v-else class="stat-value">{{ formatCurrency(summary?.avgOrderValue || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card">
        <template #content>
          <div class="stat-label">Voided Today</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="60px" height="36px" /></div>
          <div v-else class="stat-value void">{{ formatNumber(summary?.voidedCount || 0) }}</div>
        </template>
      </Card>
    </div>

    <!-- Payment Breakdown -->
    <div class="stat-grid">
      <Card class="stat-card payment-cash">
        <template #content>
          <div class="stat-label">Cash</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="100px" height="28px" /></div>
          <div v-else class="stat-value small">{{ formatCurrency(summary?.cashTotal || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card payment-card">
        <template #content>
          <div class="stat-label">Card</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="100px" height="28px" /></div>
          <div v-else class="stat-value small">{{ formatCurrency(summary?.cardTotal || 0) }}</div>
        </template>
      </Card>
      <Card class="stat-card payment-qris">
        <template #content>
          <div class="stat-label">QRIS</div>
          <div v-if="summaryLoading" class="stat-value"><Skeleton width="100px" height="28px" /></div>
          <div v-else class="stat-value small">{{ formatCurrency(summary?.qrisTotal || 0) }}</div>
        </template>
      </Card>
    </div>

    <!-- Revenue Trend -->
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Revenue Trend</h2>
        <Select v-model="days" :options="[7, 14, 30]" placeholder="Days" style="width: 100px" />
      </div>
      <div class="revenue-chart">
        <div v-if="revenueLoading" class="chart-loading">
          <Skeleton width="100%" height="200px" />
        </div>
        <div v-else-if="revenueTrend?.length" class="chart-bars">
          <div
            v-for="point in revenueTrend"
            :key="point.date"
            class="chart-bar-item"
            :title="`${point.date}: ${formatCurrency(point.revenue)} (${point.orders} orders)`"
          >
            <div
              class="chart-bar"
              :style="{ height: `${(point.revenue / maxRevenue(revenueTrend!)) * 100}%` }"
            />
            <div class="chart-label">{{ new Date(point.date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) }}</div>
          </div>
        </div>
        <div v-else class="empty-chart">No revenue data yet</div>
      </div>
    </div>

    <!-- Top Products -->
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Top Products</h2>
      </div>
      <DataTable
        :value="topProducts || []"
        :loading="topLoading"
        tableStyle="min-width: 40rem"
        stripedRows
        size="small"
      >
        <Column header="#" style="width: 50px">
          <template #body="{ index }">{{ index + 1 }}</template>
        </Column>
        <Column field="productName" header="Product" sortable />
        <Column field="totalSold" header="Sold" sortable style="width: 100px;">
          <template #body="{ data }: { data: TopProduct }">{{ formatNumber(data.totalSold) }}</template>
        </Column>
        <Column field="revenue" header="Revenue" sortable style="width: 160px;">
          <template #body="{ data }: { data: TopProduct }">{{ formatCurrency(data.revenue) }}</template>
        </Column>
      </DataTable>
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
.header-filters {
  display: flex;
  gap: 8px;
  align-items: center;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-card {
  border-radius: 12px;
}
.stat-card .p-card-content { padding: 20px; }
.stat-label {
  font-size: 13px;
  color: var(--p-text-secondary-color);
  margin-bottom: 8px;
  font-weight: 500;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--p-text-color);
}
.stat-value.small {
  font-size: 20px;
}
.stat-value.revenue {
  color: var(--p-primary-color);
}
.stat-value.void {
  color: var(--p-red-500);
}
.payment-cash { border-left: 4px solid var(--p-green-500); }
.payment-card { border-left: 4px solid var(--p-blue-500); }
.payment-qris { border-left: 4px solid var(--p-purple-500); }
.card {
  background: var(--p-surface-0);
  border-radius: 12px;
  border: 1px solid var(--p-surface-200);
  padding: 20px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--p-text-color);
}
.revenue-chart {
  height: 220px;
  display: flex;
  align-items: flex-end;
}
.chart-bars {
  display: flex;
  align-items: flex-end;
  width: 100%;
  height: 200px;
  gap: 4px;
}
.chart-bar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  justify-content: flex-end;
}
.chart-bar {
  width: 100%;
  max-width: 48px;
  background: var(--p-primary-color);
  border-radius: 4px 4px 0 0;
  min-height: 2px;
  transition: height 0.3s ease;
}
.chart-label {
  font-size: 10px;
  color: var(--p-text-secondary-color);
  margin-top: 6px;
  text-align: center;
  white-space: nowrap;
}
.empty-chart, .chart-loading {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--p-text-secondary-color);
}
</style>
