<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const error = ref('')

onMounted(async () => {
  try {
    await authStore.handleLoginCallback(window.location.href)
    const redirect = route.query.redirect as string | undefined
    router.replace(redirect || '/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Login failed'
  }
})
</script>

<template>
  <div class="callback-page">
    <div v-if="error" class="error-message">
      <h2>Authentication failed</h2>
      <p>{{ error }}</p>
      <router-link to="/login" class="login-link">Back to login</router-link>
    </div>
    <div v-else class="loading-message">
      <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
      <p>Completing sign in...</p>
    </div>
  </div>
</template>

<style scoped>
.callback-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--p-surface-50);
  padding: 24px;
}

.loading-message,
.error-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
}

.error-message h2 {
  color: var(--p-red-500);
}

.login-link {
  color: var(--p-primary-color);
  text-decoration: underline;
}
</style>
