<script setup lang="ts">
import { ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import { useToast } from 'primevue/usetoast'
import { useFuelPrices, useUpdateFuelPrice } from '../composables/useFuelPrices'
import type { FuelPrice } from '../types/api'
import { formatRupiah } from '../utils/format'

const toast = useToast()
const { data: prices, isLoading } = useFuelPrices()
const { mutate: updatePrice } = useUpdateFuelPrice()

const showDialog = ref(false)
const editingPrice = ref<FuelPrice | null>(null)
const newPrice = ref<number>(0)

function openEdit(price: FuelPrice) {
  editingPrice.value = price
  newPrice.value = price.pricePerLiter
  showDialog.value = true
}

function savePrice() {
  if (!editingPrice.value) return
  updatePrice(
    { code: editingPrice.value.code, data: { pricePerLiter: newPrice.value } },
    {
      onSuccess: () => {
        toast.add({ severity: 'success', summary: 'Saved', detail: 'Fuel price updated', life: 3000 })
        showDialog.value = false
      },
      onError: (err: any) => {
        toast.add({ severity: 'error', summary: 'Error', detail: err?.message || 'Failed to update price', life: 5000 })
      },
    }
  )
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Fuel Prices</h1>
    <DataTable :value="prices || []" :loading="isLoading" class="p-datatable-sm" stripedRows>
      <Column field="code" header="Code" />
      <Column field="name" header="Fuel Type" />
      <Column header="Price / Liter">
        <template #body="{ data }">
          {{ formatRupiah(data.pricePerLiter) }}
        </template>
      </Column>
      <Column header="Actions">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" severity="secondary" text @click="openEdit(data)" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="showDialog" header="Edit Fuel Price" modal :style="{ width: '400px' }">
      <div class="flex flex-col gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">{{ editingPrice?.name }}</label>
          <InputNumber v-model="newPrice" mode="currency" currency="IDR" locale="id-ID" class="w-full" />
        </div>
        <Button label="Save" @click="savePrice" />
      </div>
    </Dialog>
  </div>
</template>
