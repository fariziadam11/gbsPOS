<script setup lang="ts">
import Button from 'primevue/button'
import Menu from 'primevue/menu'
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useTheme } from '../composables/useTheme'

const props = defineProps<{
  showMenuToggle?: boolean
}>()

const emit = defineEmits<{
  toggleSidebar: []
}>()

const authStore = useAuthStore()
const { isDark, toggle } = useTheme()
const userMenu = ref<InstanceType<typeof Menu> | null>(null)

const userMenuItems = ref([
  {
    label: 'Logout',
    icon: 'pi pi-sign-out',
    command: () => {
      authStore.logout()
    },
  },
])

function toggleUserMenu(event: Event) {
  userMenu.value?.toggle(event)
}

async function logout() {
  await authStore.logout()
}
</script>

<template>
  <header class="app-header flex justify-content-between align-items-center px-3 lg:px-4 surface-section border-bottom-1 surface-border">
    <div class="header-brand flex align-items-center gap-2 lg:gap-3">
      <Button
        v-if="showMenuToggle"
        icon="pi pi-bars"
        text
        class="md:hidden mr-2"
        @click="emit('toggleSidebar')"
        title="Menu"
      />
      <i class="pi pi-play-circle brand-icon text-2xl lg:text-3xl text-primary"></i>
      <span class="brand-text text-lg lg:text-xl font-semibold text-color">GBS CMS</span>
    </div>

    <div class="header-actions flex align-items-center gap-2">
      <Button
        :icon="isDark ? 'pi pi-sun' : 'pi pi-moon'"
        text
        severity="secondary"
        @click="toggle"
        title="Toggle theme"
        class="theme-toggle"
      />

      <div class="user-info flex align-items-center">
        <Button
          type="button"
          :label="authStore.username || 'User'"
          icon="pi pi-user"
          text
          @click="toggleUserMenu"
          class="hidden md:inline-flex"
        />
        <Button
          type="button"
          icon="pi pi-user"
          text
          @click="toggleUserMenu"
          class="md:hidden"
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
  position: sticky;
  top: 0;
  z-index: 100;
}

.brand-icon {
  color: var(--p-primary-color);
}

.brand-text {
  color: var(--p-text-color);
}
</style>
