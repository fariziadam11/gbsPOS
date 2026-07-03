<script setup lang="ts">
import { ref, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import { useFuelSalesReport } from '../composables/useFuelSales'
import { formatRupiah } from '../utils/format'

const fromDate = ref<Date>(new Date())
const toDate = ref<Date>(new Date())
const triggered = ref(false)

const from = computed(() => fromDate.value.toISOString().slice(0, 10))
const to = computed(() => toDate.value.toISOString().slice(0, 10))

const { data: report, isLoading } = useFuelSalesReport(from, to, computed(() => triggered.value))

function loadReport() {
  triggered.value = true
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Fuel Sales Report</h1>

    <div class="flex flex-col sm:flex-row gap-2 items-start sm:items-end mb-4">
      <div class="w-full sm:w-auto">
        <label class="block text-sm font-medium mb-1">From</label>
        <DatePicker v-model="fromDate" dateFormat="yy-mm-dd" showIcon class="w-full" />
      </div>
      <div class="w-full sm:w-auto">
        <label class="block text-sm font-medium mb-1">To</label>
        <DatePicker v-model="toDate" dateFormat="yy-mm-dd" showIcon class="w-full" />
      </div>
      <Button label="Load" icon="pi pi-search" @click="loadReport" />
    </div>

    <div v-if="triggered && report" class="grid grid-cols-1 xl:grid-cols-2 gap-4">
      <div>
        <h2 class="text-lg font-semibold mb-2">By Fuel Type</h2>
        <div class="overflow-x-auto">
          <DataTable :value="report.summary" class="p-datatable-sm min-w-[20rem]" stripedRows>
            <Column field="fuelCode" header="Fuel Code" />
            <Column field="liters" header="Liters" />
            <Column header="Total">
              <template #body="{ data }">
                {{ formatRupiah(data.totalAmount) }}
              </template>
            </Column>
          </DataTable>
        </div>
      </div>

      <div>
        <h2 class="text-lg font-semibold mb-2">By Pump</h2>
        <div class="overflow-x-auto">
          <DataTable :value="report.pumpTotals" class="p-datatable-sm min-w-[20rem]" stripedRows>
            <Column field="pumpId" header="Pump ID" />
            <Column field="liters" header="Liters" />
            <Column header="Total">
              <template #body="{ data }">
                {{ formatRupiah(data.totalAmount) }}
              </template>
            </Column>
          </DataTable>
        </div>
      </div>
    </div>

    <div v-else-if="isLoading" class="text-center py-8">Loading...</div>
    <div v-else class="text-secondary py-8">Select a date range and click Load.</div>
  </div>
</template>
