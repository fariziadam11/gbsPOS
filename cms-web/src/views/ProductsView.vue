<script setup lang="ts">
import { ref, computed } from "vue";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Button from "primevue/button";
import Dialog from "primevue/dialog";
import InputText from "primevue/inputtext";
import InputNumber from "primevue/inputnumber";
import Select from "primevue/select";
import Tag from "primevue/tag";
import Tabs from "primevue/tabs";
import TabList from "primevue/tablist";
import Tab from "primevue/tab";
import TabPanels from "primevue/tabpanels";
import TabPanel from "primevue/tabpanel";
import Chip from "primevue/chip";
import ConfirmDialog from "primevue/confirmdialog";
import FileUpload from "primevue/fileupload";
import DatePicker from "primevue/datepicker";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { useAuthStore } from "../stores/auth";
import { useProducts, useCreateProduct, useUpdateProduct, useDeleteProduct, useImportProducts } from "../composables/useProducts";
import { useDiscounts, useCreateDiscount, useUpdateDiscount, useDeleteDiscount } from "../composables/useDiscounts";
import { getExportUrl, getVariants, createVariant, updateVariant, deleteVariant } from "../api/products";
import type { VariantItem, CreateVariantReq } from "../api/products";
import { getErrorMessage } from "../api/client";
import type { Product, CreateProductRequest, Discount, DiscountStatus, DiscountType } from "../types/api";
const authStore = useAuthStore();
const confirm = useConfirm();
const toast = useToast();

const storeType = ref<string | undefined>(undefined);
const { data: products, isLoading } = useProducts(storeType);
const { mutate: createProduct } = useCreateProduct();
const { mutate: updateProduct } = useUpdateProduct();
const { mutate: deleteProduct } = useDeleteProduct();
const { mutate: importProducts } = useImportProducts();

const selectedDiscountProduct = ref<Product | null>(null);
const selectedDiscountProductId = computed(() => selectedDiscountProduct.value?.id ?? null);
const { data: discounts, isLoading: isDiscountsLoading } = useDiscounts(selectedDiscountProductId);
const { mutate: createDiscount } = useCreateDiscount();
const { mutate: updateDiscount } = useUpdateDiscount();
const { mutate: deleteDiscount } = useDeleteDiscount();

const showDialog = ref(false);
const editingProduct = ref<Product | null>(null);
const form = ref<CreateProductRequest>({
  name: "",
  price: 0,
  category: "",
  imageUrl: "",
  storeType: "",
  stockQuantity: 0,
  lowStockThreshold: 10,
});
const dialogTitle = ref("Add Product");
const submitting = ref(false);
const activeTab = ref("0");

const variants = ref<VariantItem[]>([]);
const variantLoading = ref(false);
const showVariantDialog = ref(false);
const editingVariant = ref<VariantItem | null>(null);
const variantForm = ref<CreateVariantReq>({ name: "", attributes: {}, stockQuantity: 0 });
const attrKey = ref("");
const attrValue = ref("");
const variantSubmitting = ref(false);
type DiscountForm = {
  name: string;
  type: DiscountType;
  value: number;
  startDate: Date | null;
  endDate: Date | null;
  finalPrice: number;
};

const showDiscountDialog = ref(false);
const showDiscountFormDialog = ref(false);
const editingDiscount = ref<Discount | null>(null);
const discountSubmitting = ref(false);
const discountForm = ref<DiscountForm>({
  name: "",
  type: "PERCENTAGE",
  value: 0,
  startDate: null,
  endDate: null,
  finalPrice: 0,
});
const discountDialogTitle = computed(() => {
  if (!selectedDiscountProduct.value) return "Product Discounts";
  return `Product Discounts - ${selectedDiscountProduct.value.name}`;
});
const discountFormTitle = computed(() => (editingDiscount.value ? "Edit Discount" : "Add Discount"));
const selectedProductPrice = computed(() => selectedDiscountProduct.value?.price ?? 0);

const storeTypes = ["RETAIL", "FNB", "OUTFIT"];
const categories = ["Food", "Beverages", "Electronics", "Groceries"];
const discountTypes: DiscountType[] = ["PERCENTAGE", "FIXED"];

async function loadVariants(productId: number) {
  variantLoading.value = true;
  try {
    const res = await getVariants(productId);
    variants.value = res.success ? res.data : [];
  } catch {
    variants.value = [];
  }
  variantLoading.value = false;
}

function openCreate() {
  editingProduct.value = null;
  dialogTitle.value = "Add Product";
  activeTab.value = "0";
  form.value = {
    name: "",
    price: 0,
    category: "",
    imageUrl: "",
    storeType: storeType.value || "RETAIL",
    stockQuantity: 0,
    lowStockThreshold: 10,
  };
  variants.value = [];
  showDialog.value = true;
}

function openEdit(product: Product) {
  editingProduct.value = product;
  dialogTitle.value = "Edit Product";
  activeTab.value = "0";
  form.value = {
    name: product.name,
    price: product.price,
    category: product.category,
    imageUrl: product.imageUrl,
    storeType: product.storeType,
    stockQuantity: product.stockQuantity,
    lowStockThreshold: product.lowStockThreshold,
  };
  loadVariants(product.id);
  showDialog.value = true;
}

function openDiscountManagement(product: Product) {
  selectedDiscountProduct.value = product;
  showDiscountDialog.value = true;
}

function openCreateDiscount() {
  if (!selectedDiscountProduct.value) return;
  editingDiscount.value = null;
  discountForm.value = {
    name: "",
    type: "PERCENTAGE",
    value: 0,
    startDate: null,
    endDate: null,
    finalPrice: selectedProductPrice.value,
  };
  showDiscountFormDialog.value = true;
}

function openEditDiscount(discount: Discount) {
  editingDiscount.value = discount;
  discountForm.value = {
    name: discount.name,
    type: discount.type,
    value: discount.value,
    startDate: new Date(discount.startDate),
    endDate: new Date(discount.endDate),
    finalPrice: calculateFinalPrice(discount.type, discount.value),
  };
  showDiscountFormDialog.value = true;
}
function save() {
  if (!form.value.name || !form.value.category) {
    toast.add({ severity: "warn", summary: "Validation", detail: "Name and category are required", life: 3000 });
    return;
  }
  submitting.value = true;
  if (editingProduct.value) {
    updateProduct(
      { id: editingProduct.value.id, data: form.value },
      {
        onSuccess: () => {
          toast.add({ severity: "success", summary: "Updated", detail: "Product updated successfully", life: 3000 });
          showDialog.value = false;
        },
        onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
        onSettled: () => {
          submitting.value = false;
        },
      },
    );
  } else {
    createProduct(form.value as CreateProductRequest, {
      onSuccess: () => {
        toast.add({ severity: "success", summary: "Created", detail: "Product created successfully", life: 3000 });
        showDialog.value = false;
      },
      onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => {
        submitting.value = false;
      },
    });
  }
}

function confirmDelete(product: Product) {
  confirm.require({
    message: `Delete "${product.name}"?`,
    header: "Confirm Delete",
    icon: "pi pi-exclamation-triangle",
    rejectLabel: "Cancel",
    rejectProps: { severity: "secondary", outlined: true },
    acceptLabel: "Delete",
    acceptProps: { severity: "danger" },
    accept: () => {
      deleteProduct(product.id, {
        onSuccess: () => toast.add({ severity: "success", summary: "Deleted", detail: "Product deleted", life: 3000 }),
        onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
      });
    },
  });
}

function saveDiscount() {
  if (!selectedDiscountProduct.value) return;
  const startDate = discountForm.value.startDate;
  const endDate = discountForm.value.endDate;
  if (!discountForm.value.name || !startDate || !endDate) {
    toast.add({ severity: "warn", summary: "Validation", detail: "Name, start date, and end date are required", life: 3000 });
    return;
  }
  if (discountForm.value.value <= 0) {
    toast.add({ severity: "warn", summary: "Validation", detail: "Discount value must be greater than 0", life: 3000 });
    return;
  }

  const payload = {
    productId: selectedDiscountProduct.value.id,
    name: discountForm.value.name,
    type: discountForm.value.type,
    value: discountForm.value.value,
    startDate: dateToDateStr(startDate),
    endDate: dateToDateStr(endDate),
  };

  discountSubmitting.value = true;
  if (editingDiscount.value) {
    updateDiscount(
      { id: editingDiscount.value.id, data: payload },
      {
        onSuccess: () => {
          toast.add({ severity: "success", summary: "Updated", detail: "Discount updated successfully", life: 3000 });
          showDiscountFormDialog.value = false;
        },
        onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
        onSettled: () => {
          discountSubmitting.value = false;
        },
      },
    );
  } else {
    createDiscount(payload, {
      onSuccess: () => {
        toast.add({ severity: "success", summary: "Created", detail: "Discount created successfully", life: 3000 });
        showDiscountFormDialog.value = false;
      },
      onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => {
        discountSubmitting.value = false;
      },
    });
  }
}

function confirmDeleteDiscount(discount: Discount) {
  confirm.require({
    message: `Delete "${discount.name}"?`,
    header: "Confirm Delete",
    icon: "pi pi-exclamation-triangle",
    rejectLabel: "Cancel",
    rejectProps: { severity: "secondary", outlined: true },
    acceptLabel: "Delete",
    acceptProps: { severity: "danger" },
    accept: () => {
      deleteDiscount(discount.id, {
        onSuccess: () => toast.add({ severity: "success", summary: "Deleted", detail: "Discount deleted", life: 3000 }),
        onError: (err) => toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 }),
      });
    },
  });
}

function formatCurrency(value: number): string {
  return `Rp ${value?.toLocaleString("id-ID")}`;
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString("id-ID");
}

function dateToDateStr(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function calculateFinalPrice(type: DiscountType, value: number): number {
  const price = selectedProductPrice.value;
  const finalPrice = type === "PERCENTAGE" ? price - (price * value) / 100 : price - value;
  return clampPrice(finalPrice);
}

function clampPrice(value: number): number {
  return Math.min(selectedProductPrice.value, Math.max(0, value || 0));
}

function onDiscountTypeChange(type: DiscountType) {
  discountForm.value.type = type;
  discountForm.value.value = 0;
  discountForm.value.finalPrice = selectedProductPrice.value;
}

function onDiscountValueChange(value: number | null) {
  const price = selectedProductPrice.value;
  const normalizedValue = value ?? 0;
  discountForm.value.value = discountForm.value.type === "PERCENTAGE" ? Math.min(100, Math.max(0, normalizedValue)) : Math.min(price, Math.max(0, normalizedValue));
  discountForm.value.finalPrice = calculateFinalPrice(discountForm.value.type, discountForm.value.value);
}

function onFinalPriceChange(value: number | null) {
  const price = selectedProductPrice.value;
  const finalPrice = clampPrice(value ?? 0);
  discountForm.value.finalPrice = finalPrice;
  if (discountForm.value.type === "PERCENTAGE") {
    discountForm.value.value = price > 0 ? ((price - finalPrice) / price) * 100 : 0;
    return;
  }
  discountForm.value.value = price - finalPrice;
}

function formatDiscountValue(discount: Discount): string {
  if (discount.type === "PERCENTAGE") return `${discount.value}%`;
  return formatCurrency(discount.value);
}

const formatDiscountTag = (discount: any) => {
  if (discount.type === "PERCENTAGE") {
    return `${discount.value}%`;
  }

  return `${formatCurrency(discount.value)}`;
};

function getStatusSeverity(status: DiscountStatus): string {
  switch (status) {
    case "ACTIVE":
      return "success";
    case "PENDING":
      return "warn";
    case "EXPIRED":
      return "secondary";
    case "STOPPED":
    case "CANCELLED":
      return "danger";
    default:
      return "info";
  }
}

function getStockSeverity(product: Product): string {
  if (product.stockQuantity <= 0) return "danger";
  if (product.stockQuantity <= product.lowStockThreshold) return "warn";
  return "success";
}
function getStockLabel(product: Product): string {
  if (product.stockQuantity <= 0) return "Out";
  if (product.stockQuantity <= product.lowStockThreshold) return "Low";
  return "OK";
}

function addAttribute() {
  if (attrKey.value.trim() && attrValue.value.trim()) {
    variantForm.value.attributes = { ...variantForm.value.attributes, [attrKey.value.trim()]: attrValue.value.trim() };
    attrKey.value = "";
    attrValue.value = "";
  }
}

function removeAttribute(key: string) {
  const attrs = { ...variantForm.value.attributes };
  delete attrs[key];
  variantForm.value.attributes = attrs;
}

function openVariantCreate() {
  editingVariant.value = null;
  variantForm.value = { name: "", attributes: {}, stockQuantity: 0 };
  showVariantDialog.value = true;
}

function openVariantEdit(v: VariantItem) {
  editingVariant.value = v;
  variantForm.value = {
    sku: v.sku,
    name: v.name,
    attributes: { ...v.attributes },
    price: v.price,
    stockQuantity: v.stockQuantity,
    lowStockThreshold: v.lowStockThreshold,
    sortOrder: v.sortOrder,
  };
  showVariantDialog.value = true;
}

function saveVariant() {
  if (!variantForm.value.name) {
    toast.add({ severity: "warn", summary: "Validation", detail: "Name required", life: 3000 });
    return;
  }
  if (!editingProduct.value) return;
  variantSubmitting.value = true;
  const pid = editingProduct.value.id;
  const onDone = () => {
    variantSubmitting.value = false;
    showVariantDialog.value = false;
    loadVariants(pid);
  };
  const onErr = (err: unknown) => {
    toast.add({ severity: "error", summary: "Error", detail: getErrorMessage(err), life: 5000 });
    variantSubmitting.value = false;
  };
  if (editingVariant.value) {
    updateVariant(editingVariant.value.id, variantForm.value).then(onDone).catch(onErr);
  } else {
    createVariant(pid, variantForm.value).then(onDone).catch(onErr);
  }
}

function confirmDeleteVariant(v: VariantItem) {
  confirm.require({
    message: `Delete variant "${v.name}"?`,
    header: "Confirm Delete",
    icon: "pi pi-exclamation-triangle",
    rejectLabel: "Cancel",
    rejectProps: { severity: "secondary", outlined: true },
    acceptLabel: "Delete",
    acceptProps: { severity: "danger" },
    accept: () =>
      deleteVariant(v.id).then(() => {
        toast.add({ severity: "success", summary: "Deleted", detail: "Variant deleted", life: 3000 });
        if (editingProduct.value) loadVariants(editingProduct.value.id);
      }),
  });
}

function onImport(event: any) {
  const file = event.files?.[0] as File;
  if (!file) return;
  importProducts(
    { file, storeType: storeType.value },
    {
      onSuccess: (data) => {
        const result = data.data;
        toast.add({
          severity: result.failed > 0 ? "warn" : "success",
          summary: "Import Complete",
          detail: `Success: ${result.success}, Failed: ${result.failed}`,
          life: 5000,
        });
      },
      onError: (err) => toast.add({ severity: "error", summary: "Import Error", detail: getErrorMessage(err), life: 5000 }),
    },
  );
}

const baseExportUrl = getExportUrl();
const exportUrl = computed(() => {
  let url = baseExportUrl;
  if (storeType.value) url += `&storeType=${storeType.value}`;
  return url;
});
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
        <a :href="exportUrl" class="export-link"><Button label="Export CSV" icon="pi pi-download" text severity="secondary" /></a>
        <Button v-if="authStore.isAdmin" label="Add Product" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>
    <div class="card">
      <DataTable :value="products || []" :loading="isLoading" tableStyle="min-width:60rem" stripedRows size="small" paginator :rows="20" :rowsPerPageOptions="[10, 20, 50]">
        <Column field="id" header="ID" sortable style="width: 60px" />
        <Column field="name" header="Name" sortable />
        <!-- <Column header="Price" sortable style="width: 120px">
          <template #body="{ data }">{{ formatCurrency(data.price) }}</template>
        </Column> -->

        <!-- <Column header="Price" sortable style="width: 140px">
          <template #body="{ data }">
            <div class="price-cell">
              <span class="final-price">
                {{ formatCurrency(data.finalPrice ?? data.price) }}
              </span>

              <small v-if="data.discount" class="original-price">
                {{ formatCurrency(data.price) }}
              </small>
            </div>
          </template>
        </Column> -->

        <Column header="Price" sortable style="width: 140px">
          <template #body="{ data }">
            <div class="price-cell">
              <span class="final-price">
                {{ formatCurrency(data.finalPrice ?? data.price) }}
              </span>

              <small v-if="data.discount" class="original-price">
                {{ formatCurrency(data.price) }}
              </small>
            </div>
          </template>
        </Column>

        <Column header="Discount" style="width: 160px">
          <template #body="{ data }">
            <Tag v-if="data.discount" :value="formatDiscountTag(data.discount)" severity="success" />

            <Tag v-else value="No Discount" severity="contrast" />
          </template>
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
              <Tag :value="getStockLabel(data)" :severity="getStockSeverity(data)" style="font-size: 11px" />
            </div>
          </template>
        </Column>
        <Column v-if="authStore.isAdmin" header="Actions" style="width: 150px">
          <template #body="{ data }">
            <div class="actions">
              <Button icon="pi pi-tag" text rounded size="small" title="Manage Discount" @click="openDiscountManagement(data)" />
              <Button icon="pi pi-pencil" text rounded size="small" title="Edit Product" @click="openEdit(data)" />
              <Button icon="pi pi-trash" text rounded size="small" severity="danger" title="Delete Product" @click="confirmDelete(data)" />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="empty-state">No products found.</div>
        </template>
      </DataTable>
    </div>

    <!-- Product + Variants Dialog -->
    <Dialog v-model:visible="showDialog" :header="dialogTitle" :modal="true" :style="{ width: editingProduct ? '700px' : '500px' }">
      <Tabs v-if="editingProduct" v-model:value="activeTab">
        <TabList>
          <Tab value="0">Product</Tab>
          <Tab value="1">Variants ({{ variants.length }})</Tab>
        </TabList>
        <TabPanels>
          <TabPanel value="0">
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
          </TabPanel>
          <TabPanel value="1">
            <div style="margin-bottom: 12px">
              <Button label="Add Variant" icon="pi pi-plus" size="small" @click="openVariantCreate" />
            </div>
            <DataTable :value="variants" :loading="variantLoading" size="small" stripedRows>
              <Column field="name" header="Name" />
              <Column field="sku" header="SKU" style="width: 100px" />
              <Column header="Attributes" style="width: 150px">
                <template #body="{ data }"><Chip v-for="(v, k) in data.attributes" :key="k" :label="`${k}:${v}`" style="margin: 2px" /></template>
              </Column>
              <Column header="Price" style="width: 100px">
                <template #body="{ data }">{{ data.price ? formatCurrency(data.price) : "-" }}</template>
              </Column>
              <Column field="stockQuantity" header="Stock" style="width: 70px" />
              <Column header="Actions" style="width: 100px">
                <template #body="{ data }">
                  <div class="actions">
                    <Button icon="pi pi-pencil" text rounded size="small" @click="openVariantEdit(data)" />
                    <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="confirmDeleteVariant(data)" />
                  </div>
                </template>
              </Column>
            </DataTable>
          </TabPanel>
        </TabPanels>
      </Tabs>
      <div v-else class="form-grid">
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

    <!-- Discount Management Dialog -->
    <Dialog v-model:visible="showDiscountDialog" :header="discountDialogTitle" :modal="true" :style="{ width: '900px' }">
      <div class="discount-dialog">
        <div class="discount-toolbar">
          <div v-if="selectedDiscountProduct" class="product-price">
            <span>Product Price</span>
            <strong>{{ formatCurrency(selectedDiscountProduct.price) }}</strong>
          </div>
          <Button label="Add Discount" icon="pi pi-plus" @click="openCreateDiscount" />
        </div>

        <DataTable :value="discounts || []" :loading="isDiscountsLoading" tableStyle="min-width: 48rem" stripedRows size="small">
          <Column field="name" header="Name" />
          <Column field="type" header="Type" style="width: 130px">
            <template #body="{ data }">
              <Tag :value="data.type" severity="info" />
            </template>
          </Column>
          <Column header="Value" style="width: 130px">
            <template #body="{ data }">{{ formatDiscountValue(data) }}</template>
          </Column>
          <Column header="Start Date" style="width: 130px">
            <template #body="{ data }">{{ formatDate(data.startDate) }}</template>
          </Column>
          <Column header="End Date" style="width: 130px">
            <template #body="{ data }">{{ formatDate(data.endDate) }}</template>
          </Column>
          <Column header="Effective Status" style="width: 150px">
            <template #body="{ data }">
              <Tag :value="data.effectiveStatus" :severity="getStatusSeverity(data.effectiveStatus)" />
            </template>
          </Column>
          <Column header="Actions" style="width: 100px">
            <template #body="{ data }">
              <div class="actions">
                <Button icon="pi pi-pencil" text rounded size="small" title="Edit" @click="openEditDiscount(data)" />
                <Button icon="pi pi-trash" text rounded size="small" severity="danger" title="Delete" @click="confirmDeleteDiscount(data)" />
              </div>
            </template>
          </Column>
          <template #empty>
            <div class="empty-state">No discounts found.</div>
          </template>
        </DataTable>
      </div>
    </Dialog>

    <!-- Discount Form Dialog -->
    <Dialog v-model:visible="showDiscountFormDialog" :header="discountFormTitle" :modal="true" :style="{ width: '520px' }">
      <div class="form-grid">
        <div class="form-field">
          <label>Name *</label>
          <InputText v-model="discountForm.name" fluid />
        </div>
        <div class="form-field">
          <label>Type *</label>
          <Select v-model="discountForm.type" :options="discountTypes" fluid @change="onDiscountTypeChange($event.value)" />
        </div>
        <div class="form-row">
          <div class="form-field">
            <label>Start Date *</label>
            <DatePicker v-model="discountForm.startDate" dateFormat="yy-mm-dd" showIcon fluid />
          </div>
          <div class="form-field">
            <label>End Date *</label>
            <DatePicker v-model="discountForm.endDate" dateFormat="yy-mm-dd" showIcon fluid />
          </div>
        </div>
        <div class="form-field">
          <label>Product Price</label>
          <InputNumber :modelValue="selectedProductPrice" mode="currency" currency="IDR" disabled fluid />
        </div>
        <template v-if="discountForm.type === 'PERCENTAGE'">
          <div class="form-field">
            <label>Discount Percentage</label>
            <InputNumber :modelValue="discountForm.value" suffix="%" :min="0" :max="100" :minFractionDigits="0" :maxFractionDigits="2" fluid @update:modelValue="onDiscountValueChange" />
          </div>
          <div class="form-field">
            <label>Final Price</label>
            <InputNumber :modelValue="discountForm.finalPrice" mode="currency" currency="IDR" :min="0" :max="selectedProductPrice" fluid @update:modelValue="onFinalPriceChange" />
          </div>
        </template>
        <template v-else>
          <div class="form-field">
            <label>Discount Amount</label>
            <InputNumber :modelValue="discountForm.value" mode="currency" currency="IDR" :min="0" :max="selectedProductPrice" fluid @update:modelValue="onDiscountValueChange" />
          </div>
          <div class="form-field">
            <label>Final Price</label>
            <InputNumber :modelValue="discountForm.finalPrice" mode="currency" currency="IDR" :min="0" :max="selectedProductPrice" fluid @update:modelValue="onFinalPriceChange" />
          </div>
        </template>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showDiscountFormDialog = false" />
        <Button label="Save" :loading="discountSubmitting" @click="saveDiscount" />
      </template>
    </Dialog>

    <!-- Variant Form Dialog -->
    <Dialog v-model:visible="showVariantDialog" :header="editingVariant ? 'Edit Variant' : 'Add Variant'" :modal="true" :style="{ width: '450px' }">
      <div class="form-grid">
        <div class="form-field">
          <label>Name *</label>
          <InputText v-model="variantForm.name" fluid />
        </div>
        <div class="form-field">
          <label>SKU</label>
          <InputText v-model="variantForm.sku" fluid />
        </div>
        <div class="form-field">
          <label>Price</label>
          <InputNumber v-model="variantForm.price" mode="currency" currency="IDR" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Stock</label>
          <InputNumber v-model="variantForm.stockQuantity" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Low Stock Threshold</label>
          <InputNumber v-model="variantForm.lowStockThreshold" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Sort Order</label>
          <InputNumber v-model="variantForm.sortOrder" :min="0" fluid />
        </div>
        <div class="form-field">
          <label>Attributes</label>
          <div class="attr-chips"><Chip v-for="(v, k) in variantForm.attributes" :key="k" :label="`${k}: ${v}`" removable @remove="removeAttribute(k)" style="margin: 2px" /></div>
          <div style="display: flex; gap: 8px; margin-top: 6px">
            <InputText v-model="attrKey" placeholder="Key (e.g. Size)" style="flex: 1" size="small" />
            <InputText v-model="attrValue" placeholder="Value (e.g. L)" style="flex: 1" size="small" />
            <Button icon="pi pi-plus" size="small" @click="addAttribute" />
          </div>
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
.discount-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.discount-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.product-price {
  display: flex;
  flex-direction: column;
  gap: 2px;
  color: var(--p-text-secondary-color);
  font-size: 13px;
}
.product-price strong {
  color: var(--p-text-color);
  font-size: 16px;
}

.discount-price {
  margin-top: 4px;
}

.original-price {
  display: block;
  text-decoration: line-through;
  opacity: 0.6;
}

.final-price {
  font-weight: 600;
}

.price-cell {
  position: relative;
  display: inline-block;
  padding-right: 50px; /* kasih ruang kanan biar gak nabrak */
}

.final-price {
  font-weight: 600;
  font-size: 14px;
  line-height: 1;
}

.original-price {
  position: absolute;
  top: -6px;
  right: 0;

  font-size: 11px;
  color: #9ca3af;
  text-decoration: line-through;
  white-space: nowrap;
}

.attr-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

@media (max-width: 640px) {
  .header-actions,
  .discount-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
