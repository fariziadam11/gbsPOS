<script setup lang="ts">
import Drawer from 'primevue/drawer'
import Button from 'primevue/button'
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const visible = ref(false)

const navItems = [
  { label: 'Dashboard', icon: 'pi pi-list', route: '/', requiresAdmin: false },
  { label: 'Upload Ad', icon: 'pi pi-upload', route: '/upload', requiresAdmin: true },
]

const filteredNavItems = navItems.filter(
  (item) => !item.requiresAdmin || authStore.isAdmin
)

function isActive(path: string) {
  return route.path === path
}

function navigateTo(path: string) {
  router.push(path)
  visible.value = false
}

function openSidebar() {
  visible.value = true
}
</script>

<template>
  <div class="sidebar-wrapper">
    <Button
      icon="pi pi-bars"
      text
      @click="openSidebar"
      class="sidebar-toggle"
      title="Menu"
    />

    <Drawer v-model:visible="visible" header="Menu" class="mobile-drawer">
      <nav class="nav-menu">
        <a
          v-for="item in filteredNavItems"
          :key="item.route"
          class="nav-link"
          :class="{ active: isActive(item.route) }"
          @click="navigateTo(item.route)"
        >
          <i :class="item.icon"></i>
          <span>{{ item.label }}</span>
        </a>
      </nav>
    </Drawer>

    <aside class="desktop-sidebar">
      <nav class="nav-menu">
        <a
          v-for="item in filteredNavItems"
          :key="item.route"
          class="nav-link"
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
.sidebar-wrapper {
  display: contents;
}

.sidebar-toggle {
  display: none;
  position: fixed;
  top: 12px;
  left: 12px;
  z-index: 101;
}

.desktop-sidebar {
  width: 220px;
  min-height: calc(100vh - 64px);
  background: var(--p-surface-0);
  border-right: 1px solid var(--p-surface-200);
  padding: 16px 0;
  position: sticky;
  top: 64px;
  flex-shrink: 0;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 12px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 8px;
  color: var(--p-text-secondary-color);
  text-decoration: none;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  font-size: 14px;
  font-weight: 500;
}

.nav-link:hover {
  background: var(--p-surface-100);
  color: var(--p-text-color);
}

.nav-link.active {
  background: var(--p-primary-50);
  color: var(--p-primary-color);
}

.nav-link i {
  font-size: 16px;
}

@media (max-width: 768px) {
  .sidebar-toggle {
    display: flex;
  }

  .desktop-sidebar {
    display: none;
  }
}
</style>
