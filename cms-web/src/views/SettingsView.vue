<script setup lang="ts">
import { ref, watch } from 'vue'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import { useSettings, useUpdateSettings } from '../composables/useSettings'
import { getErrorMessage } from '../api/client'

const toast = useToast()
const { data: settingsData, isLoading } = useSettings()
const { mutate: updateSettings } = useUpdateSettings()

const form = ref<Record<string, string>>({})
const saving = ref(false)

watch(settingsData, (val) => {
  if (val?.settings) {
    form.value = { ...val.settings }
  }
})

function save() {
  saving.value = true
  updateSettings({ settings: form.value }, {
    onSuccess: () => {
      toast.add({ severity: 'success', summary: 'Saved', detail: 'Settings updated', life: 3000 })
    },
    onError: (err) => {
      toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 })
    },
    onSettled: () => { saving.value = false },
  })
}
</script>

<template>
  <div class="flex flex-column gap-3 lg:gap-4">
    <div class="flex flex-column md:flex-row md:align-items-start justify-content-between gap-3">
      <div>
        <h1 class="text-2xl lg:text-3xl font-semibold text-color m-0">Settings</h1>
        <p class="text-sm text-color-secondary mt-1 mb-0">Configure POS application settings</p>
      </div>
      <div class="flex flex-wrap align-items-center gap-2">
        <Button label="Save Changes" icon="pi pi-check" :loading="saving" @click="save" />
      </div>
    </div>

    <Message v-if="isLoading" severity="info" style="margin-bottom: 16px;">Loading settings...</Message>

    <div class="surface-0 border-round-xl border-1 surface-border p-4">
      <h2 class="text-base font-semibold text-color mt-0 mb-3">General</h2>
      <div class="flex flex-column gap-3">
        <div class="form-field">
          <label>Store Name</label>
          <InputText v-model="form.store_name" fluid />
        </div>
        <div class="form-field">
          <label>Currency</label>
          <InputText v-model="form.currency" fluid disabled />
        </div>
      </div>
    </div>

    <div class="surface-0 border-round-xl border-1 surface-border p-4">
      <h2 class="text-base font-semibold text-color mt-0 mb-3">Tax</h2>
      <div class="flex flex-column gap-3">
        <div class="form-field">
          <label>Tax Rate</label>
          <InputText v-model="form.tax_rate" placeholder="0.10" fluid />
        </div>
      </div>
      <p class="text-xs text-color-secondary mt-2 mb-0">Tax rate in decimal format. 0.10 = 10% PPN.</p>
    </div>

    <div class="surface-0 border-round-xl border-1 surface-border p-4">
      <h2 class="text-base font-semibold text-color mt-0 mb-3">Receipt</h2>
      <div class="flex flex-column gap-3">
        <div class="form-field">
          <label>Receipt Header</label>
          <InputText v-model="form.receipt_header" fluid />
        </div>
        <div class="form-field">
          <label>Receipt Footer</label>
          <InputText v-model="form.receipt_footer" fluid />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.form-field { display: flex; flex-direction: column; gap: 4px; }
.form-field label { font-size: 14px; font-weight: 500; color: var(--p-text-color); }
</style>
