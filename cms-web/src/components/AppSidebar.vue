<script setup lang="ts">
import Drawer from 'primevue/drawer'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const visible = defineModel<boolean>('visible', { default: false })

const navItems = [
  { label: 'Dashboard', icon: 'pi pi-chart-bar', route: '/', requiresAdmin: false },
  { label: 'Products', icon: 'pi pi-box', route: '/products', requiresAdmin: false },
  { label: 'Orders', icon: 'pi pi-receipt', route: '/orders', requiresAdmin: false },
  { label: 'Customers', icon: 'pi pi-users', route: '/customers', requiresAdmin: false },
  { label: 'Ads', icon: 'pi pi-video', route: '/ads', requiresAdmin: false },
  { label: 'Upload Ad', icon: 'pi pi-upload', route: '/upload', requiresAdmin: true },
  { label: 'Users', icon: 'pi pi-id-card', route: '/users', requiresAdmin: true },
  { label: 'Settings', icon: 'pi pi-cog', route: '/settings', requiresAdmin: true },
]

const filteredNavItems = navItems.filter(
  (item) => !item.requiresAdmin || authStore.isAdmin
)

function isActive(path: string) {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path === path || route.path.startsWith(`${path}/`)
}

function navigateTo(path: string) {
  router.push(path)
  visible.value = false
}
</script>

<template>
  <div class="sidebar-wrapper">
    <Drawer v-model:visible="visible" header="Menu" class="mobile-drawer md:hidden">
      <nav class="nav-menu flex flex-column gap-1 p-3">
        <a
          v-for="item in filteredNavItems"
          :key="item.route"
          class="nav-link flex align-items-center gap-3 px-3 py-2 border-round-lg text-sm font-medium"
          :class="{ active: isActive(item.route) }"
          @click="navigateTo(item.route)"
        >
          <i :class="item.icon"></i>
          <span>{{ item.label }}</span>
        </a>
      </nav>
    </Drawer>

    <aside class="desktop-sidebar hidden md:flex flex-column surface-section border-right-1 surface-border p-0">
      <nav class="nav-menu flex flex-column gap-1 p-3">
        <a
          v-for="item in filteredNavItems"
          :key="item.route"
          class="nav-link flex align-items-center gap-3 px-3 py-2 border-round-lg text-sm font-medium"
          :class="{ active: isActive(item.route) }"
          @click="navigateTo(item.route)"
        >
          <i :class="item.icon"></i>
          <span>{{ item.label }}</span>
        </a>
      </nav>
    </aside>
  </div>
</template>

<style scoped>
.desktop-sidebar {
  width: 220px;
  min-height: calc(100vh - 64px);
  position: sticky;
  top: 64px;
  flex-shrink: 0;
}

.nav-link {
  color: var(--p-text-secondary-color);
  text-decoration: none;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}

.nav-link:hover {
  background: light-dark(var(--p-surface-100), var(--p-surface-800));
  color: var(--p-text-color);
}

.nav-link.active {
  background: light-dark(var(--p-primary-50), var(--p-primary-900));
  color: var(--p-primary-color);
}

.nav-link i {
  font-size: 16px;
}
</style>
