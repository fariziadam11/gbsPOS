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
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Chip from 'primevue/chip'
import ConfirmDialog from 'primevue/confirmdialog'
import FileUpload from 'primevue/fileupload'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '../stores/auth'
import { useProducts, useCreateProduct, useUpdateProduct, useDeleteProduct, useImportProducts } from '../composables/useProducts'
import { getExportUrl, getVariants, createVariant, updateVariant, deleteVariant } from '../api/products'
import type { VariantItem, CreateVariantReq } from '../api/products'
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
  name: '', price: 0, category: '', imageUrl: '', storeType: '', stockQuantity: 0, lowStockThreshold: 10,
})
const dialogTitle = ref('Add Product')
const submitting = ref(false)
const activeTab = ref('0')

// Variant state
const variants = ref<VariantItem[]>([])
const variantLoading = ref(false)
const showVariantDialog = ref(false)
const editingVariant = ref<VariantItem | null>(null)
const variantForm = ref<CreateVariantReq>({ attributes: {}, stockQuantity: 0 })
const attrKey = ref('')
const attrValue = ref('')
const variantSubmitting = ref(false)

const storeTypes = ['RETAIL', 'FNB', 'OUTFIT']
const categories = ['Food', 'Beverages', 'Electronics', 'Groceries']

async function loadVariants(productId: number) {
  variantLoading.value = true
  try {
    const res = await getVariants(productId)
    variants.value = res.success ? res.data : []
  } catch { variants.value = [] }
  variantLoading.value = false
}

function openCreate() {
  editingProduct.value = null; dialogTitle.value = 'Add Product'; activeTab.value = '0'
  form.value = { name: '', price: 0, category: '', imageUrl: '', storeType: storeType.value || 'RETAIL', stockQuantity: 0, lowStockThreshold: 10 }
  variants.value = []
  showDialog.value = true
}

function openEdit(product: Product) {
  editingProduct.value = product; dialogTitle.value = 'Edit Product'; activeTab.value = '0'
  form.value = { name: product.name, price: product.price, category: product.category, imageUrl: product.imageUrl, storeType: product.storeType, stockQuantity: product.stockQuantity, lowStockThreshold: product.lowStockThreshold }
  loadVariants(product.id)
  showDialog.value = true
}

function save() {
  if (!form.value.name || !form.value.category) {
    toast.add({ severity: 'warn', summary: 'Validation', detail: 'Name and category are required', life: 3000 }); return
  }
  submitting.value = true
  if (editingProduct.value) {
    updateProduct({ id: editingProduct.value.id, data: form.value }, {
      onSuccess: () => { toast.add({ severity: 'success', summary: 'Updated', detail: 'Product updated', life: 3000 }); showDialog.value = false },
      onError: (err: any) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  } else {
    createProduct(form.value as CreateProductRequest, {
      onSuccess: () => { toast.add({ severity: 'success', summary: 'Created', detail: 'Product created', life: 3000 }); showDialog.value = false },
      onError: (err: any) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  }
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
function formatCurrency(value: number): string { return `Rp ${value.toLocaleString('id-ID')}` }

// Variant CRUD
function addAttribute() {
  if (attrKey.value.trim() && attrValue.value.trim()) {
    variantForm.value.attributes = { ...variantForm.value.attributes, [attrKey.value.trim()]: attrValue.value.trim() }
    attrKey.value = ''; attrValue.value = ''
  }
}
function removeAttribute(key: string) {
  const attrs = { ...variantForm.value.attributes }; delete attrs[key]; variantForm.value.attributes = attrs
}
function openVariantCreate() {
  editingVariant.value = null
  variantForm.value = { attributes: {}, stockQuantity: 0 }
  showVariantDialog.value = true
}
function openVariantEdit(v: VariantItem) {
  editingVariant.value = v
  variantForm.value = { sku: v.sku, attributes: { ...v.attributes }, price: v.price, stockQuantity: v.stockQuantity, lowStockThreshold: v.lowStockThreshold, sortOrder: v.sortOrder }
  showVariantDialog.value = true
}
function saveVariant() {
  if (!editingProduct.value) return
  variantSubmitting.value = true
  const pid = editingProduct.value.id
  const onDone = () => { variantSubmitting.value = false; showVariantDialog.value = false; loadVariants(pid) }
  const onErr = (err: any) => { toast.add({ severity: 'error', detail: getErrorMessage(err), life: 5000 }); variantSubmitting.value = false }
  if (editingVariant.value) {
    updateVariant(editingVariant.value.id, variantForm.value).then(onDone).catch(onErr)
  } else {
    createVariant(pid, variantForm.value).then(onDone).catch(onErr)
  }
}
function confirmDeleteVariant(v: VariantItem) {
  const label = v.sku || Object.entries(v.attributes).map(([k, val]) => `${k}: ${val}`).join(', ') || `Variant #${v.id}`
  confirm.require({
    message: `Delete variant "${label}"?`, header: 'Confirm', icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel', rejectProps: { severity: 'secondary', outlined: true },
    acceptLabel: 'Delete', acceptProps: { severity: 'danger' },
    accept: () => deleteVariant(v.id).then(() => { toast.add({ severity: 'success', detail: 'Variant deleted', life: 3000 }); if (editingProduct.value) loadVariants(editingProduct.value.id) })
  })
}
function confirmDelete(p: Product) {
  confirm.require({ message: `Delete "${p.name}"?`, header: 'Confirm', icon: 'pi pi-exclamation-triangle', rejectLabel: 'Cancel', rejectProps: { severity: 'secondary', outlined: true }, acceptLabel: 'Delete', acceptProps: { severity: 'danger' },
    accept: () => deleteProduct(p.id, {
      onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'Product deleted', life: 3000 }),
      onError: (err: any) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
    }) })
}
function onImport(event: any) {
  const file = event.files?.[0] as File; if (!file) return
  importProducts({ file, storeType: storeType.value }, {
    onSuccess: (data: any) => toast.add({ severity: data.data.failed > 0 ? 'warn' : 'success', summary: 'Import', detail: `Success: ${data.data.success}, Failed: ${data.data.failed}`, life: 5000 }),
    onError: (err: any) => toast.add({ severity: 'error', detail: getErrorMessage(err), life: 5000 }),
  })
}
const baseExportUrl = getExportUrl()
const exportUrl = computed(() => {
  let url = baseExportUrl; if (storeType.value) url += `&storeType=${storeType.value}`; return url
})
</script>

<template>
  <div class="products-page">
    <div class="page-header">
      <div><h1 class="page-title">Products</h1><p class="page-subtitle">Manage product catalog and inventory</p></div>
      <div class="header-actions">
        <Select v-model="storeType" :options="storeTypes" showClear placeholder="All Stores" style="width:140px" />
        <FileUpload mode="basic" accept=".csv" :maxFileSize="10000000" customUpload :auto="true" @uploader="onImport" chooseLabel="Import CSV" />
        <a :href="exportUrl" class="export-link"><Button label="Export CSV" icon="pi pi-download" text severity="secondary" /></a>
        <Button v-if="authStore.isAdmin" label="Add Product" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>
    <div class="card">
      <DataTable :value="products || []" :loading="isLoading" tableStyle="min-width:60rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10,20,50]">
        <Column field="id" header="ID" sortable style="width:60px" />
        <Column field="name" header="Name" sortable />
        <Column header="Price" sortable style="width:120px"><template #body="{data}">{{ formatCurrency(data.price) }}</template></Column>
        <Column field="category" header="Category" sortable style="width:120px"><template #body="{data}"><Tag :value="data.category" severity="info" /></template></Column>
        <Column field="storeType" header="Store" sortable style="width:90px"><template #body="{data}"><Tag :value="data.storeType" severity="secondary" /></template></Column>
        <Column header="Stock" style="width:100px"><template #body="{data}"><div class="stock-cell"><span>{{data.stockQuantity}}</span><Tag :value="getStockLabel(data)" :severity="getStockSeverity(data)" style="font-size:11px" /></div></template></Column>
        <Column v-if="authStore.isAdmin" header="Actions" style="width:120px"><template #body="{data}"><div class="actions"><Button icon="pi pi-pencil" text rounded size="small" @click="openEdit(data)" /><Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="confirmDelete(data)" /></div></template></Column>
        <template #empty><div class="empty-state">No products found.</div></template>
      </DataTable>
    </div>

    <!-- Product + Variants Dialog -->
    <Dialog v-model:visible="showDialog" :header="dialogTitle" :modal="true" :style="{ width: editingProduct ? '700px' : '500px' }">
      <Tabs v-if="editingProduct" v-model:value="activeTab">
        <TabList><Tab value="0">Product</Tab><Tab value="1">Variants ({{ variants.length }})</Tab></TabList>
        <TabPanels>
          <TabPanel value="0">
            <div class="form-grid">
              <div class="form-field"><label>Name *</label><InputText v-model="form.name" fluid /></div>
              <div class="form-field"><label>Price *</label><InputNumber v-model="form.price" mode="currency" currency="IDR" :min="0" fluid /></div>
              <div class="form-field"><label>Category *</label><Select v-model="form.category" :options="categories" editable fluid /></div>
              <div class="form-field"><label>Store Type</label><Select v-model="form.storeType" :options="storeTypes" fluid /></div>
              <div class="form-field"><label>Image URL</label><InputText v-model="form.imageUrl" fluid /></div>
              <div class="form-field"><label>Stock Quantity</label><InputNumber v-model="form.stockQuantity" :min="0" fluid /></div>
              <div class="form-field"><label>Low Stock Threshold</label><InputNumber v-model="form.lowStockThreshold" :min="0" fluid /></div>
            </div>
          </TabPanel>
          <TabPanel value="1">
            <div style="margin-bottom:12px">
              <Button label="Add Variant" icon="pi pi-plus" size="small" @click="openVariantCreate" />
            </div>
            <DataTable :value="variants" :loading="variantLoading" size="small" stripedRows>
              <Column field="sku" header="SKU" style="width:100px" />
              <Column header="Attributes" style="width:150px"><template #body="{data}"><Chip v-for="(v,k) in data.attributes" :key="k" :label="`${k}:${v}`" style="margin:2px" /></template></Column>
              <Column header="Price" style="width:100px"><template #body="{data}">{{ data.price ? formatCurrency(data.price) : '-' }}</template></Column>
              <Column field="stockQuantity" header="Stock" style="width:70px" />
              <Column header="Actions" style="width:100px"><template #body="{data}"><div class="actions"><Button icon="pi pi-pencil" text rounded size="small" @click="openVariantEdit(data)" /><Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="confirmDeleteVariant(data)" /></div></template></Column>
            </DataTable>
          </TabPanel>
        </TabPanels>
      </Tabs>
      <div v-else class="form-grid">
        <div class="form-field"><label>Name *</label><InputText v-model="form.name" fluid /></div>
        <div class="form-field"><label>Price *</label><InputNumber v-model="form.price" mode="currency" currency="IDR" :min="0" fluid /></div>
        <div class="form-field"><label>Category *</label><Select v-model="form.category" :options="categories" editable fluid /></div>
        <div class="form-field"><label>Store Type</label><Select v-model="form.storeType" :options="storeTypes" fluid /></div>
        <div class="form-field"><label>Image URL</label><InputText v-model="form.imageUrl" fluid /></div>
        <div class="form-field"><label>Stock Quantity</label><InputNumber v-model="form.stockQuantity" :min="0" fluid /></div>
        <div class="form-field"><label>Low Stock Threshold</label><InputNumber v-model="form.lowStockThreshold" :min="0" fluid /></div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showDialog = false" />
        <Button label="Save" :loading="submitting" @click="save" />
      </template>
    </Dialog>

    <!-- Variant Form Dialog -->
    <Dialog v-model:visible="showVariantDialog" :header="editingVariant ? 'Edit Variant' : 'Add Variant'" :modal="true" :style="{ width: '450px' }">
      <div class="form-grid">
        <div class="form-field"><label>SKU</label><InputText v-model="variantForm.sku" fluid /></div>
        <div class="form-field"><label>Price</label><InputNumber v-model="variantForm.price" mode="currency" currency="IDR" :min="0" fluid /></div>
        <div class="form-field"><label>Stock</label><InputNumber v-model="variantForm.stockQuantity" :min="0" fluid /></div>
        <div class="form-field"><label>Low Stock Threshold</label><InputNumber v-model="variantForm.lowStockThreshold" :min="0" fluid /></div>
        <div class="form-field"><label>Sort Order</label><InputNumber v-model="variantForm.sortOrder" :min="0" fluid /></div>
        <div class="form-field">
          <label>Attributes</label>
          <div class="attr-chips"><Chip v-for="(v,k) in variantForm.attributes" :key="k" :label="`${k}: ${v}`" removable @remove="removeAttribute(k)" style="margin:2px" /></div>
          <div style="display:flex;gap:8px;margin-top:6px"><InputText v-model="attrKey" placeholder="Key (e.g. Size)" style="flex:1" size="small" /><InputText v-model="attrValue" placeholder="Value (e.g. L)" style="flex:1" size="small" /><Button icon="pi pi-plus" size="small" @click="addAttribute" /></div>
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showVariantDialog = false" />
        <Button label="Save" :loading="variantSubmitting" @click="saveVariant" />
      </template>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>

<style scoped>
.products-page { display:flex;flex-direction:column;gap:24px }
.page-header { display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:16px }
.page-title { margin:0;font-size:28px;font-weight:600;color:var(--p-text-color) }
.page-subtitle { margin:4px 0 0;color:var(--p-text-secondary-color);font-size:14px }
.header-actions { display:flex;gap:8px;align-items:center }
.export-link { text-decoration:none }
.card { background:var(--p-surface-0);border-radius:12px;border:1px solid var(--p-surface-200);padding:16px }
.actions { display:flex;gap:4px }
.stock-cell { display:flex;align-items:center;gap:6px }
.empty-state { text-align:center;padding:40px;color:var(--p-text-secondary-color) }
.form-grid { display:flex;flex-direction:column;gap:16px }
.form-field { display:flex;flex-direction:column;gap:4px }
.form-field label { font-size:14px;font-weight:500;color:var(--p-text-color) }
.attr-chips { display:flex;flex-wrap:wrap;gap:4px }
</style>
