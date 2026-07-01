<script setup lang="ts">
import { ref, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import { usePumps, useCreatePump, useUpdatePump, useDeletePump } from '../composables/usePumps'
import { useNozzles, useCreateNozzle, useUpdateNozzle, useDeleteNozzle } from '../composables/useNozzles'
import type { Pump, Nozzle, CreatePumpRequest, CreateNozzleRequest } from '../types/api'

const toast = useToast()
const confirm = useConfirm()

const { data: pumps, isLoading: pumpsLoading } = usePumps()
const { data: nozzles, isLoading: nozzlesLoading } = useNozzles()
const { mutate: createPump } = useCreatePump()
const { mutate: updatePump } = useUpdatePump()
const { mutate: deletePump } = useDeletePump()
const { mutate: createNozzle } = useCreateNozzle()
const { mutate: updateNozzle } = useUpdateNozzle()
const { mutate: deleteNozzle } = useDeleteNozzle()

const fuelCodes = ['PERTALITE', 'PERTAMAX', 'PERTAMAX_GREEN', 'PERTAMAX_TURBO', 'PERTAMINA_DEX', 'PERTAMINA_DEXLITE']

const showPumpDialog = ref(false)
const pumpForm = ref<CreatePumpRequest>({ id: '', name: '' })
const editingPump = ref<Pump | null>(null)

const showNozzleDialog = ref(false)
const nozzleForm = ref<CreateNozzleRequest>({ id: '', pumpId: '', name: '', fuelCode: fuelCodes[0] })
const editingNozzle = ref<Nozzle | null>(null)

const selectedPump = ref<string | null>(null)
const filteredNozzles = computed(() => {
  if (!nozzles.value) return []
  return selectedPump.value ? nozzles.value.filter((n) => n.pumpId === selectedPump.value) : nozzles.value
})

function openCreatePump() {
  editingPump.value = null
  pumpForm.value = { id: '', name: '' }
  showPumpDialog.value = true
}

function openEditPump(pump: Pump) {
  editingPump.value = pump
  pumpForm.value = { id: pump.id, name: pump.name }
  showPumpDialog.value = true
}

function savePump() {
  if (editingPump.value) {
    updatePump(
      { id: editingPump.value.id, data: { name: pumpForm.value.name } },
      { onSuccess: () => { toast.add({ severity: 'success', summary: 'Saved', detail: 'Pump updated', life: 3000 }); showPumpDialog.value = false } }
    )
  } else {
    createPump(pumpForm.value, {
      onSuccess: () => { toast.add({ severity: 'success', summary: 'Created', detail: 'Pump created', life: 3000 }); showPumpDialog.value = false },
    })
  }
}

function confirmDeletePump(pump: Pump) {
  confirm.require({
    message: `Delete pump ${pump.name}?`,
    header: 'Confirm',
    icon: 'pi pi-exclamation-triangle',
    accept: () => deletePump(pump.id, { onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'Pump deleted', life: 3000 }) }),
  })
}

function openCreateNozzle() {
  editingNozzle.value = null
  nozzleForm.value = { id: '', pumpId: selectedPump.value || '', name: '', fuelCode: fuelCodes[0] }
  showNozzleDialog.value = true
}

function openEditNozzle(nozzle: Nozzle) {
  editingNozzle.value = nozzle
  nozzleForm.value = { id: nozzle.id, pumpId: nozzle.pumpId, name: nozzle.name, fuelCode: nozzle.fuelCode }
  showNozzleDialog.value = true
}

function saveNozzle() {
  if (editingNozzle.value) {
    updateNozzle(
      { id: editingNozzle.value.id, data: { name: nozzleForm.value.name, fuelCode: nozzleForm.value.fuelCode } },
      { onSuccess: () => { toast.add({ severity: 'success', summary: 'Saved', detail: 'Nozzle updated', life: 3000 }); showNozzleDialog.value = false } }
    )
  } else {
    createNozzle(nozzleForm.value, {
      onSuccess: () => { toast.add({ severity: 'success', summary: 'Created', detail: 'Nozzle created', life: 3000 }); showNozzleDialog.value = false },
    })
  }
}

function confirmDeleteNozzle(nozzle: Nozzle) {
  confirm.require({
    message: `Delete nozzle ${nozzle.name}?`,
    header: 'Confirm',
    icon: 'pi pi-exclamation-triangle',
    accept: () => deleteNozzle(nozzle.id, { onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'Nozzle deleted', life: 3000 }) }),
  })
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Pumps & Nozzles</h1>

    <div class="mb-6">
      <div class="flex justify-between items-center mb-2">
        <h2 class="text-lg font-semibold">Pumps</h2>
        <Button label="Add Pump" icon="pi pi-plus" @click="openCreatePump" />
      </div>
      <DataTable :value="pumps || []" :loading="pumpsLoading" class="p-datatable-sm" stripedRows>
        <Column field="id" header="ID" />
        <Column field="name" header="Name" />
        <Column field="isActive" header="Active">
          <template #body="{ data }">
            <Checkbox :modelValue="data.isActive" :binary="true" disabled />
          </template>
        </Column>
        <Column header="Actions">
          <template #body="{ data }">
            <Button icon="pi pi-pencil" severity="secondary" text @click="openEditPump(data)" />
            <Button icon="pi pi-trash" severity="danger" text @click="confirmDeletePump(data)" />
          </template>
        </Column>
      </DataTable>
    </div>

    <div>
      <div class="flex justify-between items-center mb-2 gap-2 flex-wrap">
        <h2 class="text-lg font-semibold">Nozzles</h2>
        <div class="flex gap-2">
          <Select v-model="selectedPump" :options="pumps || []" optionLabel="name" optionValue="id" placeholder="Filter by pump" showClear class="w-48" />
          <Button label="Add Nozzle" icon="pi pi-plus" @click="openCreateNozzle" />
        </div>
      </div>
      <DataTable :value="filteredNozzles" :loading="nozzlesLoading" class="p-datatable-sm" stripedRows>
        <Column field="id" header="ID" />
        <Column field="pumpId" header="Pump" />
        <Column field="name" header="Name" />
        <Column field="fuelCode" header="Fuel" />
        <Column field="isActive" header="Active">
          <template #body="{ data }">
            <Checkbox :modelValue="data.isActive" :binary="true" disabled />
          </template>
        </Column>
        <Column header="Actions">
          <template #body="{ data }">
            <Button icon="pi pi-pencil" severity="secondary" text @click="openEditNozzle(data)" />
            <Button icon="pi pi-trash" severity="danger" text @click="confirmDeleteNozzle(data)" />
          </template>
        </Column>
      </DataTable>
    </div>

    <Dialog v-model:visible="showPumpDialog" :header="editingPump ? 'Edit Pump' : 'Add Pump'" modal :style="{ width: '400px' }">
      <div class="flex flex-col gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">ID</label>
          <InputText v-model="pumpForm.id" class="w-full" :disabled="!!editingPump" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Name</label>
          <InputText v-model="pumpForm.name" class="w-full" />
        </div>
        <Button label="Save" @click="savePump" />
      </div>
    </Dialog>

    <Dialog v-model:visible="showNozzleDialog" :header="editingNozzle ? 'Edit Nozzle' : 'Add Nozzle'" modal :style="{ width: '400px' }">
      <div class="flex flex-col gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">ID</label>
          <InputText v-model="nozzleForm.id" class="w-full" :disabled="!!editingNozzle" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Pump</label>
          <Select v-model="nozzleForm.pumpId" :options="pumps || []" optionLabel="name" optionValue="id" placeholder="Select pump" class="w-full" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Name</label>
          <InputText v-model="nozzleForm.name" class="w-full" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Fuel</label>
          <Select v-model="nozzleForm.fuelCode" :options="fuelCodes" placeholder="Select fuel" class="w-full" />
        </div>
        <Button label="Save" @click="saveNozzle" />
      </div>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>
