<script setup lang="ts">
import { ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '../composables/useUsers'
import { getErrorMessage } from '../api/client'
import type { UserListItem, CreateUserRequest } from '../types/api'

const confirm = useConfirm()
const toast = useToast()

const { data: users, isLoading } = useUsers()
const { mutate: createUser } = useCreateUser()
const { mutate: updateUser } = useUpdateUser()
const { mutate: deleteUser } = useDeleteUser()

const showDialog = ref(false)
const editingUser = ref<UserListItem | null>(null)
const dialogTitle = ref('Add User')
const submitting = ref(false)
const form = ref<CreateUserRequest>({
  username: '',
  password: '',
  name: '',
  role: 'CASHIER',
  gender: '',
})
const editForm = ref({
  name: '',
  role: 'CASHIER' as string,
  password: '',
  gender: '',
})

const roles = ['ADMIN', 'CASHIER']

function openCreate() {
  editingUser.value = null
  dialogTitle.value = 'Add User'
  form.value = { username: '', password: '', name: '', role: 'CASHIER', gender: '' }
  showDialog.value = true
}

function openEdit(user: UserListItem) {
  editingUser.value = user
  dialogTitle.value = 'Edit User'
  editForm.value = { name: user.name, role: user.role, password: '', gender: user.gender || '' }
  showDialog.value = true
}

function save() {
  if (editingUser.value) {
    if (!editForm.value.name) return
    submitting.value = true
    updateUser({ id: editingUser.value.id, data: editForm.value }, {
      onSuccess: () => {
        toast.add({ severity: 'success', summary: 'Updated', detail: 'User updated', life: 3000 })
        showDialog.value = false
      },
      onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  } else {
    if (!form.value.username || !form.value.password) {
      toast.add({ severity: 'warn', summary: 'Validation', detail: 'Username and password required', life: 3000 })
      return
    }
    submitting.value = true
    createUser(form.value, {
      onSuccess: () => {
        toast.add({ severity: 'success', summary: 'Created', detail: 'User created', life: 3000 })
        showDialog.value = false
      },
      onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      onSettled: () => { submitting.value = false },
    })
  }
}

function confirmDelete(user: UserListItem) {
  confirm.require({
    message: `Delete user "${user.username}"?`,
    header: 'Confirm Delete',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel',
    rejectProps: { severity: 'secondary', outlined: true },
    acceptLabel: 'Delete',
    acceptProps: { severity: 'danger' },
    accept: () => {
      deleteUser(user.id, {
        onSuccess: () => toast.add({ severity: 'success', summary: 'Deleted', detail: 'User deleted', life: 3000 }),
        onError: (err) => toast.add({ severity: 'error', summary: 'Error', detail: getErrorMessage(err), life: 5000 }),
      })
    },
  })
}
</script>

<template>
  <div class="flex flex-column gap-3 lg:gap-4">
    <div class="flex flex-column md:flex-row md:align-items-start justify-content-between gap-3">
      <div>
        <h1 class="text-2xl lg:text-3xl font-semibold text-color m-0">Users</h1>
        <p class="text-sm text-color-secondary mt-1 mb-0">Manage POS user accounts</p>
      </div>
      <div class="flex flex-wrap align-items-center gap-2">
        <Button label="Add User" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>

    <div class="surface-0 border-round-xl border-1 surface-border p-3">
      <DataTable :value="users || []" :loading="isLoading" tableStyle="min-width: 40rem" stripedRows size="small" paginator :rows="20">
        <Column field="username" header="Username" sortable />
        <Column field="name" header="Name" sortable />
        <Column header="Role" sortable style="width: 100px">
          <template #body="{ data }">
            <Tag :value="data.role" :severity="data.role === 'ADMIN' ? 'warn' : 'info'" />
          </template>
        </Column>
        <Column field="gender" header="Gender" style="width: 100px">
          <template #body="{ data }">{{ data.gender || '-' }}</template>
        </Column>
        <Column header="Created" sortable style="width: 150px">
          <template #body="{ data }">{{ new Date(data.createdAt).toLocaleDateString('id-ID') }}</template>
        </Column>
        <Column header="Actions" style="width: 120px">
          <template #body="{ data }">
            <div class="flex gap-1">
              <Button icon="pi pi-pencil" text rounded size="small" @click="openEdit(data)" />
              <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="confirmDelete(data)" />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="text-center p-5 text-color-secondary">No users found.</div>
        </template>
      </DataTable>
    </div>

    <!-- User Form Dialog -->
    <Dialog v-model:visible="showDialog" :header="dialogTitle" :modal="true" :style="{ width: '95vw', maxWidth: '450px' }">
      <template v-if="!editingUser">
        <div class="grid">
          <div class="col-12 md:col-6 form-field">
            <label>Username *</label>
            <InputText v-model="form.username" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Password *</label>
            <InputText v-model="form.password" type="password" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Name</label>
            <InputText v-model="form.name" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Role</label>
            <Select v-model="form.role" :options="roles" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Gender</label>
            <InputText v-model="form.gender" fluid />
          </div>
        </div>
      </template>
      <template v-else>
        <div class="grid">
          <div class="col-12 md:col-6 form-field">
            <label>Name</label>
            <InputText v-model="editForm.name" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Role</label>
            <Select v-model="editForm.role" :options="roles" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>New Password (leave blank to keep)</label>
            <InputText v-model="editForm.password" type="password" fluid />
          </div>
          <div class="col-12 md:col-6 form-field">
            <label>Gender</label>
            <InputText v-model="editForm.gender" fluid />
          </div>
        </div>
      </template>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="showDialog = false" />
        <Button label="Save" :loading="submitting" @click="save" />
      </template>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>

<style scoped>
.form-field { display: flex; flex-direction: column; gap: 4px; }
.form-field label { font-size: 14px; font-weight: 500; color: var(--p-text-color); }
</style>
