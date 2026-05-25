<script setup lang="ts">
import Button from 'primevue/button'
import Menu from 'primevue/menu'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const userMenu = ref<InstanceType<typeof Menu> | null>(null)

const userMenuItems = ref([
  {
    label: 'Logout',
    icon: 'pi pi-sign-out',
    command: () => {
      authStore.logout()
      router.push('/login')
    },
  },
])

function toggleUserMenu(event: Event) {
  userMenu.value?.toggle(event)
}

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <header class="app-header">
    <div class="header-brand">
      <i class="pi pi-play-circle brand-icon"></i>
      <span class="brand-text">GBS CMS</span>
    </div>

    <div class="header-actions">
      <div class="user-info">
        <Button
          type="button"
          :label="authStore.username || 'User'"
          icon="pi pi-user"
          text
          @click="toggleUserMenu"
        />
        <Menu ref="userMenu" :model="userMenuItems" popup />
      </div>
      <Button
        icon="pi pi-sign-out"
        text
        severity="secondary"
        @click="logout"
        title="Logout"
      />
    </div>
  </header>
</template>

<style scoped>
.app-header {
  height: 64px;
  background: var(--p-surface-0);
  border-bottom: 1px solid var(--p-surface-200);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-icon {
  font-size: 28px;
  color: var(--p-primary-color);
}

.brand-text {
  font-size: 20px;
  font-weight: 600;
  color: var(--p-text-color);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-info {
  display: flex;
  align-items: center;
}
</style>
