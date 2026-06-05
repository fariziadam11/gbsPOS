<script setup lang="ts">
import { ref, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import ConfirmDialog from 'primevue/confirmdialog'
import FileUpload from 'primevue/fileupload'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '../stores/auth'
import { useProducts, useCreateProduct, useUpdateProduct, useDeleteProduct, useImportProducts } from '../composables/useProducts'
import { getExportUrl } from '../api/products'
import { getErrorMessage } from '../api/client'
import type { Product, CreateProductRequest } from '../types/api'

const authStore = useAuthStore()
const confirm = useConfirm()
const toast = useToast()

const storeType = ref<string | undefined>(undefined)
const { data: products, isLoading } = useProducts(storeType)
const { mutate: createProduct } = useCreateProduct()
const { mutate: updateProduct } = useUpdateProduct()
const { mutate: deleteProduct } = useDeleteProduct()
const { mutate: importProducts } = useImportProducts()

const showDialog = ref(false)
const editingProduct = ref<Product | null>(null)
const form = ref<CreateProductRequest>({
  name: '',
  price: 0,
  category: '',
  imageUrl: '',
  storeType: '',
  stockQuantity: 0,
  lowStockThreshold: 10,
})
const dialogTitle = ref('Add Product')
const submitting = ref(false)

const storeTypes = ['RETAIL', 'FNB', 'OUTFIT']
const categories = ['Food', 'Beverages', 'Electronics', 'Groceries']

function openCreate() {
  editingProduct.value = null
  dialogTitle.value = 'Add Product'
  form.value = {
    name: '',
    price: 0,
    category: '',
    imageUrl: '',
    storeType: storeType.value || 'RETAIL',
    stockQuantity: 0,
    lowStockThreshold: 10,
  }
  showDialog.value = true
}

function openEdit(product: Product) {
  editingProduct.value = product
  dialogTitle.value = 'Edit Product'
  form.value = {
    name: product.name,
    price: product.price,
    category: product.category,
    imageUrl: product.imageUrl,
    storeType: product.storeType,
    stockQuantity: product.stockQuantity,
    lowStockThreshold: product.lowStockThreshold,
  }
  showDialog.value = true
}

function save() {
  if (!form.value.name || !form.value.category) {
    toast.add({ severity: 'warn', summary: 'Validation', detail: 'Name and category are required', life: 3000 })
    return
  }
  submitting.value = true
  if (editingProduct.value) {
    updateProduct({ id: editingProduct.value.id, data: form.value }, {
      onSuccess: () => {
        toast.add({ severity: 'success', summary: 'Updated', detail: 'Product updated successfully', life: 3000 })
        showDialog.value = false
      },
      onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  } else {
    createProduct(form.value as CreateProductRequest, {
      onSuccess: () => {
        toast.add({ severity: 'success', summary: 'Created', detail: 'Product created successfully', life: 3000 })
        showDialog.value = false
      },
      onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  }
}

function confirmDelete(product: Product) {
  confirm.require({
    message: `Delete "${product.name}"?`,
    header: 'Confirm Delete',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel',
    rejectProps: { severity: 'secondary', outlined: true },
    acceptLabel: 'Delete',
    acceptProps: { severity: 'danger' },
    accept: () => {
      deleteProduct(product.id, {
        onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'Product deleted', life: 3000 }),
        onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      })
    },
  })
}

function formatCurrency(value: number): string {
  return `Rp ${value.toLocaleString('id-ID')}`
}

function getStockSeverity(product: Product): string {
  if (product.stockQuantity <= 0) return 'danger'
  if (product.stockQuantity <= product.lowStockThreshold) return 'warn'
  return 'success'
}

function getStockLabel(product: Product): string {
  if (product.stockQuantity <= 0) return 'Out'
  if (product.stockQuantity <= product.lowStockThreshold) return 'Low'
  return 'OK'
}

function onImport(event: any) {
  const file = event.files?.[0] as File
  if (!file) return
  importProducts({ file, storeType: storeType.value }, {
    onSuccess: (data) => {
      const result = data.data
      toast.add({
        severity: result.failed > 0 ? 'warn' : 'success',
        summary: 'Import Complete',
        detail: `Success: ${result.success}, Failed: ${result.failed}`,
        life: 5000,
      })
    },
    onError: (err) => toast.add({ severity: 'error', summary: 'Import Error', detail: getErrorMessage(err), life: 5000 }),
  })
}

const baseExportUrl = getExportUrl()
const exportUrl = computed(() => {
  let url = baseExportUrl
  if (storeType.value) url += `&storeType=${storeType.value}`
  return url
})
</script>

<template>
  <div class="products-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Products</h1>
        <p class="page-subtitle">Manage product catalog and inventory</p>
      </div>
      <div class="header-actions">
        <Select v-model="storeType" :options="storeTypes" showClear placeholder="All Stores" style="width: 140px" />
        <FileUpload mode="basic" accept=".csv" :maxFileSize="10000000" customUpload :auto="true" @uploader="onImport" chooseLabel="Import CSV" />
        <a :href="exportUrl" class="export-link">
          <Button label="Export CSV" icon="pi pi-download" text severity="secondary" />
        </a>
        <Button v-if="authStore.isAdmin" label="Add Product" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>

    <div class="card">
      <DataTable :value="products || []" :loading="isLoading" tableStyle="min-width: 60rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
        <Column field="id" header="ID" sortable style="width: 60px" />
        <Column field="name" header="Name" sortable />
        <Column header="Price" sortable style="width: 120px">
          <template #body="{ data }">{{ formatCurrency(data.price) }}</template>
        </Column>
        <Column field="category" header="Category" sortable style="width: 120px">
          <template #body="{ data }">
            <Tag :value="data.category" severity="info" />
          </template>
        </Column>
        <Column field="storeType" header="Store" sortable style="width: 90px">
          <template #body="{ data }">
            <Tag :value="data.storeType" severity="secondary" />
          </template>
        </Column>
        <Column header="Stock" style="width: 100px">
          <template #body="{ data }">
            <div class="stock-cell">
              <span>{{ data.stockQuantity }}</span>
              <Tag :value="getStockLabel(data)" :severity="getStockSeverity(data)" style="font-size: 11px;" />
            </div>
          </template>
        </Column>
        <Column v-if="authStore.isAdmin" header="Actions" style="width: 120px">
          <template #body="{ data }">
            <div class="actions">
              <Button icon="pi pi-pencil" text rounded size="small" title="Edit" @click="openEdit(data)" />
              <Button icon="pi pi-trash" text rounded size="small" severity="danger" title="Delete" @click="confirmDelete(data)" />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No products found.</div>
        </template>
      </DataTable>
    </div>

    <!-- Product Form Dialog -->
    <Dialog v-model:visible="showDialog" :header="dialogTitle" :modal="true" :style="{ width: '500px' }">
      <div class="form-grid">
        <div class="form-field">
          <label>Name *</label>
          <InputText v-model="form.name" fluid />
        </div>
        <div class="form-field">
          <label>Price *</label>
          <InputNumber v-model="form.price" mode="currency" currency="IDR" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Category *</label>
          <Select v-model="form.category" :options="categories" editable fluid />
        </div>
        <div class="form-field">
          <label>Store Type</label>
          <Select v-model="form.storeType" :options="storeTypes" fluid />
        </div>
        <div class="form-field">
          <label>Image URL</label>
          <InputText v-model="form.imageUrl" fluid />
        </div>
        <div class="form-field">
          <label>Stock Quantity</label>
          <InputNumber v-model="form.stockQuantity" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Low Stock Threshold</label>
          <InputNumber v-model="form.lowStockThreshold" :min="0" fluid />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showDialog = false" />
        <Button label="Save" :loading="submitting" @click="save" />
      </template>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>

<style scoped>
.products-page {
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
.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.export-link {
  text-decoration: none;
}
.card {
  background: var(--p-surface-0);
  border-radius: 12px;
  border: 1px solid var(--p-surface-200);
  padding: 16px;
}
.actions {
  display: flex;
  gap: 4px;
}
.stock-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--p-text-secondary-color);
}
.form-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.form-field label {
  font-size: 14px;
  font-weight: 500;
  color: var(--p-text-color);
}
</style>
