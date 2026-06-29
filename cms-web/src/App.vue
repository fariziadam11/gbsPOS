<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

import Toast from "primevue/toast";
import ConfirmDialog from "primevue/confirmdialog";

import AppHeader from "./components/AppHeader.vue";
import AppSidebar from "./components/AppSidebar.vue";

const route = useRoute();
const sidebarVisible = ref(false);

const isAuthenticatedPage = computed(() => {
  return route.meta.requiresAuth;
});

function toggleSidebar() {
  sidebarVisible.value = !sidebarVisible.value;
}
</script>

<template>
  <div class="app-layout flex flex-column min-h-screen">
    <Toast position="top-right" />
    <ConfirmDialog />

    <AppHeader
      v-if="isAuthenticatedPage"
      :show-menu-toggle="true"
      @toggle-sidebar="toggleSidebar"
    />

    <div class="app-body flex flex-1">
      <AppSidebar
        v-if="isAuthenticatedPage"
        v-model:visible="sidebarVisible"
      />

      <main class="app-main flex-1 p-3 lg:p-4 overflow-y-auto surface-ground">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style>
</style>
