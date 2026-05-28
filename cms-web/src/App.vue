<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";

import Toast from "primevue/toast";
import ConfirmDialog from "primevue/confirmdialog";

import AppHeader from "./components/AppHeader.vue";
import AppSidebar from "./components/AppSidebar.vue";

const route = useRoute();

const isAuthenticatedPage = computed(() => {
  return route.meta.requiresAuth;
});
</script>

<template>
  <div class="app-layout">
    <Toast position="top-right" />
    <ConfirmDialog />

    <AppHeader v-if="isAuthenticatedPage" />

    <div class="app-body">
      <AppSidebar v-if="isAuthenticatedPage" />

      <main class="app-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-body {
  display: flex;
  flex: 1;
}

.app-main {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  background: var(--p-surface-50);
}
</style>
