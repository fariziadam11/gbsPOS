<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import DatePicker from 'primevue/datepicker'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { useCreateDiscount, useDeleteDiscount, useDiscounts, useUpdateDiscount } from '../composables/useDiscounts'
import { useProducts } from '../composables/useProducts'
import { getErrorMessage } from '../api/client'
import type { CreateDiscountRequest, Discount, DiscountScope, DiscountStatus, DiscountType, Product, UpdateDiscountRequest } from '../types/api'

type ScopeFilter = 'ALL' | DiscountScope
type StatusFilter = 'ALL' | DiscountStatus

interface DiscountForm {
  scope: DiscountScope
  productId: number | null
  voucherCode: string
  minTransaction: number
  name: string
  type: DiscountType
  value: number
  startDate: Date | null
  endDate: Date | null
}

const route = useRoute()
const router = useRouter()
const confirm = useConfirm()
const toast = useToast()

const scopeOptions: ScopeFilter[] = ['ALL', 'PRODUCT', 'TRANSACTION', 'VOUCHER']
const formScopeOptions: DiscountScope[] = ['PRODUCT', 'TRANSACTION', 'VOUCHER']
const statusOptions: StatusFilter[] = ['ALL', 'ACTIVE', 'PENDING', 'EXPIRED', 'STOPPED', 'CANCELLED']
const discountTypes: DiscountType[] = ['PERCENTAGE', 'FIXED']

const scopeFilter = ref<ScopeFilter>('ALL')
const statusFilter = ref<StatusFilter>('ALL')
const selectedProductId = ref<number | null>(null)
const startDateFilter = ref<Date | null>(null)
const endDateFilter = ref<Date | null>(null)

const discountQuery = computed(() => ({
  productId: scopeFilter.value === 'PRODUCT' && selectedProductId.value ? selectedProductId.value : undefined,
}))
const { data: discounts, isLoading } = useDiscounts(discountQuery)
const { data: products, isLoading: productsLoading } = useProducts()
const { mutate: createDiscount } = useCreateDiscount()
const { mutate: updateDiscount } = useUpdateDiscount()
const { mutate: deleteDiscount } = useDeleteDiscount()

const showDialog = ref(false)
const editingDiscount = ref<Discount | null>(null)
const submitting = ref(false)
const form = ref<DiscountForm>(defaultForm())

const productOptions = computed(() =>
  (products.value || []).map((product) => ({
    label: `${product.name} - ${formatCurrency(product.price)}`,
    value: product.id,
  })),
)

const productMap = computed(() => {
  const map = new Map<number, Product>()
  for (const product of products.value || []) {
    map.set(product.id, product)
  }
  return map
})

const selectedFormProduct = computed(() => {
  if (!form.value.productId) return null
  return productMap.value.get(form.value.productId) ?? null
})

const dialogTitle = computed(() => (editingDiscount.value ? 'Edit Discount' : 'Add Discount'))

const filteredDiscounts = computed(() => {
  return (discounts.value || []).filter((discount) => {
    if (scopeFilter.value !== 'ALL' && discount.scope !== scopeFilter.value) return false
    if (statusFilter.value !== 'ALL' && discount.effectiveStatus !== statusFilter.value) return false
    if (scopeFilter.value === 'PRODUCT' && selectedProductId.value && discount.productId !== selectedProductId.value) return false
    if (startDateFilter.value && new Date(discount.startDate) < startOfDay(startDateFilter.value)) return false
    if (endDateFilter.value && new Date(discount.endDate) > endOfDay(endDateFilter.value)) return false
    return true
  })
})

watch(
  () => route.query,
  () => applyQueryFilters(),
  { immediate: true },
)

watch([scopeFilter, selectedProductId], () => updateRouteQuery())

function applyQueryFilters() {
  const queryScope = normalizeScope(route.query.scope)
  const productId = parseQueryNumber(route.query.productId)
  if (queryScope) scopeFilter.value = queryScope
  if (productId) selectedProductId.value = productId
  if (productId && !queryScope) scopeFilter.value = 'PRODUCT'
}

function normalizeScope(value: unknown): DiscountScope | null {
  const raw = Array.isArray(value) ? value[0] : value
  if (raw === 'PRODUCT' || raw === 'TRANSACTION' || raw === 'VOUCHER') return raw
  return null
}

function parseQueryNumber(value: unknown): number | null {
  const raw = Array.isArray(value) ? value[0] : value
  const parsed = Number(raw)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

function updateRouteQuery() {
  const query: Record<string, string> = {}
  if (scopeFilter.value !== 'ALL') query.scope = scopeFilter.value
  if (scopeFilter.value === 'PRODUCT' && selectedProductId.value) query.productId = String(selectedProductId.value)
  const currentScope = Array.isArray(route.query.scope) ? route.query.scope[0] : route.query.scope
  const currentProductId = Array.isArray(route.query.productId) ? route.query.productId[0] : route.query.productId
  if ((currentScope ?? '') === (query.scope ?? '') && (currentProductId ?? '') === (query.productId ?? '')) return
  router.replace({ path: '/discounts', query })
}

function onScopeFilterChange(scope: ScopeFilter) {
  scopeFilter.value = scope
  if (scope !== 'PRODUCT') selectedProductId.value = null
}

function onProductFilterChange(productId: number | null) {
  selectedProductId.value = productId
  if (productId) scopeFilter.value = 'PRODUCT'
}

function openCreate() {
  editingDiscount.value = null
  form.value = defaultForm()
  if (scopeFilter.value !== 'ALL') form.value.scope = scopeFilter.value
  if (form.value.scope === 'PRODUCT' && selectedProductId.value) form.value.productId = selectedProductId.value
  showDialog.value = true
}

function openEdit(discount: Discount) {
  editingDiscount.value = discount
  form.value = {
    scope: discount.scope,
    productId: discount.productId ?? null,
    voucherCode: discount.voucherCode ?? '',
    minTransaction: discount.minTransaction ?? 0,
    name: discount.name,
    type: discount.type,
    value: discount.value,
    startDate: new Date(discount.startDate),
    endDate: new Date(discount.endDate),
  }
  showDialog.value = true
}

function defaultForm(): DiscountForm {
  return {
    scope: 'PRODUCT',
    productId: null,
    voucherCode: '',
    minTransaction: 0,
    name: '',
    type: 'PERCENTAGE',
    value: 0,
    startDate: null,
    endDate: null,
  }
}

function onFormScopeChange(scope: DiscountScope) {
  form.value.scope = scope
  if (scope !== 'PRODUCT') form.value.productId = null
  if (scope !== 'VOUCHER') form.value.voucherCode = ''
  if (scope === 'PRODUCT') form.value.minTransaction = 0
}

function saveDiscount() {
  const validationMessage = validateForm()
  if (validationMessage) {
    toast.add({ severity: 'warn', summary: 'Validation', detail: validationMessage, life: 3000 })
    return
  }

  const payload = buildPayload()
  submitting.value = true
  if (editingDiscount.value) {
    updateDiscount(
      { id: editingDiscount.value.id, data: payload },
      {
        onSuccess: () => {
          toast.add({ severity: 'success', summary: 'Updated', detail: 'Discount updated successfully', life: 3000 })
          showDialog.value = false
        },
        onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
        onSettled: () => {
          submitting.value = false
        },
      },
    )
    return
  }

  createDiscount(payload as CreateDiscountRequest, {
    onSuccess: () => {
      toast.add({ severity: 'success', summary: 'Created', detail: 'Discount created successfully', life: 3000 })
      showDialog.value = false
    },
    onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
    onSettled: () => {
      submitting.value = false
    },
  })
}

function validateForm(): string | null {
  if (!form.value.name.trim()) return 'Name is required'
  if (!form.value.startDate || !form.value.endDate) return 'Start date and end date are required'
  if (form.value.value <= 0) return 'Discount value must be greater than 0'
  if (form.value.type === 'PERCENTAGE' && form.value.value > 100) return 'Percentage discount must be 100 or less'
  if (form.value.scope === 'PRODUCT' && !form.value.productId) return 'Product is required for product discounts'
  if (form.value.scope === 'VOUCHER' && !form.value.voucherCode.trim()) return 'Voucher code is required'
  if ((form.value.scope === 'TRANSACTION' || form.value.scope === 'VOUCHER') && form.value.minTransaction < 0) return 'Minimum transaction must be 0 or greater'
  if (form.value.startDate > form.value.endDate) return 'Start date must be before or equal to end date'
  return null
}

function buildPayload(): CreateDiscountRequest | UpdateDiscountRequest {
  const payload: CreateDiscountRequest | UpdateDiscountRequest = {
    scope: form.value.scope,
    name: form.value.name.trim(),
    type: form.value.type,
    value: form.value.value,
    startDate: dateToDateStr(form.value.startDate!),
    endDate: dateToDateStr(form.value.endDate!),
  }
  if (form.value.scope === 'PRODUCT') {
    payload.productId = form.value.productId!
    return payload
  }
  payload.minTransaction = form.value.minTransaction || 0
  if (form.value.scope === 'VOUCHER') payload.voucherCode = form.value.voucherCode.trim().toUpperCase()
  return payload
}

function confirmDelete(discount: Discount) {
  confirm.require({
    message: `Delete "${discount.name}"?`,
    header: 'Confirm Delete',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel',
    rejectProps: { severity: 'secondary', outlined: true },
    acceptLabel: 'Delete',
    acceptProps: { severity: 'danger' },
    accept: () => {
      deleteDiscount(discount.id, {
        onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'Discount deleted', life: 3000 }),
        onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      })
    },
  })
}

function clearFilters() {
  scopeFilter.value = 'ALL'
  statusFilter.value = 'ALL'
  selectedProductId.value = null
  startDateFilter.value = null
  endDateFilter.value = null
}

function formatCurrency(value: number): string {
  return `Rp ${value?.toLocaleString('id-ID')}`
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString('id-ID')
}

function formatDiscountValue(discount: Discount): string {
  if (discount.type === 'PERCENTAGE') return `${discount.value}%`
  return formatCurrency(discount.value)
}

function getProductLabel(productId?: number | null): string {
  if (!productId) return '-'
  return productMap.value.get(productId)?.name ?? `#${productId}`
}

function getScopeSeverity(scope: DiscountScope): string {
  switch (scope) {
    case 'PRODUCT':
      return 'info'
    case 'TRANSACTION':
      return 'success'
    case 'VOUCHER':
      return 'warn'
    default:
      return 'secondary'
  }
}

function getStatusSeverity(status: DiscountStatus): string {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'PENDING':
      return 'warn'
    case 'EXPIRED':
      return 'secondary'
    case 'STOPPED':
    case 'CANCELLED':
      return 'danger'
    default:
      return 'info'
  }
}

function dateToDateStr(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function startOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

function endOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59, 999)
}
</script>

<template>
  <div class="discounts-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Discounts</h1>
        <p class="page-subtitle">Manage product, transaction, and voucher discounts</p>
      </div>
      <Button label="Add Discount" icon="pi pi-plus" @click="openCreate" />
    </div>

    <div class="filter-bar">
      <Select :modelValue="scopeFilter" :options="scopeOptions" placeholder="Scope" style="width: 150px" @update:modelValue="onScopeFilterChange" />
      <Select v-model="statusFilter" :options="statusOptions" placeholder="Status" style="width: 160px" />
      <Select
        v-if="scopeFilter === 'PRODUCT'"
        :modelValue="selectedProductId"
        :options="productOptions"
        :loading="productsLoading"
        showClear
        filter
        optionLabel="label"
        optionValue="value"
        placeholder="Product"
        class="product-filter"
        @update:modelValue="onProductFilterChange"
      />
      <DatePicker v-model="startDateFilter" dateFormat="yy-mm-dd" showIcon showClear placeholder="From" style="width: 160px" />
      <DatePicker v-model="endDateFilter" dateFormat="yy-mm-dd" showIcon showClear placeholder="To" style="width: 160px" />
      <Button label="Clear" icon="pi pi-filter-slash" text severity="secondary" @click="clearFilters" />
    </div>

    <div class="card">
      <DataTable :value="filteredDiscounts" :loading="isLoading" tableStyle="min-width: 70rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
        <Column field="name" header="Name" sortable />
        <Column header="Scope" sortable style="width: 130px">
          <template #body="{ data }: { data: Discount }">
            <Tag :value="data.scope" :severity="getScopeSeverity(data.scope)" />
          </template>
        </Column>
        <Column header="Product" style="width: 180px">
          <template #body="{ data }: { data: Discount }">{{ data.scope === 'PRODUCT' ? getProductLabel(data.productId) : '-' }}</template>
        </Column>
        <Column header="Voucher" style="width: 140px">
          <template #body="{ data }: { data: Discount }">{{ data.scope === 'VOUCHER' ? data.voucherCode : '-' }}</template>
        </Column>
        <Column header="Min Transaction" style="width: 150px">
          <template #body="{ data }: { data: Discount }">
            {{ data.scope === 'PRODUCT' ? '-' : formatCurrency(data.minTransaction || 0) }}
          </template>
        </Column>
        <Column field="type" header="Type" style="width: 130px">
          <template #body="{ data }: { data: Discount }">
            <Tag :value="data.type" severity="info" />
          </template>
        </Column>
        <Column header="Value" style="width: 130px">
          <template #body="{ data }: { data: Discount }">{{ formatDiscountValue(data) }}</template>
        </Column>
        <Column header="Start Date" sortable style="width: 130px">
          <template #body="{ data }: { data: Discount }">{{ formatDate(data.startDate) }}</template>
        </Column>
        <Column header="End Date" sortable style="width: 130px">
          <template #body="{ data }: { data: Discount }">{{ formatDate(data.endDate) }}</template>
        </Column>
        <Column header="Status" sortable style="width: 130px">
          <template #body="{ data }: { data: Discount }">
            <Tag :value="data.effectiveStatus" :severity="getStatusSeverity(data.effectiveStatus)" />
          </template>
        </Column>
        <Column header="Actions" style="width: 100px">
          <template #body="{ data }: { data: Discount }">
            <div class="actions">
              <Button icon="pi pi-pencil" text rounded size="small" title="Edit" @click="openEdit(data)" />
              <Button icon="pi pi-trash" text rounded size="small" severity="danger" title="Delete" @click="confirmDelete(data)" />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No discounts found.</div>
        </template>
      </DataTable>
    </div>

    <Dialog v-model:visible="showDialog" :header="dialogTitle" :modal="true" :style="{ width: '560px' }">
      <div class="form-grid">
        <div class="form-field">
          <label>Scope *</label>
          <Select v-model="form.scope" :options="formScopeOptions" fluid @change="onFormScopeChange($event.value)" />
        </div>

        <div v-if="form.scope === 'PRODUCT'" class="form-field">
          <label>Product *</label>
          <Select v-model="form.productId" :options="productOptions" :loading="productsLoading" filter optionLabel="label" optionValue="value" placeholder="Select product" fluid />
        </div>

        <div v-if="form.scope === 'PRODUCT' && selectedFormProduct" class="product-context">
          <span>Product ID: {{ selectedFormProduct.id }}</span>
          <strong>{{ formatCurrency(selectedFormProduct.price) }}</strong>
        </div>

        <div v-if="form.scope === 'VOUCHER'" class="form-field">
          <label>Voucher Code *</label>
          <InputText v-model="form.voucherCode" fluid />
        </div>

        <div v-if="form.scope === 'TRANSACTION' || form.scope === 'VOUCHER'" class="form-field">
          <label>Minimum Transaction</label>
          <InputNumber v-model="form.minTransaction" mode="currency" currency="IDR" :min="0" fluid />
        </div>

        <div class="form-field">
          <label>Name *</label>
          <InputText v-model="form.name" fluid />
        </div>

        <div class="form-field">
          <label>Type *</label>
          <Select v-model="form.type" :options="discountTypes" fluid />
        </div>

        <div class="form-field">
          <label>{{ form.type === 'PERCENTAGE' ? 'Discount Percentage' : 'Discount Amount' }} *</label>
          <InputNumber v-if="form.type === 'PERCENTAGE'" v-model="form.value" suffix="%" :min="0" :max="100" :minFractionDigits="0" :maxFractionDigits="2" fluid />
          <InputNumber v-else v-model="form.value" mode="currency" currency="IDR" :min="0" fluid />
        </div>

        <div class="form-row">
          <div class="form-field">
            <label>Start Date *</label>
            <DatePicker v-model="form.startDate" dateFormat="yy-mm-dd" showIcon fluid />
          </div>
          <div class="form-field">
            <label>End Date *</label>
            <DatePicker v-model="form.endDate" dateFormat="yy-mm-dd" showIcon fluid />
          </div>
        </div>
      </div>

      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showDialog = false" />
        <Button label="Save" :loading="submitting" @click="saveDiscount" />
      </template>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>

<style scoped>
.discounts-page {
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
.filter-bar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
.product-filter {
  width: 280px;
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
.form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
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
.product-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--p-surface-50);
  color: var(--p-text-secondary-color);
  font-size: 13px;
}
.product-context strong {
  color: var(--p-text-color);
  font-size: 14px;
}
@media (max-width: 640px) {
  .filter-bar,
  .page-header {
    align-items: stretch;
    flex-direction: column;
  }
  .product-filter {
    width: 100%;
  }
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
